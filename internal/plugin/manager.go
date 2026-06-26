package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"

	grpcpb "minimalpanel/pluginsdk/grpc/proto"
	wasmpb "minimalpanel/pluginsdk/wasm/proto"

	"minimalpanel/internal/netx"

	goplugin "github.com/SteelDrEgg/go-plugin"
	"google.golang.org/grpc"
)

// handshake is shared with gRPC plugins. Plugins must use the same values.
var handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MINIMALPANEL_PLUGIN",
	MagicCookieValue: "minimalpanel",
}

// Options configures a Manager.
type Options struct {
	// TempDir is where plugin packages are extracted. Required.
	TempDir string
	// Mux receives plugin HTTP routes. Required.
	Mux *http.ServeMux
	// Socket is the global Socket.IO server plugins attach namespaces to. Required.
	Socket *netx.Socket
	// Logger is used for host and plugin logs. Optional.
	Logger *slog.Logger
}

// Manager loads plugins and exposes the shared host API to them.
type Manager struct {
	inner    *goplugin.Manager
	kv       *KV
	api      *HostAPI
	registry *Registry
	socket   *socketBridge
	mux      *http.ServeMux
	log      *slog.Logger

	hostGRPC     *grpcHostServer
	hostGRPCAddr string

	mu      sync.Mutex
	plugins map[string]*loadedPlugin
}

type loadedPlugin struct {
	handle    *goplugin.Handle
	conn      pluginConn
	record    *PluginRecord
	grpcToken string
}

// NewManager builds a plugin manager and starts the gRPC host callback server.
func NewManager(opts Options) (*Manager, error) {
	if opts.TempDir == "" {
		return nil, fmt.Errorf("TempDir is required")
	}
	if opts.Mux == nil {
		return nil, fmt.Errorf("Mux is required")
	}
	if opts.Socket == nil {
		return nil, fmt.Errorf("Socket is required")
	}
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}

	if err := os.MkdirAll(opts.TempDir, 0o755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	m := &Manager{
		kv:      NewKV(),
		mux:     opts.Mux,
		log:     log,
		plugins: make(map[string]*loadedPlugin),
	}
	m.registry = NewRegistry(m.kv)
	m.socket = newSocketBridge(opts.Socket, log)
	m.api = NewHostAPI(m.kv, m.socket, log)

	m.hostGRPC = newGRPCHostServer(m.api)
	addr, err := m.hostGRPC.Start()
	if err != nil {
		return nil, err
	}
	m.hostGRPCAddr = addr

	inner, err := goplugin.NewManager(goplugin.Config{
		TempDir: opts.TempDir,
		GRPC: &goplugin.GRPCConfig{
			HandshakeConfig:  handshake,
			AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
			SyncStderr:       os.Stderr,
			LoaderWithBroker: func(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
				return grpcpb.NewPluginClient(c), nil
			},
		},
		WASM: &goplugin.WASMConfig{
			Loader: m.wasmLoader,
		},
	})
	if err != nil {
		m.hostGRPC.Stop()
		return nil, err
	}
	m.inner = inner
	return m, nil
}

// KV exposes the shared key-value store (e.g. for host-side seeding).
func (m *Manager) KV() *KV { return m.kv }

// Registry exposes the plugin registry.
func (m *Manager) Registry() *Registry { return m.registry }

func (m *Manager) wasmLoader(ctx context.Context, modulePath string, _ goplugin.Info) (any, func(context.Context) error, error) {
	loader, err := wasmpb.NewPluginPlugin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("new wasm loader: %w", err)
	}
	client, err := loader.Load(ctx, modulePath, wasmHostFns{api: m.api})
	if err != nil {
		return nil, nil, fmt.Errorf("load wasm module: %w", err)
	}
	return client, func(ctx context.Context) error { return client.Close(ctx) }, nil
}

// LoadDir loads every *.plg package found in dir (non-recursively).
func (m *Manager) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			m.log.Warn("plugin directory does not exist; skipping", "dir", dir)
			return nil
		}
		return fmt.Errorf("read plugin dir: %w", err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".plg" {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)

	for _, p := range paths {
		if _, err := m.Load(p); err != nil {
			m.log.Error("failed to load plugin", "path", p, "err", err)
		}
	}
	return nil
}

