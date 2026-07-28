// Package secrets holds secret-manager's domain logic: secret metadata,
// persistence, encryption, caller/ACL semantics, and the request-handling
// service shared by every transport. Nothing here depends on WASM or any
// other Arupa backend adapter, so it can be exercised with plain go test.
package secrets

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// secretMeta is the persisted, non-secret metadata for one secret.
type secretMeta struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	AllowedPlugins []string `json:"allowed_plugins"`
	UpdatedAt      string   `json:"updated_at"`
	Encryption     string   `json:"encryption,omitempty"`
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

func allowedPlugin(plugins []string, source string) bool {
	for _, plugin := range plugins {
		if plugin == source {
			return true
		}
	}
	return false
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
