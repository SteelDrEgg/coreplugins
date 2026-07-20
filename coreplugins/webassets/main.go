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
				{Prefix: "/assets/css/", Directory: "$PLUGIN_ROOT/assets/css"},
				{Prefix: "/assets/icon/", Directory: "$PLUGIN_ROOT/assets/icon"},
				{Prefix: "/assets/vendor/", Directory: "$PLUGIN_ROOT/assets/vendor"},
			},
		},
	}
	wasm.Register(plugin)
}

const webAssetsNamespace = "web-assets"
