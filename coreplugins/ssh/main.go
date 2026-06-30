package main

import hcplugin "github.com/hashicorp/go-plugin"

// pluginName is the dispense name expected by the host's default gRPC preset.
const pluginName = "default_grpc"

// handshake must match the host-side minimalpanel plugin manager handshake.
var handshake = hcplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MINIMALPANEL_PLUGIN",
	MagicCookieValue: "minimalpanel",
}

// main serves the SSH terminal plugin as a HashiCorp go-plugin gRPC process.
func main() {
	hcplugin.Serve(&hcplugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]hcplugin.Plugin{
			pluginName: &sshPlugin{},
		},
		GRPCServer: hcplugin.DefaultGRPCServer,
	})
}
