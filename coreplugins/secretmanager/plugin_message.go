//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

const (
	topicSecretGet    = "secret.get"
	topicSecretList   = "secret.list"
	topicSecretAdd    = "secret.add"
	topicSecretUpdate = "secret.update"
	topicSecretDelete = "secret.delete"
)

type secretGetRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

type pluginMessageError string

func (e pluginMessageError) Error() string { return string(e) }

func newSecretManagerPlugin() *secretManagerPlugin {
	p := &secretManagerPlugin{messages: arupa.NewMessageListener(), store: newParamsStore()}
	for topic, handler := range map[string]arupa.MessageHandler{
		topicSecretGet:    p.handleSecretGetMessage,
		topicSecretList:   p.handleSecretListMessage,
		topicSecretAdd:    p.handleSecretAddMessage,
		topicSecretUpdate: p.handleSecretUpdateMessage,
		topicSecretDelete: p.handleSecretDeleteMessage,
	} {
		if err := p.messages.On(topic, handler); err != nil {
			panic(fmt.Sprintf("register secret-manager message handler: %v", err))
		}
	}
	if err := p.messages.OnAny(func(context.Context, arupa.IncomingMessage) (string, error) {
		return "", pluginMessageError("unsupported topic")
	}); err != nil {
		panic(fmt.Sprintf("register secret-manager fallback message handler: %v", err))
	}
	return p
}

func (p *secretManagerPlugin) handleSecretGetMessage(_ context.Context, message arupa.IncomingMessage) (string, error) {
	var payload secretGetRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	if !p.allowed(payload.Name, message.Source) {
		return "", pluginMessageError("plugin is not allowed to access this secret")
	}
	if _, err := p.store.encryption(payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	value, err := p.decryptSecret(payload.Name, payload.Passphrase)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}
	return value, nil
}

func (p *secretManagerPlugin) handleSecretListMessage(_ context.Context, message arupa.IncomingMessage) (string, error) {
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	keys, err := p.store.listSecrets()
	if err != nil {
		return "", pluginMessageError(err.Error())
	}
	visibleKeys := make([]secretMeta, 0, len(keys))
	for _, key := range keys {
		if allowedPlugin(key.AllowedPlugins, source) {
			visibleKeys = append(visibleKeys, key)
		}
	}
	return pluginMessageJSON(map[string]any{"keys": visibleKeys})
}

func (p *secretManagerPlugin) handleSecretAddMessage(ctx context.Context, message arupa.IncomingMessage) (string, error) {
	return p.writeSecretMessage(ctx, message, false)
}

func (p *secretManagerPlugin) handleSecretUpdateMessage(ctx context.Context, message arupa.IncomingMessage) (string, error) {
	return p.writeSecretMessage(ctx, message, true)
}

func (p *secretManagerPlugin) writeSecretMessage(ctx context.Context, message arupa.IncomingMessage, update bool) (string, error) {
	var payload secretWriteRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	if exists := p.store.hasSecret(payload.Name); exists != update {
		if update {
			return "", pluginMessageError("secret not found")
		}
		return "", pluginMessageError("secret already exists")
	}
	if update && !p.allowed(payload.Name, source) {
		return "", pluginMessageError("plugin is not allowed to access this secret")
	}

	allowedPlugins := payload.AllowedPlugins
	if !update {
		allowedPlugins = append(allowedPlugins, source)
	}
	allowedPlugins, err = normalizePlugins(allowedPlugins)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	var ciphertext string
	encryption := secretEncryptionIdentity
	if update && payload.Value == "*" {
		if payload.Passphrase != "" {
			return "", pluginMessageError("you must edit both value and passphrase")
		}
		var ok bool
		ciphertext, ok = p.store.ciphertext(payload.Name)
		if !ok {
			return "", pluginMessageError("secret not found")
		}
		encryption, err = p.store.encryption(payload.Name)
		if err != nil {
			return "", pluginMessageError(err.Error())
		}
	} else {
		ciphertext, encryption, err = p.encryptSecret(payload.Value, payload.Passphrase)
		if err != nil {
			return "", pluginMessageError(err.Error())
		}
	}

	meta := secretMeta{
		Name:           payload.Name,
		Description:    strings.TrimSpace(payload.Description),
		AllowedPlugins: allowedPlugins,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		Encryption:     encryption,
	}
	if err := p.store.putSecret(ctx, ciphertext, meta); err != nil {
		return "", pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": payload.Name})
}

func (p *secretManagerPlugin) handleSecretDeleteMessage(ctx context.Context, message arupa.IncomingMessage) (string, error) {
	var payload secretNameRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	if err := validateSecretName(payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	if !p.store.hasSecret(payload.Name) {
		return "", pluginMessageError("secret not found")
	}
	if !p.allowed(payload.Name, source) {
		return "", pluginMessageError("plugin is not allowed to access this secret")
	}
	if err := p.store.deleteSecret(ctx, payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": payload.Name})
}

func pluginMessageSource(message arupa.IncomingMessage) (string, error) {
	source := strings.TrimSpace(message.Source)
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

func pluginMessageJSON(payload any) (string, error) {
	message, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(message), nil
}
