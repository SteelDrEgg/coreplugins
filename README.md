<div align="center"><h1>Arupa Core Plugins</h1></div>

Core plugin repository for Arupa.

This repository owns the plugin contract, generated plugin SDKs, and the
official core plugins that Arupa can install as `.plg` packages. Each
plugin keeps its own runtime metadata, source code, pages, and assets inside
`coreplugins/<plugin>/`.

## Layout

- `proto/panel.proto` is the host/plugin contract.
- `pluginsdk/grpc` and `pluginsdk/wasm` are generated SDKs used by plugin code.
- `coreplugins/<plugin>/info.yaml` is the runtime metadata loaded by the host.
- `coreplugins/<plugin>/Makefile` builds and packages one plugin.
- `Makefile` is the repository-level entrypoint that delegates to each plugin.

## Build

Generate SDK from `proto/panel.proto`:

```sh
make proto
```

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

Or build from the plugin directory:

```sh
make -C coreplugins/ssh package
```

Packages are written to `plugins/*.plg`. Temporary build output is written to
`dist/`. Override `PLUGIN_DIR` to send packages elsewhere:

```sh
make plugins PLUGIN_DIR=../minimalpanel/plugins
```

Each plugin's `RegisterReply.Version` is injected at build time with `ldflags`
from `coreplugins/<plugin>/info.yaml`.

## Conventions

All plugin have their own namespace, and should not touch anything outside the namespace.
An SSH plugin may have the following assets, but they must all stay under namespace `/ssh`.  

```text
/ssh/pages/terimnal.html
/ssh/icon/terminal.svg
/ssh/js/terminal.js
```

Though there is nothing stopping you from doing the opposite, but it prevents collision.
