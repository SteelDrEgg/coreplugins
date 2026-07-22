//go:build wasip1

package main

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
	paramIdentity      = "secretmgr.identity"
	paramSecretsPrefix = "secrets."
	paramSecretField   = "secret"
	paramPolicyField   = "policy"
	paramMetaField     = "meta"
)

// paramsStore is the secret manager's persisted state. It owns both the host
// Params client and the local snapshot, so callers never manipulate parameter
// keys or a raw Params map directly.
type paramsStore struct {
	mu     sync.RWMutex
	client arupa.ParamsClient
	values map[string]string
}

func newParamsStore() *paramsStore {
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

func (s *paramsStore) encryption(name string) (string, error) {
	s.mu.RLock()
	params := arupa.CloneParams(s.values)
	s.mu.RUnlock()
	return secretEncryptionFromParams(params, name)
}

func (s *paramsStore) allows(name, plugin string) bool {
	plugin = strings.TrimSpace(plugin)
	if plugin == "" {
		return false
	}
	s.mu.RLock()
	raw := s.values[secretParamKey(name, paramPolicyField)]
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
	policyJSON, err := json.Marshal(meta.AllowedPlugins)
	if err != nil {
		return fmt.Errorf("encode access policy: %w", err)
	}
	return s.patch(ctx, map[string]string{
		secretParamKey(meta.Name, paramSecretField): ciphertext,
		secretParamKey(meta.Name, paramPolicyField): string(policyJSON),
		secretParamKey(meta.Name, paramMetaField):   string(metaJSON),
	}, nil)
}

func (s *paramsStore) deleteSecret(ctx context.Context, name string) error {
	return s.patch(ctx, nil, []string{
		secretParamKey(name, paramSecretField),
		secretParamKey(name, paramPolicyField),
		secretParamKey(name, paramMetaField),
	})
}

func secretEncryptionFromParams(params map[string]string, name string) (string, error) {
	raw := params[secretParamKey(name, paramMetaField)]
	if raw == "" {
		return secretEncryptionIdentity, nil
	}

	var meta secretMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return "", fmt.Errorf("invalid metadata for secret %q", name)
	}
	return normalizeSecretEncryption(meta.Encryption)
}

func normalizeSecretEncryption(encryption string) (string, error) {
	if encryption == "" {
		return secretEncryptionIdentity, nil
	}
	if encryption != secretEncryptionIdentity && encryption != secretEncryptionScrypt {
		return "", fmt.Errorf("unsupported secret encryption %q", encryption)
	}
	return encryption, nil
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
		plugins, err := decodePlugins(params[secretParamKey(name, paramPolicyField)])
		if err != nil {
			return nil, fmt.Errorf("invalid access policy for secret %q", name)
		}
		meta.Name = name
		meta.AllowedPlugins = plugins
		result = append(result, meta)
	}
	return result, nil
}

func decodePlugins(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var plugins []string
	if err := json.Unmarshal([]byte(raw), &plugins); err != nil {
		return nil, err
	}
	return normalizePlugins(plugins)
}

func normalizePlugins(plugins []string) ([]string, error) {
	seen := make(map[string]struct{}, len(plugins))
	result := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		plugin = strings.TrimSpace(plugin)
		if plugin == "" {
			return nil, fmt.Errorf("allowed plugin names cannot be empty")
		}
		if _, ok := seen[plugin]; ok {
			continue
		}
		seen[plugin] = struct{}{}
		result = append(result, plugin)
	}
	sort.Strings(result)
	return result, nil
}

func validateSecretName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("secret name is too long")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("secret name cannot contain '..'")
	}
	for _, r := range name {
		if r == '/' || r == '_' || r == '-' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		return fmt.Errorf("secret name contains unsupported character %q", r)
	}
	return nil
}
