package main

import (
	"bytes"
	"context"
	"io"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

type operation int

const (
	opList operation = iota
	opCreate
	opUpdate
	opReveal
	opDelete
	// Allow up to six-byte JSON escaping of the 1 MiB plaintext plus metadata.
	maxRequestSize = 8 * secrets.MaxSecretSize
)

type secretNameRequest struct {
	Name string `json:"name"`
}
type secretRevealRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

type secretCommand struct {
	op         operation
	create     secrets.CreateSecretInput
	update     secrets.UpdateSecretInput
	name       string
	passphrase string
}

// execute decodes ISC payloads. HTTP supplies names from resource paths;
// both adapters dispatch typed commands through executeCommand.
func (p *secretManagerPlugin) execute(ctx context.Context, caller secrets.Caller, op operation, body io.Reader) (response, error) {
	command := secretCommand{op: op}
	switch op {
	case opList:
		if body != nil {
			if err := decodeRequest(body, &struct{}{}); err != nil {
				return response{}, err
			}
		}
	case opCreate:
		if err := decodeRequest(body, &command.create); err != nil {
			return response{}, err
		}
	case opUpdate:
		if err := decodeRequest(body, &command.update); err != nil {
			return response{}, err
		}
	case opReveal:
		var input secretRevealRequest
		if err := decodeRequest(body, &input); err != nil {
			return response{}, err
		}
		command.name, command.passphrase = input.Name, input.Passphrase
	case opDelete:
		var input secretNameRequest
		if err := decodeRequest(body, &input); err != nil {
			return response{}, err
		}
		command.name = input.Name
	default:
		return response{}, secrets.NewError(secrets.ErrInvalidInput, "unsupported operation")
	}
	return p.executeCommand(ctx, caller, command)
}

func (p *secretManagerPlugin) executeCommand(ctx context.Context, caller secrets.Caller, command secretCommand) (response, error) {
	if p.secrets == nil {
		return response{}, secrets.NewError(secrets.ErrInternal, "secret-manager is not initialized")
	}
	switch command.op {
	case opList:
		keys, err := p.secrets.ListSecrets(ctx, caller)
		if err != nil {
			return response{}, err
		}
		return response{Success: true, Keys: &keys}, nil
	case opCreate:
		info, err := p.secrets.CreateSecret(ctx, caller, command.create)
		if err != nil {
			return response{}, err
		}
		return response{Success: true, Name: info.Name, Secret: &info}, nil
	case opUpdate:
		info, err := p.secrets.UpdateSecret(ctx, caller, command.update)
		if err != nil {
			return response{}, err
		}
		return response{Success: true, Name: info.Name, Secret: &info}, nil
	case opReveal:
		value, err := p.secrets.GetSecret(ctx, caller, command.name, command.passphrase)
		if err != nil {
			return response{}, err
		}
		name, _ := secrets.NormalizeName(command.name)
		return response{Success: true, Name: name, Value: &value}, nil
	case opDelete:
		if err := p.secrets.DeleteSecret(ctx, caller, command.name); err != nil {
			return response{}, err
		}
		name, _ := secrets.NormalizeName(command.name)
		return response{Success: true, Name: name}, nil
	}
	return response{}, secrets.NewError(secrets.ErrInvalidInput, "unsupported operation")
}

func decodeRequest(body io.Reader, target any) error {
	if body == nil {
		return secrets.NewError(secrets.ErrInvalidInput, "JSON object required")
	}
	raw, err := io.ReadAll(io.LimitReader(body, maxRequestSize+1))
	if err != nil {
		return secrets.NewError(secrets.ErrInvalidInput, "cannot read request body")
	}
	if len(raw) > maxRequestSize {
		return secrets.NewError(secrets.ErrInvalidInput, "request body is too large")
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return secrets.NewError(secrets.ErrInvalidInput, "JSON object required")
	}
	if err := secrets.DecodeJSON(raw, target); err != nil {
		return secrets.NewError(secrets.ErrInvalidInput, "invalid request: "+err.Error())
	}
	return nil
}
