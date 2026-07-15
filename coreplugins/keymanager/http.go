//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"
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

func (p *keyManagerPlugin) HandleHTTP(ctx context.Context, req *panel.HTTPRequest) (*panel.HTTPResponse, error) {
	path := strings.TrimRight(req.GetPath(), "/")
	if path == "" {
		path = "/"
	}

	switch {
	case req.GetMethod() == http.MethodGet && path == "/keys":
		return p.listResponse()
	case req.GetMethod() == http.MethodPost && path == "/keys/add":
		return p.addResponse(ctx, req.GetBody())
	case req.GetMethod() == http.MethodPost && path == "/keys/update":
		return p.updateResponse(ctx, req.GetBody())
	case req.GetMethod() == http.MethodPost && path == "/keys/reveal":
		return p.revealResponse(req.GetBody())
	case req.GetMethod() == http.MethodPost && path == "/keys/delete":
		return p.deleteResponse(ctx, req.GetBody())
	default:
		return jsonResponse(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Not found",
		})
	}
}

func (p *keyManagerPlugin) listResponse() (*panel.HTTPResponse, error) {
	p.mu.RLock()
	params := cloneParams(p.params)
	p.mu.RUnlock()

	keys, err := listSecretMeta(params)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
	}

	return jsonResponse(http.StatusOK, map[string]any{
		"success": true,
		"keys":    keys,
	})
}

func (p *keyManagerPlugin) addResponse(ctx context.Context, body []byte) (*panel.HTTPResponse, error) {
	return p.writeResponse(ctx, body, false)
}

func (p *keyManagerPlugin) updateResponse(ctx context.Context, body []byte) (*panel.HTTPResponse, error) {
	return p.writeResponse(ctx, body, true)
}

func (p *keyManagerPlugin) writeResponse(ctx context.Context, body []byte, update bool) (*panel.HTTPResponse, error) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	var payload secretWriteRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
	}
	if err := validateSecretName(payload.Name); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
	}
	if exists := p.secretExists(payload.Name); exists != update {
		if update {
			return jsonResponse(http.StatusNotFound, map[string]any{"success": false, "message": "Secret not found"})
		}
		return jsonResponse(http.StatusConflict, map[string]any{"success": false, "message": "Secret already exists"})
	}
	allowedPlugins, err := normalizePlugins(payload.AllowedPlugins)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
	}

	var ciphertext string
	encryption := secretEncryptionIdentity
	if update && payload.Value == "*" {
		if payload.Passphrase != "" {
			return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": "You must edit both value and passphrase"})
		}
		var ok bool
		ciphertext, ok = p.secretCiphertext(payload.Name)
		if !ok {
			return jsonResponse(http.StatusNotFound, map[string]any{"success": false, "message": "Secret not found"})
		}
		encryption, err = p.secretEncryption(payload.Name)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		}
	} else {
		ciphertext, encryption, err = p.encryptSecret(payload.Value, payload.Passphrase)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
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
		return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": "Encode metadata"})
	}
	policyJSON, err := json.Marshal(allowedPlugins)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": "Encode access policy"})
	}

	if err := p.patchParams(ctx, map[string]string{
		paramSecretPrefix + payload.Name: ciphertext,
		paramPolicyPrefix + payload.Name: string(policyJSON),
		paramMetaPrefix + payload.Name:   string(metaJSON),
	}, nil); err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
	}

	status := http.StatusOK
	if !update {
		status = http.StatusCreated
	}
	return jsonResponse(status, map[string]any{"success": true, "name": payload.Name})
}

func (p *keyManagerPlugin) revealResponse(body []byte) (*panel.HTTPResponse, error) {
	var payload secretRevealRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
	}
	if err := validateSecretName(payload.Name); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
	}

	value, err := p.decryptSecret(payload.Name, payload.Passphrase)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, errPassphraseRequired) || errors.Is(err, errInvalidPassphrase) {
			status = http.StatusBadRequest
		}
		return jsonResponse(status, map[string]any{"success": false, "message": err.Error()})
	}
	return jsonResponse(http.StatusOK, map[string]any{
		"success": true,
		"name":    payload.Name,
		"value":   value,
	})
}

func (p *keyManagerPlugin) deleteResponse(ctx context.Context, body []byte) (*panel.HTTPResponse, error) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	var payload secretNameRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid JSON body"})
	}
	if err := validateSecretName(payload.Name); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
	}

	if err := p.patchParams(ctx, nil, []string{
		paramSecretPrefix + payload.Name,
		paramPolicyPrefix + payload.Name,
		paramMetaPrefix + payload.Name,
	}); err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
	}
	return jsonResponse(http.StatusOK, map[string]any{"success": true, "name": payload.Name})
}
