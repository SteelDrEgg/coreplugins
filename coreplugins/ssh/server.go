package main

import (
	"context"
	"net/http"
	"sync"

	"google.golang.org/grpc"

	panel "minimalpanel/pluginsdk/grpc/proto"
)

const (
	// pluginDisplayName and pluginVersion are returned to the host registry.
	pluginDisplayName = "ssh"
	pluginVersion     = "0.1.0"

	// socketNamespace is the Socket.IO namespace served by this plugin.
	socketNamespace = "/ssh"

	eventConnectSSH    = "connect_ssh"
	eventTerminalInput = "terminal_input"
	eventResize        = "resize"
	eventDisconnect    = "disconnect"

	eventSSHConnected    = "ssh_connected"
	eventSSHDisconnected = "ssh_disconnected"
	eventSSHError        = "ssh_error"
	eventTerminalOutput  = "terminal_output"
)

// sshServer implements the minimalpanel Plugin gRPC service for SSH terminals.
//
// It owns active SSH sessions keyed by Socket.IO socket id and uses the host
// callback service to emit terminal output back to the browser.
type sshServer struct {
	panel.UnimplementedPluginServer

	mu            sync.RWMutex
	host          panel.HostClient
	hostConn      *grpc.ClientConn
	sshConfigPath string
	sessions      map[string]*sshSession
}

// newSSHServer constructs a plugin server with an empty session table.
func newSSHServer() *sshServer {
	return &sshServer{
		sessions: make(map[string]*sshSession),
	}
}

// Register declares the plugin's static terminal page and Socket.IO namespace.
func (s *sshServer) Register(ctx context.Context, req *panel.RegisterRequest) (*panel.RegisterReply, error) {
	if err := s.configureHostCallback(ctx, req); err != nil {
		return nil, err
	}

	s.sshConfigPath = req.GetParams()["ssh_config_path"]
	s.log(ctx, "info", "ssh plugin registered")

	return &panel.RegisterReply{
		Name:    pluginDisplayName,
		Version: pluginVersion,
		StaticMounts: []*panel.StaticMount{
			{
				Prefix:    "/pages/terminal.html",
				Directory: "$PLUGIN_ROOT/pages/terminal.html",
				Protected: true,
			},
			{
				Prefix:    "/assets/terminal/",
				Directory: "$PLUGIN_ROOT/assets/terminal",
				Protected: true,
			},
		},
		SocketNamespaces: []*panel.SocketNamespace{
			{
				Name:      socketNamespace,
				Events:    []string{eventConnectSSH, eventTerminalInput, eventResize, eventDisconnect},
				Protected: true,
			},
		},
	}, nil
}

// HandleHTTP is unused because this plugin serves only static files.
func (s *sshServer) HandleHTTP(context.Context, *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	return &panel.HTTPResponse{Status: http.StatusNotFound}, nil
}

// HandleSocketEvent routes browser terminal events to the SSH session layer.
func (s *sshServer) HandleSocketEvent(ctx context.Context, ev *panel.SocketEvent) (*panel.SocketEventReply, error) {
	switch ev.GetEvent() {
	case eventConnectSSH:
		return &panel.SocketEventReply{}, s.connectSSH(ctx, ev)
	case eventTerminalInput:
		return &panel.SocketEventReply{}, s.writeInput(ctx, ev)
	case eventResize:
		return &panel.SocketEventReply{}, s.resize(ctx, ev)
	case eventDisconnect:
		s.cleanup(ev.GetSocketId())
		return &panel.SocketEventReply{}, nil
	default:
		return &panel.SocketEventReply{}, nil
	}
}
