package plugin

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"minimalpanel/internal/auth"
	"minimalpanel/internal/netx"
)

// maxRequestBody caps the request body forwarded to plugins (8 MiB).
const maxRequestBody = 8 << 20

type httpRouteKey struct {
	pattern string
	method  string
}

// ServeHTTP dispatches requests that did not match host routes to the current
// plugin HTTP/static route table. It chooses the longest matching plugin
// pattern; static mounts win ties so an exact static file mount can serve its
// asset without being shadowed by an HTTP route of the same length.
func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, routePattern, routeConn, routeLen, allowedMethods := m.matchPluginRoute(r.Method, r.URL.Path)
	mount, staticHandler, staticLen := m.matchPluginStatic(r.URL.Path)

	if staticHandler != nil && staticLen >= routeLen {
		m.handlePluginStatic(mount, staticHandler, w, r)
		return
	}
	if routeConn != nil {
		m.handlePluginRoute(routePattern, route, routeConn, w, r)
		return
	}
	if routeLen >= 0 {
		writeMethodNotAllowed(w, allowedMethods)
		return
	}
	http.NotFound(w, r)
}

// registerRoute wires a plugin HTTP route into the plugin dispatch table. Route
// ownership is exclusive for each pattern/method while the owning plugin is
// loaded; Stop removes the entry so another plugin can claim it later.
func (m *Manager) registerRoute(owner string, route HTTPRoute, conn pluginConn) error {
	if route.Pattern == "" {
		return fmt.Errorf("http route pattern is required")
	}

	pattern := route.Pattern
	method := normalizeRouteMethod(route.Method)
	route.Method = method
	key := httpRouteKey{pattern: pattern, method: method}

	m.routeMu.Lock()
	defer m.routeMu.Unlock()
	for existingKey, binding := range m.routes {
		if existingKey.pattern != pattern || binding == nil || binding.owner == owner {
			continue
		}
		if routeMethodsConflict(existingKey.method, method) {
			return fmt.Errorf("http route %s %q already owned by plugin %q", formatRouteMethod(existingKey.method), pattern, binding.owner)
		}
	}

	m.staticMu.RLock()
	staticBinding, staticExists := m.static[pattern]
	m.staticMu.RUnlock()
	if staticExists && staticBinding.owner != owner {
		return fmt.Errorf("http route pattern %q already owned by static mount from plugin %q", pattern, staticBinding.owner)
	}

	if binding, ok := m.routes[key]; ok {
		binding.owner = owner
		binding.route = route
		binding.conn = conn
		return nil
	}

	m.routes[key] = &httpRouteBinding{owner: owner, route: route, conn: conn}
	return nil
}

func (m *Manager) unregisterRoutes(owner string) {
	m.routeMu.Lock()
	defer m.routeMu.Unlock()
	for key, binding := range m.routes {
		if binding.owner == owner {
			delete(m.routes, key)
		}
	}
}

// matchPluginRoute returns the longest HTTP route pattern that matches path,
// then selects the handler by method within that pattern. Empty route methods
// are wildcard handlers. Patterns ending in "/" are prefix matches; all other
// patterns are exact matches.
func (m *Manager) matchPluginRoute(method, path string) (HTTPRoute, string, pluginConn, int, []string) {
	m.routeMu.RLock()
	defer m.routeMu.RUnlock()

	requestMethod := normalizeRouteMethod(method)
	bestPattern := ""
	for key, binding := range m.routes {
		if binding == nil || binding.conn == nil {
			continue
		}
		if !pathMatchesPattern(path, key.pattern) {
			continue
		}
		if len(key.pattern) > len(bestPattern) {
			bestPattern = key.pattern
		}
	}
	if bestPattern == "" {
		return HTTPRoute{}, "", nil, -1, nil
	}

	var exact *httpRouteBinding
	var wildcard *httpRouteBinding
	allowed := make(map[string]struct{})
	for key, binding := range m.routes {
		if key.pattern != bestPattern || binding == nil || binding.conn == nil {
			continue
		}
		if key.method == "" {
			wildcard = binding
			continue
		}
		allowed[key.method] = struct{}{}
		if key.method == requestMethod {
			exact = binding
		}
	}
	if exact != nil {
		return exact.route, bestPattern, exact.conn, len(bestPattern), nil
	}
	if wildcard != nil {
		return wildcard.route, bestPattern, wildcard.conn, len(bestPattern), nil
	}
	return HTTPRoute{}, bestPattern, nil, len(bestPattern), sortedMethods(allowed)
}

// handlePluginRoute serializes an HTTP request into the plugin contract, calls
// the owning plugin, and writes the plugin's response back to the client.
func (m *Manager) handlePluginRoute(pattern string, route HTTPRoute, conn pluginConn, w http.ResponseWriter, r *http.Request) {
	method := normalizeRouteMethod(route.Method)
	if method != "" && !strings.EqualFold(r.Method, method) {
		writeMethodNotAllowed(w, []string{method})
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

func normalizeRouteMethod(method string) string {
	return strings.ToUpper(strings.TrimSpace(method))
}

func routeMethodsConflict(a, b string) bool {
	return a == b || a == "" || b == ""
}

func formatRouteMethod(method string) string {
	if method == "" {
		return "ANY"
	}
	return method
}

func sortedMethods(methods map[string]struct{}) []string {
	out := make([]string, 0, len(methods))
	for method := range methods {
		out = append(out, method)
	}
	sort.Strings(out)
	return out
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods []string) {
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	}
	_ = netx.WriteMethodNotAllowed(w)
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
