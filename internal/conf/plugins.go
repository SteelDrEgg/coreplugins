package conf

import "strings"

const pluginsDefaultKey = "default"

// PluginAutoStart reports whether a plugin should be auto-started.
//
// Per-plugin policy under [Plugins.<name>] overrides [Plugins.default].
// If neither exists, it falls back to auto-start enabled for compatibility.
func (c Config) PluginAutoStart(name string) bool {
	policy, ok := c.Plugins[name]
	if ok && strings.TrimSpace(policy.Restart) != "" {
		return parseRestart(policy.Restart)
	}

	def, ok := c.Plugins[pluginsDefaultKey]
	if ok && strings.TrimSpace(def.Restart) != "" {
		return parseRestart(def.Restart)
	}

	return true
}

// PluginParams returns the config params a plugin should receive.
//
// Params from [Plugins.default.params] are used as a base, then overridden by
// per-plugin [Plugins.<name>.params] entries.
func (c Config) PluginParams(name string) map[string]string {
	out := make(map[string]string)
	if def, ok := c.Plugins[pluginsDefaultKey]; ok {
		for k, v := range def.Params {
			out[k] = v
		}
	}
	if policy, ok := c.Plugins[name]; ok {
		for k, v := range policy.Params {
			out[k] = v
		}
	}
	return out
}

// PluginRunAsUser returns the OS user a gRPC plugin should run as.
//
// Per-plugin policy under [Plugins.<name>] overrides [Plugins.default].
// An empty result means the current minimalpanel process user.
func (c Config) PluginRunAsUser(name string) string {
	if policy, ok := c.Plugins[name]; ok {
		if user := strings.TrimSpace(policy.RunAsUser); user != "" {
			return user
		}
	}

	if def, ok := c.Plugins[pluginsDefaultKey]; ok {
		return strings.TrimSpace(def.RunAsUser)
	}

	return ""
}

func parseRestart(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "no", "false", "off", "disable", "disabled", "never", "0":
		return false
	default:
		return true
	}
}
