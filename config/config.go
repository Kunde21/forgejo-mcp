package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	RemoteURL string `mapstructure:"remote_url"`
	AuthToken string `mapstructure:"auth_token"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 3000)
	viper.SetDefault("remote_url", "")
	viper.SetDefault("auth_token", "")

	// Environment variables
	viper.BindEnv("host", "MCP_HOST")
	viper.BindEnv("port", "MCP_PORT")
	viper.BindEnv("remote_url", "FORGEJO_REMOTE_URL")
	viper.BindEnv("auth_token", "FORGEJO_AUTH_TOKEN")

	// Config file support (optional)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Read config file if it exists (ignore errors for tests)
	viper.ReadInConfig() // Ignore errors - config file is optional

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration has all required fields for API operations
func (c *Config) Validate() error {
	if c.RemoteURL == "" {
		return &ValidationError{Field: "RemoteURL", Message: "FORGEJO_REMOTE_URL environment variable or config file remote_url is required"}
	}
	if c.AuthToken == "" {
		return &ValidationError{Field: "AuthToken", Message: "FORGEJO_AUTH_TOKEN environment variable or config file auth_token is required"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
