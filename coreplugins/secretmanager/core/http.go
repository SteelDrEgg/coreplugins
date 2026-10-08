package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

const methodNotAllowed secrets.ErrorKind = "method_not_allowed"

func (p *secretManagerPlugin) handleHTTP(w http.ResponseWriter, req *http.Request) {
	command, status, allow, err := decodeHTTPCommand(req)
	if err != nil {
		if allow != "" {
			w.Header().Set("Allow", allow)
		}
		if status == 0 {
			status = errorStatus(err)
		}
		writeJSONResponse(w, status, errorResponse(err))
		return
	}
	result, err := p.executeCommand(req.Context(), secrets.HumanAdminCaller(), command)
	if err != nil {
		p.logOperationError(req.Context(), err)
		writeJSONResponse(w, errorStatus(err), errorResponse(err))
		return
	}
	status = http.StatusOK
	if command.op == opCreate {
		status = http.StatusCreated
		w.Header().Set("Location", secretResourcePath(result.Name))
	}
	writeJSONResponse(w, status, result)
}

func secretResourcePath(name string) string {
	// Browsers normalize a single-dot URL segment even when it is percent-encoded.
	// '~' is outside the secret-name alphabet, so this escape is unambiguous.
	if name == "." {
		return secretCollectionPath + "/~."
	}
	return secretCollectionPath + "/" + url.PathEscape(name)
}

func decodeHTTPCommand(req *http.Request) (secretCommand, int, string, error) {
	command := secretCommand{}
	path := req.URL.Path // Already decoded by net/http and the host; never unescape twice.
	if path == secretCollectionPath {
		switch req.Method {
		case http.MethodGet:
			command.op = opList
			return command, 0, "", requireEmptyBody(req.Body)
		case http.MethodPost:
			command.op = opCreate
			return command, 0, "", decodeRequest(req.Body, &command.create)
		default:
			return command, http.StatusMethodNotAllowed, "GET, POST", secrets.NewError(methodNotAllowed, "method not allowed")
		}
	}
	if !strings.HasPrefix(path, secretCollectionPath+"/") {
		return command, http.StatusNotFound, "", secrets.NewError(secrets.ErrNotFound, "not found")
	}
	name := strings.TrimPrefix(path, secretCollectionPath+"/")
	allow := "PATCH, DELETE"
	if strings.HasSuffix(name, "/reveal") {
		allow += ", POST"
	}
	switch req.Method {
	case http.MethodPatch:
		var fields struct {
			Description    secrets.Field[string]             `json:"description"`
			AllowedPlugins secrets.Field[[]string]           `json:"allowed_plugins"`
			Value          secrets.Field[secrets.ValueInput] `json:"value"`
		}
		command.op = opUpdate
		if err := decodeRequest(req.Body, &fields); err != nil {
			return command, http.StatusBadRequest, "", err
		}
		command.update = secrets.UpdateSecretInput{
			Name: name, Description: fields.Description,
			AllowedPlugins: fields.AllowedPlugins, Value: fields.Value,
		}
	case http.MethodDelete:
		command.op, command.name = opDelete, name
		if err := requireEmptyBody(req.Body); err != nil {
			return command, http.StatusBadRequest, "", err
		}
	case http.MethodPost:
		if !strings.HasSuffix(name, "/reveal") {
			return command, http.StatusMethodNotAllowed, allow, secrets.NewError(methodNotAllowed, "method not allowed")
		}
		var input struct {
			Passphrase string `json:"passphrase"`
		}
		if err := decodeRequest(req.Body, &input); err != nil {
			return command, http.StatusBadRequest, "", err
		}
		command.op, command.name = opReveal, strings.TrimSuffix(name, "/reveal")
		command.passphrase = input.Passphrase
	default:
		return command, http.StatusMethodNotAllowed, allow, secrets.NewError(methodNotAllowed, "method not allowed")
	}
	if command.update.Name == "~." {
		command.update.Name = "."
	}
	if command.name == "~." {
		command.name = "."
	}
	return command, 0, "", nil
}

func requireEmptyBody(body io.Reader) error {
	if body == nil {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(body, 1))
	if err != nil {
		return secrets.NewError(secrets.ErrInvalidInput, "cannot read request body")
	}
	if len(data) != 0 {
		return secrets.NewError(secrets.ErrInvalidInput, "request body must be empty")
	}
	return nil
}
