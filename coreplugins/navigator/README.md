# Navigator

Navigate across plugin pages

Navigator is a WASM service with a Svelte frontend. Its backend builds the
navigation list from the running service records published in the read-only
`sys` KV namespace; it does not call the kernel service-management HTTP API.

## Entry route contract

A running service appears in navigation when it exposes exactly one HTTP route
with:

- route ID `entry`;
- method `GET`;
- a local URL path as its pattern; and
- an access policy that permits the current authenticated user.

Only the route's access policy is evaluated by Navigator. The kernel remains
responsible for enforcing access again when the entry URL is requested.

```go
arupa.Route{
    ID:          "entry",
    TransportID: "pages",
    HTTP: &arupa.HTTPRoute{
        Method:  http.MethodGet,
        Pattern: "/example/",
        Access:  arupa.AccessPolicy{RequireAuth: true},
    },
}
```

Display name and icon metadata are read from the matching service catalog
record after the running record has supplied a valid, accessible entry route.

## Example config

```toml
  [Services.navigator]
    Restart = "always"
    RunAsUser = ""
    [Services.navigator.Params]
      icon = "/Arupa.svg"
      order = "ssh,service-manager,hello"
      hide = "hello"
      server-friendly-name = "Lab Server"
```

`icon` accepts a local path or an absolute URL, including a cross-origin URL.
`order` and `hide` are comma-separated lists of service names. These values can
also be edited from the Navigator tab in the settings dialog; changes are
persisted through the service Params API and take effect immediately.

`server-friendly-name` controls the name shown in the mobile Navigator header
and defaults to `Arupa`. `GET /usr/server-friendly-name` exposes the current
name publicly; authenticated users can edit it from the General tab.

The General tab shows the current kernel version and can request a
configuration reload. Reloading configuration is presented as a dangerous
operation because it may restart, stop, or reconfigure running services.
