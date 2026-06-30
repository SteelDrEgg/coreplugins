package main

import (
	"context"
	"fmt"

	hcplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	panel "minimalpanel/pluginsdk/grpc/proto"
)

// sshPlugin adapts sshServer to the HashiCorp go-plugin gRPC interface.
type sshPlugin struct {
	hcplugin.NetRPCUnsupportedPlugin
}

// GRPCServer registers the minimalpanel Plugin service implemented by sshServer.
func (p *sshPlugin) GRPCServer(_ *hcplugin.GRPCBroker, s *grpc.Server) error {
	panel.RegisterPluginServer(s, newSSHServer())
	return nil
}

// GRPCClient is unused because this process only serves plugin RPCs.
func (p *sshPlugin) GRPCClient(context.Context, *hcplugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return nil, fmt.Errorf("plugin process does not use GRPCClient")
}
