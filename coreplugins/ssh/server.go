package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
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

// sshServer owns SSH terminal application state keyed by Socket.IO socket id.
// The SDK owns the gRPC service and host callback protocol.
type sshServer struct {
	sdk           *arupagrpc.Plugin
	settingsStore sshSettingsStore
	events        *arupa.SocketListener
	mu            sync.RWMutex
	sshConfigPath string
	sessions      map[string]*sshSession
	pending       map[string]*pendingConnection

	settingsMu      sync.RWMutex
	settings        map[string]savedConnection
	settingsWriteMu sync.Mutex
}

// newSSHServer constructs an SSH application and its SDK gRPC adapter.
func newSSHServer() *sshServer {
	s := &sshServer{
		sessions: make(map[string]*sshSession),
		pending:  make(map[string]*pendingConnection),
		settings: make(map[string]savedConnection),
	}
	events := arupa.NewSocketListener()
	_ = events.On(eventConnectSSH, s.connectSSH)
	_ = events.On(eventTerminalInput, s.writeInput)
	_ = events.On(eventResize, s.resize)
	_ = events.On(eventDisconnect, func(_ context.Context, event arupa.SocketEvent, _ arupa.Emitter) error {
		s.cleanup(event.SocketID)
		return nil
	})

	authenticated := arupa.AccessPolicy{RequireAuth: true}
	s.sdk = &arupagrpc.Plugin{
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
		Handler:    http.HandlerFunc(s.handleConnectionsHTTP),
		Events:     events,
		OnRegister: s.configure,
	}
	s.settingsStore = sshSettingsStore{params: s.sdk}
	s.events = events
	return s
}

// configure initializes SSH application state after the SDK has connected the
// host callback and captured the initial Params snapshot.
func (s *sshServer) configure(ctx context.Context) error {
	if s.sdk == nil {
		return fmt.Errorf("ssh: SDK plugin is unavailable during registration")
	}
	if err := s.configureParams(s.sdk.InitialParams()); err != nil {
		return err
	}
	// Registration must not fail merely because the host did not expose the
	// optional logging callback.
	_ = s.sdk.Info(ctx, "ssh plugin registered")
	return nil
}

// configureParams applies the SSH-specific subset of the initial plugin
// parameters. Keeping it separate from the SDK hook makes the application
// configuration independently testable.
func (s *sshServer) configureParams(params map[string]string) error {
	settings, err := s.settingsStore.Load(params)
	if err != nil {
		return err
	}
	s.sshConfigPath = settings.SSHConfigPath
	s.settingsMu.Lock()
	s.settings = settings.Connections
	s.settingsMu.Unlock()
	return nil
}
