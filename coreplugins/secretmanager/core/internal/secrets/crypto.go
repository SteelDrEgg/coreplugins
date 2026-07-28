package secrets

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"filippo.io/age"
)

const (
	maxSecretSize = 1 << 20

	secretEncryptionIdentity = "identity"
	secretEncryptionScrypt   = "scrypt"
)

var (
	errPassphraseRequired = errors.New("passphrase required")
	errInvalidPassphrase  = errors.New("invalid passphrase")
)

// Encryptor is the narrow boundary between secret business logic and the
// underlying cryptography, so tests can substitute a fake without touching
// real age/scrypt crypto.
type Encryptor interface {
	Encrypt(value, passphrase string) (ciphertext, encryption string, err error)
	Decrypt(ciphertext, encryption, passphrase string) (value string, err error)
	// EnsureIdentity loads identityText (the persisted serialized identity)
	// if non-empty, otherwise generates a new one. generated reports whether
	// a new identity was created, so the caller knows to persist newText.
	EnsureIdentity(identityText string) (newText string, generated bool, err error)
}

// ageEncryptor is the production Encryptor backed by filippo.io/age. It owns
// the plugin's own X25519 identity, used whenever a caller doesn't supply a
// passphrase.
type ageEncryptor struct {
	identityMu sync.RWMutex
	identity   *age.X25519Identity
}

// NewAgeEncryptor builds the production age/scrypt-backed Encryptor.
func NewAgeEncryptor() Encryptor {
	return &ageEncryptor{}
}

func (e *ageEncryptor) setIdentity(identity *age.X25519Identity) {
	e.identityMu.Lock()
	e.identity = identity
	e.identityMu.Unlock()
}

func (e *ageEncryptor) ownIdentity() *age.X25519Identity {
	e.identityMu.RLock()
	defer e.identityMu.RUnlock()
	return e.identity
}

func (e *ageEncryptor) EnsureIdentity(identityText string) (string, bool, error) {
	identityText = strings.TrimSpace(identityText)
	generated := identityText == ""
	if generated {
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			return "", false, fmt.Errorf("generate secrets manager identity: %w", err)
		}
		identityText = identity.String()
	}

	identity, err := age.ParseX25519Identity(identityText)
	if err != nil {
		return "", false, fmt.Errorf("parse secret manager identity: %w", err)
	}
	e.setIdentity(identity)
	return identityText, generated, nil
}

func (e *ageEncryptor) Encrypt(value, passphrase string) (string, string, error) {
	var recipient age.Recipient
	encryption := secretEncryptionIdentity
	if passphrase != "" {
		var err error
		recipient, err = age.NewScryptRecipient(passphrase)
		if err != nil {
			return "", "", fmt.Errorf("create passphrase recipient: %w", err)
		}
		encryption = secretEncryptionScrypt
	} else {
		identity := e.ownIdentity()
		if identity == nil {
			return "", "", fmt.Errorf("secrets manager is not initialized")
		}
		recipient = identity.Recipient()
	}

	var encrypted bytes.Buffer
	writer, err := age.Encrypt(&encrypted, recipient)
	if err != nil {
		return "", "", fmt.Errorf("encrypt secret: %w", err)
	}
	if _, err := io.WriteString(writer, value); err != nil {
		return "", "", fmt.Errorf("write encrypted secret: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("close encrypted secret: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted.Bytes()), encryption, nil
}

func (e *ageEncryptor) Decrypt(ciphertext, encryption, passphrase string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("secret contains invalid ciphertext")
	}

	var identities []age.Identity
	switch encryption {
	case secretEncryptionIdentity:
		identity := e.ownIdentity()
		if identity == nil {
			return "", fmt.Errorf("secrets manager is not initialized")
		}
		identities = []age.Identity{identity}
	case secretEncryptionScrypt:
		if passphrase == "" {
			return "", errPassphraseRequired
		}
		scryptIdentity, err := age.NewScryptIdentity(passphrase)
		if err != nil {
			return "", fmt.Errorf("create passphrase identity: %w", err)
		}
		identities = []age.Identity{scryptIdentity}
	default:
		return "", fmt.Errorf("unsupported secret encryption %q", encryption)
	}

	reader, err := age.Decrypt(bytes.NewReader(raw), identities...)
	if err != nil {
		if encryption == secretEncryptionScrypt {
			return "", fmt.Errorf("%w: %v", errInvalidPassphrase, err)
		}
		return "", fmt.Errorf("decrypt secret: %w", err)
	}
	cleartext, err := io.ReadAll(io.LimitReader(reader, maxSecretSize+1))
	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}
	if len(cleartext) > maxSecretSize {
		return "", fmt.Errorf("secret exceeds the %d byte limit", maxSecretSize)
	}
	return string(cleartext), nil
}
