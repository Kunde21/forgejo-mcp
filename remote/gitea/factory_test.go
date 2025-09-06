package gitea

import (
	"testing"
)

// mockConfig holds test configuration for client factory
type mockConfig struct {
	url       string
	token     string
	expectErr bool
}

func TestNewGiteaClient(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		token       string
		expectError bool
	}{
		{
			name:        "empty URL",
			url:         "",
			token:       "valid-token-123",
			expectError: true,
		},
		{
			name:        "empty token",
			url:         "https://forgejo.example.com",
			token:       "",
			expectError: true,
		},
		{
			name:        "invalid URL format",
			url:         "://invalid-url",
			token:       "valid-token-123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGiteaClient(tt.url, tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if client != nil {
					t.Errorf("Expected nil client on error, got %v", client)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Errorf("Expected valid client but got nil")
				}
				// Verify the client implements the interface
				var _ GiteaClientInterface = client
			}
		})
	}
}

func TestNewGiteaClientFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *ClientConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "empty base URL",
			config: &ClientConfig{
				BaseURL: "",
				Token:   "valid-token-123",
			},
			expectError: true,
		},
		{
			name: "empty token",
			config: &ClientConfig{
				BaseURL: "https://forgejo.example.com",
				Token:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGiteaClientFromConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if client != nil {
					t.Errorf("Expected nil client on error, got %v", client)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Errorf("Expected valid client but got nil")
				}
				// Verify the client implements the interface
				var _ GiteaClientInterface = client
			}
		})
	}
}

func TestClientConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ClientConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &ClientConfig{
				BaseURL: "https://forgejo.example.com",
				Token:   "valid-token-123",
			},
			expectError: false,
		},
		{
			name: "empty base URL",
			config: &ClientConfig{
				BaseURL: "",
				Token:   "valid-token-123",
			},
			expectError: true,
		},
		{
			name: "empty token",
			config: &ClientConfig{
				BaseURL: "https://forgejo.example.com",
				Token:   "",
			},
			expectError: true,
		},
		{
			name: "invalid URL format",
			config: &ClientConfig{
				BaseURL: "://invalid-url",
				Token:   "valid-token-123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
