package secrets

// CallerKind distinguishes who is invoking a secret-manager operation.
type CallerKind int

const (
	// CallerHumanAdmin is an authenticated human via the HTTP admin UI. The
	// host already gates these routes with AccessPolicy.RequireAuth, so the
	// admin manages each secret's allow-list directly rather than being
	// subject to it.
	CallerHumanAdmin CallerKind = iota
	// CallerPlugin is another registered service calling in via a service
	// message. Its identity comes from the host-authenticated message
	// Source, never from caller-supplied payload data.
	CallerPlugin
)

// Caller identifies the actor performing a read/write/delete. It is the
// single place that decides whether a per-secret ACL check or an
// auto-grant-on-create applies, replacing what used to be two independently
// hand-coded transport paths that happened to differ.
type Caller struct {
	Kind CallerKind
	// Plugin is the calling service name; only meaningful when
	// Kind == CallerPlugin.
	Plugin string
}

// HumanAdminCaller identifies the authenticated human admin using the HTTP API.
func HumanAdminCaller() Caller { return Caller{Kind: CallerHumanAdmin} }

// PluginCaller identifies another service calling in by service message.
// source must already be host-authenticated (e.g. IncomingServiceMessage.Source).
func PluginCaller(source string) Caller { return Caller{Kind: CallerPlugin, Plugin: source} }

// requiresACL reports whether this caller must pass the secret's per-caller
// allow-list to read, update, or delete it.
func (c Caller) requiresACL() bool { return c.Kind == CallerPlugin }

// autoGrantOnCreate reports whether creating a secret should implicitly add
// this caller to its allow-list.
func (c Caller) autoGrantOnCreate() bool { return c.Kind == CallerPlugin }
