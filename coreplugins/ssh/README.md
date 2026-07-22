# SSH Terminal Plugin

This core plugin provides the `/pages/terminal.html` SSH terminal page and the
`/ssh` Socket.IO namespace. It is implemented as a gRPC go-plugin process and
uses the Arupa Go SDK for the gRPC runtime, registration, host callbacks, HTTP
adaptation, Socket.IO event dispatch, and background emits. `internal/sshc`
handles SSH config parsing, authentication, connection setup, and PTY creation.

## Layout

- `main.go` hands the SDK plugin to the SDK-managed gRPC runtime.
- `server.go` composes the SDK plugin, its registration hook, HTTP handler, and
  EventBus.
- `connect.go` resolves host config and opens SSH sessions.
- `connections.go` is a standard `net/http` handler that validates and persists
  non-sensitive connection profiles.
- `session.go` owns PTY input, resize, output, and cleanup.
- `payload.go` contains small JSON/path helpers.

## Frontend Contract

The terminal page connects to namespace `/ssh` and emits:

- `connect_ssh`: `{ host, port, username, password?, privateKey?, passphrase? }`
- `terminal_input`: raw terminal input string
- `resize`: `{ cols, rows }`
- `disconnect`: cleanup signal

The plugin emits `ssh_connected`, `terminal_output`, `ssh_error`, and
`ssh_disconnected` back to the calling socket.

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

This builds `dist/ssh-plugin` and packages `plugins/ssh.plg` with the binary,
`pages/terminal.html`, and vendored xterm/socket.io assets under
`assets/terminal`.

For local debugging, start the panel with `go run ./cmd` and open
`/pages/terminal.html` after logging in.

## Example config

```toml
  [Plugins.ssh]
    Restart = "always"
    RunAsUser = ""
    [Plugins.ssh.Params]
      ssh_config_path = "~/.ssh/config"
      "connection.host1.host" = "localhost"
      "connection.host1.port" = "22"
      "connection.host1.username" = "root"
      "connection.host1.auth" = "{key, ~/.ssh/id_ed25519}"
```
