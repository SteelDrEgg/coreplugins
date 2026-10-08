package secrets

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Repository interface {
	Get(context.Context, string) (SecretRecord, error)
	List(context.Context) ([]SecretRecord, error)
	Create(context.Context, SecretRecord) error
	Update(context.Context, SecretRecord) error
	Delete(context.Context, string) error
}

type Cipher interface {
	Encrypt(context.Context, ValueInput) (EncryptedValue, error)
	Decrypt(context.Context, EncryptedValue, string) (string, error)
}

// Service serializes reads and mutations across both transports. The repository
// reads the host's current state; there are no versions or local state snapshots.
type Service struct {
	queue      serialQueue
	repository Repository
	cipher     Cipher
}

func NewService(repository Repository, cipher Cipher) *Service {
	return &Service{repository: repository, cipher: cipher}
}

func (s *Service) CreateSecret(ctx context.Context, caller Caller, in CreateSecretInput) (SecretInfo, error) {
	leave, err := s.queue.enter(ctx)
	if err != nil {
		return SecretInfo{}, err
	}
	defer leave()
	if err := caller.validate(); err != nil {
		return SecretInfo{}, err
	}
	name, err := NormalizeName(in.Name)
	if err != nil {
		return SecretInfo{}, NewError(ErrInvalidInput, err.Error())
	}
	if !in.Value.Set {
		return SecretInfo{}, NewError(ErrInvalidInput, "value is required")
	}
	if err := in.Value.Value.Validate(); err != nil {
		return SecretInfo{}, NewError(ErrInvalidInput, err.Error())
	}
	plugins := append([]string{}, in.AllowedPlugins.Value...)
	if caller.Kind == CallerPlugin {
		plugins = append(plugins, caller.Plugin)
	}
	plugins, err = NormalizePlugins(plugins)
	if err != nil {
		return SecretInfo{}, NewError(ErrInvalidInput, err.Error())
	}
	if _, err := s.repository.Get(ctx, name); err == nil {
		return SecretInfo{}, NewError(ErrConflict, "secret already exists")
	} else if !errors.Is(err, RecordNotFound) {
		return SecretInfo{}, internalError(err)
	}
	value, err := s.cipher.Encrypt(ctx, in.Value.Value)
	if err != nil {
		return SecretInfo{}, internalError(err)
	}
	record := SecretRecord{
		Name:           name,
		Description:    strings.TrimSpace(in.Description.Value),
		AllowedPlugins: plugins,
		EncryptedValue: value,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := s.repository.Create(ctx, record); err != nil {
		if errors.Is(err, RecordExists) {
			return SecretInfo{}, NewError(ErrConflict, "secret already exists")
		}
		return SecretInfo{}, internalError(err)
	}
	return record.Info(), nil
}

func (s *Service) UpdateSecret(ctx context.Context, caller Caller, in UpdateSecretInput) (SecretInfo, error) {
	leave, err := s.queue.enter(ctx)
	if err != nil {
		return SecretInfo{}, err
	}
	defer leave()
	record, err := s.authorizedRecord(ctx, caller, in.Name)
	if err != nil {
		return SecretInfo{}, err
	}
	if !in.Description.Set && !in.AllowedPlugins.Set && !in.Value.Set {
		return SecretInfo{}, NewError(ErrInvalidInput, "at least one update field is required")
	}
	if in.Description.Set {
		record.Description = strings.TrimSpace(in.Description.Value)
	}
	if in.AllowedPlugins.Set {
		record.AllowedPlugins, err = NormalizePlugins(in.AllowedPlugins.Value)
		if err != nil {
			return SecretInfo{}, NewError(ErrInvalidInput, err.Error())
		}
	}
	if in.Value.Set {
		if err := in.Value.Value.Validate(); err != nil {
			return SecretInfo{}, NewError(ErrInvalidInput, err.Error())
		}
		record.EncryptedValue, err = s.cipher.Encrypt(ctx, in.Value.Value)
		if err != nil {
			return SecretInfo{}, internalError(err)
		}
	}
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.repository.Update(ctx, record); err != nil {
		return SecretInfo{}, internalError(err)
	}
	return record.Info(), nil
}

func (s *Service) GetSecret(ctx context.Context, caller Caller, name, passphrase string) (string, error) {
	leave, err := s.queue.enter(ctx)
	if err != nil {
		return "", err
	}
	defer leave()
	record, err := s.authorizedRecord(ctx, caller, name)
	if err != nil {
		return "", err
	}
	value, err := s.cipher.Decrypt(ctx, record.EncryptedValue, passphrase)
	if err != nil {
		switch {
		case errors.Is(err, PassphraseRequired):
			return "", NewError(ErrPassphraseRequired, "passphrase required")
		case errors.Is(err, InvalidPassphrase):
			return "", NewError(ErrInvalidPassphrase, "invalid passphrase")
		default:
			return "", internalError(err)
		}
	}
	return value, nil
}

func (s *Service) ListSecrets(ctx context.Context, caller Caller) ([]SecretInfo, error) {
	leave, err := s.queue.enter(ctx)
	if err != nil {
		return nil, err
	}
	defer leave()
	if err := caller.validate(); err != nil {
		return nil, err
	}
	records, err := s.repository.List(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	result := make([]SecretInfo, 0, len(records))
	for _, record := range records {
		if caller.allows(record) {
			result = append(result, record.Info())
		}
	}
	return result, nil
}

func (s *Service) DeleteSecret(ctx context.Context, caller Caller, name string) error {
	leave, err := s.queue.enter(ctx)
	if err != nil {
		return err
	}
	defer leave()
	record, err := s.authorizedRecord(ctx, caller, name)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, record.Name); err != nil {
		return internalError(err)
	}
	return nil
}

func (s *Service) authorizedRecord(ctx context.Context, caller Caller, name string) (SecretRecord, error) {
	if err := caller.validate(); err != nil {
		return SecretRecord{}, err
	}
	name, err := NormalizeName(name)
	if err != nil {
		return SecretRecord{}, NewError(ErrInvalidInput, err.Error())
	}
	record, err := s.repository.Get(ctx, name)
	if errors.Is(err, RecordNotFound) {
		if caller.Kind == CallerPlugin {
			return SecretRecord{}, NewError(ErrForbidden, "plugin is not allowed to access this secret")
		}
		return SecretRecord{}, NewError(ErrNotFound, "secret not found")
	}
	if err != nil {
		return SecretRecord{}, internalError(err)
	}
	if !caller.allows(record) {
		return SecretRecord{}, NewError(ErrForbidden, "plugin is not allowed to access this secret")
	}
	return record, nil
}
