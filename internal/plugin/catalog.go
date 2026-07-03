package plugin

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goplugin "github.com/SteelDrEgg/go-plugin"
	"gopkg.in/yaml.v3"
)

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

type pluginCatalog struct {
	kv  *KV
	log *slog.Logger
}

func newPluginCatalog(kv *KV, log *slog.Logger) *pluginCatalog {
	if log == nil {
		log = slog.Default()
	}
	return &pluginCatalog{kv: kv, log: log}
}

func (c *pluginCatalog) discover(dir string) ([]DiscoveredPlugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.log.Warn("plugin directory does not exist; skipping", "dir", dir)
			return nil, nil
		}
		return nil, fmt.Errorf("read plugin dir: %w", err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".plg" {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)

	seen := make(map[string]struct{}, len(paths))
	out := make([]DiscoveredPlugin, 0, len(paths))
	for _, p := range paths {
		info, err := readPluginInfo(p)
		if err != nil {
			c.log.Error("failed to scan plugin package", "path", p, "err", err)
			continue
		}
		if _, exists := seen[info.Name]; exists {
			c.log.Error("duplicate plugin name found in packages; keeping first", "name", info.Name, "path", p)
			continue
		}
		seen[info.Name] = struct{}{}
		out = append(out, info)
	}
	return out, nil
}

func (c *pluginCatalog) publish(d DiscoveredPlugin) {
	b, err := json.Marshal(d)
	if err != nil {
		return
	}
	c.kv.SystemSet(SysNamespace, registryKVPrefix+"catalog/"+d.Name, b)
}

func (c *pluginCatalog) unpublish(name string) {
	c.kv.SystemDelete(SysNamespace, registryKVPrefix+"catalog/"+name)
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
