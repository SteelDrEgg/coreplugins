package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/internal/secrets"
)

func writeJSONResponse(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// writeServiceError maps a secrets.ServiceError (or any other error) to the
// same {"success": false, "message": ...} envelope every HTTP handler used
// to build ad hoc, with a single place deciding the status code.
func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var serviceErr *secrets.ServiceError
	if errors.As(err, &serviceErr) {
		switch serviceErr.Kind {
		case secrets.ErrInvalidInput, secrets.ErrPassphraseRequired, secrets.ErrInvalidPassphrase:
			status = http.StatusBadRequest
		case secrets.ErrNotFound:
			status = http.StatusNotFound
		case secrets.ErrConflict:
			status = http.StatusConflict
		case secrets.ErrForbidden:
			status = http.StatusForbidden
		}
	}
	writeJSONResponse(w, status, map[string]any{"success": false, "message": err.Error()})
}
