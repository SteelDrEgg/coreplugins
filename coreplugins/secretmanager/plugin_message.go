//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

type pluginMessageRoute struct {
	topic   string
	handler func(*secretManagerPlugin, context.Context, *panel.PluginMessage) (*panel.PluginMessageReply, error)
}

var pluginMessageRoutes = []pluginMessageRoute{
	{
		topic:   "secret-manager.secret.get",
		handler: (*secretManagerPlugin).handleSecretGetMessage,
	},
	{
		topic:   "secret-manager.secret.list",
		handler: (*secretManagerPlugin).handleSecretListMessage,
	},
	{
		topic:   "secret-manager.secret.add",
		handler: (*secretManagerPlugin).handleSecretAddMessage,
	},
	{
		topic:   "secret-manager.secret.update",
		handler: (*secretManagerPlugin).handleSecretUpdateMessage,
	},
	{
		topic:   "secret-manager.secret.delete",
		handler: (*secretManagerPlugin).handleSecretDeleteMessage,
	},
}

type secretGetRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

func (p *secretManagerPlugin) HandlePluginMessage(ctx context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	for _, route := range pluginMessageRoutes {
		if req.GetTopic() == route.topic {
			return route.handler(p, ctx, req)
		}
	}
	return pluginMessageError("unsupported topic")
}

func (p *secretManagerPlugin) handleSecretGetMessage(_ context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
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
	if _, err := p.secretEncryption(payload.Name); err != nil {
		return pluginMessageError(err.Error())
	}
	value, err := p.decryptSecret(payload.Name, payload.Passphrase)
	if err != nil {
		return pluginMessageError(err.Error())
	}
	return &panel.PluginMessageReply{Message: value}, nil
}

func (p *secretManagerPlugin) handleSecretListMessage(_ context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	source, err := pluginMessageSource(req)
	if err != nil {
		return pluginMessageError(err.Error())
	}

	p.mu.RLock()
	params := cloneParams(p.params)
	p.mu.RUnlock()

	keys, err := listSecretMeta(params)
	if err != nil {
		return pluginMessageError(err.Error())
	}
	visibleKeys := make([]secretMeta, 0, len(keys))
	for _, key := range keys {
		if allowedPlugin(key.AllowedPlugins, source) {
			visibleKeys = append(visibleKeys, key)
		}
	}
	return pluginMessageJSON(map[string]any{"keys": visibleKeys})
}

func (p *secretManagerPlugin) handleSecretAddMessage(ctx context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	return p.writeSecretMessage(ctx, req, false)
}

func (p *secretManagerPlugin) handleSecretUpdateMessage(ctx context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	return p.writeSecretMessage(ctx, req, true)
}

func (p *secretManagerPlugin) writeSecretMessage(ctx context.Context, req *panel.PluginMessage, update bool) (*panel.PluginMessageReply, error) {
	var payload secretWriteRequest
	if err := json.Unmarshal(req.GetPayload(), &payload); err != nil {
		return pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return pluginMessageError(err.Error())
	}
	source, err := pluginMessageSource(req)
	if err != nil {
		return pluginMessageError(err.Error())
	}

	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	if exists := p.secretExists(payload.Name); exists != update {
		if update {
			return pluginMessageError("secret not found")
		}
		return pluginMessageError("secret already exists")
	}
	if update && !p.allowed(payload.Name, source) {
		return pluginMessageError("plugin is not allowed to access this secret")
	}

	allowedPlugins := payload.AllowedPlugins
	if !update {
		allowedPlugins = append(allowedPlugins, source)
	}
	allowedPlugins, err = normalizePlugins(allowedPlugins)
	if err != nil {
		return pluginMessageError(err.Error())
	}

	var ciphertext string
	encryption := secretEncryptionIdentity
	if update && payload.Value == "*" {
		if payload.Passphrase != "" {
			return pluginMessageError("you must edit both value and passphrase")
		}
		var ok bool
		ciphertext, ok = p.secretCiphertext(payload.Name)
		if !ok {
			return pluginMessageError("secret not found")
		}
		encryption, err = p.secretEncryption(payload.Name)
		if err != nil {
			return pluginMessageError(err.Error())
		}
	} else {
		ciphertext, encryption, err = p.encryptSecret(payload.Value, payload.Passphrase)
		if err != nil {
			return pluginMessageError(err.Error())
		}
	}

	meta := secretMeta{
		Name:           payload.Name,
		Description:    strings.TrimSpace(payload.Description),
		AllowedPlugins: allowedPlugins,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		Encryption:     encryption,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return pluginMessageError("encode metadata")
	}
	policyJSON, err := json.Marshal(allowedPlugins)
	if err != nil {
		return pluginMessageError("encode access policy")
	}
	if err := p.patchParams(ctx, map[string]string{
		paramSecretPrefix + payload.Name: ciphertext,
		paramPolicyPrefix + payload.Name: string(policyJSON),
		paramMetaPrefix + payload.Name:   string(metaJSON),
	}, nil); err != nil {
		return pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": payload.Name})
}

func (p *secretManagerPlugin) handleSecretDeleteMessage(ctx context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	var payload secretNameRequest
	if err := json.Unmarshal(req.GetPayload(), &payload); err != nil {
		return pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return pluginMessageError(err.Error())
	}
	source, err := pluginMessageSource(req)
	if err != nil {
		return pluginMessageError(err.Error())
	}

	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	if !p.secretExists(payload.Name) {
		return pluginMessageError("secret not found")
	}
	if !p.allowed(payload.Name, source) {
		return pluginMessageError("plugin is not allowed to access this secret")
	}
	if err := p.patchParams(ctx, nil, []string{
		paramSecretPrefix + payload.Name,
		paramPolicyPrefix + payload.Name,
		paramMetaPrefix + payload.Name,
	}); err != nil {
		return pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": payload.Name})
}

func pluginMessageSource(req *panel.PluginMessage) (string, error) {
	source := strings.TrimSpace(req.GetSource())
	if source == "" {
		return "", fmt.Errorf("requesting plugin is required")
	}
	return source, nil
}

func allowedPlugin(plugins []string, source string) bool {
	for _, plugin := range plugins {
		if plugin == source {
			return true
		}
	}
	return false
}

func pluginMessageJSON(payload any) (*panel.PluginMessageReply, error) {
	message, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &panel.PluginMessageReply{Message: string(message)}, nil
}

func pluginMessageError(message string) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{Error: message}, nil
}
