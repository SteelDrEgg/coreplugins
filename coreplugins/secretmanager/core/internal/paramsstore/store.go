// Package paramsstore persists each secret as a single host Params value.
package paramsstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

const IdentityKey = "secretmgr.identity"
const secretPrefix = "secrets."

// Repository has no local copy of Params. Service serializes all its operations.
type Repository struct{ client arupa.ParamsClient }

var _ secrets.Repository = (*Repository)(nil)

func New(client arupa.ParamsClient) *Repository { return &Repository{client: client} }

func (r *Repository) Identity(ctx context.Context) (string, error) {
	params, err := r.client.Params(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(params[IdentityKey]), nil
}

func (r *Repository) SetIdentity(ctx context.Context, identity string) error {
	return r.client.PatchParams(ctx, arupa.ParamsPatch{Set: map[string]string{IdentityKey: identity}})
}

func (r *Repository) Get(ctx context.Context, name string) (secrets.SecretRecord, error) {
	params, err := r.client.Params(ctx)
	if err != nil {
		return secrets.SecretRecord{}, err
	}
	raw, exists := params[secretPrefix+name]
	if !exists {
		return secrets.SecretRecord{}, secrets.RecordNotFound
	}
	return decodeRecord(name, raw)
}

func (r *Repository) List(ctx context.Context) ([]secrets.SecretRecord, error) {
	params, err := r.client.Params(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for key := range params {
		if strings.HasPrefix(key, secretPrefix) {
			names = append(names, strings.TrimPrefix(key, secretPrefix))
		}
	}
	sort.Strings(names)
	records := make([]secrets.SecretRecord, 0, len(names))
	for _, name := range names {
		record, err := decodeRecord(name, params[secretPrefix+name])
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func decodeRecord(name, raw string) (secrets.SecretRecord, error) {
	var record secrets.SecretRecord
	if err := secrets.DecodeJSON([]byte(raw), &record); err != nil {
		return record, fmt.Errorf("invalid record %q: expected the single-record secret format", name)
	}
	record.Name = name
	if err := record.Validate(); err != nil {
		return record, fmt.Errorf("invalid record %q: %w", name, err)
	}
	record.AllowedPlugins, _ = secrets.NormalizePlugins(record.AllowedPlugins)
	return record, nil
}

func (r *Repository) Create(ctx context.Context, record secrets.SecretRecord) error {
	if _, err := r.Get(ctx, record.Name); err == nil {
		return secrets.RecordExists
	} else if !errors.Is(err, secrets.RecordNotFound) {
		return err
	}
	return r.put(ctx, record)
}

func (r *Repository) Update(ctx context.Context, record secrets.SecretRecord) error {
	if _, err := r.Get(ctx, record.Name); err != nil {
		return err
	}
	return r.put(ctx, record)
}

func (r *Repository) put(ctx context.Context, record secrets.SecretRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return r.client.PatchParams(ctx, arupa.ParamsPatch{Set: map[string]string{secretPrefix + record.Name: string(raw)}})
}

func (r *Repository) Delete(ctx context.Context, name string) error {
	if _, err := r.Get(ctx, name); err != nil {
		return err
	}
	return r.client.PatchParams(ctx, arupa.ParamsPatch{Delete: []string{secretPrefix + name}})
}
