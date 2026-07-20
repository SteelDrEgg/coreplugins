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
				{Prefix: "/assets/js/sdk.js", Directory: "$PLUGIN_ROOT/assets/sdk.js"},
			},
		},
	}
	wasm.Register(plugin)
}

const webAssetsNamespace = "web-sdk"