// Load extracts, loads, registers and wires a single plugin package.
func (m *Manager) Load(path string) (*loadedPlugin, error) {
	handle, err := m.inner.Load(path)
	if err != nil {
		return nil, err
	}

	info := handle.Info()
	conn, err := m.connFor(info.Type, handle.Client())
	if err != nil {
		_ = m.inner.Unload(handle)
		return nil, err
	}

	instanceID := info.Name
	if m.registry.Has(instanceID) {
		_ = m.inner.Unload(handle)
		return nil, fmt.Errorf("plugin instance %q already loaded", instanceID)
	}

	req := RegisterRequest{InstanceID: instanceID}
	var grpcToken string
	if info.Type == "grpc" {
		token, err := m.hostGRPC.issueToken(instanceID)
		if err != nil {
			_ = m.inner.Unload(handle)
			return nil, fmt.Errorf("issue host callback token: %w", err)
		}
		grpcToken = token
		req.HostCallbackAddr = m.hostGRPCAddr
		req.HostCallbackToken = token
	}

	reg, err := conn.Register(context.Background(), req)
	if err != nil {
		if grpcToken != "" {
			m.hostGRPC.revokeToken(grpcToken)
		}
		_ = m.inner.Unload(handle)
		return nil, fmt.Errorf("register plugin %q: %w", instanceID, err)
	}

	record := &PluginRecord{
		InstanceID: instanceID,
		Name:       reg.Name,
		Version:    reg.Version,
		Type:       info.Type,
		Path:       path,
		Routes:     reg.Routes,
		Namespaces: reg.Namespaces,
	}

	for _, route := range reg.Routes {
		if err := m.registerRoute(route, conn); err != nil {
			m.log.Error("failed to register plugin route", "plugin", instanceID, "pattern", route.Pattern, "err", err)
		}
	}
	for _, ns := range reg.Namespaces {
		if err := m.socket.register(ns, conn); err != nil {
			m.log.Error("failed to register plugin socket namespace", "plugin", instanceID, "namespace", ns.Name, "err", err)
		}
	}

	lp := &loadedPlugin{handle: handle, conn: conn, record: record, grpcToken: grpcToken}
	m.mu.Lock()
	m.plugins[instanceID] = lp
	m.mu.Unlock()
	m.registry.Add(record)

	m.log.Info("loaded plugin", "name", reg.Name, "version", reg.Version, "type", info.Type,
		"routes", len(reg.Routes), "namespaces", len(reg.Namespaces))
	return lp, nil
}

func (m *Manager) connFor(pluginType string, client any) (pluginConn, error) {
	switch pluginType {
	case "wasm":
		pc, ok := client.(wasmpb.Plugin)
		if !ok {
			return nil, fmt.Errorf("unexpected wasm plugin client type %T", client)
		}
		return wasmConn{client: pc}, nil
	case "grpc":
		pc, ok := client.(grpcpb.PluginClient)
		if !ok {
			return nil, fmt.Errorf("unexpected grpc plugin client type %T", client)
		}
		return grpcConn{client: pc}, nil
	default:
		return nil, fmt.Errorf("unsupported plugin type %q", pluginType)
	}
}

// Close unloads all plugins and stops the host callback server.
func (m *Manager) Close() error {
	m.mu.Lock()
	plugins := make([]*loadedPlugin, 0, len(m.plugins))
	for _, lp := range m.plugins {
		plugins = append(plugins, lp)
	}
	m.plugins = make(map[string]*loadedPlugin)
	m.mu.Unlock()

	for _, lp := range plugins {
		if lp.grpcToken != "" {
			m.hostGRPC.revokeToken(lp.grpcToken)
		}
		m.registry.Remove(lp.record.InstanceID)
		if err := m.inner.Unload(lp.handle); err != nil {
			m.log.Error("failed to unload plugin", "plugin", lp.record.InstanceID, "err", err)
		}
	}

	m.hostGRPC.Stop()
	return nil
}
