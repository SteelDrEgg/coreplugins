//go:build wasip1

package main

import (
	"context"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

func main() {}

func init() {
	panel.RegisterPlugin(pluginManagerPlugin{})
}

type pluginManagerPlugin struct{}

func (pluginManagerPlugin) Register(_ context.Context, _ *panel.RegisterRequest) (*panel.RegisterReply, error) {
	return &panel.RegisterReply{
		Name:    "plugin-manager",
		Version: "0.1.0",
		StaticMounts: []*panel.StaticMount{
			{
				Prefix:    "/plugins/pages/plugins.html",
				Directory: "$PLUGIN_ROOT/pages/plugins.html",
				Protected: true,
			},
			{
				Prefix:    "/plugins/icon/",
				Directory: "$PLUGIN_ROOT/icon",
				Protected: true,
			},
		},
	}, nil
}

func (pluginManagerPlugin) HandleHTTP(_ context.Context, _ *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	return &panel.HTTPResponse{Status: 404}, nil
}

func (pluginManagerPlugin) HandleSocketEvent(_ context.Context, _ *panel.SocketEvent) (*panel.SocketEventReply, error) {
	return &panel.SocketEventReply{}, nil
}

func (pluginManagerPlugin) HandlePluginMessage(_ context.Context, _ *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{}, nil
}
