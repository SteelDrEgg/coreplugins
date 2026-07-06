<div align="center"><h1>MinimalPanel Core Plugins</h1></div>

Core plugin repository for MinimalPanel.

This repository owns the plugin contract, generated plugin SDKs, and the
official core plugins that MinimalPanel can install as `.plg` packages. Each
plugin keeps its own runtime metadata, source code, pages, and assets inside
`coreplugins/<plugin>/`.

## Layout

- `proto/panel.proto` is the host/plugin contract.
- `pluginsdk/grpc` and `pluginsdk/wasm` are generated SDKs used by plugin code.
- `coreplugins/<plugin>/info.yaml` is the runtime metadata loaded by the host.
- `Makefile` builds each plugin and writes a `.plg` package.

## Build

Build every core plugin:

```sh
make plugins
```

Build one plugin:

```sh
make ssh
make login
make web-assets
```

Packages are written to `plugins/*.plg`. Temporary build output is written to
`dist/`.

## Regenerate SDKs

Only needed after changing `proto/panel.proto`:

```sh
make proto
```

`make proto` installs protobuf generators into the local `.bin/` directory.
