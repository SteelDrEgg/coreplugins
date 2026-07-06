# SSH Terminal Plugin

This core plugin provides the `/pages/terminal.html` SSH terminal page and the
`/ssh` Socket.IO namespace. It is implemented as a gRPC go-plugin process and
uses `internal/sshc` for SSH config parsing, authentication, connection setup,
and PTY creation.

## Layout

- `main.go` starts the go-plugin gRPC server.
- `plugin.go` adapts the server to HashiCorp go-plugin.
- `server.go` declares static mounts, Socket.IO events, and dispatch.
- `connect.go` resolves host config and opens SSH sessions.
- `session.go` owns PTY input, resize, output, and cleanup.
- `host.go` wraps host callback logging and Socket.IO emits.
- `payload.go` contains small JSON/path helpers.

## Frontend Contract

The terminal page connects to namespace `/ssh` and emits:

- `connect_ssh`: `{ host, port, username, password?, privateKey?, passphrase? }`
- `terminal_input`: raw terminal input string
- `resize`: `{ cols, rows }`
- `disconnect`: cleanup signal

The plugin emits `ssh_connected`, `terminal_output`, `ssh_error`, and
`ssh_disconnected` back to the calling socket.

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
```