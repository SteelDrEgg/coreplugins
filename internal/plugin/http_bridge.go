package plugin

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"minimalpanel/internal/auth"
	"minimalpanel/internal/netx"
)

// maxRequestBody caps the request body forwarded to plugins (8 MiB).
const maxRequestBody = 8 << 20

// ServeHTTP dispatches requests that did not match host routes to the current
// plugin HTTP/static route table. It chooses the longest matching plugin
// pattern; static mounts win ties so an exact static file mount can serve its
// asset without being shadowed by an HTTP route of the same length.
func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, routePattern, routeConn, routeLen := m.matchPluginRoute(r.URL.Path)
	mount, staticHandler, staticLen := m.matchPluginStatic(r.URL.Path)

	if staticHandler != nil && staticLen >= routeLen {
		m.handlePluginStatic(mount, staticHandler, w, r)
		return
	}
	if routeConn != nil {
		m.handlePluginRoute(routePattern, route, routeConn, w, r)
		return
	}
	http.NotFound(w, r)
}

// registerRoute wires a plugin HTTP route into the plugin dispatch table. Route
// ownership is exclusive while the owning plugin is loaded; Stop removes the
// entry so another plugin can claim the same pattern later.
func (m *Manager) registerRoute(owner string, route HTTPRoute, conn pluginConn) error {
	if route.Pattern == "" {
		return fmt.Errorf("http route pattern is required")
	}

	pattern := route.Pattern
	m.routeMu.Lock()
	defer m.routeMu.Unlock()
	if binding, ok := m.routes[pattern]; ok && binding.owner != owner {
		return fmt.Errorf("http route pattern %q already owned by plugin %q", pattern, binding.owner)
	}

	m.staticMu.RLock()
	staticBinding, staticExists := m.static[pattern]
	m.staticMu.RUnlock()
	if staticExists && staticBinding.owner != owner {
		return fmt.Errorf("http route pattern %q already owned by static mount from plugin %q", pattern, staticBinding.owner)
	}

	if binding, ok := m.routes[pattern]; ok {
		binding.owner = owner
		binding.route = route
		binding.conn = conn
		return nil
	}

	m.routes[pattern] = &httpRouteBinding{owner: owner, route: route, conn: conn}
	return nil
}

func (m *Manager) unregisterRoutes(owner string) {
	m.routeMu.Lock()
	defer m.routeMu.Unlock()
	for pattern, binding := range m.routes {
		if binding.owner == owner {
			delete(m.routes, pattern)
		}
	}
}

// matchPluginRoute returns the longest HTTP route pattern that matches path.
// Patterns ending in "/" are prefix matches; all other patterns are exact
// matches.
func (m *Manager) matchPluginRoute(path string) (HTTPRoute, string, pluginConn, int) {
	m.routeMu.RLock()
	defer m.routeMu.RUnlock()

	var best *httpRouteBinding
	bestPattern := ""
	for pattern, binding := range m.routes {
		if binding == nil || binding.conn == nil {
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
		return HTTPRoute{}, "", nil, -1
	}
	return best.route, bestPattern, best.conn, len(bestPattern)
}

// handlePluginRoute serializes an HTTP request into the plugin contract, calls
// the owning plugin, and writes the plugin's response back to the client.
func (m *Manager) handlePluginRoute(pattern string, route HTTPRoute, conn pluginConn, w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	if method != "" && !strings.EqualFold(r.Method, method) {
		_ = netx.WriteMethodNotAllowed(w)
		return
	}
	if route.Protected {
		if _, ok := auth.IsAuthenticated(r); !ok {
			_ = netx.WriteUnauthorized(w, "authentication required")
			return
		}
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
	if err != nil {
		_ = netx.WriteBadRequest(w, "failed to read request body")
		return
	}

	req := &HTTPRequest{
		RoutePattern: pattern,
		Method:       r.Method,
		Path:         r.URL.Path,
		Query:        r.URL.RawQuery,
		Headers:      flattenHeaders(r.Header),
		Body:         body,
		RemoteAddr:   r.RemoteAddr,
	}

	resp, err := conn.HandleHTTP(r.Context(), req)
	if err != nil {
		_ = netx.WriteError(w, http.StatusBadGateway, "plugin handler failed", err)
		return
	}

	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	status := resp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(resp.Body)
}

// pathMatchesPattern implements the plugin dispatcher's small matching model:
// trailing-slash patterns are subtree prefixes, and every other pattern is an
// exact match.
func pathMatchesPattern(path, pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}
	return path == pattern
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}
