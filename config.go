package main

import (
	"os"
	"strconv"
)

type Config struct {
	Host      string
	Port      int
	RemoteURL string
	AuthToken string
}

func LoadConfig() *Config {
	config := &Config{
		Host:      getEnv("MCP_HOST", "localhost"),
		Port:      getEnvAsInt("MCP_PORT", 3000),
		RemoteURL: getEnv("FORGEJO_REMOTE_URL", ""),
		AuthToken: getEnv("FORGEJO_AUTH_TOKEN", ""),
	}
	return config
}

// Validate checks if the configuration has all required fields for API operations
func (c *Config) Validate() error {
	if c.RemoteURL == "" {
		return &ValidationError{Field: "RemoteURL", Message: "FORGEJO_REMOTE_URL environment variable is required"}
	}
	if c.AuthToken == "" {
		return &ValidationError{Field: "AuthToken", Message: "FORGEJO_AUTH_TOKEN environment variable is required"}
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
