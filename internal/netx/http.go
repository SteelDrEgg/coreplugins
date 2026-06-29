package netx

import (
	"fmt"
	"net/http"
)

// HandleSafe registers a handler and converts ServeMux duplicate-pattern panics
// into a regular error for callers.
func HandleSafe(mux *http.ServeMux, pattern string, handler http.Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("register http pattern %q: %v", pattern, r)
		}
	}()
	mux.Handle(pattern, handler)
	return nil
}
