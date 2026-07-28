package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

const (
	paramIdentity            = "secretmgr.identity"
	paramSecretsPrefix       = "secrets."
	paramSecretField         = "secret"
	paramAllowedPluginsField = "allowed_plugins"
	paramMetaField           = "meta"
)

// paramsStore is the secret manager's persisted state. It owns both the host
// Params client and the local snapshot, so callers never manipulate parameter
// keys or a raw Params map directly.
type paramsStore struct {
	mu     sync.RWMutex
	client arupa.ParamsClient
	values map[string]string
}

// NewParamsStore builds an empty paramsStore; call load once the host's
// Params snapshot is available (typically during OnRegister).
func NewParamsStore() *paramsStore {
	return &paramsStore{values: make(map[string]string)}
}

func secretParamKey(name, field string) string {
	return paramSecretsPrefix + name + "." + field
}

func (s *paramsStore) load(client arupa.ParamsClient, params map[string]string) {
	s.mu.Lock()
	s.client = client
	s.values = arupa.CloneParams(params)
	s.mu.Unlock()
}

func (s *paramsStore) patch(ctx context.Context, set map[string]string, deleteKeys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("secret-manager Params store is not configured")
	}
	patch := arupa.ParamsPatch{
		Set:    arupa.CloneParams(set),
		Delete: append([]string(nil), deleteKeys...),
	}
	if err := s.client.PatchParams(ctx, patch); err != nil {
		return err
	}

	for key, value := range patch.Set {
		s.values[key] = value
	}
	for _, key := range patch.Delete {
		delete(s.values, key)
	}
	return nil
}

func (s *paramsStore) identity() string {
	s.mu.RLock()
	identity := s.values[paramIdentity]
	s.mu.RUnlock()
	return identity
}

func (s *paramsStore) setIdentity(ctx context.Context, identity string) error {
	return s.patch(ctx, map[string]string{paramIdentity: identity}, nil)
}

func (s *paramsStore) hasSecret(name string) bool {
	s.mu.RLock()
	_, exists := s.values[secretParamKey(name, paramSecretField)]
	s.mu.RUnlock()
	return exists
}

func (s *paramsStore) ciphertext(name string) (string, bool) {
	s.mu.RLock()
	ciphertext, exists := s.values[secretParamKey(name, paramSecretField)]
	s.mu.RUnlock()
	return ciphertext, exists
}

// encryption looks up the single meta key for name directly, rather than
// cloning the entire Params snapshot just to read one field.
func (s *paramsStore) encryption(name string) (string, error) {
	s.mu.RLock()
	raw := s.values[secretParamKey(name, paramMetaField)]
	s.mu.RUnlock()
	return secretEncryptionFromRaw(raw)
}

func (s *paramsStore) allows(name, plugin string) bool {
	plugin = strings.TrimSpace(plugin)
	if plugin == "" {
		return false
	}
	s.mu.RLock()
	raw := s.values[secretParamKey(name, paramAllowedPluginsField)]
	s.mu.RUnlock()
	plugins, err := decodePlugins(raw)
	return err == nil && allowedPlugin(plugins, plugin)
}

func (s *paramsStore) listSecrets() ([]secretMeta, error) {
	s.mu.RLock()
	params := arupa.CloneParams(s.values)
	s.mu.RUnlock()
	return listSecretMeta(params)
}

func (s *paramsStore) putSecret(ctx context.Context, ciphertext string, meta secretMeta) error {
	if err := validateSecretName(meta.Name); err != nil {
		return err
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	allowedPluginsJSON, err := json.Marshal(meta.AllowedPlugins)
	if err != nil {
		return fmt.Errorf("encode allowed plugins: %w", err)
	}
	return s.patch(ctx, map[string]string{
		secretParamKey(meta.Name, paramSecretField):         ciphertext,
		secretParamKey(meta.Name, paramAllowedPluginsField): string(allowedPluginsJSON),
		secretParamKey(meta.Name, paramMetaField):           string(metaJSON),
	}, nil)
}

func (s *paramsStore) deleteSecret(ctx context.Context, name string) error {
	return s.patch(ctx, nil, []string{
		secretParamKey(name, paramSecretField),
		secretParamKey(name, paramAllowedPluginsField),
		secretParamKey(name, paramMetaField),
	})
}

func secretEncryptionFromRaw(raw string) (string, error) {
	if raw == "" {
		return secretEncryptionIdentity, nil
	}
	var meta secretMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return "", fmt.Errorf("invalid metadata")
	}
	return normalizeSecretEncryption(meta.Encryption)
}

func listSecretMeta(params map[string]string) ([]secretMeta, error) {
	keys := make([]string, 0)
	for key := range params {
		if strings.HasPrefix(key, paramSecretsPrefix) && strings.HasSuffix(key, "."+paramSecretField) {
			name := strings.TrimSuffix(strings.TrimPrefix(key, paramSecretsPrefix), "."+paramSecretField)
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)

	result := make([]secretMeta, 0, len(keys))
	for _, name := range keys {
		if err := validateSecretName(name); err != nil {
			return nil, err
		}
		meta := secretMeta{Name: name}
		if raw := params[secretParamKey(name, paramMetaField)]; raw != "" {
			if err := json.Unmarshal([]byte(raw), &meta); err != nil {
				return nil, fmt.Errorf("invalid metadata for secret %q", name)
			}
		}
		encryption, err := normalizeSecretEncryption(meta.Encryption)
		if err != nil {
			return nil, fmt.Errorf("invalid encryption for secret %q: %w", name, err)
		}
		meta.Encryption = encryption
		plugins, err := decodePlugins(params[secretParamKey(name, paramAllowedPluginsField)])
		if err != nil {
			return nil, fmt.Errorf("invalid allowed plugins for secret %q", name)
		}
		meta.Name = name
		meta.AllowedPlugins = plugins
		result = append(result, meta)
	}
	return result, nil
}
