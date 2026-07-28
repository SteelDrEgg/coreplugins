package secrets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

// ErrorKind classifies a ServiceError so transport adapters can map it to a
// status code (HTTP) or a plain string (service messages) without
// re-deriving what went wrong.
type ErrorKind int

const (
	ErrInvalidInput ErrorKind = iota
	ErrNotFound
	ErrConflict
	ErrForbidden
	ErrPassphraseRequired
	ErrInvalidPassphrase
	ErrInternal
)

// ServiceError is the single error type returned by every Service method, so
// HTTP and message adapters share one error taxonomy instead of each having
// its own ad-hoc shape.
type ServiceError struct {
	Kind ErrorKind
	Msg  string
}

func (e *ServiceError) Error() string { return e.Msg }

func newError(kind ErrorKind, format string, args ...any) *ServiceError {
	return &ServiceError{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}

// Service is the shared domain logic behind every secret-manager transport.
// The HTTP handlers and service-message handlers are both thin adapters that
// build a Caller and call into this type, instead of each re-implementing
// the add/update/get/list/delete flow.
type Service struct {
	writeMu   sync.Mutex
	store     *paramsStore
	encryptor Encryptor
	logger    arupa.Logger
}

// NewService builds a Service. logger may be nil (see SetLogger) since the
// SDK's Logger is often the same value being composed around this Service.
func NewService(store *paramsStore, encryptor Encryptor, logger arupa.Logger) *Service {
	return &Service{store: store, encryptor: encryptor, logger: logger}
}

// SetLogger assigns the logger after construction, for callers that only
// obtain their Logger (e.g. the SDK service value) after building the
// Service that needs to log through it.
func (s *Service) SetLogger(logger arupa.Logger) { s.logger = logger }

// Load applies the host's Params snapshot, typically during OnRegister.
func (s *Service) Load(client arupa.ParamsClient, params map[string]string) {
	s.store.load(client, params)
}

// EnsureIdentity loads the persisted encryption identity, generating and
// persisting a new one on first run.
func (s *Service) EnsureIdentity(ctx context.Context) error {
	identityText, generated, err := s.encryptor.EnsureIdentity(s.store.identity())
	if err != nil {
		return err
	}
	if generated {
		if err := s.store.setIdentity(ctx, identityText); err != nil {
			return fmt.Errorf("persist secrets manager identity: %w", err)
		}
		s.logInfo(ctx, "generated new secret-manager encryption identity")
	}
	return nil
}

// WriteSecretInput carries the fields shared by add and update.
type WriteSecretInput struct {
	Name           string
	Description    string
	Value          string // "*" on an update means "keep the existing ciphertext"
	Passphrase     string
	AllowedPlugins []string
	Update         bool
}

// WriteSecret adds or updates a secret. It is the single code path used by
// both the HTTP admin UI and the plugin message API; the differences between
// those two callers (ACL enforcement, auto-grant-on-create) are expressed
// through Caller rather than duplicated per transport.
func (s *Service) WriteSecret(ctx context.Context, caller Caller, in WriteSecretInput) (secretMeta, error) {
	if err := validateSecretName(in.Name); err != nil {
		return secretMeta{}, newError(ErrInvalidInput, "%s", err.Error())
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	exists := s.store.hasSecret(in.Name)
	if exists != in.Update {
		if in.Update {
			return secretMeta{}, newError(ErrNotFound, "secret not found")
		}
		return secretMeta{}, newError(ErrConflict, "secret already exists")
	}

	if in.Update && caller.requiresACL() && !s.store.allows(in.Name, caller.Plugin) {
		s.logWarn(ctx, fmt.Sprintf("plugin %q denied write access to secret %q", caller.Plugin, in.Name))
		return secretMeta{}, newError(ErrForbidden, "plugin is not allowed to access this secret")
	}

	allowedPlugins := in.AllowedPlugins
	if !in.Update && caller.autoGrantOnCreate() {
		allowedPlugins = append(allowedPlugins, caller.Plugin)
	}
	allowedPlugins, err := normalizePlugins(allowedPlugins)
	if err != nil {
		return secretMeta{}, newError(ErrInvalidInput, "%s", err.Error())
	}

	var ciphertext, encryption string
	if in.Update && in.Value == "*" {
		if in.Passphrase != "" {
			return secretMeta{}, newError(ErrInvalidInput, "you must edit both value and passphrase")
		}
		var ok bool
		ciphertext, ok = s.store.ciphertext(in.Name)
		if !ok {
			return secretMeta{}, newError(ErrNotFound, "secret not found")
		}
		encryption, err = s.store.encryption(in.Name)
		if err != nil {
			return secretMeta{}, newError(ErrInternal, "%s", err.Error())
		}
	} else {
		ciphertext, encryption, err = s.encryptor.Encrypt(in.Value, in.Passphrase)
		if err != nil {
			s.logError(ctx, fmt.Sprintf("encrypt secret %q: %v", in.Name, err))
			return secretMeta{}, newError(ErrInternal, "%s", err.Error())
		}
	}

	meta := secretMeta{
		Name:           in.Name,
		Description:    strings.TrimSpace(in.Description),
		AllowedPlugins: allowedPlugins,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		Encryption:     encryption,
	}
	if err := s.store.putSecret(ctx, ciphertext, meta); err != nil {
		return secretMeta{}, newError(ErrInternal, "%s", err.Error())
	}
	return meta, nil
}

// GetSecret decrypts and returns a secret's value. The ACL check runs before
// any existence check for callers that require it, so a plugin without
// access learns nothing about whether the secret exists.
func (s *Service) GetSecret(ctx context.Context, caller Caller, name, passphrase string) (string, error) {
	if err := validateSecretName(name); err != nil {
		return "", newError(ErrInvalidInput, "%s", err.Error())
	}
	if caller.requiresACL() && !s.store.allows(name, caller.Plugin) {
		s.logWarn(ctx, fmt.Sprintf("plugin %q denied read access to secret %q", caller.Plugin, name))
		return "", newError(ErrForbidden, "plugin is not allowed to access this secret")
	}

	ciphertext, ok := s.store.ciphertext(name)
	if !ok {
		return "", newError(ErrNotFound, "secret %q was not found", name)
	}
	encryption, err := s.store.encryption(name)
	if err != nil {
		return "", newError(ErrInternal, "%s", err.Error())
	}
	value, err := s.encryptor.Decrypt(ciphertext, encryption, passphrase)
	if err != nil {
		switch {
		case errors.Is(err, errPassphraseRequired):
			return "", newError(ErrPassphraseRequired, "passphrase required")
		case errors.Is(err, errInvalidPassphrase):
			return "", newError(ErrInvalidPassphrase, "invalid passphrase")
		default:
			return "", newError(ErrInternal, "%s", err.Error())
		}
	}
	return value, nil
}

// ListSecrets returns metadata for every secret visible to caller. A human
// admin sees everything; a plugin sees only secrets that allow it.
func (s *Service) ListSecrets(ctx context.Context, caller Caller) ([]secretMeta, error) {
	all, err := s.store.listSecrets()
	if err != nil {
		return nil, newError(ErrInternal, "%s", err.Error())
	}
	if !caller.requiresACL() {
		return all, nil
	}
	visible := make([]secretMeta, 0, len(all))
	for _, meta := range all {
		if allowedPlugin(meta.AllowedPlugins, caller.Plugin) {
			visible = append(visible, meta)
		}
	}
	return visible, nil
}

// DeleteSecret removes a secret, subject to the same per-caller ACL as
// WriteSecret's update path.
func (s *Service) DeleteSecret(ctx context.Context, caller Caller, name string) error {
	if err := validateSecretName(name); err != nil {
		return newError(ErrInvalidInput, "%s", err.Error())
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if !s.store.hasSecret(name) {
		return newError(ErrNotFound, "secret not found")
	}
	if caller.requiresACL() && !s.store.allows(name, caller.Plugin) {
		s.logWarn(ctx, fmt.Sprintf("plugin %q denied delete access to secret %q", caller.Plugin, name))
		return newError(ErrForbidden, "plugin is not allowed to access this secret")
	}
	if err := s.store.deleteSecret(ctx, name); err != nil {
		return newError(ErrInternal, "%s", err.Error())
	}
	return nil
}

func (s *Service) logInfo(ctx context.Context, msg string)  { s.log(ctx, arupa.LogInfo, msg) }
func (s *Service) logWarn(ctx context.Context, msg string)  { s.log(ctx, arupa.LogWarn, msg) }
func (s *Service) logError(ctx context.Context, msg string) { s.log(ctx, arupa.LogError, msg) }

func (s *Service) log(ctx context.Context, level arupa.LogLevel, msg string) {
	if s.logger == nil {
		return
	}
	_ = s.logger.Log(ctx, level, msg)
}
