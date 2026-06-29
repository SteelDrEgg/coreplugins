package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"minimalpanel/internal/conf"
	"minimalpanel/internal/netx"
	"minimalpanel/internal/plugin"
	"minimalpanel/internal/web"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := conf.LoadConfig("config.toml"); err != nil {
		logger.Warn("failed to load config; using defaults", "err", err)
	}
	cfg := conf.Read()

	mux := http.NewServeMux()

	// Socket.IO server (plugins attach their own namespaces).
	socketServer := netx.GetGlobalServer()
	mux.Handle("/socket.io/", socketServer.Handler())
	// Keep auth endpoints so protected plugin routes/events can be used.
	web.StartLogin(mux)

	pm, err := plugin.NewManager(plugin.Options{
		TempDir:        cfg.PluginTempDir,
		Mux:            mux,
		Socket:         socketServer,
		Logger:         logger,
		ParamsResolver: cfg.PluginParams,
	})
	if err != nil {
		logger.Error("failed to create plugin manager", "err", err)
		os.Exit(1)
	}
	defer pm.Close()

	if err := pm.ScanDir(cfg.PluginDir); err != nil {
		logger.Error("failed to scan plugin directory", "dir", cfg.PluginDir, "err", err)
	}
	if err := pm.StartMatching(func(d plugin.DiscoveredPlugin) bool {
		return cfg.PluginAutoStart(d.Name)
	}); err != nil {
		logger.Error("failed to start configured plugins", "err", err)
	}
	for _, d := range pm.Discovered() {
		logger.Info("discovered plugin",
			"name", d.Name,
			"version", d.Version,
			"type", d.Type,
			"auto_start", cfg.PluginAutoStart(d.Name),
			"package", d.PackagePath,
		)
	}

	srv := &http.Server{Addr: cfg.Listen, Handler: mux}

	go func() {
		logger.Info("minimalpanel listening", "addr", cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
