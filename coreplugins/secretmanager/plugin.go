//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"filippo.io/age"
	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

const (
	paramIdentity     = "secretmgr.identity"
	paramSecretPrefix = "secretmgr.secret."
	paramPolicyPrefix = "secretmgr.policy."
	paramMetaPrefix   = "secretmgr.meta."

	topicSecretGet = "secret-manager.secret.get"

	maxSecretSize = 1 << 20

	secretEncryptionIdentity = "identity"
	secretEncryptionScrypt   = "scrypt"
)

type secretManagerPlugin struct {
	mu       sync.RWMutex
	writeMu  sync.Mutex
	params   map[string]string
	identity *age.X25519Identity
}

type secretMeta struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	AllowedPlugins []string `json:"allowed_plugins"`
	UpdatedAt      string   `json:"updated_at"`
	Encryption     string   `json:"encryption,omitempty"`
}

type secretGetRequest struct {
	Name string `json:"name"`
}

func (p *secretManagerPlugin) Register(ctx context.Context, req *panel.RegisterRequest) (*panel.RegisterReply, error) {
	params := cloneParams(req.GetParams())
	identityText := strings.TrimSpace(params[paramIdentity])
	if identityText == "" {
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			return nil, fmt.Errorf("generate secrets manager identity: %w", err)
		}
		identityText = identity.String()
		params[paramIdentity] = identityText

		if err := p.patchParams(ctx, map[string]string{paramIdentity: identityText}, nil); err != nil {
			return nil, fmt.Errorf("persist secrets manager identity: %w", err)
		}
	}

	identity, err := age.ParseX25519Identity(identityText)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", paramIdentity, err)
	}

	p.mu.Lock()
	p.params = params
	p.identity = identity
	p.mu.Unlock()

	access := &panel.AccessPolicy{RequireAuth: true}
	return &panel.RegisterReply{
		Name:    "secret-manager",
		Version: pluginVersion,
		HttpRoutes: []*panel.HTTPRoute{
			{Method: http.MethodGet, Pattern: "/keys", Access: access},
			{Method: http.MethodPost, Pattern: "/keys/add", Access: access},
			{Method: http.MethodPost, Pattern: "/keys/update", Access: access},
			{Method: http.MethodPost, Pattern: "/keys/reveal", Access: access},
			{Method: http.MethodPost, Pattern: "/keys/delete", Access: access},
		},
		StaticMounts: []*panel.StaticMount{
			{
				Prefix:    "/keys/pages/index.html",
				Directory: "$PLUGIN_ROOT/pages/index.html",
				Access:    access,
			},
			{
				Prefix:    "/keys/icon/",
				Directory: "$PLUGIN_ROOT/icon",
				Access:    access,
			},
		},
	}, nil
}

func (p *secretManagerPlugin) HandlePluginMessage(_ context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	if req.GetTopic() != topicSecretGet {
		return pluginMessageError("unsupported topic")
	}

	var payload secretGetRequest
	if err := json.Unmarshal(req.GetPayload(), &payload); err != nil {
		return pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return pluginMessageError(err.Error())
	}
	if !p.allowed(payload.Name, req.GetSource()) {
		return pluginMessageError("plugin is not allowed to access this secret")
	}
	encryption, err := p.secretEncryption(payload.Name)
	if err != nil {
		return pluginMessageError(err.Error())
	}
	if encryption == secretEncryptionScrypt {
		return pluginMessageError("secret requires a passphrase")
	}

	value, err := p.decryptSecret(payload.Name, "")
	if err != nil {
		return pluginMessageError(err.Error())
	}
	return &panel.PluginMessageReply{Message: value}, nil
}

func (p *secretManagerPlugin) HandleSocketEvent(context.Context, *panel.SocketEvent) (*panel.SocketEventReply, error) {
	return &panel.SocketEventReply{}, nil
}
