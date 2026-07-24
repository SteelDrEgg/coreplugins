package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/internal/secrets"
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

// newSecretMessageListener wires the inter-plugin message API. Every handler
// here is a thin adapter: decode the payload, build a secrets.PluginCaller
// from the host-authenticated message source, call into secrets.Service, and
// translate the result to the plain-string reply this API has always used.
func newSecretMessageListener(p *secretManagerPlugin) *arupa.ServiceMessageListener {
	listener := arupa.NewServiceMessageListener()
	handlers := map[string]arupa.ServiceMessageHandler{
		topicSecretGet:    p.handleSecretGetMessage,
		topicSecretList:   p.handleSecretListMessage,
		topicSecretAdd:    p.handleSecretAddMessage,
		topicSecretUpdate: p.handleSecretUpdateMessage,
		topicSecretDelete: p.handleSecretDeleteMessage,
	}
	for topic, handler := range handlers {
		if err := listener.On(topic, handler); err != nil {
			panic(fmt.Sprintf("register secret-manager message handler: %v", err))
		}
	}
	if err := listener.OnAny(func(context.Context, arupa.IncomingServiceMessage) (string, error) {
		return "", pluginMessageError("unsupported topic")
	}); err != nil {
		panic(fmt.Sprintf("register secret-manager fallback message handler: %v", err))
	}
	return listener
}

func (p *secretManagerPlugin) handleSecretGetMessage(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
	var payload secretGetRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	value, err := p.secrets.GetSecret(ctx, secrets.PluginCaller(source), payload.Name, payload.Passphrase)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}
	return value, nil
}

func (p *secretManagerPlugin) handleSecretListMessage(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	keys, err := p.secrets.ListSecrets(ctx, secrets.PluginCaller(source))
	if err != nil {
		return "", pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"keys": keys})
}

func (p *secretManagerPlugin) handleSecretAddMessage(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
	return p.writeSecretMessage(ctx, message, false)
}

func (p *secretManagerPlugin) handleSecretUpdateMessage(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
	return p.writeSecretMessage(ctx, message, true)
}

func (p *secretManagerPlugin) writeSecretMessage(ctx context.Context, message arupa.IncomingServiceMessage, update bool) (string, error) {
	var payload secretWriteRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	meta, err := p.secrets.WriteSecret(ctx, secrets.PluginCaller(source), secrets.WriteSecretInput{
		Name:           payload.Name,
		Description:    payload.Description,
		Value:          payload.Value,
		Passphrase:     payload.Passphrase,
		AllowedPlugins: payload.AllowedPlugins,
		Update:         update,
	})
	if err != nil {
		return "", pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": meta.Name})
}

func (p *secretManagerPlugin) handleSecretDeleteMessage(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
	var payload secretNameRequest
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", pluginMessageError("invalid request payload")
	}
	source, err := pluginMessageSource(message)
	if err != nil {
		return "", pluginMessageError(err.Error())
	}

	if err := p.secrets.DeleteSecret(ctx, secrets.PluginCaller(source), payload.Name); err != nil {
		return "", pluginMessageError(err.Error())
	}
	return pluginMessageJSON(map[string]any{"success": true, "name": payload.Name})
}

func pluginMessageSource(message arupa.IncomingServiceMessage) (string, error) {
	source := strings.TrimSpace(message.Source)
	if source == "" {
		return "", fmt.Errorf("requesting plugin is required")
	}
	return source, nil
}

func pluginMessageJSON(payload any) (string, error) {
	message, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(message), nil
}
