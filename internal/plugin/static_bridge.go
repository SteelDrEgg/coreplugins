package plugin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"minimalpanel/internal/auth"
	"minimalpanel/internal/netx"
)

// registerStatic wires a plugin-declared static directory or file into the
// plugin dispatch table. Directory mounts are normalized to subtree patterns
// ending in "/", while file mounts remain exact patterns.
func (m *Manager) registerStatic(owner, pluginRoot string, mount StaticMount) error {
	prefix := strings.TrimSpace(mount.Prefix)
	if prefix == "" {
		return fmt.Errorf("static mount prefix is required")
	}
	if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf("static mount prefix %q must start with '/'", prefix)
	}

	dir := strings.TrimSpace(mount.Directory)
	if dir == "" {
		return fmt.Errorf("static mount path is required")
	}
	if pluginRoot != "" {
		dir = strings.ReplaceAll(dir, "$PLUGIN_ROOT", pluginRoot)
	}
	dir = filepath.Clean(dir)

	st, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("stat static mount path %q: %w", dir, err)
	}
	pattern := prefix
	var handler http.Handler
	if st.IsDir() {
		if !strings.HasSuffix(pattern, "/") {
			pattern += "/"
		}

		stripPrefix := strings.TrimSuffix(pattern, "/")
		if stripPrefix == "" {
			stripPrefix = "/"
		}
		handler = http.StripPrefix(stripPrefix, http.FileServer(http.Dir(dir)))
	} else {
		if strings.HasSuffix(pattern, "/") {
			return fmt.Errorf("static file mount prefix %q must not end with '/'", pattern)
		}
		filePath := dir
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filePath)
		})
	}

	m.routeMu.RLock()
	routeOwner, routeExists := m.routePatternOwnedByOtherLocked(pattern, owner)
	m.routeMu.RUnlock()
	if routeExists {
		return fmt.Errorf("static mount prefix %q already owned by http route from plugin %q", pattern, routeOwner)
	}

	m.staticMu.Lock()
	defer m.staticMu.Unlock()
	if binding, ok := m.static[pattern]; ok {
		if binding.owner != owner {
			return fmt.Errorf("static mount prefix %q already owned by plugin %q", pattern, binding.owner)
		}
		binding.owner = owner
		binding.mount = mount
		binding.handler = handler
		return nil
	}

	m.static[pattern] = &staticMountBinding{owner: owner, mount: mount, handler: handler}
	return nil
}

func (m *Manager) routePatternOwnedByOtherLocked(pattern, owner string) (string, bool) {
	for key, binding := range m.routes {
		if key.pattern == pattern && binding != nil && binding.owner != owner {
			return binding.owner, true
		}
	}
	return "", false
}

func (m *Manager) unregisterStatic(owner string) {
	m.staticMu.Lock()
	defer m.staticMu.Unlock()
	for pattern, binding := range m.static {
		if binding.owner == owner {
			delete(m.static, pattern)
		}
	}
}

// matchPluginStatic returns the longest static mount pattern that matches path.
// The returned handler already knows whether it should serve a file or strip a
// directory prefix.
func (m *Manager) matchPluginStatic(path string) (StaticMount, http.Handler, int) {
	m.staticMu.RLock()
	defer m.staticMu.RUnlock()

	var best *staticMountBinding
	bestPattern := ""
	for pattern, binding := range m.static {
		if binding == nil || binding.handler == nil {
			continue
		}
		if !pathMatchesPattern(path, pattern) {
			continue
		}
		if len(pattern) > len(bestPattern) {
			best = binding
			bestPattern = pattern
		}
	}
	if best == nil {
		return StaticMount{}, nil, -1
	}
	return best.mount, best.handler, len(bestPattern)
}

// handlePluginStatic applies mount-level auth and delegates file serving to the
// prepared static handler.
func (m *Manager) handlePluginStatic(mount StaticMount, handler http.Handler, w http.ResponseWriter, r *http.Request) {
	if mount.Protected {
		if _, ok := auth.IsAuthenticated(r); !ok {
			_ = netx.WriteUnauthorized(w, "authentication required")
			return
		}
	}
	handler.ServeHTTP(w, r)
}
