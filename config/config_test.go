package config

import (
	"os"
	"testing"
)

func TestLoadConfig_WithNewFields(t *testing.T) {
	// Set up environment variables
	os.Setenv("FORGEJO_REMOTE_URL", "https://forgejo.example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token-123")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.RemoteURL != "https://forgejo.example.com" {
		t.Errorf("Expected RemoteURL to be 'https://forgejo.example.com', got '%s'", config.RemoteURL)
	}

	if config.AuthToken != "test-token-123" {
		t.Errorf("Expected AuthToken to be 'test-token-123', got '%s'", config.AuthToken)
	}
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("FORGEJO_REMOTE_URL")
	os.Unsetenv("FORGEJO_AUTH_TOKEN")

	config, err := Load()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Assuming defaults: RemoteURL empty or some default, AuthToken empty
	if config.RemoteURL != "" {
		t.Errorf("Expected default RemoteURL to be empty, got '%s'", config.RemoteURL)
	}

	if config.AuthToken != "" {
		t.Errorf("Expected default AuthToken to be empty, got '%s'", config.AuthToken)
	}
}

func TestLoadConfig_Validation(t *testing.T) {
	// Test missing required fields
	os.Unsetenv("FORGEJO_REMOTE_URL")
	os.Setenv("FORGEJO_AUTH_TOKEN", "token")
	defer os.Unsetenv("FORGEJO_AUTH_TOKEN")

	config, err := Load()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// This should fail validation if RemoteURL is required
	err = config.Validate()
	if err == nil {
		t.Error("Expected validation error for missing RemoteURL")
	}
}
