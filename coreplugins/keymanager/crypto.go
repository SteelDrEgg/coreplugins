//go:build wasip1

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
)

func (p *keyManagerPlugin) encryptSecret(value string) (string, error) {
	p.mu.RLock()
	identity := p.identity
	p.mu.RUnlock()
	if identity == nil {
		return "", fmt.Errorf("secrets manager is not initialized")
	}

	var encrypted bytes.Buffer
	writer, err := age.Encrypt(&encrypted, identity.Recipient())
	if err != nil {
		return "", fmt.Errorf("encrypt secret: %w", err)
	}
	if _, err := io.WriteString(writer, value); err != nil {
		return "", fmt.Errorf("write encrypted secret: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close encrypted secret: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted.Bytes()), nil
}

func (p *keyManagerPlugin) decryptSecret(name string) (string, error) {
	p.mu.RLock()
	params := cloneParams(p.params)
	identity := p.identity
	p.mu.RUnlock()
	if identity == nil {
		return "", fmt.Errorf("secrets manager is not initialized")
	}

	encoded, ok := params[paramSecretPrefix+name]
	if !ok || encoded == "" {
		return "", fmt.Errorf("secret %q was not found", name)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("secret %q contains invalid ciphertext", name)
	}
	reader, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
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

func (p *keyManagerPlugin) allowed(name, plugin string) bool {
	plugin = strings.TrimSpace(plugin)
	if plugin == "" {
		return false
	}
	p.mu.RLock()
	raw := p.params[paramPolicyPrefix+name]
	p.mu.RUnlock()
	plugins, err := decodePlugins(raw)
	if err != nil {
		return false
	}
	for _, candidate := range plugins {
		if candidate == plugin {
			return true
		}
	}
	return false
}
