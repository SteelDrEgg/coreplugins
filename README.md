<div align="center">

<h1>Arupa Core Plugins</h1>

</div>

[![Release](https://github.com/SteelDrEgg/coreplugins/actions/workflows/release.yml/badge.svg?branch=main)](https://github.com/SteelDrEgg/coreplugins/actions/workflows/release.yml)

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

This regenerates `pluginsdk/grpc` and `pluginsdk/wasm` from `proto/panel.proto`
before packaging plugins.

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

## Batch Releases

GitHub Actions publishes full plugin batches as GitHub Releases. Release tags
keep the `dist-YYYY-MM-DD-HHMMSS` format; each plugin keeps its own version in
`info.yaml`.

Latest plugin packages are available from stable URLs:

```text
https://github.com/<owner>/<repo>/releases/latest/download/ssh.plg
https://github.com/<owner>/<repo>/releases/latest/download/login.plg
https://github.com/<owner>/<repo>/releases/latest/download/web-assets.plg
```

## Conventions

All plugin have their own namespace, and should not touch anything outside the namespace.
An SSH plugin may have the following assets, but they must all stay under namespace `/ssh`.  

```text
/ssh/pages/terimnal.html
/ssh/icon/terminal.svg
/ssh/js/terminal.js
```

Though there is nothing stopping you from doing the opposite, but it prevents collision.
