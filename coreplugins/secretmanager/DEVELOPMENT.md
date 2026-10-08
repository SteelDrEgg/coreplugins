# Secret Manager Development Guide

This guide describes integration and implementation for the `secret-manager`
WASM service, plugin version 0.3.0 and service contract version 2. See the
[README](README.md) for user instructions, access-control endpoints, and Config.

## Contract and validation

HTTP and ISC use the same domain operations and value schema. HTTP identifies
existing secrets through resource paths; ISC includes `name` in its JSON payload.
All application requests with a payload require one UTF-8 JSON object. Unknown
fields, JSON null, duplicate object keys, and trailing JSON documents are rejected.
Payloads are limited to 8 MiB to accommodate JSON escaping of the plaintext limit.

Secret names are immutable, trimmed, and contain 1–128 ASCII characters. Letters,
digits, `/`, `_`, `-`, and `.` are allowed; `..` is forbidden. Descriptions are
trimmed. Allowed service names are trimmed, deduplicated, sorted, and must match
the caller's registered service name exactly; empty entries are rejected.

Plaintext is limited to 1 MiB of UTF-8 bytes, checked before encryption and after
decryption. An explicitly empty plaintext is supported. `"*"` is an ordinary
value with no special update meaning.

### Create input

```json
{
  "name": "database-password",
  "description": "Database credentials",
  "allowed_plugins": ["database"],
  "value": {
    "plaintext": "secret-value",
    "protection": "identity"
  }
}
```

`name` and `value` are required. Omitted `description` and `allowed_plugins`
default to an empty string and empty list. The `value` object must explicitly
provide both `plaintext` and `protection`.

| Protection | Meaning | Passphrase input |
|------------|---------|------------------|
| `identity` | Age X25519 encryption with the manager's identity | A non-empty `value.passphrase` is rejected |
| `passphrase` | Age scrypt encryption | A non-empty `value.passphrase` is required |

For example, a passphrase-protected value is supplied as:

```json
{
  "plaintext": "secret-value",
  "protection": "passphrase",
  "passphrase": "example-passphrase"
}
```

Use this object as the request's `value` field. Passphrases are never persisted.
A duplicate creation fails with `conflict`; it does not replace the existing
secret.

### Partial updates

Only `description`, `allowed_plugins`, and `value` can be changed. At least one
of these fields must be supplied.

| Input | Result |
|-------|--------|
| Omit a mutable field | Keep its existing value |
| `"description": ""` | Clear the description |
| `"allowed_plugins": []` | Remove all service access |
| Supply `value` | Replace both the encrypted value and its protection mode |

A replacement `value` must explicitly contain `plaintext` and `protection`,
including when saving an empty plaintext. Omitting `value` preserves the existing
ciphertext and protection. Metadata-only updates do not decrypt the old value
and do not require its passphrase.

## Authorization

HTTP routes require host authentication. Calls accepted by the HTTP management
interface operate as a human administrator and can manage all secrets, regardless
of their allowed service lists. Revealing a passphrase-protected value still
requires its passphrase.

ISC identity comes from the host-provided source service name. Payloads cannot
specify or override the caller. A missing source identity is rejected.

`allowed_plugins` grants combined read and management access. An allowed service
can reveal, update, delete, and change the permissions of a secret. There are no
separate read and write roles. An empty list permits no service access to that
secret, and listing only returns records accessible to the caller.

Any authenticated service may create a secret through ISC. Its source name is
automatically added to the new record's allowed list. Updating the allowed list
replaces it without automatically retaining the caller; a service can revoke
its own access. Saved permissions apply to subsequent operations without restart.

For an ISC caller, an unknown name and an inaccessible record both return
`forbidden`. For an HTTP administrator, an unknown name returns `not_found`.

## HTTP API

All routes share the namespace derived from the service name, `/secret-manager`.
Secret resources are under `/secret-manager/secrets`; page and icon paths are
separate so they cannot shadow a valid secret name.

| Method | Endpoint | Request body | Successful response |
|--------|----------|--------------|---------------------|
| `GET` | `/secret-manager/secrets` | None | HTTP 200: `success`, `keys` |
| `POST` | `/secret-manager/secrets` | Create input | HTTP 201: `success`, `name`, `secret` |
| `PATCH` | `/secret-manager/secrets/{name}` | Mutable fields only | HTTP 200: `success`, `name`, `secret` |
| `DELETE` | `/secret-manager/secrets/{name}` | None | HTTP 200: `success`, `name` |
| `POST` | `/secret-manager/secrets/{name}/reveal` | `{}` or optional `passphrase` | HTTP 200: `success`, `name`, `value` |

