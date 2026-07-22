//go:build wasip1

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"filippo.io/age"
	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	arupawasm "github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

const (
	maxSecretSize = 1 << 20

	secretEncryptionIdentity = "identity"
	secretEncryptionScrypt   = "scrypt"
)

type secretManagerPlugin struct {
	plugin     *arupawasm.Plugin
	identityMu sync.RWMutex
	writeMu    sync.Mutex
	identity   *age.X25519Identity
	store      *paramsStore
	messages   *arupa.MessageListener
}

type secretMeta struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	AllowedPlugins []string `json:"allowed_plugins"`
	UpdatedAt      string   `json:"updated_at"`
	Encryption     string   `json:"encryption,omitempty"`
}

// initialize loads the registration snapshot and persists an encryption
// identity before the plugin accepts requests. The regular wasm.Plugin owns
// protocol registration, so this package never needs generated ABI types.
func (p *secretManagerPlugin) initialize(ctx context.Context) error {
	if p.plugin == nil {
		return fmt.Errorf("secret-manager plugin is not configured")
	}

	p.store.load(p.plugin, p.plugin.InitialParams())
	identityText := strings.TrimSpace(p.store.identity())
	if identityText == "" {
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			return fmt.Errorf("generate secrets manager identity: %w", err)
		}
		identityText = identity.String()

		if err := p.store.setIdentity(ctx, identityText); err != nil {
			return fmt.Errorf("persist secrets manager identity: %w", err)
		}
	}

	identity, err := age.ParseX25519Identity(identityText)
	if err != nil {
		return fmt.Errorf("parse secret manager identity: %w", err)
	}

	p.identityMu.Lock()
	p.identity = identity
	p.identityMu.Unlock()
	return nil
}

var authenticatedAccess = arupa.AccessPolicy{RequireAuth: true}
