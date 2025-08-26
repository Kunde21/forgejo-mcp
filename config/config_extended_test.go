package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadFromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("FORGEJO_MCP_FORGEJO_URL", "https://test.forgejo.com")
	os.Setenv("FORGEJO_MCP_AUTH_TOKEN", "test-token")
	os.Setenv("FORGEJO_MCP_TEA_PATH", "/custom/tea/path")
	os.Setenv("FORGEJO_MCP_DEBUG", "true")
	os.Setenv("FORGEJO_MCP_LOG_LEVEL", "debug")

	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("FORGEJO_MCP_FORGEJO_URL")
		os.Unsetenv("FORGEJO_MCP_AUTH_TOKEN")
		os.Unsetenv("FORGEJO_MCP_TEA_PATH")
		os.Unsetenv("FORGEJO_MCP_DEBUG")
		os.Unsetenv("FORGEJO_MCP_LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ForgejoURL != "https://test.forgejo.com" {
		t.Errorf("Expected ForgejoURL to be 'https://test.forgejo.com', got '%s'", cfg.ForgejoURL)
	}

	if cfg.AuthToken != "test-token" {
		t.Errorf("Expected AuthToken to be 'test-token', got '%s'", cfg.AuthToken)
	}

	if cfg.TeaPath != "/custom/tea/path" {
		t.Errorf("Expected TeaPath to be '/custom/tea/path', got '%s'", cfg.TeaPath)
	}

	if cfg.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", cfg.Debug)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestConfigLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `forgejo_url: "https://file-test.forgejo.com"
auth_token: "file-test-token"
tea_path: "/file/test/tea"
debug: true
log_level: "debug"`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to the temporary directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Restore original directory after test
	defer func() {
		os.Chdir(oldDir)
	}()

	// Test loading config from file
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ForgejoURL != "https://file-test.forgejo.com" {
		t.Errorf("Expected ForgejoURL to be 'https://file-test.forgejo.com', got '%s'", cfg.ForgejoURL)
	}

	if cfg.AuthToken != "file-test-token" {
		t.Errorf("Expected AuthToken to be 'file-test-token', got '%s'", cfg.AuthToken)
	}

	if cfg.TeaPath != "/file/test/tea" {
		t.Errorf("Expected TeaPath to be '/file/test/tea', got '%s'", cfg.TeaPath)
	}

	if cfg.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", cfg.Debug)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				ForgejoURL: "https://forgejo.com",
				AuthToken:  "token123",
				TeaPath:    "tea",
				Debug:      false,
				LogLevel:   "info",
			},
			expectError: false,
		},
		{
			name: "missing forgejo url",
			config: &Config{
				ForgejoURL: "",
				AuthToken:  "token123",
				TeaPath:    "tea",
				Debug:      false,
				LogLevel:   "info",
			},
			expectError: true,
		},
		{
			name: "missing auth token",
			config: &Config{
				ForgejoURL: "https://forgejo.com",
				AuthToken:  "",
				TeaPath:    "tea",
				Debug:      false,
				LogLevel:   "info",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
