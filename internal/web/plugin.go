package web

import (
	"encoding/json"
	"fmt"
	"minimalpanel/internal/auth"
	"minimalpanel/internal/conf"
	"minimalpanel/internal/netx"
	"minimalpanel/internal/plugin"
	"net/http"
	"os"
	"sort"
	"strings"
)

type pluginActionRequest struct {
	Name string `json:"name"`
}

type pluginConfigRequest struct {
	PluginDir     string `json:"plugin_dir"`
	PluginTempDir string `json:"plugin_temp_dir"`
}

type pluginView struct {
	Name            string         `json:"name"`
	Version         string         `json:"version"`
	Type            string         `json:"type"`
	ContractVersion int            `json:"contract_version"`
	Command         string         `json:"command"`
	PackagePath     string         `json:"package_path"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	Loaded          bool           `json:"loaded"`
}

// StartPlugin registers plugin management endpoints used by frontend pages.
// All plugin endpoints are protected.
func StartPlugin(mux *http.ServeMux, pm *plugin.Manager) {
	mux.HandleFunc("/api/plugins", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginList(w, r, pm)
	}))
	mux.HandleFunc("/api/plugins/scan", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginScan(w, r, pm)
	}))
	mux.HandleFunc("/api/plugins/start", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginStart(w, r, pm)
	}))
	mux.HandleFunc("/api/plugins/stop", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginStop(w, r, pm)
	}))
	mux.HandleFunc("/api/plugins/restart", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginRestart(w, r, pm)
	}))
	mux.HandleFunc("/api/plugins/config", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlePluginConfig(w, r, pm)
	}))
}

func handlePluginList(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	if r.Method != http.MethodGet {
		netx.WriteMethodNotAllowed(w)
		return
	}

	running := pm.Registry().List()
	loadedByName := make(map[string]bool, len(running))
	for _, rec := range running {
		loadedByName[rec.InstanceID] = true
	}

	discovered := pm.Discovered()
	plugins := make([]pluginView, 0, len(discovered))
	for _, d := range discovered {
		plugins = append(plugins, pluginView{
			Name:            d.Name,
			Version:         d.Version,
			Type:            d.Type,
			ContractVersion: d.ContractVersion,
			Command:         d.Command,
			PackagePath:     d.PackagePath,
			Metadata:        d.Metadata,
			Loaded:          loadedByName[d.Name],
		})
	}

	sort.Slice(running, func(i, j int) bool {
		return running[i].InstanceID < running[j].InstanceID
	})

	cfg := conf.GetPlugin()
	netx.WriteSuccess(w, "Plugin state fetched", map[string]any{
		"plugin_dir":      cfg.PluginDir,
		"plugin_temp_dir": cfg.PluginTempDir,
		"discovered":      plugins,
		"running":         running,
	})
}

func handlePluginScan(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	if r.Method != http.MethodPost {
		netx.WriteMethodNotAllowed(w)
		return
	}

	pluginCfg := conf.GetPlugin()
	if err := pm.ScanDir(pluginCfg.PluginDir); err != nil {
		netx.WriteInternalServerError(w, "Failed to scan plugin directory", err)
		return
	}

	netx.WriteSuccess(w, "Plugin directory scanned", map[string]any{
		"plugin_dir": pluginCfg.PluginDir,
		"count":      len(pm.Discovered()),
	})
}

func handlePluginStart(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	if r.Method != http.MethodPost {
		netx.WriteMethodNotAllowed(w)
		return
	}

	name, ok := readPluginActionName(w, r)
	if !ok {
		return
	}

	if err := pm.Start(name); err != nil {
		netx.WriteBadRequest(w, fmt.Sprintf("Failed to start plugin %q: %v", name, err))
		return
	}

	netx.WriteSuccess(w, "Plugin started", map[string]any{
		"name": name,
	})
}

func handlePluginStop(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	if r.Method != http.MethodPost {
		netx.WriteMethodNotAllowed(w)
		return
	}

	name, ok := readPluginActionName(w, r)
	if !ok {
		return
	}

	if err := pm.Stop(name); err != nil {
		netx.WriteBadRequest(w, fmt.Sprintf("Failed to stop plugin %q: %v", name, err))
		return
	}

	netx.WriteSuccess(w, "Plugin stopped", map[string]any{
		"name": name,
	})
}

func handlePluginRestart(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	if r.Method != http.MethodPost {
		netx.WriteMethodNotAllowed(w)
		return
	}

	name, ok := readPluginActionName(w, r)
	if !ok {
		return
	}

	if err := pm.Restart(name); err != nil {
		netx.WriteBadRequest(w, fmt.Sprintf("Failed to restart plugin %q: %v", name, err))
		return
	}

	netx.WriteSuccess(w, "Plugin restarted", map[string]any{
		"name": name,
	})
}

func readPluginActionName(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req pluginActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		netx.WriteBadRequest(w, "Invalid request body")
		return "", false
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		netx.WriteBadRequest(w, "Plugin name is required")
		return "", false
	}
	return name, true
}

func handlePluginConfig(w http.ResponseWriter, r *http.Request, pm *plugin.Manager) {
	switch r.Method {
	case http.MethodGet:
		cfg := conf.GetPlugin()
		netx.WriteSuccess(w, "Plugin config fetched", map[string]any{
			"plugin_dir":      cfg.PluginDir,
			"plugin_temp_dir": cfg.PluginTempDir,
		})
	case http.MethodPut:
		var req pluginConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			netx.WriteBadRequest(w, "Invalid request body")
			return
		}

		req.PluginDir = strings.TrimSpace(req.PluginDir)
		req.PluginTempDir = strings.TrimSpace(req.PluginTempDir)
		if req.PluginDir == "" || req.PluginTempDir == "" {
			netx.WriteBadRequest(w, "plugin_dir and plugin_temp_dir are required")
			return
		}

		// Validate/create directories first, then persist config.
		if err := os.MkdirAll(req.PluginDir, 0o755); err != nil {
			netx.WriteBadRequest(w, fmt.Sprintf("Invalid plugin_dir: %v", err))
			return
		}
		if err := os.MkdirAll(req.PluginTempDir, 0o755); err != nil {
			netx.WriteBadRequest(w, fmt.Sprintf("Invalid plugin_temp_dir: %v", err))
			return
		}

		oldCfg := conf.GetPlugin()
		newCfg, err := conf.SetPluginPaths(req.PluginDir, req.PluginTempDir)
		if err != nil {
			netx.WriteInternalServerError(w, "Failed to persist plugin config", err)
			return
		}

		if err := pm.ScanDir(newCfg.Plugin.PluginDir); err != nil {
			netx.WriteInternalServerError(w, "Plugin config saved, but scan failed", err)
			return
		}

		tempDirChanged := oldCfg.PluginTempDir != newCfg.Plugin.PluginTempDir
		netx.WriteSuccess(w, "Plugin config updated", map[string]any{
			"plugin_dir":                      newCfg.Plugin.PluginDir,
			"plugin_temp_dir":                 newCfg.Plugin.PluginTempDir,
			"temp_dir_requires_restart":       tempDirChanged,
			"temp_dir_restart_reason":         "plugin manager temp dir is initialized at startup",
			"discovered_plugin_count":         len(pm.Discovered()),
			"scan_path_effective_immediately": true,
		})
	default:
		netx.WriteMethodNotAllowed(w)
	}
}
