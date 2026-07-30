package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

func (s *navigatorService) handleHTTP(writer http.ResponseWriter, request *http.Request) {
	validRoute := request.Method == http.MethodGet && request.URL.Path == entriesAPIPath ||
		(request.Method == http.MethodGet || request.Method == http.MethodPut) && request.URL.Path == configAPIPath ||
		(request.Method == http.MethodGet || request.Method == http.MethodPut) && request.URL.Path == serverFriendlyNamePath
	if !validRoute {
		writeJSON(writer, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
		return
	}

	publicServerFriendlyNameRead := request.Method == http.MethodGet &&
		request.URL.Path == serverFriendlyNamePath
	user, authenticated := arupa.UserFromContext(request.Context())
	if !publicServerFriendlyNameRead && !authenticated {
		writeJSON(writer, http.StatusUnauthorized, map[string]any{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	if request.URL.Path == serverFriendlyNamePath {
		s.handleServerFriendlyName(writer, request)
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

func (s *navigatorService) handleServerFriendlyName(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		config, err := s.reloadConfig(request.Context())
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Failed to read server friendly name",
			})
			return
		}
		writeServerFriendlyName(writer, config.ServerFriendlyName)
		return
	}

	var update serverFriendlyNameUpdate
	if err := decodeJSONBody(request, 4<<10, &update); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid server friendly name",
		})
		return
	}
	config, err := s.updateServerFriendlyName(request.Context(), update)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to save server friendly name",
		})
		return
	}
	writeServerFriendlyName(writer, config.ServerFriendlyName)
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
	if err := decodeJSONBody(request, 64<<10, &update); err != nil {
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

func decodeJSONBody(request *http.Request, limit int64, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, limit))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return ensureJSONEnd(decoder)
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
			"hide":  sortedSet(config.Hide),
		},
	})
}

func writeServerFriendlyName(writer http.ResponseWriter, name string) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"name": name,
		},
	})
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
