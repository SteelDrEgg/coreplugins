package plugin

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	grpcpb "minimalpanel/pluginsdk/grpc/proto"
	wasmpb "minimalpanel/pluginsdk/wasm/proto"

	"minimalpanel/internal/netx"

	goplugin "github.com/SteelDrEgg/go-plugin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
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
	// ParamsResolver returns config params passed to a plugin at registration,
	// keyed by plugin name. Optional.
	ParamsResolver func(name string) map[string]string
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

	paramsResolver func(name string) map[string]string

	mu      sync.Mutex
	plugins map[string]*loadedPlugin

	scanMu     sync.RWMutex
	discovered map[string]DiscoveredPlugin // plugin name -> descriptor
}

type loadedPlugin struct {
	handle    *goplugin.Handle
	conn      pluginConn
	record    *PluginRecord
	grpcToken string
}

// DiscoveredPlugin is metadata scanned from a .plg package's info.yaml without
// loading the plugin runtime.
type DiscoveredPlugin struct {
	Name            string
	Version         string
	Type            string
	ContractVersion int
	Command         string
	Metadata        map[string]any
	PackagePath     string
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
		kv:             NewKV(),
		mux:            opts.Mux,
		log:            log,
		plugins:        make(map[string]*loadedPlugin),
		discovered:     make(map[string]DiscoveredPlugin),
		paramsResolver: opts.ParamsResolver,
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

// ScanDir scans *.plg packages in dir (non-recursively), reads info.yaml and
// caches metadata without starting plugins.
func (m *Manager) ScanDir(dir string) error {
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

	next := make(map[string]DiscoveredPlugin, len(paths))
	for _, p := range paths {
		info, err := readPluginInfo(p)
		if err != nil {
			m.log.Error("failed to scan plugin package", "path", p, "err", err)
			continue
		}
		if _, exists := next[info.Name]; exists {
			m.log.Error("duplicate plugin name found in packages; keeping first", "name", info.Name, "path", p)
			continue
		}
		next[info.Name] = info
	}

	m.scanMu.Lock()
	prev := m.discovered
	m.discovered = next
	m.scanMu.Unlock()

	for name := range prev {
		if _, ok := next[name]; !ok {
			m.kv.SystemDelete(SysNamespace, registryKVPrefix+"catalog/"+name)
		}
	}
	for _, info := range next {
		m.publishDiscovered(info)
	}
	return nil
}

// Discovered returns scanned plugin metadata snapshot.
func (m *Manager) Discovered() []DiscoveredPlugin {
	m.scanMu.RLock()
	defer m.scanMu.RUnlock()
	out := make([]DiscoveredPlugin, 0, len(m.discovered))
	for _, d := range m.discovered {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// StartByName starts a previously scanned plugin by name.
func (m *Manager) StartByName(name string) (*loadedPlugin, error) {
	m.scanMu.RLock()
	d, ok := m.discovered[name]
	m.scanMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin %q not found in scan results", name)
	}
	return m.Load(d.PackagePath)
}

// Start starts a previously scanned plugin by name.
func (m *Manager) Start(name string) error {
	_, err := m.StartByName(name)
	return err
}

// StartMatching starts scanned plugins where shouldStart returns true.
func (m *Manager) StartMatching(shouldStart func(DiscoveredPlugin) bool) error {
	discovered := m.Discovered()
	for _, d := range discovered {
		if shouldStart != nil && !shouldStart(d) {
			m.log.Info("plugin auto-start disabled by config", "name", d.Name)
			continue
		}
		if _, err := m.StartByName(d.Name); err != nil {
			m.log.Error("failed to start plugin", "name", d.Name, "path", d.PackagePath, "err", err)
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
	if m.paramsResolver != nil {
		req.Params = m.paramsResolver(instanceID)
	}
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
		Static:     reg.Static,
		Namespaces: reg.Namespaces,
	}

	for _, route := range reg.Routes {
		if err := m.registerRoute(route, conn); err != nil {
			m.log.Error("failed to register plugin route", "plugin", instanceID, "pattern", route.Pattern, "err", err)
		}
	}
	for _, mount := range reg.Static {
		if err := m.registerStatic(handle.RootPath(), mount); err != nil {
			m.log.Error("failed to register plugin static mount", "plugin", instanceID, "prefix", mount.Prefix, "dir", mount.Directory, "err", err)
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
		"routes", len(reg.Routes), "static_mounts", len(reg.Static), "namespaces", len(reg.Namespaces))
	return lp, nil
}

func readPluginInfo(path string) (DiscoveredPlugin, error) {
	f, err := os.Open(path)
	if err != nil {
		return DiscoveredPlugin{}, fmt.Errorf("open plugin package: %w", err)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return DiscoveredPlugin{}, fmt.Errorf("stat plugin package: %w", err)
	}

	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		return DiscoveredPlugin{}, fmt.Errorf("read zip plugin package: %w", err)
	}

	var info goplugin.Info
	for _, zf := range zr.File {
		if filepath.Clean(zf.Name) != "info.yaml" {
			continue
		}
		r, err := zf.Open()
		if err != nil {
			return DiscoveredPlugin{}, fmt.Errorf("open info.yaml: %w", err)
		}
		b, err := io.ReadAll(r)
		_ = r.Close()
		if err != nil {
			return DiscoveredPlugin{}, fmt.Errorf("read info.yaml: %w", err)
		}
		if err := yaml.Unmarshal(b, &info); err != nil {
			return DiscoveredPlugin{}, fmt.Errorf("parse info.yaml: %w", err)
		}
		break
	}

	if strings.TrimSpace(info.Name) == "" {
		return DiscoveredPlugin{}, fmt.Errorf("info.yaml Name is required")
	}
	if strings.TrimSpace(info.Version) == "" {
		return DiscoveredPlugin{}, fmt.Errorf("info.yaml Version is required")
	}
	if info.Type != "grpc" && info.Type != "wasm" {
		return DiscoveredPlugin{}, fmt.Errorf("info.yaml Type must be grpc or wasm")
	}
	if info.ContractVersion == 0 {
		return DiscoveredPlugin{}, fmt.Errorf("info.yaml ContractVersion is required")
	}
	if strings.TrimSpace(info.Command) == "" {
		return DiscoveredPlugin{}, fmt.Errorf("info.yaml Command is required")
	}

	return DiscoveredPlugin{
		Name:            info.Name,
		Version:         info.Version,
		Type:            info.Type,
		ContractVersion: info.ContractVersion,
		Command:         info.Command,
		Metadata:        info.Metadata,
		PackagePath:     path,
	}, nil
}

func (m *Manager) publishDiscovered(d DiscoveredPlugin) {
	// Keep scanned info available through read-only sys KV.
	b, err := json.Marshal(d)
	if err != nil {
		return
	}
	m.kv.SystemSet(SysNamespace, registryKVPrefix+"catalog/"+d.Name, b)
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
