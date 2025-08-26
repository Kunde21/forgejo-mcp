package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

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
	v := viper.New()

	// Set default values
	v.SetDefault("forgejo_url", "https://example.forgejo.com")
	v.SetDefault("auth_token", "placeholder-token")
	v.SetDefault("tea_path", "tea")
	v.SetDefault("debug", false)
	v.SetDefault("log_level", "info")

	// Set environment variables
	v.SetEnvPrefix("FORGEJO_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set config file settings
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	v.AddConfigPath(".")
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".forgejo-mcp"))
	v.AddConfigPath("/etc/forgejo-mcp")

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
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
