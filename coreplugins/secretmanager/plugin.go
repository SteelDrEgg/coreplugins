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
	paramIdentity     = "secretmgr.identity"
	paramSecretPrefix = "secretmgr.secret."
	paramPolicyPrefix = "secretmgr.policy."
	paramMetaPrefix   = "secretmgr.meta."

	maxSecretSize = 1 << 20

	secretEncryptionIdentity = "identity"
	secretEncryptionScrypt   = "scrypt"
)

type secretManagerPlugin struct {
	plugin   *arupawasm.Plugin
	mu       sync.RWMutex
	writeMu  sync.Mutex
	params   map[string]string
	identity *age.X25519Identity
	messages *arupa.MessageListener
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

	params := p.plugin.InitialParams()
	identityText := strings.TrimSpace(params[paramIdentity])
	if identityText == "" {
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			return fmt.Errorf("generate secrets manager identity: %w", err)
		}
		identityText = identity.String()
		params[paramIdentity] = identityText

		if err := p.plugin.PatchParams(ctx, arupa.ParamsPatch{Set: map[string]string{paramIdentity: identityText}}); err != nil {
			return fmt.Errorf("persist secrets manager identity: %w", err)
		}
	}

	identity, err := age.ParseX25519Identity(identityText)
	if err != nil {
		return fmt.Errorf("parse %s: %w", paramIdentity, err)
	}

	p.mu.Lock()
	p.params = params
	p.identity = identity
	p.mu.Unlock()
	return nil
}

var authenticatedAccess = arupa.AccessPolicy{RequireAuth: true}
