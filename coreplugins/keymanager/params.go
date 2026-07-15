//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

func (p *keyManagerPlugin) patchParams(ctx context.Context, set map[string]string, deleteKeys []string) error {
	reply, err := panel.NewHost().PatchParams(ctx, &panel.ParamsPatchRequest{Set: set, Delete: deleteKeys})
	if err != nil {
		return err
	}
	if reply.GetError() != "" {
		return fmt.Errorf("host rejected Params update: %s", reply.GetError())
	}

	p.mu.Lock()
	if p.params == nil {
		p.params = make(map[string]string)
	}
	for key, value := range set {
		p.params[key] = value
	}
	for _, key := range deleteKeys {
		delete(p.params, key)
	}
	p.mu.Unlock()
	return nil
}

func (p *keyManagerPlugin) secretExists(name string) bool {
	p.mu.RLock()
	_, exists := p.params[paramSecretPrefix+name]
	p.mu.RUnlock()
	return exists
}

func (p *keyManagerPlugin) secretCiphertext(name string) (string, bool) {
	p.mu.RLock()
	ciphertext, exists := p.params[paramSecretPrefix+name]
	p.mu.RUnlock()
	return ciphertext, exists
}

func (p *keyManagerPlugin) secretEncryption(name string) (string, error) {
	p.mu.RLock()
	params := cloneParams(p.params)
	p.mu.RUnlock()
	return secretEncryptionFromParams(params, name)
}

func secretEncryptionFromParams(params map[string]string, name string) (string, error) {
	raw := params[paramMetaPrefix+name]
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
		if strings.HasPrefix(key, paramSecretPrefix) {
			keys = append(keys, strings.TrimPrefix(key, paramSecretPrefix))
		}
	}
	sort.Strings(keys)

	result := make([]secretMeta, 0, len(keys))
	for _, name := range keys {
		if err := validateSecretName(name); err != nil {
			return nil, err
		}
		meta := secretMeta{Name: name}
		if raw := params[paramMetaPrefix+name]; raw != "" {
			if err := json.Unmarshal([]byte(raw), &meta); err != nil {
				return nil, fmt.Errorf("invalid metadata for secret %q", name)
			}
		}
		encryption, err := normalizeSecretEncryption(meta.Encryption)
		if err != nil {
			return nil, fmt.Errorf("invalid encryption for secret %q: %w", name, err)
		}
		meta.Encryption = encryption
		plugins, err := decodePlugins(params[paramPolicyPrefix+name])
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

func cloneParams(params map[string]string) map[string]string {
	result := make(map[string]string, len(params))
	for key, value := range params {
		result[key] = value
	}
	return result
}
