package main

import (
	"encoding/json"
	"net/http"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

func (s *navigatorService) handleHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet || request.URL.Path != entriesAPIPath {
		writeJSON(writer, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
		return
	}

	user, authenticated := arupa.UserFromContext(request.Context())
	if !authenticated {
		writeJSON(writer, http.StatusUnauthorized, map[string]any{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}
	entries, err := s.navigationEntries(request.Context(), user)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to load navigation entries",
		})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"entries":   entries,
			"languages": s.config.Languages,
		},
	})
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
