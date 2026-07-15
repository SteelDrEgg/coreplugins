//go:build wasip1

package main

import (
	"encoding/json"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

func jsonResponse(status int, payload any) (*panel.HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &panel.HTTPResponse{
		Status: int32(status),
		Headers: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Cache-Control": "no-store",
		},
		Body: body,
	}, nil
}

func pluginMessageError(message string) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{Error: message}, nil
}
