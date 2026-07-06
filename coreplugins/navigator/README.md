# Navigator

Navigate across plugin pages

## Example config

```toml
  [Plugins.navigator]
    Restart = "always"
    RunAsUser = ""
    [Plugins.navigator.Params]
      order = "ssh,plugin-manager,hello"
      ignore = "login,web-assets,navigator"
```