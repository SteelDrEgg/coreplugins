package plugin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxRequestBody caps the request body forwarded to plugins (8 MiB).
const maxRequestBody = 8 << 20

// registerRoute wires a plugin HTTP route into the host's ServeMux. The handler
// serializes the request, forwards it to the plugin, and writes the reply.
func (m *Manager) registerRoute(route HTTPRoute, conn pluginConn) error {
	if route.Pattern == "" {
		return fmt.Errorf("http route pattern is required")
	}

	method := strings.ToUpper(strings.TrimSpace(route.Method))
	pattern := route.Pattern

	handler := func(w http.ResponseWriter, r *http.Request) {
		if method != "" && !strings.EqualFold(r.Method, method) {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
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
			http.Error(w, "plugin error: "+err.Error(), http.StatusBadGateway)
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

	// http.ServeMux panics on duplicate patterns; surface that as an error.
	if err := safeHandle(m.mux, pattern, http.HandlerFunc(handler)); err != nil {
		return err
	}
	return nil
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

func safeHandle(mux *http.ServeMux, pattern string, handler http.Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("register http pattern %q: %v", pattern, r)
		}
	}()
	mux.Handle(pattern, handler)
	return nil
}
