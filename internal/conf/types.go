package conf

type Config struct {
	SSHConfigPath string
	Listen        string
	Auth
	Web
	Plugin
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
