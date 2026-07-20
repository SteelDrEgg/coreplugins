//go:build wasip1

package main

import (
	"context"
	"encoding/json"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

type pluginMessageRoute struct {
	topic   string
	handler func(*secretManagerPlugin, *panel.PluginMessage) (*panel.PluginMessageReply, error)
}

var pluginMessageRoutes = []pluginMessageRoute{
	{
		topic:   "secret-manager.secret.get",
		handler: (*secretManagerPlugin).handleSecretGetMessage,
	},
}

type secretGetRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

func (p *secretManagerPlugin) HandlePluginMessage(_ context.Context, req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	for _, route := range pluginMessageRoutes {
		if req.GetTopic() == route.topic {
			return route.handler(p, req)
		}
	}
	return pluginMessageError("unsupported topic")
}

func (p *secretManagerPlugin) handleSecretGetMessage(req *panel.PluginMessage) (*panel.PluginMessageReply, error) {
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

func pluginMessageError(message string) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{Error: message}, nil
}
