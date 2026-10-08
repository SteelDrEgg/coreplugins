# Secret Manager

A web interface for storing and managing encrypted secrets in
[Arupa](https://github.com/SteelDrEgg/Arupa). The `secret-manager` WASM service lets
you add, edit, reveal, and delete secrets, and choose which services can access
them. Values are protected by the manager's identity or a passphrase.

Start the service, sign in, and open `/secret-manager/pages/`.

## HTTP Endpoints

Use these methods and paths when configuring HTTP access control. All endpoints
require authentication.

| Endpoint | Method | Description | Require auth |
|----------|--------|-------------|--------------|
| `/secret-manager/pages/` | `GET` | Management page and bundled assets | yes |
| `/secret-manager/icon/` | `GET` | Navigation icons | yes |
| `/secret-manager/secrets` | `GET` | List secret metadata | yes |
| `/secret-manager/secrets` | `POST` | Create a secret | yes |
| `/secret-manager/secrets/{name}` | `PATCH` | Edit a secret and its permissions | yes |
| `/secret-manager/secrets/{name}` | `DELETE` | Delete a secret | yes |
| `/secret-manager/secrets/{name}/reveal` | `POST` | Reveal a secret value | yes |

`{name}` is a placeholder. For access rules covering all individual secrets, use
the `/secret-manager/secrets/` path prefix with the desired HTTP method.

## Config

No manual Params settings are required for a new installation. Manage secrets
and their allowed service lists through the page. The service generates and saves
its encryption identity on first start.

To supply your own age X25519 identity, configure it directly or through an
environment reference:

```toml
[Services.secret-manager.Params]
"secretmgr.identity" = "env://SECRET_KEY"
```

Preserve the configured or generated identity when backing up or moving an
installation. Passphrases are entered through the page and are never stored.

For integration, storage details, upgrades, and build instructions, see the
[development guide](DEVELOPMENT.md).
