//go:build wasip1

package main

import (
	"context"
	"encoding/json"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
)

func main() {}

func init() {
	panel.RegisterPlugin(webAssetsPlugin{})
}

type webAssetsPlugin struct{}

const webAssetsNamespace = "web-assets"

func (webAssetsPlugin) Register(ctx context.Context, _ *panel.RegisterRequest) (*panel.RegisterReply, error) {
	host := panel.NewHost()
	urls := map[string]string{
		"css_prefix":  "/assets/css/",
		"icon_prefix": "/assets/icon/",
		"scheme_css":  "/assets/css/scheme.css",
	}
	urlsJSON, err := json.Marshal(urls)
	if err == nil {
		_, _ = host.KVSet(ctx, &panel.KVSetRequest{
			Namespace: webAssetsNamespace,
			Key:       "urls",
			Value:     urlsJSON,
		})
	}
	for k, v := range urls {
		_, _ = host.KVSet(ctx, &panel.KVSetRequest{
			Namespace: webAssetsNamespace,
			Key:       k,
			Value:     []byte(v),
		})
	}

	return &panel.RegisterReply{
		Name:    "web-assets",
		Version: pluginVersion,
		StaticMounts: []*panel.StaticMount{
			{
				Prefix:    "/assets/css/",
				Directory: "$PLUGIN_ROOT/assets/css",
			},
			{
				Prefix:    "/assets/icon/",
				Directory: "$PLUGIN_ROOT/assets/icon",
			},
			//{
			//	Prefix:    "/pages/login.html",
			//	Directory: "$PLUGIN_ROOT/assets/icon/terminal.svg",
			//},
		},
	}, nil
}

func (webAssetsPlugin) HandleHTTP(_ context.Context, _ *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	return &panel.HTTPResponse{Status: 404}, nil
}

func (webAssetsPlugin) HandleSocketEvent(_ context.Context, _ *panel.SocketEvent) (*panel.SocketEventReply, error) {
	return &panel.SocketEventReply{}, nil
}

func (webAssetsPlugin) HandlePluginMessage(_ context.Context, _ *panel.PluginMessage) (*panel.PluginMessageReply, error) {
	return &panel.PluginMessageReply{}, nil
}
