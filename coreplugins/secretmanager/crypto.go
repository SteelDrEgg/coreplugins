//go:build wasip1

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"filippo.io/age"
)

var (
	errPassphraseRequired = errors.New("passphrase required")
	errInvalidPassphrase  = errors.New("invalid passphrase")
)

func (p *secretManagerPlugin) encryptSecret(value, passphrase string) (string, string, error) {
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
		p.identityMu.RLock()
		identity := p.identity
		p.identityMu.RUnlock()
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

func (p *secretManagerPlugin) decryptSecret(name, passphrase string) (string, error) {
	encoded, ok := p.store.ciphertext(name)
	p.identityMu.RLock()
	identity := p.identity
	p.identityMu.RUnlock()
	if !ok || encoded == "" {
		return "", fmt.Errorf("secret %q was not found", name)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("secret %q contains invalid ciphertext", name)
	}
	encryption, err := p.store.encryption(name)
	if err != nil {
		return "", err
	}
	var identities []age.Identity
	switch encryption {
	case secretEncryptionIdentity:
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

	reader, err := age.Decrypt(bytes.NewReader(ciphertext), identities...)
	if err != nil {
		if encryption == secretEncryptionScrypt {
			return "", fmt.Errorf("%w: %v", errInvalidPassphrase, err)
		}
		return "", fmt.Errorf("decrypt secret %q: %w", name, err)
	}
	cleartext, err := io.ReadAll(io.LimitReader(reader, maxSecretSize+1))
	if err != nil {
		return "", fmt.Errorf("read secret %q: %w", name, err)
	}
	if len(cleartext) > maxSecretSize {
		return "", fmt.Errorf("secret %q exceeds the %d byte limit", name, maxSecretSize)
	}
	return string(cleartext), nil
}

func (p *secretManagerPlugin) allowed(name, plugin string) bool {
	return p.store.allows(name, plugin)
}
