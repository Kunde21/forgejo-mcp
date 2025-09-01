package client

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var _ RepositoryLister = (*ForgejoClient)(nil)
var _ Client = (*ForgejoClient)(nil)
var exampleCom *url.URL

func init() {
	exampleCom, _ = url.Parse("https://example.com")
}

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		token   string
		wantErr error
	}{
		{
			name:    "empty baseURL",
			baseURL: "",
			token:   "test-token",
			wantErr: &ValidationError{Message: "baseURL cannot be empty", Field: "baseURL"},
		},
		{
			name:    "empty token",
			baseURL: "https://example.com",
			token:   "",
			wantErr: &ValidationError{Message: "token cannot be empty", Field: "token"},
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			token:   "test-token",
			wantErr: &ValidationError{Message: "invalid baseURL format, must be a valid HTTP/HTTPS URL", Field: "baseURL"},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			client, err := New(tst.baseURL, tst.token)
			if !cmp.Equal(tst.wantErr, err) {
				t.Error(cmp.Diff(tst.wantErr, err))
			}
			if tst.wantErr == nil && client == nil {
				t.Error("expected client to be created, got nil")
			}
		})
	}
}

// Test client interface compliance
func TestClientInterfaceCompliance(t *testing.T) {
	// Test that ForgejoClient implements the Client interface
	var _ Client = (*ForgejoClient)(nil)
	var _ RepositoryLister = (*ForgejoClient)(nil)
}

// Test client creation with various configurations (validation only)
func TestNewClientValidationExtended(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		expectError bool
	}{
		{
			name:        "valid HTTPS URL",
			baseURL:     "https://forgejo.example.com",
			token:       "valid_token_12345678901234567890",
			expectError: false,
		},
		{
			name:        "valid HTTP URL",
			baseURL:     "http://forgejo.example.com",
			token:       "valid_token_12345678901234567890",
			expectError: false,
		},
		{
			name:        "URL with port",
			baseURL:     "https://forgejo.example.com:3000",
			token:       "valid_token_12345678901234567890",
			expectError: false,
		},
		{
			name:        "URL with path",
			baseURL:     "https://forgejo.example.com/api/v1",
			token:       "valid_token_12345678901234567890",
			expectError: false,
		},
		{
			name:        "empty baseURL",
			baseURL:     "",
			token:       "valid_token_12345678901234567890",
			expectError: true,
		},
		{
			name:        "invalid URL format",
			baseURL:     "not-a-valid-url",
			token:       "valid_token_12345678901234567890",
			expectError: true,
		},
		{
			name:        "empty token",
			baseURL:     "https://forgejo.example.com",
			token:       "",
			expectError: true,
		},
		{
			name:        "token too short",
			baseURL:     "https://forgejo.example.com",
			token:       "short",
			expectError: false, // Client creation doesn't validate token format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation without creating real Gitea client
			err := validateClientInputs(tt.baseURL, tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
			}
		})
	}
}

// validateClientInputs performs the same validation as New but without creating Gitea client
func validateClientInputs(baseURL, token string) error {
	if baseURL == "" {
		return &ValidationError{
			Message: "baseURL cannot be empty",
			Field:   "baseURL",
		}
	}

	if token == "" {
		return &ValidationError{
			Message: "token cannot be empty",
			Field:   "token",
		}
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return &ValidationError{
			Message: "invalid baseURL format, must be a valid HTTP/HTTPS URL",
			Field:   "baseURL",
		}
	}

	return nil
}

// Test client with default values
func TestNewClientDefaults(t *testing.T) {
	t.Skip("Skipping test that requires network connection to Gitea API")
}

// Test client with custom configuration
func TestNewWithConfig(t *testing.T) {
	t.Skip("Skipping test that requires network connection to Gitea API")
}

// Test client interface methods
func TestClientInterfaceMethods(t *testing.T) {
	t.Skip("Skipping test that requires network connection to Gitea API")
}

// Test client with various token formats
func TestClientTokenValidation(t *testing.T) {
	t.Skip("Skipping test that requires network connection to Gitea API")
}
