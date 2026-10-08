package secrets

import "strings"

// CallerKind distinguishes who is invoking a secret-manager operation.
type CallerKind int

const (
	CallerUnknown CallerKind = iota
	// CallerHumanAdmin is an authenticated human via the HTTP admin UI. The
	// host already gates these routes with AccessPolicy.RequireAuth, so the
	// admin manages each secret's allow-list directly rather than being
	// subject to it.
	CallerHumanAdmin
	// CallerPlugin is another registered service calling in via a service
	// message. Its identity comes from the host-authenticated message
	// Source, never from caller-supplied payload data.
	CallerPlugin
)

// Caller identifies the actor performing a read/write/delete. It is the
// common permission boundary used by every transport.
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
func PluginCaller(source string) Caller {
	return Caller{Kind: CallerPlugin, Plugin: strings.TrimSpace(source)}
}

func (c Caller) validate() error {
	if c.Kind == CallerHumanAdmin || c.Kind == CallerPlugin && c.Plugin != "" && c.Plugin == strings.TrimSpace(c.Plugin) {
		return nil
	}
	return NewError(ErrForbidden, "invalid caller identity")
}

func (c Caller) allows(record SecretRecord) bool {
	if c.Kind == CallerHumanAdmin {
		return true
	}
	if c.Kind == CallerPlugin && c.Plugin != "" {
		for _, plugin := range record.AllowedPlugins {
			if plugin == c.Plugin {
				return true
			}
		}
	}
	return false
}
