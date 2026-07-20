//go:build wasip1

package main

import pluginv1 "github.com/SteelDrEgg/arupa-sdk/golang/gen/wasm/proto"

func main() {}

func init() {
	// The SDK owns HTTP and plugin-message protocol adaptation. Registration
	// remains here because this plugin initializes and persists its identity
	// from the host-provided parameter set.
	pluginv1.RegisterPlugin(&secretManagerPlugin{})
}
