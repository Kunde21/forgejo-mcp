package main

import (
	"os"
	"strconv"
)

type Config struct {
	Host string
	Port int
}

func LoadConfig() *Config {
	config := &Config{
		Host: getEnv("MCP_HOST", "localhost"),
		Port: getEnvAsInt("MCP_PORT", 3000),
	}
	return config
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
