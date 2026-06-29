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

// registerStatic wires a plugin-declared static directory or file into the host mux.
func (m *Manager) registerStatic(pluginRoot string, mount StaticMount) error {
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

	if mount.Protected {
		next := handler
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := auth.IsAuthenticated(r); !ok {
				_ = netx.WriteUnauthorized(w, "authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	if err := netx.HandleSafe(m.mux, pattern, handler); err != nil {
		return fmt.Errorf("register static mount %q: %w", pattern, err)
	}
	return nil
}
