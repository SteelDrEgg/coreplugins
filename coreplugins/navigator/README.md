# Navigator

A home page for navigating between [Arupa](https://github.com/SteelDrEgg/Arupa)
services.

Browse accessible service pages and open them from the navigation menu. Customize
the icon, service order, hidden entries, and server name from the settings dialog.
The General tab also shows the kernel version and offers configuration reload.
Reloading configuration may restart, stop, or reconfigure running services.

## Usage

Start the `navigator` service, sign in, and open `/`.

A running service appears in the menu when it provides exactly one `entry` HTTP
route using `GET`, with a local URL path and access allowed for the current user.

## Info

```text
Name: navigator
Type: wasm
```

## HTTP Endpoints

| Endpoint | Method | Description | Require auth |
|----------|--------|-------------|--------------|
| `/` | `GET` | Entry and bundled page assets | yes |
| `/navigator/api/entries` | `GET` | List accessible navigation entries | yes |
| `/navigator/api/config` | `GET`, `PUT` | Read and update navigation settings | yes |
| `/usr/server-friendly-name` | `GET` | Read the server display name | no |
| `/usr/server-friendly-name` | `PUT` | Update the server display name | yes |

## Config

```toml
[Services.navigator.Params]
icon = "/Arupa.svg"
order = "ssh,service-manager,hello"
hide = "hello"
server-friendly-name = "Lab Server"
```

`icon` accepts a local path or an absolute URL and defaults to `/Arupa.svg`.
`order` and `hide` are comma-separated lists of service names. Hiding an entry
only removes it from the menu; it does not change access permissions.
`server-friendly-name` defaults to `Arupa` and is shown in the mobile header
and login pages.

Settings saved through the dialog take effect immediately.

## Build

```sh
make build
```

Run from this directory. `make build` produces `dist/navigator.wasm` in the
repository root and static files in `ui/build/`. `make package` creates
`plugins/navigator.plg` in the repository root.
Rescan the service directory and restart the service to load a new build.
