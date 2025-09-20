package config

import (
	"os"
	"testing"
)

func TestLoadConfig_ClientType(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    string
		expectError bool
	}{
		{
			name:        "gitea client type",
			envValue:    "gitea",
			expected:    "gitea",
			expectError: false,
		},
		{
			name:        "forgejo client type",
			envValue:    "forgejo",
			expected:    "forgejo",
			expectError: false,
		},
		{
			name:        "auto client type",
			envValue:    "auto",
			expected:    "auto",
			expectError: false,
		},
		{
			name:        "empty client type defaults to auto",
			envValue:    "",
			expected:    "auto",
			expectError: false,
		},
		{
			name:        "invalid client type",
			envValue:    "invalid",
			expected:    "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
			os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
			if tt.envValue != "" {
				os.Setenv("FORGEJO_CLIENT_TYPE", tt.envValue)
			} else {
				os.Unsetenv("FORGEJO_CLIENT_TYPE")
			}
			defer func() {
				os.Unsetenv("FORGEJO_REMOTE_URL")
				os.Unsetenv("FORGEJO_AUTH_TOKEN")
				if tt.envValue != "" {
					os.Unsetenv("FORGEJO_CLIENT_TYPE")
				}
			}()

			config, err := Load()
			if err != nil {
				t.Fatalf("LoadConfig failed: %v", err)
			}

			if config.ClientType != tt.expected {
				t.Errorf("Expected ClientType to be '%s', got '%s'", tt.expected, config.ClientType)
			}

			// Test validation
			err = config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestConfig_Validate_ClientType(t *testing.T) {
	tests := []struct {
		name        string
		clientType  string
		expectError bool
	}{
		{
			name:        "valid gitea client type",
			clientType:  "gitea",
			expectError: false,
		},
		{
			name:        "valid forgejo client type",
			clientType:  "forgejo",
			expectError: false,
		},
		{
			name:        "valid auto client type",
			clientType:  "auto",
			expectError: false,
		},
		{
			name:        "invalid client type",
			clientType:  "invalid",
			expectError: true,
		},
		{
			name:        "empty client type is valid",
			clientType:  "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				RemoteURL:  "https://example.com",
				AuthToken:  "token",
				ClientType: tt.clientType,
			}

			err := config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestLoadConfig_WithNewFields(t *testing.T) {
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
	os.Unsetenv("FORGEJO_REMOTE_URL")
	os.Unsetenv("FORGEJO_AUTH_TOKEN")

	config, err := Load()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.RemoteURL != "" {
		t.Errorf("Expected default RemoteURL to be empty, got '%s'", config.RemoteURL)
	}

	if config.AuthToken != "" {
		t.Errorf("Expected default AuthToken to be empty, got '%s'", config.AuthToken)
	}
}

func TestLoadConfig_Validation(t *testing.T) {
	os.Unsetenv("FORGEJO_REMOTE_URL")
	os.Setenv("FORGEJO_AUTH_TOKEN", "token")
	defer os.Unsetenv("FORGEJO_AUTH_TOKEN")

	config, err := Load()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected validation error for missing RemoteURL")
	}
}
