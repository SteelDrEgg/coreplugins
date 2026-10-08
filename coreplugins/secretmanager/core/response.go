package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

type response struct {
	Success bool                  `json:"success"`
	Code    secrets.ErrorKind     `json:"code,omitempty"`
	Message string                `json:"message,omitempty"`
	Name    string                `json:"name,omitempty"`
	Secret  *secrets.SecretInfo   `json:"secret,omitempty"`
	Keys    *[]secrets.SecretInfo `json:"keys,omitempty"`
	Value   *string               `json:"value,omitempty"`
}

func errorResponse(err error) response {
	var domain *secrets.ServiceError
	if errors.As(err, &domain) {
		return response{Code: domain.Kind, Message: domain.Msg}
	}
	return response{Code: secrets.ErrInternal, Message: "secret operation failed"}
}

func errorStatus(err error) int {
	switch errorResponse(err).Code {
	case secrets.ErrInvalidInput, secrets.ErrPassphraseRequired, secrets.ErrInvalidPassphrase:
		return http.StatusBadRequest
	case secrets.ErrNotFound:
		return http.StatusNotFound
	case secrets.ErrConflict:
		return http.StatusConflict
	case secrets.ErrForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (p *secretManagerPlugin) logOperationError(ctx context.Context, err error) {
	if p.sdk == nil {
		return
	}
	var domain *secrets.ServiceError
	if errors.As(err, &domain) && domain.Cause != nil {
		_ = p.sdk.LogError(ctx, domain.Cause.Error())
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, payload response) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
