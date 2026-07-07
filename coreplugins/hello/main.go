//go:build wasip1

// Command hello is a minimal WASM example plugin.
//
// It demonstrates the full plugin contract:
//   - Register: declares an HTTP route and a Socket.IO namespace, logs through
//     the host, and seeds a value into the shared KV store.
//   - HandleHTTP: reads back the seeded KV value, and persists a new greeting
//     into Params through Host.PatchParams.
//   - HandleSocketEvent: replies to "ping" by emitting "pong" back to the
//     calling socket (the emit-in-reply pattern used by WASM plugins).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

func main() {}

func init() {
	panel.RegisterPlugin(helloPlugin{})
}

const (
	kvNamespace = "hello"
	kvGreeting  = "greeting"
)

type helloPlugin struct{}

func (helloPlugin) Register(ctx context.Context, req *panel.RegisterRequest) (*panel.RegisterReply, error) {
	host := panel.NewHost()

	_, _ = host.Log(ctx, &panel.LogRequest{
		Level:   "info",
		Message: "hello plugin registering, instance=" + req.GetInstanceId(),
	})

	// Plugins can read host-provided config params directly.
	greeting := "Hello from the hello plugin!"
	if g, ok := req.GetParams()["greeting"]; ok && g != "" {
		greeting = g
	}

	if _, err := host.KVSet(ctx, &panel.KVSetRequest{
		Namespace: kvNamespace,
		Key:       kvGreeting,
		Value:     []byte(greeting),
	}); err != nil {
		return nil, err
	}

	return &panel.RegisterReply{
		Name:    "hello",
		Version: pluginVersion,
		HttpRoutes: []*panel.HTTPRoute{
			{Method: "GET", Pattern: "/hello"},
			{Method: "GET", Pattern: "/hello/private", Protected: true},
			{Method: "POST", Pattern: "/hello/greeting", Protected: true},
		},
		SocketNamespaces: []*panel.SocketNamespace{
			{
				Name:            "/hello",
				Events:          []string{"ping", "ping_private"},
				ProtectedEvents: []string{"ping_private"},
			},
		},
	}, nil
}

func (helloPlugin) HandleHTTP(ctx context.Context, req *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	host := panel.NewHost()

	if req.GetMethod() == "POST" && req.GetPath() == "/hello/greeting" {
		return updateGreeting(ctx, host, req)
	}

	greeting := "Hello!"
	if reply, err := host.KVGet(ctx, &panel.KVGetRequest{Namespace: kvNamespace, Key: kvGreeting}); err == nil && reply.GetFound() {
		greeting = string(reply.GetValue())
	}

	body := fmt.Sprintf("%s\nYou requested: %s %s\n", greeting, req.GetMethod(), req.GetPath())
	if req.GetPath() == "/hello/private" {
		body += "Protected route access granted.\n"
	}
	return &panel.HTTPResponse{
		Status:  200,
		Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		Body:    []byte(body),
	}, nil
}

func updateGreeting(ctx context.Context, host panel.Host, req *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	greeting := strings.TrimSpace(string(req.GetBody()))
	if greeting == "" {
		return &panel.HTTPResponse{
			Status:  400,
			Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			Body:    []byte("Greeting body cannot be empty.\n"),
		}, nil
	}

	reply, err := host.PatchParams(ctx, &panel.ParamsPatchRequest{
		Set: map[string]string{kvGreeting: greeting},
	})
	if err != nil {
		return nil, err
	}
	if reply.GetError() != "" {
		return nil, fmt.Errorf("persist greeting param: %s", reply.GetError())
	}

	if _, err := host.KVSet(ctx, &panel.KVSetRequest{
		Namespace: kvNamespace,
		Key:       kvGreeting,
		Value:     []byte(greeting),
	}); err != nil {
		return nil, err
	}

	return &panel.HTTPResponse{
		Status:  200,
		Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		Body:    []byte("Greeting persisted to Params.\n"),
	}, nil
}

func (helloPlugin) HandleSocketEvent(ctx context.Context, ev *panel.SocketEvent) (*panel.SocketEventReply, error) {
	msg := "pong from hello plugin"
	if ev.GetEvent() == "ping_private" {
		msg = "protected pong from hello plugin"
	}
	payload, err := json.Marshal([]any{map[string]string{"message": msg}})
	if err != nil {
		return nil, err
	}

	return &panel.SocketEventReply{
		Emits: []*panel.EmitInstruction{
			{
				Namespace: ev.GetNamespace(),
				Target:    ev.GetSocketId(),
				Event:     "pong",
				Payload:   payload,
			},
		},
	}, nil
}

func (helloPlugin) HandlePluginMessage(_ context.Context, _ *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{}, nil
}
