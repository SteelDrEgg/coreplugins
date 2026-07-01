package conf

type Config struct {
	SSHConfigPath string
	Listen        string
	Auth
	Web
	Plugin
	Plugins map[string]PluginPolicy
}

type Auth struct {
	Users map[string]string
}

type Web struct {
	RootPath string
}

// Plugin holds plugin-system configuration.
type Plugin struct {
	// PluginDir is the directory scanned for *.plg plugin packages.
	PluginDir string
	// PluginTempDir is where plugin packages are extracted at load time.
	PluginTempDir string
}

// PluginPolicy controls plugin runtime behavior from [Plugins.<name>].
type PluginPolicy struct {
	// Restart controls auto-start behavior at host startup.
	// Typical values: "always", "yes", "true", "on", "no", "false", "off".
	Restart string
	// RunAsUser controls the OS user used to start gRPC plugin processes.
	// Empty means the plugin runs as the current minimalpanel process user.
	RunAsUser string
	// Params are arbitrary string config values passed directly to the plugin
	// at registration, from [Plugins.<name>.params].
	Params map[string]string
}
