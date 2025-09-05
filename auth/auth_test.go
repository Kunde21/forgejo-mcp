package auth

import (
	"testing"

	"github.com/Kunde21/forgejo-mcp/config"
)

func TestConfigAuthentication(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
	}{
		{
			name: "valid authentication config",
			cfg: &config.Config{
				ForgejoURL: "https://forgejo.example.com",
				AuthToken:  "valid-token-123",
				Host:       "localhost", // Required for config validation
				Port:       3000,        // Required for config validation
			},
			expectError: false,
		},
		{
			name: "missing auth token",
			cfg: &config.Config{
				ForgejoURL: "https://forgejo.example.com",
				AuthToken:  "",
			},
			expectError: true,
		},
		{
			name: "missing forgejo URL",
			cfg: &config.Config{
				ForgejoURL: "",
				AuthToken:  "valid-token-123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for %s, got: %v", tt.name, err)
				}
			}
		})
	}
}

func TestAuthenticationFlow(t *testing.T) {
	// Test that authentication configuration is properly loaded
	cfg := &config.Config{
		ForgejoURL: "https://test.forgejo.com",
		AuthToken:  "test-auth-token",
		Host:       "localhost", // Required for config validation
		Port:       3000,        // Required for config validation
	}

	// Validate configuration
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// Verify auth token is accessible
	if cfg.AuthToken == "" {
		t.Error("Auth token should not be empty")
	}

	if cfg.AuthToken != "test-auth-token" {
		t.Errorf("Expected auth token 'test-auth-token', got '%s'", cfg.AuthToken)
	}
}
