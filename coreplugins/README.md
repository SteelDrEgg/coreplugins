These are the official MinimalPanel core plugins.

Each plugin directory is self-contained:

- `info.yaml` describes the plugin for the host runtime.
- Go source files implement the plugin backend.
- Local `pages/`, `assets/`, or `internal/` directories belong to that plugin.

Use `make plugins` from the repository root to build every `.plg` package.
