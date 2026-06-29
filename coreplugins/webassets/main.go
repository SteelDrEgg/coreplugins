//go:build wasip1

package main

import (
	"context"

	panel "minimalpanel/pluginsdk/wasm/proto"
)

func main() {}

func init() {
	panel.RegisterPlugin(webAssetsPlugin{})
}

type webAssetsPlugin struct{}

func (webAssetsPlugin) Register(_ context.Context, _ *panel.RegisterRequest) (*panel.RegisterReply, error) {
	return &panel.RegisterReply{
		Name:    "web-assets",
		Version: "0.1.0",
		StaticMounts: []*panel.StaticMount{
			{
				Prefix:    "/assets/",
				Directory: "$PLUGIN_ROOT/assets",
			},
		},
	}, nil
}

func (webAssetsPlugin) HandleHTTP(_ context.Context, _ *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	return &panel.HTTPResponse{Status: 404}, nil
}

func (webAssetsPlugin) HandleSocketEvent(_ context.Context, _ *panel.SocketEvent) (*panel.SocketEventReply, error) {
	return &panel.SocketEventReply{}, nil
}
