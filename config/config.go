package config

// Config represents the forgejo-mcp application configuration
type Config struct {
	ForgejoURL string `mapstructure:"forgejo_url"`
	AuthToken  string `mapstructure:"auth_token"`
	TeaPath    string `mapstructure:"tea_path"`
	Debug      bool   `mapstructure:"debug"`
	LogLevel   string `mapstructure:"log_level"`
}

// Load loads the configuration from environment variables and config files
func Load() (*Config, error) {
	// Placeholder implementation - will be expanded later
	return &Config{
		ForgejoURL: "https://example.forgejo.com",
		AuthToken:  "placeholder-token",
		TeaPath:    "tea",
		Debug:      false,
		LogLevel:   "info",
	}, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Placeholder implementation - will be expanded later
	if c.ForgejoURL == "" {
		return &ValidationError{"forgejo_url is required"}
	}
	if c.AuthToken == "" {
		return &ValidationError{"auth_token is required"}
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