### Paths and bodies

Encode the entire secret name as one URL component, including `/`. The name
`database/password` uses `/secret-manager/secrets/database%2Fpassword`. Leading,
trailing, and repeated slashes belong to the name and are preserved. The host
forwards a decoded path; the HTTP adapter does not decode it a second time.

The name `.` uses `/secret-manager/secrets/~.` because browsers normalize
single-dot path segments, including percent-encoded ones. `~` is outside the
secret-name alphabet, so this escape is unambiguous.

PATCH, DELETE, and reveal take the name only from the path. Repeating `name` in
a PATCH or reveal body is rejected; DELETE and list GET require an empty body.
There is no full-replacement PUT endpoint. Unsupported HTTP methods return 405
with an `Allow` header.

For example, `PATCH /secret-manager/secrets/database-password` accepts:

```json
{
  "description": "",
  "allowed_plugins": []
}
```

This clears the description and permissions while preserving the encrypted value.

For identity protection, reveal with
`POST /secret-manager/secrets/database-password/reveal` and `{}`. For passphrase
protection, use the same endpoint with:

```json
{
  "passphrase": "example-passphrase"
}
```

Keep the passphrase in the JSON body. An empty decrypted plaintext is explicitly
returned as `"value": ""`.

### Responses

Creating a secret also returns a `Location` header with its encoded resource
path. Create and update return public metadata in `secret`:

```json
{
  "success": true,
  "name": "database-password",
  "secret": {
    "name": "database-password",
    "description": "Database credentials",
    "allowed_plugins": ["database"],
    "protection": "identity",
    "updated_at": "2026-10-07T00:00:00Z"
  }
}
```

List returns this metadata in a `keys` array, sorted by name. An empty list
returns `"keys": []`. Metadata replies contain neither ciphertext nor plaintext.
Only reveal responses contain the decrypted `value`.

Application JSON responses include `Content-Type: application/json; charset=utf-8`
and `Cache-Control: no-store`.

## ISC

