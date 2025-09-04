// Package server provides tests for authentication edge cases and error scenarios
package server

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/auth"
	"github.com/Kunde21/forgejo-mcp/config"
)

// TestAuthEdgeCases_TokenFormatValidation tests various token format edge cases
func TestAuthEdgeCases_TokenFormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
		errorField  string
	}{
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			errorField:  "token",
		},
		{
			name:        "token too short",
			token:       "short",
			expectError: true,
			errorField:  "token",
		},
		{
			name:        "token with invalid characters",
			token:       "invalid@token#123",
			expectError: true,
			errorField:  "token",
		},
		{
			name:        "token with spaces",
			token:       "token with spaces",
			expectError: true,
			errorField:  "token",
		},
		{
			name:        "valid token minimum length",
			token:       strings.Repeat("a", 20),
			expectError: false,
		},
		{
			name:        "valid token with special characters",
			token:       "valid_testing-auth-token-123.456_789",
			expectError: false,
		},
		{
			name:        "very long valid token",
			token:       strings.Repeat("a", 200),
			expectError: false,
		},
		{
			name:        "token with only numbers",
			token:       "12345678901234567890",
			expectError: false,
		},
		{
			name:        "token with mixed case",
			token:       "AbCdEfGhIjKlMnOpQrSt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateTokenFormat(tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				var tokenErr *auth.TokenValidationError
				if !errors.As(err, &tokenErr) {
					t.Errorf("expected TokenValidationError, got %T", err)
					return
				}

				if tokenErr.Field != tt.errorField {
					t.Errorf("expected error field %q, got %q", tt.errorField, tokenErr.Field)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAuthEdgeCases_TokenMasking tests token masking edge cases
func TestAuthEdgeCases_TokenMasking(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "empty token",
			token:    "",
			expected: "",
		},
		{
			name:     "very short token",
			token:    "ab",
			expected: "**",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "token length 4",
			token:    "abcd",
			expected: "****",
		},
		{
			name:     "token length 5",
			token:    "abcde",
			expected: "ab*de",
		},
		{
			name:     "token length 8",
			token:    "abcdefgh",
			expected: "ab****gh",
		},
		{
			name:     "token length 9",
			token:    "abcdefghi",
			expected: "abcd*fghi",
		},
		{
			name:     "normal token",
			token:    "abcdefghij",
			expected: "abcd**ghij",
		},
		{
			name:     "long token",
			token:    "abcdefghijklmnopqrstuvwxyz0123456789",
			expected: "abcd****************************6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.MaskToken(tt.token)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestAuthEdgeCases_ErrorWrapping tests error wrapping with context
func TestAuthEdgeCases_ErrorWrapping(t *testing.T) {
	tests := []struct {
		name           string
		originalError  error
		context        string
		operation      string
		token          string
		expectContains []string
	}{
		{
			name:           "wrap token validation error",
			originalError:  &auth.TokenValidationError{Message: "invalid format", Field: "token"},
			context:        "tool execution",
			operation:      "authentication",
			token:          "secret-testing-auth-token-123",
			expectContains: []string{"authentication failed during tool execution", "invalid format"},
		},
		{
			name:           "wrap network error",
			originalError:  auth.NewAuthNetworkError("connection failed", "https://example.com"),
			context:        "API call",
			operation:      "validation",
			token:          "masked-token-456",
			expectContains: []string{"validation failed during API call", "connection failed"},
		},
		{
			name:           "wrap timeout error",
			originalError:  auth.NewAuthTimeoutError(5 * time.Second),
			context:        "server request",
			operation:      "token check",
			token:          "timeout-token-789",
			expectContains: []string{"token check failed during server request", "timed out after 5s"},
		},
		{
			name:           "wrap server error",
			originalError:  auth.NewAuthServerError(500, "Internal Server Error"),
			context:        "validation",
			operation:      "API request",
			token:          "server-error-token",
			expectContains: []string{"API request failed during validation", "500 - Internal Server Error"},
		},
		{
			name:           "wrap with empty token",
			originalError:  errors.New("generic error"),
			context:        "test",
			operation:      "operation",
			token:          "",
			expectContains: []string{"operation failed during test", "generic error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := auth.WrapErrorWithContext(tt.originalError, tt.context, tt.operation, tt.token)

			if wrappedErr == nil {
				t.Error("expected wrapped error, got nil")
				return
			}

			errorMsg := wrappedErr.Error()

			// Check that all expected strings are contained in the error message
			for _, expected := range tt.expectContains {
				if !strings.Contains(errorMsg, expected) {
					t.Errorf("expected error message to contain %q, got %q", expected, errorMsg)
				}
			}

			// Ensure token is not exposed in the error message
			if tt.token != "" && strings.Contains(errorMsg, tt.token) {
				t.Errorf("error message should not contain the original token %q", tt.token)
			}
		})
	}
}

// TestAuthEdgeCases_CacheKeyGeneration tests cache key generation edge cases
func TestAuthEdgeCases_CacheKeyGeneration(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		token    string
		expected string
	}{
		{
			name:     "normal case",
			baseURL:  "https://example.com",
			token:    "normal-testing-auth-token-123",
			expected: "https://example.com:norm*********************-123",
		},
		{
			name:     "empty baseURL",
			baseURL:  "",
			token:    "testing-auth-token-123",
			expected: ":test**************-123",
		},
		{
			name:     "empty token",
			baseURL:  "https://example.com",
			token:    "",
			expected: "https://example.com:",
		},
		{
			name:     "both empty",
			baseURL:  "",
			token:    "",
			expected: ":",
		},
		{
			name:     "very long baseURL",
			baseURL:  "https://very-long-subdomain.example.com/api/v1/forgejo",
			token:    "testing-auth-token-123",
			expected: "https://very-long-subdomain.example.com/api/v1/forgejo:test**************-123",
		},
		{
			name:     "baseURL with special characters",
			baseURL:  "https://example.com:8080/path?query=value",
			token:    "testing-auth-token-123",
			expected: "https://example.com:8080/path?query=value:test**************-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.CacheKey(tt.baseURL, tt.token)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestAuthEdgeCases_ServerConfigurationEdgeCases tests server configuration edge cases
func TestAuthEdgeCases_ServerConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "config with very long URLs",
			config: &config.Config{
				ForgejoURL:   "https://very-long-subdomain.very-long-domain.com/api/v1/forgejo/with/very/long/path",
				AuthToken:    "testing-auth-token-123",
				TeaPath:      "/very/long/path/to/tea/binary",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  300,
				WriteTimeout: 300,
				LogLevel:     "debug",
			},
			expectError: false,
		},
		{
			name: "config with minimal timeouts",
			config: &config.Config{
				ForgejoURL:   "https://example.com",
				AuthToken:    "testing-auth-token-123",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  1,
				WriteTimeout: 1,
				LogLevel:     "error",
			},
			expectError: false,
		},
		{
			name: "config with maximum timeouts",
			config: &config.Config{
				ForgejoURL:   "https://example.com",
				AuthToken:    "testing-auth-token-123",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  3600,
				WriteTimeout: 3600,
				LogLevel:     "trace",
			},
			expectError: false,
		},
		{
			name: "config with special characters in paths",
			config: &config.Config{
				ForgejoURL:   "https://example.com",
				AuthToken:    "testing-auth-token-123",
				TeaPath:      "/path/with spaces/and-dashes_123",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if server == nil {
				t.Error("expected non-nil server")
				return
			}

			// Verify server configuration is properly set
			if server.config.ForgejoURL != tt.config.ForgejoURL {
				t.Errorf("expected ForgejoURL %q, got %q", tt.config.ForgejoURL, server.config.ForgejoURL)
			}

			if server.config.AuthToken != tt.config.AuthToken {
				t.Errorf("expected AuthToken %q, got %q", tt.config.AuthToken, server.config.AuthToken)
			}
		})
	}
}
