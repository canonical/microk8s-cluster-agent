package internal

// TraefikConfiguration is traefik-compatible configuration.
type TraefikConfiguration struct {
	EntryPoints ExpectedEntryPoints `yaml:"entryPoints"`
	Providers   ExpectedProviders   `yaml:"providers"`
}

// ExpectedEntryPoints is traefik-compatible configuration.
type ExpectedEntryPoints struct {
	APIServer EntryPoint `yaml:"apiserver"`
}

// EntryPoint is traefik-compatible configuration.
type EntryPoint struct {
	Address string `yaml:"address"`
}

// ExpectedProviders is traefik-compatible configuration.
type ExpectedProviders struct {
	File FileProvider `yaml:"file"`
}

// FileProvider is traefik-compatible configuration.
type FileProvider struct {
	Filename string `yaml:"filename"`
	Watch    bool   `yaml:"watch"`
}
