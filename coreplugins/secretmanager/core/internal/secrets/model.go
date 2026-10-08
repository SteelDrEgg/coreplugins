// Package secrets defines secret-manager's records, permissions and operations.
// Host persistence and cryptography are supplied through small interfaces.
package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const MaxSecretSize = 1 << 20

type Protection string

const (
	ProtectionIdentity   Protection = "identity"
	ProtectionPassphrase Protection = "passphrase"
)

type EncryptedValue struct {
	Protection Protection `json:"protection"`
	Ciphertext string     `json:"ciphertext"`
}

// SecretRecord is persisted as a whole. Its immutable name comes from the key.
type SecretRecord struct {
	Name           string         `json:"-"`
	Description    string         `json:"description"`
	AllowedPlugins []string       `json:"allowed_plugins"`
	EncryptedValue EncryptedValue `json:"encrypted_value"`
	UpdatedAt      string         `json:"updated_at"`
}

// SecretInfo excludes ciphertext and is the public result of list/create/update.
type SecretInfo struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	AllowedPlugins []string   `json:"allowed_plugins"`
	Protection     Protection `json:"protection"`
	UpdatedAt      string     `json:"updated_at"`
}

func (r SecretRecord) Info() SecretInfo {
	return SecretInfo{
		Name:           r.Name,
		Description:    r.Description,
		AllowedPlugins: append([]string{}, r.AllowedPlugins...),
		Protection:     r.EncryptedValue.Protection,
		UpdatedAt:      r.UpdatedAt,
	}
}

// Field distinguishes omission from an explicitly empty update; null is invalid.
type Field[T any] struct {
	Value T
	Set   bool
}

func Present[T any](value T) Field[T] { return Field[T]{Value: value, Set: true} }

func (f *Field[T]) UnmarshalJSON(raw []byte) error {
	var value T
	if err := DecodeJSON(raw, &value); err != nil {
		return err
	}
	*f = Present(value)
	return nil
}

// ValueInput replaces a whole value, including its explicit protection mode.
type ValueInput struct {
	Plaintext  string
	Protection Protection
	Passphrase string
}

func (v *ValueInput) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Plaintext  Field[string]     `json:"plaintext"`
		Protection Field[Protection] `json:"protection"`
		Passphrase Field[string]     `json:"passphrase"`
	}
	if err := DecodeJSON(raw, &wire); err != nil {
		return err
	}
	if !wire.Plaintext.Set || !wire.Protection.Set {
		return fmt.Errorf("value requires plaintext and protection")
	}
	*v = ValueInput{Plaintext: wire.Plaintext.Value, Protection: wire.Protection.Value, Passphrase: wire.Passphrase.Value}
	return nil
}

func (v ValueInput) Validate() error {
	if len(v.Plaintext) > MaxSecretSize {
		return fmt.Errorf("secret exceeds the %d byte limit", MaxSecretSize)
	}
	switch v.Protection {
	case ProtectionIdentity:
		if v.Passphrase != "" {
			return fmt.Errorf("identity protection does not accept a passphrase")
		}
	case ProtectionPassphrase:
		if v.Passphrase == "" {
			return fmt.Errorf("passphrase protection requires a passphrase")
		}
	default:
		return fmt.Errorf("unsupported protection mode")
	}
	return nil
}

type CreateSecretInput struct {
	Name           string            `json:"name"`
	Description    Field[string]     `json:"description"`
	AllowedPlugins Field[[]string]   `json:"allowed_plugins"`
	Value          Field[ValueInput] `json:"value"`
}

type UpdateSecretInput struct {
	Name           string            `json:"name"`
	Description    Field[string]     `json:"description"`
	AllowedPlugins Field[[]string]   `json:"allowed_plugins"`
	Value          Field[ValueInput] `json:"value"`
}

// DecodeJSON rejects unknown fields, null, duplicate keys and trailing documents.
// It is shared by request and persisted-record decoding.
func DecodeJSON(raw []byte, target any) error {
	if err := validateJSON(raw); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(target)
}

func validateJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var read func() error
	read = func() error {
		token, err := d.Token()
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("null is not allowed")
		}
		if delimiter, ok := token.(json.Delim); ok {
			switch delimiter {
			case '{':
				seen := make(map[string]bool)
				for d.More() {
					key, err := d.Token()
					if err != nil {
						return err
					}
					name, ok := key.(string)
					if !ok {
						return fmt.Errorf("invalid object key")
					}
					if seen[name] {
						return fmt.Errorf("duplicate object key")
					}
					seen[name] = true
					if err := read(); err != nil {
						return err
					}
				}
			case '[':
				for d.More() {
					if err := read(); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("invalid JSON delimiter")
			}
			_, err = d.Token()
			return err
		}
		return nil
	}
	if err := read(); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("secret name cannot be empty")
	}
	if len(name) > 128 {
		return "", fmt.Errorf("secret name is too long")
	}
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("secret name cannot contain '..'")
	}
	for _, r := range name {
		if r == '/' || r == '_' || r == '-' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		return "", fmt.Errorf("secret name contains unsupported characters")
	}
	return name, nil
}

func NormalizePlugins(plugins []string) ([]string, error) {
	seen := make(map[string]bool, len(plugins))
	result := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		plugin = strings.TrimSpace(plugin)
		if plugin == "" {
			return nil, fmt.Errorf("allowed plugin names cannot be empty")
		}
		if !seen[plugin] {
			seen[plugin] = true
			result = append(result, plugin)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (r SecretRecord) Validate() error {
	name, err := NormalizeName(r.Name)
	if err != nil {
		return err
	}
	if name != r.Name {
		return fmt.Errorf("stored name must be normalized")
	}
	if r.AllowedPlugins == nil {
		return fmt.Errorf("allowed_plugins must be an array")
	}
	if _, err := NormalizePlugins(r.AllowedPlugins); err != nil {
		return err
	}
	if r.EncryptedValue.Protection != ProtectionIdentity && r.EncryptedValue.Protection != ProtectionPassphrase {
		return fmt.Errorf("invalid protection mode")
	}
	if r.EncryptedValue.Ciphertext == "" {
		return fmt.Errorf("missing ciphertext")
	}
	if _, err := time.Parse(time.RFC3339, r.UpdatedAt); err != nil {
		return fmt.Errorf("invalid updated_at")
	}
	return nil
}
