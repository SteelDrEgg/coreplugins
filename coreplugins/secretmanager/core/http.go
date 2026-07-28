package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

type secretWriteRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Value          string   `json:"value"`
	Passphrase     string   `json:"passphrase"`
	AllowedPlugins []string `json:"allowed_plugins"`
}

type secretNameRequest struct {
	Name string `json:"name"`
}

type secretRevealRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

// handleHTTP is the admin HTTP API used by pages/index.html. Every handler
// here is a thin adapter: decode the request, call into secrets.Service as
// secrets.HumanAdminCaller(), and translate the result to JSON.
func (p *secretManagerPlugin) handleHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimRight(req.URL.Path, "/")
	if path == "" {
		path = "/"
	}

	switch {
	case req.Method == http.MethodGet && path == "/keys":
		p.listResponse(w, req.Context())
	case req.Method == http.MethodPost && path == "/keys/add":
		p.addResponse(w, req.Context(), req.Body)
	case req.Method == http.MethodPost && path == "/keys/update":
		p.updateResponse(w, req.Context(), req.Body)
	case req.Method == http.MethodPost && path == "/keys/reveal":
		p.revealResponse(w, req.Context(), req.Body)
	case req.Method == http.MethodPost && path == "/keys/delete":
		p.deleteResponse(w, req.Context(), req.Body)
	default:
		writeJSONResponse(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
	}
}

func (p *secretManagerPlugin) listResponse(w http.ResponseWriter, ctx context.Context) {
	keys, err := p.secrets.ListSecrets(ctx, secrets.HumanAdminCaller())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]any{"success": true, "keys": keys})
}

func (p *secretManagerPlugin) addResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	p.writeResponse(w, ctx, body, false)
}

func (p *secretManagerPlugin) updateResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	p.writeResponse(w, ctx, body, true)
}

func (p *secretManagerPlugin) writeResponse(w http.ResponseWriter, ctx context.Context, body io.Reader, update bool) {
	var payload secretWriteRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}

	meta, err := p.secrets.WriteSecret(ctx, secrets.HumanAdminCaller(), secrets.WriteSecretInput{
		Name:           payload.Name,
		Description:    payload.Description,
		Value:          payload.Value,
		Passphrase:     payload.Passphrase,
		AllowedPlugins: payload.AllowedPlugins,
		Update:         update,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	status := http.StatusOK
	if !update {
		status = http.StatusCreated
	}
	writeJSONResponse(w, status, map[string]any{"success": true, "name": meta.Name})
}

func (p *secretManagerPlugin) revealResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	var payload secretRevealRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}

	value, err := p.secrets.GetSecret(ctx, secrets.HumanAdminCaller(), payload.Name, payload.Passphrase)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"success": true,
		"name":    payload.Name,
		"value":   value,
	})
}

func (p *secretManagerPlugin) deleteResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	var payload secretNameRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}

	if err := p.secrets.DeleteSecret(ctx, secrets.HumanAdminCaller(), payload.Name); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]any{"success": true, "name": payload.Name})
}
