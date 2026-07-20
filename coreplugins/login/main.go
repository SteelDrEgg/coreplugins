//go:build wasip1

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
				{Prefix: "/pages/login.html", Directory: "$PLUGIN_ROOT/pages/login.html"},
				{Prefix: "/pages/logout.html", Directory: "$PLUGIN_ROOT/pages/logout.html"},
			},
		},
	}
	wasm.Register(plugin)
}

const webAssetsNamespace = "login"
