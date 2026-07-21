//go:build wasip1

package main

import (
	"net/http"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

func main() {}

func init() {
	manager := newSecretManagerPlugin()
	plugin := &wasm.Plugin{
		Registration: secretManagerRegistration(),
		Handler:      http.HandlerFunc(manager.handleHTTP),
		Messages:     manager.messages,
	}
	manager.plugin = plugin
	wasm.Register(plugin)
}

func secretManagerRegistration() arupa.Registration {
	return arupa.Registration{
		Name:    "secret-manager",
		Version: pluginVersion,
		HTTPRoutes: []arupa.HTTPRoute{
			{Method: http.MethodGet, Pattern: "/keys", Access: authenticatedAccess},
			{Method: http.MethodPost, Pattern: "/keys/add", Access: authenticatedAccess},
			{Method: http.MethodPost, Pattern: "/keys/update", Access: authenticatedAccess},
			{Method: http.MethodPost, Pattern: "/keys/reveal", Access: authenticatedAccess},
			{Method: http.MethodPost, Pattern: "/keys/delete", Access: authenticatedAccess},
		},
		StaticMounts: []arupa.StaticMount{
			{Prefix: "/keys/pages/index.html", Directory: "$PLUGIN_ROOT/pages/index.html", Access: authenticatedAccess},
			{Prefix: "/keys/icon/", Directory: "$PLUGIN_ROOT/icon", Access: authenticatedAccess},
		},
	}
}
