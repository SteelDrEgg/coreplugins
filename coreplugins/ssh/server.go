package main

import (
	"context"
	"net/http"
	"sync"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	pluginv1 "github.com/SteelDrEgg/arupa-sdk/golang/gen/grpc"
	arupagrpc "github.com/SteelDrEgg/arupa-sdk/golang/grpc"
)

const (
	pluginDisplayName = "ssh"

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

// sshServer implements the Plugin gRPC service for SSH terminals.
//
// It owns active SSH sessions keyed by Socket.IO socket id and uses the host
// callback service to emit terminal output back to the browser.
type sshServer struct {
	pluginv1.UnimplementedPluginServer

	sdk           *arupagrpc.HTTPPlugin
	host          hostBridge
	mu            sync.RWMutex
	sshConfigPath string
	sessions      map[string]*sshSession
	pending       map[string]*pendingConnection

	settingsMu      sync.RWMutex
	settings        map[string]savedConnection
	settingsWriteMu sync.Mutex
}

// newSSHServer constructs a plugin server with an empty session table.
func newSSHServer() *sshServer {
	s := &sshServer{
		sessions: make(map[string]*sshSession),
		pending:  make(map[string]*pendingConnection),
		settings: make(map[string]savedConnection),
	}
	events := arupa.NewEventBus()
	_ = events.On(eventConnectSSH, s.connectSSH)
	_ = events.On(eventTerminalInput, s.writeInput)
	_ = events.On(eventResize, s.resize)
	_ = events.On(eventDisconnect, func(_ context.Context, event arupa.SocketEvent, _ arupa.Emitter) error {
		s.cleanup(event.SocketID)
		return nil
	})

	authenticated := arupa.AccessPolicy{RequireAuth: true}
	s.sdk = &arupagrpc.HTTPPlugin{
		Registration: arupa.Registration{
			Name:    pluginDisplayName,
			Version: pluginVersion,
			StaticMounts: []arupa.StaticMount{
				{
					Prefix:    "/ssh/pages/terminal.html",
					Directory: "$PLUGIN_ROOT/pages/terminal.html",
					Access:    authenticated,
				},
				{
					Prefix:    "/ssh/assets/",
					Directory: "$PLUGIN_ROOT/assets",
					Access:    authenticated,
				},
			},
			HTTPRoutes: []arupa.HTTPRoute{
				{Method: http.MethodGet, Pattern: savedConnectionsPath, Access: authenticated},
				{Method: http.MethodPost, Pattern: savedConnectionsPath, Access: authenticated},
			},
			SocketNamespaces: []arupa.SocketNamespace{
				{
					Name:   socketNamespace,
					Events: []string{eventConnectSSH, eventTerminalInput, eventResize, eventDisconnect},
					Access: authenticated,
				},
			},
		},
		Handler: http.HandlerFunc(s.handleConnectionsHTTP),
		Events:  events,
	}
	return s
}

// Register declares the plugin's static terminal page and Socket.IO namespace.
func (s *sshServer) Register(ctx context.Context, req *pluginv1.RegisterRequest) (*pluginv1.RegisterReply, error) {
	reply, err := s.sdk.Register(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.host.configure(ctx, req); err != nil {
		return nil, err
	}

	s.sshConfigPath = req.GetParams()["ssh_config_path"]
	if err := s.loadSavedConnections(req.GetParams()[savedConnectionsParam]); err != nil {
		return nil, err
	}
	s.log(ctx, "info", "ssh plugin registered")
	return reply, nil
}

// HandleHTTP delegates protocol conversion to the SDK and application routing
// to a standard net/http handler.
func (s *sshServer) HandleHTTP(ctx context.Context, req *pluginv1.HTTPRequest) (*pluginv1.HTTPResponse, error) {
	return s.sdk.HandleHTTP(ctx, req)
}

// HandleSocketEvent delegates protocol conversion and event dispatch to the
// SDK EventBus.
func (s *sshServer) HandleSocketEvent(ctx context.Context, event *pluginv1.SocketEvent) (*pluginv1.SocketEventReply, error) {
	return s.sdk.HandleSocketEvent(ctx, event)
}

// HandlePluginMessage is unused by the SSH plugin.
func (s *sshServer) HandlePluginMessage(context.Context, *pluginv1.PluginMessage) (*pluginv1.PluginMessageReply, error) {
	return &pluginv1.PluginMessageReply{}, nil
}
