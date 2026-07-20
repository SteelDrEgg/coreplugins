//go:build wasip1

//
//package main
//
//import (
//	"context"
//	"encoding/json"
//
//	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
//)
//
//func main() {}
//
//func init() {
//	panel.RegisterPlugin(webAssetsPlugin{})
//}
//
//type webAssetsPlugin struct{}
//
//const webAssetsNamespace = "web-sdk"
//
//func (webAssetsPlugin) Register(ctx context.Context, _ *panel.RegisterRequest) (*panel.RegisterReply, error) {
//	host := panel.NewHost()
//	urls := map[string]string{
//		"sdk":       "/assets/js/sdk.js",
//		"languages": "/assets/js/lang.json",
//	}
//	urlsJSON, err := json.Marshal(urls)
//	if err == nil {
//		_, _ = host.KVSet(ctx, &panel.KVSetRequest{
//			Namespace: webAssetsNamespace,
//			Key:       "urls",
//			Value:     urlsJSON,
//		})
//	}
//	for k, v := range urls {
//		_, _ = host.KVSet(ctx, &panel.KVSetRequest{
//			Namespace: webAssetsNamespace,
//			Key:       k,
//			Value:     []byte(v),
//		})
//	}
//
//	return &panel.RegisterReply{
//		Name:    "web-sdk",
//		Version: pluginVersion,
//		StaticMounts: []*panel.StaticMount{
//			{
//				Prefix:    "/assets/js/sdk.js",
//				Directory: "$PLUGIN_ROOT/assets/sdk.js",
//				//Directory: "coreplugins/websdk/assets/sdk.js",
//			},
//		},
//	}, nil
//}
//
//func (webAssetsPlugin) HandleHTTP(_ context.Context, _ *panel.HTTPRequest) (*panel.HTTPResponse, error) {
//	return &panel.HTTPResponse{Status: 404}, nil
//}
//
//func (webAssetsPlugin) HandleSocketEvent(_ context.Context, _ *panel.SocketEvent) (*panel.SocketEventReply, error) {
//	return &panel.SocketEventReply{}, nil
//}
//
//func (webAssetsPlugin) HandlePluginMessage(_ context.Context, _ *panel.PluginMessage) (*panel.PluginMessageReply, error) {
//	return &panel.PluginMessageReply{}, nil
//}
//

package main

import (
	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

func main() {}

func init() {
	plugin := &wasm.Plugin{
		Registration: arupa.Registration{
			Name:    webAssetsNamespace,
			Version: pluginVersion,
			StaticMounts: []arupa.StaticMount{
				{Prefix: "/assets/js/sdk.js", Directory: "$PLUGIN_ROOT/assets/sdk.js"},
			},
		},
	}
	wasm.Register(plugin)
}

const webAssetsNamespace = "web-sdk"