Send a UTF-8 JSON payload to target `secret-manager` using
[inter-service communication (ISC)](https://docs.arupa.dev/docs/arupa/v0/en/service/isc).
Topics directly name actions and are local to the target service.

| Topic | Payload | Successful JSON reply |
|-------|---------|-----------------------|
| `list` | `{}` | `success`, `keys` |
| `create` | Create input | `success`, `name`, `secret` |
| `update` | `name` and mutable fields | `success`, `name`, `secret` |
| `reveal` | `name`, optional `passphrase` | `success`, `name`, `value` |
| `delete` | `name` | `success`, `name` |

Replies use the same JSON shapes as HTTP, including for reveal. An ISC update
includes `name` in the payload:

```json
{
  "name": "database-password",
  "description": "",
  "allowed_plugins": []
}
```

Messages are synchronous request/reply calls, and the target must be running.
The host does not queue messages for offline services or retry them automatically.
Unsupported topics and malformed payloads return errors.

For a Go SDK service, where `service` is the SDK service instance and `ctx` is the
call context:

```go
reply, err := service.SendServiceJSON(ctx, "secret-manager", "reveal", map[string]string{
    "name": "database-password",
})
if err != nil {
    return err
}
var result struct {
    Success bool   `json:"success"`
    Value   string `json:"value"`
}
if err := json.Unmarshal([]byte(reply), &result); err != nil {
    return err
}
// Use result.Value as the decrypted secret; do not log it.
```

Import `encoding/json`. Include `passphrase` in the payload when required.

## Errors

Application failures use a JSON object with `success: false`, `code`, and
`message`:

```json
{
  "success": false,
  "code": "forbidden",
  "message": "plugin is not allowed to access this secret"
}
```

| Code | HTTP status | Meaning |
|------|-------------|---------|
| `invalid_input` | 400 | Invalid payload, name, value, or unsupported ISC topic |
| `passphrase_required` | 400 | A protected value requires a passphrase |
| `invalid_passphrase` | 400 | The supplied passphrase cannot decrypt the value |
| `forbidden` | 403 | The caller is not allowed to access the secret |
| `not_found` | 404 | The requested HTTP resource does not exist |
| `method_not_allowed` | 405 | An HTTP method is not supported for this path |
| `conflict` | 409 | A secret with this name already exists |
| `internal_error` | 500 | Persistence, cryptography, or another internal operation failed |

ISC failures return an SDK error whose remote message contains the JSON failure
object. Internal causes are logged and are not exposed in the application
response. Authentication, routing, and transport errors returned by the host
before invoking the service are outside this application error schema.

## Storage format

The identity and secrets are stored under `[Services.secret-manager.Params]`.
Each `secrets.<name>` parameter contains one complete JSON record. Its name comes
from the parameter key and is not duplicated inside the record. The access list
is stored once, in `allowed_plugins`.

```toml
[Services.secret-manager.Params]
"secretmgr.identity" = "env://SECRET_KEY"
"secrets.database-password" = '''
{
  "description": "Database credentials",
  "allowed_plugins": ["database", "ssh"],
  "encrypted_value": {
    "protection": "identity",
    "ciphertext": "<base64 age ciphertext>"
  },
  "updated_at": "2026-10-07T00:00:00Z"
}
'''
```

The ciphertext above is a placeholder. Create valid records through the page or
service interfaces.

| Stored field | Meaning |
|--------------|---------|
| `secretmgr.identity` | Age X25519 private identity, stored directly or supplied through an environment reference |
| `description` | Trimmed description |
| `allowed_plugins` | Registered service names with combined read and management access |
| `encrypted_value.protection` | `identity` for X25519 or `passphrase` for age scrypt encryption |
| `encrypted_value.ciphertext` | Base64-encoded age ciphertext |
| `updated_at` | Time of the last change in UTC RFC 3339 format |

Only the value is encrypted. Descriptions, allowed service names, and timestamps
remain readable in Params. Plaintext values and passphrases are not persisted.

### Initialization and persistence

Startup validates all stored records before accepting requests. On first start,
a missing identity is generated and persisted before exposing the service.
If identity-protected records already exist and the identity is missing, startup
fails without generating a replacement. Preserve the identity to keep access to
those values; passphrase-protected values require their original passphrases.

The repository reads current host Params without maintaining a separate local
copy. A mutation writes or deletes the whole record in a single Params patch.
A failed persistence operation does not leave an updated local copy behind.

HTTP and ISC reads and mutations enter the same queue and execute serially in
queue order. Ordering starts when an operation enters the domain service, after
payload decoding. Concurrent same-name creation allows one winner; later creates
return `conflict`. Successive valid updates apply to the latest stored record.
There are no revision fields, conditional writes, or local state snapshots.

## Compatibility and upgrades

Version 0.3.0 does not load or migrate the previous `.secret`, `.allowed_plugins`,
and `.meta` three-key format. Old or invalid records prevent startup without
modifying data. Back up and export existing secrets using the previous version,
retain the identity, and recreate the secrets through the new interfaces.

The old HTTP namespace and action endpoints have no compatibility aliases.
The old prefixed ISC topics are not accepted. Reveal replies are JSON rather
than the previous plaintext reply, and `"*"` no longer preserves an existing
value during an update.

## Implementation and verification

[core/service.go](core/service.go) initializes persistence and cryptography and
registers transports. [core/http.go](core/http.go) and
[core/plugin_message.go](core/plugin_message.go) decode their transport-specific
requests and share the command adapter in [core/operations.go](core/operations.go).

| Package | Responsibility |
|---------|----------------|
| [internal/secrets](core/internal/secrets/service.go) | Records, validation, permissions, and serial domain operations; defines Repository and Cipher interfaces |
| [internal/paramsstore](core/internal/paramsstore/store.go) | Repository implementation using host Params |
| [internal/agecrypto](core/internal/agecrypto/crypto.go) | Cipher implementation using age |

Secrets are not stored in shared KV and have no Socket.IO interface.

Run from the coreplugins repository root:

```sh
go test -race ./coreplugins/secretmanager/core/...
npm run check:secretmanager
```

The tests cover shared HTTP/ISC state, all action topics, authorization, partial
updates, literal and empty values, request validation, encoded names, FIFO
execution, failed persistence, startup failures, and encryption round trips.

## Build and package

Run from `coreplugins/secretmanager`:

```sh
make build
make package
```

`make build` produces `dist/secret_manager.wasm` in the repository root and static
files in `ui/build/`. `make package` creates `dist/secret-manager.plg` by default;
`PLUGIN_DIR` can override the package output directory. Rescan the service
directory and restart the service to load a new build.
