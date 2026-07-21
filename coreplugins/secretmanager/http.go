//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
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

func (p *secretManagerPlugin) handleHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimRight(req.URL.Path, "/")
	if path == "" {
		path = "/"
	}

	switch {
	case req.Method == http.MethodGet && path == "/keys":
		p.listResponse(w)
	case req.Method == http.MethodPost && path == "/keys/add":
		p.addResponse(w, req.Context(), req.Body)
	case req.Method == http.MethodPost && path == "/keys/update":
		p.updateResponse(w, req.Context(), req.Body)
	case req.Method == http.MethodPost && path == "/keys/reveal":
		p.revealResponse(w, req.Body)
	case req.Method == http.MethodPost && path == "/keys/delete":
		p.deleteResponse(w, req.Context(), req.Body)
	default:
		writeJSONResponse(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
	}
}

func (p *secretManagerPlugin) listResponse(w http.ResponseWriter) {
	p.mu.RLock()
	params := cloneParams(p.params)
	p.mu.RUnlock()

	keys, err := listSecretMeta(params)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"success": true,
		"keys":    keys,
	})
}

func (p *secretManagerPlugin) addResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	p.writeResponse(w, ctx, body, false)
}

func (p *secretManagerPlugin) updateResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	p.writeResponse(w, ctx, body, true)
}

func (p *secretManagerPlugin) writeResponse(w http.ResponseWriter, ctx context.Context, body io.Reader, update bool) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	var payload secretWriteRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}
	if err := validateSecretName(payload.Name); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}
	if exists := p.secretExists(payload.Name); exists != update {
		if update {
			writeJSONResponse(w, http.StatusNotFound, map[string]any{"success": false, "message": "Secret not found"})
			return
		}
		writeJSONResponse(w, http.StatusConflict, map[string]any{"success": false, "message": "Secret already exists"})
		return
	}
	allowedPlugins, err := normalizePlugins(payload.AllowedPlugins)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}

	var ciphertext string
	encryption := secretEncryptionIdentity
	if update && payload.Value == "*" {
		if payload.Passphrase != "" {
			writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "You must edit both value and passphrase"})
			return
		}
		var ok bool
		ciphertext, ok = p.secretCiphertext(payload.Name)
		if !ok {
			writeJSONResponse(w, http.StatusNotFound, map[string]any{"success": false, "message": "Secret not found"})
			return
		}
		encryption, err = p.secretEncryption(payload.Name)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
			return
		}
	} else {
		ciphertext, encryption, err = p.encryptSecret(payload.Value, payload.Passphrase)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
			return
		}
	}

	meta := secretMeta{
		Name:           payload.Name,
		Description:    strings.TrimSpace(payload.Description),
		AllowedPlugins: allowedPlugins,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		Encryption:     encryption,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Encode metadata"})
		return
	}
	policyJSON, err := json.Marshal(allowedPlugins)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Encode access policy"})
		return
	}

	if err := p.patchParams(ctx, map[string]string{
		paramSecretPrefix + payload.Name: ciphertext,
		paramPolicyPrefix + payload.Name: string(policyJSON),
		paramMetaPrefix + payload.Name:   string(metaJSON),
	}, nil); err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		return
	}

	status := http.StatusOK
	if !update {
		status = http.StatusCreated
	}
	writeJSONResponse(w, status, map[string]any{"success": true, "name": payload.Name})
}

func (p *secretManagerPlugin) revealResponse(w http.ResponseWriter, body io.Reader) {
	var payload secretRevealRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}
	if err := validateSecretName(payload.Name); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}

	value, err := p.decryptSecret(payload.Name, payload.Passphrase)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, errPassphraseRequired) || errors.Is(err, errInvalidPassphrase) {
			status = http.StatusBadRequest
		}
		writeJSONResponse(w, status, map[string]any{"success": false, "message": err.Error()})
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"success": true,
		"name":    payload.Name,
		"value":   value,
	})
}

func (p *secretManagerPlugin) deleteResponse(w http.ResponseWriter, ctx context.Context, body io.Reader) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	var payload secretNameRequest
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
		return
	}
	if err := validateSecretName(payload.Name); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}

	if err := p.patchParams(ctx, nil, []string{
		paramSecretPrefix + payload.Name,
		paramPolicyPrefix + payload.Name,
		paramMetaPrefix + payload.Name,
	}); err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]any{"success": true, "name": payload.Name})
}
