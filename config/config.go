package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/viper"
)

// Config represents the forgejo-mcp application configuration
type Config struct {
	// Forgejo configuration
	ForgejoURL string `mapstructure:"forgejo_url"`
	AuthToken  string `mapstructure:"auth_token"`
	TeaPath    string `mapstructure:"tea_path"`

	// Gitea SDK client configuration
	ClientTimeout int    `mapstructure:"client_timeout"` // seconds
	UserAgent     string `mapstructure:"user_agent"`

	// Server configuration
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // seconds
	WriteTimeout int    `mapstructure:"write_timeout"` // seconds

	// Logging configuration
	Debug    bool   `mapstructure:"debug"`
	LogLevel string `mapstructure:"log_level"`
}

// Load loads the configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("forgejo_url", "https://example.forgejo.com")
	v.SetDefault("auth_token", "placeholder-token")
	v.SetDefault("tea_path", "tea")
	v.SetDefault("host", "localhost")
	v.SetDefault("port", 8080)
	v.SetDefault("read_timeout", 30)
	v.SetDefault("write_timeout", 30)
	v.SetDefault("client_timeout", 30)
	v.SetDefault("user_agent", "forgejo-mcp-client/1.0.0")
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

// Validate validates the configuration using ozzo-validation
func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ForgejoURL, validation.Required, is.URL),
		validation.Field(&c.AuthToken, validation.Required),
		validation.Field(&c.TeaPath),
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required, validation.Min(1), validation.Max(65535)),
		validation.Field(&c.ReadTimeout, validation.Min(0)),
		validation.Field(&c.WriteTimeout, validation.Min(0)),
		validation.Field(&c.ClientTimeout, validation.Min(1), validation.Max(300)), // 1-300 seconds
		validation.Field(&c.UserAgent, validation.Length(1, 100)),
		validation.Field(&c.LogLevel, validation.In("trace", "debug", "info", "warn", "error", "fatal", "panic")),
	)
}

// ValidateForGiteaClient validates configuration specifically for Gitea SDK client usage
func (c *Config) ValidateForGiteaClient() error {
	// First run standard validation
	if err := c.Validate(); err != nil {
		return err
	}

	// Additional Gitea client specific validations
	if c.ClientTimeout <= 0 {
		return fmt.Errorf("client_timeout must be greater than 0")
	}

	if c.UserAgent == "" {
		return fmt.Errorf("user_agent is required for Gitea client")
	}

	// Validate that ForgejoURL is accessible (basic format check)
	if !strings.HasPrefix(c.ForgejoURL, "http://") && !strings.HasPrefix(c.ForgejoURL, "https://") {
		return fmt.Errorf("forgejo_url must start with http:// or https://")
	}

	return nil
}
