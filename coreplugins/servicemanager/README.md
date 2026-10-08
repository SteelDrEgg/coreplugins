# Service Manager

A web interface for managing [Arupa](https://github.com/SteelDrEgg/Arupa) services.

Browse installed and running services, compare versions, search and filter the list,
and start, stop, or restart services. View service details, edit settings and
parameters, manage shared defaults, and change service directories.

Sign in and open `/service-manager/pages/` after starting the service. The service
list refreshes every five seconds while visible. Stop and restart require confirmation.
Saving settings does not start or stop services.

## Usage

This service requires [Service API](https://docs.arupa.dev/docs/arupa/v0/en/kernel/api-service).

```toml
[API]
Service = true
```

## Info

```text
Name: service-manager
Type: static
```

## HTTP Endpoints

| Endpoint | Method | Description | Require auth |
|----------|--------|-------------|--------------|
| `/service-manager/pages/` | `GET` | Entry and bundled page assets | yes |
| `/service-manager/icon/` | `GET` | Navigation icons | yes |

## Config

There are no `[Services.service-manager.Params]` settings.

## Build

```sh
make build
```

Run from this directory. `make build` produces `dist/service-manager.plg` in the
repository root. `make package` also copies it to `plugins/service-manager.plg`.
Rescan the service directory and restart the service to load a new build.
