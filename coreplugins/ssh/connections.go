package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	savedConnectionsParam = "ssh.connections"
	savedConnectionsPath  = "/ssh/api/connections"

	authTypePassword     = "password"
	authTypeKey          = "key"
	legacyAuthTypeSecret = "secret"
)

// savedConnection is the non-sensitive portion of an SSH connection profile.
// Passwords, secret values, and passphrases deliberately have no field here.
type savedConnection struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       string `json:"port"`
	Username   string `json:"username"`
	AuthType   string `json:"auth_type"`
	PrivateKey string `json:"private_key,omitempty"`
	SecretName string `json:"secret_name,omitempty"`
}

type savedConnectionsResponse struct {
	Success     bool              `json:"success"`
	Connections []savedConnection `json:"connections"`
}

func (s *sshServer) handleConnectionsHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.TrimRight(req.URL.Path, "/") != savedConnectionsPath {
		writeConnectionJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
		return
	}

	switch req.Method {
	case http.MethodGet:
		writeConnectionJSON(w, http.StatusOK, savedConnectionsResponse{
			Success:     true,
			Connections: s.savedConnections(),
		})
	case http.MethodPost:
		s.saveConnectionResponse(w, req)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		writeConnectionJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"message": "Method not allowed",
		})
	}
}

func (s *sshServer) loadSavedConnections(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var connections []savedConnection
	if err := json.Unmarshal([]byte(raw), &connections); err != nil {
		return fmt.Errorf("parse %s: %w", savedConnectionsParam, err)
	}

	loaded := make(map[string]savedConnection, len(connections))
	for _, connection := range connections {
		normalized, err := normalizeSavedConnection(connection)
		if err != nil {
			return fmt.Errorf("parse %s: %w", savedConnectionsParam, err)
		}
		if _, exists := loaded[normalized.Name]; exists {
			return fmt.Errorf("parse %s: duplicate connection %q", savedConnectionsParam, normalized.Name)
		}
		loaded[normalized.Name] = normalized
	}

	s.settingsMu.Lock()
	s.settings = loaded
	s.settingsMu.Unlock()
	return nil
}

func (s *sshServer) savedConnections() []savedConnection {
	s.settingsMu.RLock()
	connections := make([]savedConnection, 0, len(s.settings))
	for _, connection := range s.settings {
		connections = append(connections, connection)
	}
	s.settingsMu.RUnlock()

	sort.Slice(connections, func(i, j int) bool {
		return strings.ToLower(connections[i].Name) < strings.ToLower(connections[j].Name)
	})
	return connections
}

func (s *sshServer) saveConnectionResponse(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeConnectionJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	var connection savedConnection
	if err := json.Unmarshal(body, &connection); err != nil {
		writeConnectionJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid JSON body",
		})
		return
	}

	normalized, err := normalizeSavedConnection(connection)
	if err != nil {
		writeConnectionJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	s.settingsWriteMu.Lock()
	defer s.settingsWriteMu.Unlock()

	next := make(map[string]savedConnection)
	s.settingsMu.RLock()
	for name, existing := range s.settings {
		next[name] = existing
	}
	s.settingsMu.RUnlock()
	next[normalized.Name] = normalized

	connections := make([]savedConnection, 0, len(next))
	for _, item := range next {
		connections = append(connections, item)
	}
	sort.Slice(connections, func(i, j int) bool {
		return strings.ToLower(connections[i].Name) < strings.ToLower(connections[j].Name)
	})

	encoded, err := json.Marshal(connections)
	if err != nil {
		writeConnectionJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to encode SSH connections",
		})
		return
	}
	if err := s.patchConnectionParams(req.Context(), string(encoded)); err != nil {
		writeConnectionJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	s.settingsMu.Lock()
	s.settings = next
	s.settingsMu.Unlock()
	writeConnectionJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"connection": normalized,
	})
}

func (s *sshServer) patchConnectionParams(ctx context.Context, encoded string) error {
	err := s.host.patchParams(ctx, map[string]string{savedConnectionsParam: encoded})
	if err != nil {
		return fmt.Errorf("persist SSH connections: %w", err)
	}
	return nil
}

func normalizeSavedConnection(connection savedConnection) (savedConnection, error) {
	connection.Name = strings.TrimSpace(connection.Name)
	connection.Host = strings.TrimSpace(connection.Host)
	connection.Port = strings.TrimSpace(connection.Port)
	connection.Username = strings.TrimSpace(connection.Username)
	connection.AuthType = strings.TrimSpace(connection.AuthType)
	connection.PrivateKey = strings.TrimSpace(connection.PrivateKey)
	connection.SecretName = strings.TrimSpace(connection.SecretName)

	if connection.Name == "" || len(connection.Name) > 80 {
		return savedConnection{}, fmt.Errorf("connection name is required and must not exceed 80 characters")
	}
	if connection.Host == "" || connection.Username == "" {
		return savedConnection{}, fmt.Errorf("host and username are required")
	}
	if connection.Port == "" {
		connection.Port = "22"
	}
	port, err := strconv.Atoi(connection.Port)
	if err != nil || port < 1 || port > 65535 {
		return savedConnection{}, fmt.Errorf("port must be between 1 and 65535")
	}

	switch connection.AuthType {
	case authTypePassword:
		connection.PrivateKey = ""
	case authTypeKey:
		connection.SecretName = ""
	case legacyAuthTypeSecret:
		connection.AuthType = authTypePassword
		connection.PrivateKey = ""
		if connection.SecretName == "" {
			return savedConnection{}, fmt.Errorf("secret name is required for a Secret Manager password")
		}
	default:
		return savedConnection{}, fmt.Errorf("unsupported authentication type %q", connection.AuthType)
	}
	return connection, nil
}

func writeConnectionJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
