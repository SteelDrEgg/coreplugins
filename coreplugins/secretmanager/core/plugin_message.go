package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

// Domain failures remain SDK errors, carrying the same JSON code/message as
// HTTP failures. All successful replies, including reveal, are JSON objects.
type pluginMessageError struct{ response response }

func (e pluginMessageError) Error() string { raw, _ := json.Marshal(e.response); return string(raw) }

func newSecretMessageListener(p *secretManagerPlugin) *arupa.ServiceMessageListener {
	listener := arupa.NewServiceMessageListener()
	for topic, op := range map[string]operation{
		"list": opList, "create": opCreate, "update": opUpdate,
		"reveal": opReveal, "delete": opDelete,
	} {
		if err := listener.On(topic, func(ctx context.Context, message arupa.IncomingServiceMessage) (string, error) {
			source := strings.TrimSpace(message.Source)
			if source == "" {
				return "", pluginMessageError{errorResponse(secrets.NewError(secrets.ErrForbidden, "requesting plugin is required"))}
			}
			result, err := p.execute(ctx, secrets.PluginCaller(source), op, bytes.NewReader(message.Payload))
			if err != nil {
				p.logOperationError(ctx, err)
				return "", pluginMessageError{errorResponse(err)}
			}
			raw, err := json.Marshal(result)
			return string(raw), err
		}); err != nil {
			panic(fmt.Sprintf("register secret-manager handler: %v", err))
		}
	}
	if err := listener.OnAny(func(context.Context, arupa.IncomingServiceMessage) (string, error) {
		return "", pluginMessageError{errorResponse(secrets.NewError(secrets.ErrInvalidInput, "unsupported topic"))}
	}); err != nil {
		panic(fmt.Sprintf("register secret-manager fallback: %v", err))
	}
	return listener
}
