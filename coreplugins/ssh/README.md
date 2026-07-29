# SSH Terminal Plugin

This core plugin provides the authenticated SSH terminal application under
`/ssh/`. It uses the v2 Arupa SDK to receive an inherited Unix listener, serves
its own HTTP server on that listener, and registers one inherited proxy
transport plus the `/ssh/` route with the host.

Navigation discovery uses the registered `entry` GET route at `/ssh/pages/`;
the page entry is no longer declared in `info.yaml`.

Terminal traffic uses a native WebSocket at `/ssh/ws`; it does not use the
host's Socket.IO server. The legacy Socket.IO adapter remains in the codebase
for now, but the v2 service does not register it.

## Layout

- `core/` contains the gRPC service, inherited HTTP server, WebSocket transport,
  SSH connection handling, saved profiles, and tests.
- `ui/` is a SvelteKit static application. Reusable components live under
  `ui/src/lib/components`, transport and API helpers under `ui/src/lib/utils`,
  and the terminal entry route under `ui/src/routes/index.html`.
- `ui/locale/` contains the Paraglide English and Chinese catalogs. Locale
  resolution uses `localStorage["arupa.language"]`, then the browser language,
  then English.
- `ui/static/` contains the plugin icons and the vendored xterm browser runtime.

## Frontend Contract

The terminal page opens `/ssh/ws`. Every text frame is a JSON envelope:

```json
{"event":"connect_ssh","data":{"host":"example.com","port":"22","username":"alice"}}
```

Client events are:

- `connect_ssh`: `{ host, port, username, password?, privateKey?, passphrase? }`
- `terminal_input`: raw terminal input string
- `resize`: `{ cols, rows }`
- `disconnect`: cleanup signal

Server events use the same envelope and are `ssh_connected`, `terminal_output`,
`ssh_error`, and `ssh_disconnected`. One WebSocket owns at most one SSH session;
closing the WebSocket cancels an in-progress connection and closes the session.

The terminal page reads Secret Manager metadata from `GET /keys` and reveals a
selected password through `POST /keys/reveal`. Secret Manager is a password
source, not a separate SSH authentication method. The authenticated browser
performs these same-origin HTTP calls, and only the revealed value is sent as
the SSH password.

## Saved connections

Authenticated clients can list or upsert connection profiles at:

```text
GET  /ssh/api/connections
POST /ssh/api/connections
```

Profiles are persisted as one readable Param group per profile, rather than a
JSON blob. For a profile named `host1`, the entries are:

```text
connection.host1.host = "localhost"
connection.host1.port = "22"
connection.host1.username = "root"
connection.host1.auth = "{password, local-password}"
```

`auth` uses `{password}` or `{password, secret-name}` for password
authentication, and `{key}` or `{key, /path/to/private-key}` for key
authentication. The second value is always a Secret Manager reference or a
key path—never a password, private-key value, or passphrase.

## Build

Run:

```sh
make ssh
```

This builds the Svelte UI and `dist/ssh-plugin`, then packages `plugins/ssh.plg`
with the binary and `ui/build`.

For local debugging, put `plugins/ssh.plg` in the panel's configured
`ServiceDir`, enable `Services.ssh`, start the panel, and open
`/ssh/pages/index.html` after logging in.

## Example config

```toml
[Services.ssh]
Restart = "always"
RunAsUser = ""

[Services.ssh.Params]
ssh_config_path = "~/.ssh/config"
"connection.host1.host" = "localhost"
"connection.host1.port" = "22"
"connection.host1.username" = "root"
"connection.host1.auth" = "{key, ~/.ssh/id_ed25519}"
```
