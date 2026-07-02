package plugin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"minimalpanel/internal/conf"
	grpcpb "minimalpanel/pluginsdk/grpc/proto"
	wasmpb "minimalpanel/pluginsdk/wasm/proto"

	goplugin "github.com/SteelDrEgg/go-plugin"
	"google.golang.org/grpc"
)

// handshake is shared with gRPC plugins. Plugins must use the same values.
var handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MINIMALPANEL_PLUGIN",
	MagicCookieValue: "minimalpanel",
}

type pluginLoaderOptions struct {
	TempDir      string
	API          *HostAPI
	HostGRPC     *grpcHostServer
	HostGRPCAddr string
}

type pluginLoader struct {
	inner        *goplugin.Manager
	tempDir      string
	api          *HostAPI
	hostGRPC     *grpcHostServer
	hostGRPCAddr string
}

type pluginLoadResult struct {
	loaded       *loadedPlugin
	registration *RegisterResult
	rootPath     string
	runAsUser    string
}

type loadedPlugin struct {
	loader    *goplugin.Manager
	handle    *goplugin.Handle
	conn      pluginConn
	record    *PluginRecord
	grpcToken string
}

func newPluginLoader(opts pluginLoaderOptions) (*pluginLoader, error) {
	l := &pluginLoader{
		tempDir:      opts.TempDir,
		api:          opts.API,
		hostGRPC:     opts.HostGRPC,
		hostGRPCAddr: opts.HostGRPCAddr,
	}
	inner, err := l.newInner("")
	if err != nil {
		return nil, err
	}
	l.inner = inner
	return l, nil
}

func (l *pluginLoader) load(scanned DiscoveredPlugin, cfg conf.Plugin) (*pluginLoadResult, error) {
	loader := l.inner
	runAsUser := ""
	if scanned.Type == "grpc" {
		runAsUser = strings.TrimSpace(cfg.RunAsUser)
	}
	if runAsUser != "" {
		var err error
		loader, err = l.newInner(runAsUser)
		if err != nil {
			return nil, err
		}
	}

	handle, err := loader.Load(scanned.PackagePath)
	if err != nil {
		return nil, err
	}

	info := handle.Info()
	conn, err := l.connFor(info.Type, handle.Client())
	if err != nil {
		_ = loader.Unload(handle)
		return nil, err
	}

	instanceID := info.Name
	if instanceID != scanned.Name {
		_ = loader.Unload(handle)
		return nil, fmt.Errorf("plugin package name changed from %q to %q while loading", scanned.Name, instanceID)
	}

	req := RegisterRequest{InstanceID: instanceID, Params: cfg.Params}
	var grpcToken string
	if info.Type == "grpc" {
		if l.hostGRPC == nil {
			_ = loader.Unload(handle)
			return nil, fmt.Errorf("host callback server is not configured")
		}
		token, err := l.hostGRPC.issueToken(instanceID)
		if err != nil {
			_ = loader.Unload(handle)
			return nil, fmt.Errorf("issue host callback token: %w", err)
		}
		grpcToken = token
		req.HostCallbackAddr = l.hostGRPCAddr
		req.HostCallbackToken = token
	}

	reg, err := conn.Register(context.Background(), req)
	if err != nil {
		if grpcToken != "" {
			l.hostGRPC.revokeToken(grpcToken)
		}
		_ = loader.Unload(handle)
		return nil, fmt.Errorf("register plugin %q: %w", instanceID, err)
	}

	record := &PluginRecord{
		InstanceID: instanceID,
		Name:       reg.Name,
		Version:    reg.Version,
		Type:       info.Type,
		Path:       scanned.PackagePath,
		Routes:     reg.Routes,
		Static:     reg.Static,
		Namespaces: reg.Namespaces,
	}

	lp := &loadedPlugin{
		loader:    loader,
		handle:    handle,
		conn:      conn,
		record:    record,
		grpcToken: grpcToken,
	}
	return &pluginLoadResult{
		loaded:       lp,
		registration: reg,
		rootPath:     handle.RootPath(),
		runAsUser:    runAsUser,
	}, nil
}

func (l *pluginLoader) newInner(runAsUser string) (*goplugin.Manager, error) {
	return goplugin.NewManager(goplugin.Config{
		TempDir: l.tempDir,
		GRPC: &goplugin.GRPCConfig{
			HandshakeConfig:  handshake,
			RunAsUser:        strings.TrimSpace(runAsUser),
			AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
			SyncStderr:       os.Stderr,
			LoaderWithBroker: func(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
				return grpcpb.NewPluginClient(c), nil
			},
		},
		WASM: &goplugin.WASMConfig{
			Loader: l.wasmLoader,
		},
	})
}

func (l *pluginLoader) wasmLoader(ctx context.Context, modulePath string, info goplugin.Info) (any, func(context.Context) error, error) {
	loader, err := wasmpb.NewPluginPlugin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("new wasm loader: %w", err)
	}
	client, err := loader.Load(ctx, modulePath, wasmHostFns{api: l.api, source: info.Name})
	if err != nil {
		return nil, nil, fmt.Errorf("load wasm module: %w", err)
	}
	return client, func(ctx context.Context) error { return client.Close(ctx) }, nil
}

func (l *pluginLoader) connFor(pluginType string, client any) (pluginConn, error) {
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

func (l *pluginLoader) revoke(lp *loadedPlugin) {
	if lp != nil && lp.grpcToken != "" && l.hostGRPC != nil {
		l.hostGRPC.revokeToken(lp.grpcToken)
	}
}

func (l *pluginLoader) unload(lp *loadedPlugin) error {
	if lp == nil || lp.loader == nil || lp.handle == nil {
		return nil
	}
	return lp.loader.Unload(lp.handle)
}
