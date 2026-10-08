// Package agecrypto implements identity and passphrase protection using age.
package agecrypto

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

type Cipher struct{ identity *age.X25519Identity }

var _ secrets.Cipher = (*Cipher)(nil)

// New loads an identity or generates one for a new installation. The caller
// persists Identity before exposing any operations.
func New(identityText string) (*Cipher, error) {
	var identity *age.X25519Identity
	var err error
	if strings.TrimSpace(identityText) == "" {
		identity, err = age.GenerateX25519Identity()
	} else {
		identity, err = age.ParseX25519Identity(strings.TrimSpace(identityText))
	}
	if err != nil {
		return nil, fmt.Errorf("initialize secret-manager identity: %w", err)
	}
	return &Cipher{identity: identity}, nil
}

func (c *Cipher) Identity() string { return c.identity.String() }

func (c *Cipher) Encrypt(ctx context.Context, in secrets.ValueInput) (secrets.EncryptedValue, error) {
	if err := ctx.Err(); err != nil {
		return secrets.EncryptedValue{}, err
	}
	if err := in.Validate(); err != nil {
		return secrets.EncryptedValue{}, err
	}
	var recipient age.Recipient
	if in.Protection == secrets.ProtectionPassphrase {
		var err error
		recipient, err = age.NewScryptRecipient(in.Passphrase)
		if err != nil {
			return secrets.EncryptedValue{}, err
		}
	} else {
		recipient = c.identity.Recipient()
	}
	var encrypted bytes.Buffer
	writer, err := age.Encrypt(&encrypted, recipient)
	if err != nil {
		return secrets.EncryptedValue{}, err
	}
	if _, err := io.WriteString(writer, in.Plaintext); err != nil {
		return secrets.EncryptedValue{}, err
	}
	if err := writer.Close(); err != nil {
		return secrets.EncryptedValue{}, err
	}
	if err := ctx.Err(); err != nil {
		return secrets.EncryptedValue{}, err
	}
	return secrets.EncryptedValue{Protection: in.Protection, Ciphertext: base64.StdEncoding.EncodeToString(encrypted.Bytes())}, nil
}

func (c *Cipher) Decrypt(ctx context.Context, value secrets.EncryptedValue, passphrase string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(value.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext encoding")
	}
	var identity age.Identity
	switch value.Protection {
	case secrets.ProtectionIdentity:
		identity = c.identity
	case secrets.ProtectionPassphrase:
		if passphrase == "" {
			return "", secrets.PassphraseRequired
		}
		identity, err = age.NewScryptIdentity(passphrase)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid protection mode")
	}
	reader, err := age.Decrypt(bytes.NewReader(raw), identity)
	if err != nil {
		var noMatch *age.NoIdentityMatchError
		if value.Protection == secrets.ProtectionPassphrase && errors.As(err, &noMatch) {
			return "", secrets.InvalidPassphrase
		}
		return "", fmt.Errorf("decrypt secret: %w", err)
	}
	cleartext, err := io.ReadAll(io.LimitReader(reader, secrets.MaxSecretSize+1))
	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}
	if len(cleartext) > secrets.MaxSecretSize {
		return "", fmt.Errorf("secret exceeds the %d byte limit", secrets.MaxSecretSize)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return string(cleartext), nil
}
