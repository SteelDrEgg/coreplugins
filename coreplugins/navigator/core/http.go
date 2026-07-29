package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

func (s *navigatorService) handleHTTP(writer http.ResponseWriter, request *http.Request) {
	validRoute := request.Method == http.MethodGet && request.URL.Path == entriesAPIPath ||
		(request.Method == http.MethodGet || request.Method == http.MethodPut) && request.URL.Path == configAPIPath
	if !validRoute {
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

	if request.URL.Path == configAPIPath {
		s.handleConfig(writer, request)
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
			"entries": entries,
		},
	})
}

func (s *navigatorService) handleConfig(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		config, err := s.reloadConfig(request.Context())
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Failed to read Navigator settings",
			})
			return
		}
		writeConfig(writer, config)
		return
	}

	var update navigatorConfigUpdate
	decoder := json.NewDecoder(io.LimitReader(request.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&update); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid Navigator settings",
		})
		return
	}
	if err := ensureJSONEnd(decoder); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid Navigator settings",
		})
		return
	}

	config, err := s.updateConfig(request.Context(), update)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to save Navigator settings",
		})
		return
	}
	writeConfig(writer, config)
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}

func writeConfig(writer http.ResponseWriter, config navigatorConfig) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"icon":  config.Icon,
			"order": config.Order,
		},
	})
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
