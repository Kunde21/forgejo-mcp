package auth

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestGetTokenFromEnv(t *testing.T) {
	// Save original value to restore later
	originalToken := os.Getenv("GITEA_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITEA_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITEA_TOKEN")
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid token present",
			envValue:  "gitea_token_12345678901234567890",
			wantToken: "gitea_token_12345678901234567890",
			wantErr:   false,
		},
		{
			name:        "empty token",
			envValue:    "",
			wantToken:   "",
			wantErr:     true,
			errContains: "GITEA_TOKEN environment variable is not set",
		},
		{
			name:      "token with special characters",
			envValue:  "gitea_token_abc123.-_valid",
			wantToken: "gitea_token_abc123.-_valid",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv("GITEA_TOKEN", tt.envValue)
			} else {
				os.Unsetenv("GITEA_TOKEN")
			}

			// Test the function
			token, err := GetTokenFromEnv()

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTokenFromEnv() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("GetTokenFromEnv() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetTokenFromEnv() unexpected error = %v", err)
					return
				}
				if token != tt.wantToken {
					t.Errorf("GetTokenFromEnv() = %v, want %v", token, tt.wantToken)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsStringHelper(str, substr))
}

func containsStringHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidateTokenFormat(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid token",
			token:   "gitea_token_12345678901234567890",
			wantErr: false,
		},
		{
			name:        "empty token",
			token:       "",
			wantErr:     true,
			errContains: "token cannot be empty",
		},
		{
			name:        "token too short",
			token:       "short",
			wantErr:     true,
			errContains: "token is too short",
		},
		{
			name:        "token with spaces",
			token:       "gitea token 12345678901234567890",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:        "token with special characters",
			token:       "gitea@token#12345678901234567890",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:        "token with newlines",
			token:       "gitea_token_12345678901234567890\n",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:    "token with only special chars",
			token:   "._-._-._-._-._-._-._-._-._-._-._-._-._-",
			wantErr: false,
		},
		{
			name:    "token exactly 20 chars",
			token:   "abcdefghijklmnopqrst",
			wantErr: false,
		},
		{
			name:        "token 19 chars",
			token:       "abcdefghijklmnopqrs",
			wantErr:     true,
			errContains: "token is too short",
		},
		{
			name:    "very long token",
			token:   "gitea_token_1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:        "token with unicode characters",
			token:       "gitea_tökén_12345678901234567890",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenFormat(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenFormat() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenFormat() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenFormat() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
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
			name:     "short token",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "standard token",
			token:    "gitea_token_12345678901234567890",
			expected: "gite************************7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskToken(tt.token)
			if result != tt.expected {
				t.Errorf("MaskToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateAndMaskToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		wantMasked  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid token",
			token:      "gitea_token_12345678901234567890",
			wantMasked: "gite************************7890",
			wantErr:    false,
		},
		{
			name:        "empty token",
			token:       "",
			wantMasked:  "",
			wantErr:     true,
			errContains: "token cannot be empty",
		},
		{
			name:        "token too short",
			token:       "short",
			wantMasked:  "",
			wantErr:     true,
			errContains: "token is too short",
		},
		{
			name:        "token with invalid characters",
			token:       "forgejo token 12345678901234567890",
			wantMasked:  "",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:       "token exactly 20 chars",
			token:      "abcdefghijklmnopqrst",
			wantMasked: "abcd************qrst",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked, err := ValidateAndMaskToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAndMaskToken() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateAndMaskToken() error = %v, expected to contain %v", err, tt.errContains)
				}
				if masked != "" {
					t.Errorf("ValidateAndMaskToken() expected empty masked token on error, got %v", masked)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAndMaskToken() unexpected error = %v", err)
					return
				}
				if masked != tt.wantMasked {
					t.Errorf("ValidateAndMaskToken() masked = %v, want %v", masked, tt.wantMasked)
				}
			}
		})
	}
}

func TestGetValidatedToken(t *testing.T) {
	// Save original value to restore later
	originalToken := os.Getenv("GITEA_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITEA_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITEA_TOKEN")
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid token in environment",
			envValue:  "gitea_token_12345678901234567890",
			wantToken: "gitea_token_12345678901234567890",
			wantErr:   false,
		},
		{
			name:        "missing environment variable",
			envValue:    "",
			wantToken:   "",
			wantErr:     true,
			errContains: "GITEA_TOKEN environment variable is not set",
		},
		{
			name:        "invalid token format",
			envValue:    "short",
			wantToken:   "",
			wantErr:     true,
			errContains: "token is too short",
		},
		{
			name:        "token with invalid characters",
			envValue:    "gitea token 12345678901234567890",
			wantToken:   "",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:      "token exactly 20 chars",
			envValue:  "abcdefghijklmnopqrst",
			wantToken: "abcdefghijklmnopqrst",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv("GITEA_TOKEN", tt.envValue)
			} else {
				os.Unsetenv("GITEA_TOKEN")
			}

			// Test the function
			token, err := GetValidatedToken()

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetValidatedToken() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("GetValidatedToken() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetValidatedToken() unexpected error = %v", err)
					return
				}
				if token != tt.wantToken {
					t.Errorf("GetValidatedToken() = %v, want %v", token, tt.wantToken)
				}
			}
		})
	}
}

func TestTokenValidationError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		field    string
		expected string
	}{
		{
			name:     "standard error",
			message:  "token is invalid",
			field:    "token",
			expected: "token validation failed: token is invalid",
		},
		{
			name:     "empty message",
			message:  "",
			field:    "GITEA_TOKEN",
			expected: "token validation failed: ",
		},
		{
			name:     "empty field",
			message:  "validation failed",
			field:    "",
			expected: "token validation failed: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &TokenValidationError{
				Message: tt.message,
				Field:   tt.field,
			}
			if err.Error() != tt.expected {
				t.Errorf("TokenValidationError.Error() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestTokenSourceConstants(t *testing.T) {
	// Test that constants are defined and have expected values
	if TokenSourceEnv != "environment" {
		t.Errorf("TokenSourceEnv = %v, want %v", TokenSourceEnv, "environment")
	}
	if TokenSourceConfig != "config" {
		t.Errorf("TokenSourceConfig = %v, want %v", TokenSourceConfig, "config")
	}
}

func TestMaskTokenEdgeCases(t *testing.T) {
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
			name:     "single character",
			token:    "a",
			expected: "*",
		},
		{
			name:     "two characters",
			token:    "ab",
			expected: "**",
		},
		{
			name:     "three characters",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "four characters",
			token:    "abcd",
			expected: "****",
		},
		{
			name:     "five characters",
			token:    "abcde",
			expected: "ab*de",
		},
		{
			name:     "six characters",
			token:    "abcdef",
			expected: "ab**ef",
		},
		{
			name:     "seven characters",
			token:    "abcdefg",
			expected: "ab***fg",
		},
		{
			name:     "eight characters",
			token:    "abcdefgh",
			expected: "ab****gh",
		},
		{
			name:     "nine characters",
			token:    "abcdefghi",
			expected: "abcd*fghi",
		},
		{
			name:     "ten characters",
			token:    "abcdefghij",
			expected: "abcd**ghij",
		},
		{
			name:     "standard token",
			token:    "forgejo_token_12345678901234567890",
			expected: "forg**************************7890",
		},
		{
			name:     "very long token",
			token:    "very_long_token_with_many_characters_1234567890123456789012345678901234567890",
			expected: "very*********************************************************************7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskToken(tt.token)
			if result != tt.expected {
				t.Errorf("MaskToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestValidateTokenWithClient tests token validation using Gitea SDK client
func TestValidateTokenWithClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		setupMock   func() *mockClient
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid token",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true}
			},
			wantErr: false,
		},
		{
			name:    "invalid token",
			baseURL: "https://forgejo.example.com",
			token:   "invalid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: false}
			},
			wantErr:     true,
			errContains: "authentication failed",
		},
		{
			name:        "empty baseURL",
			baseURL:     "",
			token:       "valid_token_12345678901234567890",
			setupMock:   func() *mockClient { return &mockClient{} },
			wantErr:     true,
			errContains: "baseURL cannot be empty",
		},
		{
			name:        "empty token",
			baseURL:     "https://forgejo.example.com",
			token:       "",
			setupMock:   func() *mockClient { return &mockClient{} },
			wantErr:     true,
			errContains: "token cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()

			err := ValidateTokenWithClient(tt.baseURL, tt.token, mockClient)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenWithClient() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenWithClient() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenWithClient() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateTokenWithTimeout tests token validation with timeout
func TestValidateTokenWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		timeout     time.Duration
		setupMock   func() *mockClient
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful validation within timeout",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			timeout: 5 * time.Second,
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true, delay: 100 * time.Millisecond}
			},
			wantErr: false,
		},
		{
			name:    "validation timeout",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			timeout: 100 * time.Millisecond,
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true, delay: 200 * time.Millisecond}
			},
			wantErr:     true,
			errContains: "timed out",
		},
		{
			name:    "network error",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			timeout: 5 * time.Second,
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: false, networkError: true}
			},
			wantErr:     true,
			errContains: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()

			err := ValidateTokenWithTimeout(tt.baseURL, tt.token, tt.timeout, mockClient)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenWithTimeout() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenWithTimeout() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenWithTimeout() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateTokenCached tests cached token validation
func TestValidateTokenCached(t *testing.T) {
	// Clear cache before test
	ClearValidationCache()

	tests := []struct {
		name        string
		baseURL     string
		token       string
		setupMock   func() *mockClient
		callCount   int // expected number of times the mock client is called
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful validation gets cached",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true}
			},
			callCount: 1,
			wantErr:   false,
		},
		{
			name:    "different token requires new validation",
			baseURL: "https://forgejo.example.com",
			token:   "different_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true}
			},
			callCount: 1,
			wantErr:   false,
		},
		{
			name:    "failed validation not cached",
			baseURL: "https://forgejo.example.com",
			token:   "invalid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: false}
			},
			callCount:   1,
			wantErr:     true,
			errContains: "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearValidationCache()

			callCount := 0
			mockClient := &mockClient{
				shouldSucceed: tt.setupMock().shouldSucceed,
				delay:         tt.setupMock().delay,
				networkError:  tt.setupMock().networkError,
				callCount:     &callCount,
			}

			err := ValidateTokenCached(tt.baseURL, tt.token, mockClient)

			if callCount != tt.callCount {
				t.Errorf("ValidateTokenCached() mock called %d times, expected %d", callCount, tt.callCount)
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenCached() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenCached() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenCached() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateTokenCachedReuse tests that cached validations work
func TestValidateTokenCachedReuse(t *testing.T) {
	ClearValidationCache()

	baseURL := "https://forgejo.example.com"
	token := "valid_token_12345678901234567890"

	// First call should validate and cache
	callCount1 := 0
	mockClient1 := &mockClient{
		shouldSucceed: true,
		callCount:     &callCount1,
	}
	err1 := ValidateTokenCached(baseURL, token, mockClient1)
	if err1 != nil {
		t.Errorf("First validation failed: %v", err1)
	}
	if callCount1 != 1 {
		t.Errorf("First call should have called mock once, got %d", callCount1)
	}

	// Second call should use cache
	callCount2 := 0
	mockClient2 := &mockClient{
		shouldSucceed: true,
		callCount:     &callCount2,
	}
	err2 := ValidateTokenCached(baseURL, token, mockClient2)
	if err2 != nil {
		t.Errorf("Second validation failed: %v", err2)
	}
	if callCount2 != 0 {
		t.Errorf("Second call should have used cache, got %d calls", callCount2)
	}
}

func TestValidateTokenCachedBasic(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		setupMock   func() *mockClient
		callCount   int // expected number of times the mock client is called
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful validation gets cached",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true}
			},
			callCount: 1,
			wantErr:   false,
		},
		{
			name:    "different token requires new validation",
			baseURL: "https://forgejo.example.com",
			token:   "different_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true}
			},
			callCount: 1,
			wantErr:   false,
		},
		{
			name:    "failed validation not cached",
			baseURL: "https://forgejo.example.com",
			token:   "invalid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: false}
			},
			callCount:   1,
			wantErr:     true,
			errContains: "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearValidationCache()

			callCount := 0
			mockClient := &mockClient{
				shouldSucceed: tt.setupMock().shouldSucceed,
				delay:         tt.setupMock().delay,
				networkError:  tt.setupMock().networkError,
				callCount:     &callCount,
			}

			err := ValidateTokenCached(tt.baseURL, tt.token, mockClient)

			if callCount != tt.callCount {
				t.Errorf("ValidateTokenCached() mock called %d times, expected %d", callCount, tt.callCount)
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenCached() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenCached() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenCached() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateTokenWithTimeoutDefault tests the default timeout wrapper
func TestValidateTokenWithTimeoutDefault(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		setupMock   func() *mockClient
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful validation with default timeout",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true, delay: 100 * time.Millisecond}
			},
			wantErr: false,
		},
		{
			name:    "timeout with default 5 seconds",
			baseURL: "https://forgejo.example.com",
			token:   "valid_token_12345678901234567890",
			setupMock: func() *mockClient {
				return &mockClient{shouldSucceed: true, delay: 6 * time.Second}
			},
			wantErr:     true,
			errContains: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()

			err := ValidateTokenWithTimeoutDefault(tt.baseURL, tt.token, mockClient)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTokenWithTimeoutDefault() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateTokenWithTimeoutDefault() error = %v, expected to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTokenWithTimeoutDefault() unexpected error = %v", err)
				}
			}
		})
	}
}

// mockClient implements a mock client for testing
type mockClient struct {
	shouldSucceed bool
	delay         time.Duration
	networkError  bool
	validateFunc  func() error
	callCount     *int
}

func (m *mockClient) ValidateToken(baseURL, token string) error {
	if m.callCount != nil {
		*m.callCount++
	}

	if m.validateFunc != nil {
		return m.validateFunc()
	}

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.networkError {
		return fmt.Errorf("network error: connection failed")
	}

	if !m.shouldSucceed {
		return fmt.Errorf("authentication failed: invalid token")
	}

	return nil
}

// TestAuthErrorTypes tests custom error types for authentication failures
func TestAuthErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantType    string
		wantMessage string
	}{
		{
			name:        "TokenValidationError",
			err:         &TokenValidationError{Message: "invalid token", Field: "token"},
			wantType:    "*auth.TokenValidationError",
			wantMessage: "token validation failed: invalid token",
		},
		{
			name:        "AuthNetworkError",
			err:         &AuthNetworkError{Message: "connection timeout", URL: "https://forgejo.example.com"},
			wantType:    "*auth.AuthNetworkError",
			wantMessage: "authentication network error: connection timeout (URL: https://forgejo.example.com)",
		},
		{
			name:        "AuthTimeoutError",
			err:         &AuthTimeoutError{Timeout: 5 * time.Second},
			wantType:    "*auth.AuthTimeoutError",
			wantMessage: "authentication timed out after 5s",
		},
		{
			name:        "AuthServerError",
			err:         &AuthServerError{StatusCode: 500, Message: "internal server error"},
			wantType:    "*auth.AuthServerError",
			wantMessage: "authentication server error: 500 - internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if fmt.Sprintf("%T", tt.err) != tt.wantType {
				t.Errorf("Error type = %T, want %s", tt.err, tt.wantType)
			}
			if tt.err.Error() != tt.wantMessage {
				t.Errorf("Error message = %q, want %q", tt.err.Error(), tt.wantMessage)
			}
		})
	}
}

// TestSecureErrorFormatting tests that error messages never expose sensitive tokens
func TestSecureErrorFormatting(t *testing.T) {
	token := "sensitive_token_12345678901234567890"

	tests := []struct {
		name             string
		inputError       error
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "TokenValidationError with token in message",
			inputError:       fmt.Errorf("validation failed for token: %s", token),
			shouldContain:    []string{"validation failed"},
			shouldNotContain: []string{token},
		},
		{
			name:             "NetworkError with token in URL",
			inputError:       fmt.Errorf("failed to connect to https://api.example.com?token=%s", token),
			shouldContain:    []string{"failed to connect"},
			shouldNotContain: []string{token},
		},
		{
			name:             "TimeoutError with token context",
			inputError:       fmt.Errorf("request timed out while validating token: %s", token),
			shouldContain:    []string{"timed out", "validating"},
			shouldNotContain: []string{token},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secureErr := SecureErrorMessage(tt.inputError, token)

			for _, shouldContain := range tt.shouldContain {
				if !containsString(secureErr, shouldContain) {
					t.Errorf("SecureErrorMessage() should contain %q, but got: %s", shouldContain, secureErr)
				}
			}

			for _, shouldNotContain := range tt.shouldNotContain {
				if containsString(secureErr, shouldNotContain) {
					t.Errorf("SecureErrorMessage() should not contain sensitive data %q, but got: %s", shouldNotContain, secureErr)
				}
			}

			// Verify that if token appears, it's properly masked
			if containsString(secureErr, token) {
				t.Errorf("SecureErrorMessage() should never contain the raw token, got: %s", secureErr)
			}
		})
	}
}

// TestErrorWrappingWithContext tests error wrapping while maintaining security
func TestErrorWrappingWithContext(t *testing.T) {
	token := "sensitive_token_12345678901234567890"

	tests := []struct {
		name           string
		originalErr    error
		context        string
		operation      string
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:           "wrap validation error",
			originalErr:    &TokenValidationError{Message: "invalid format", Field: "token"},
			context:        "user authentication",
			operation:      "ValidateToken",
			wantContain:    []string{"user authentication", "ValidateToken", "invalid format"},
			wantNotContain: []string{token},
		},
		{
			name:           "wrap network error",
			originalErr:    fmt.Errorf("connection failed to %s", "https://api.example.com"),
			context:        "token validation",
			operation:      "HTTP POST",
			wantContain:    []string{"token validation", "HTTP POST", "connection failed"},
			wantNotContain: []string{token},
		},
		{
			name:           "wrap timeout error",
			originalErr:    &AuthTimeoutError{Timeout: 5 * time.Second},
			context:        "server communication",
			operation:      "API call",
			wantContain:    []string{"server communication", "API call", "timed out"},
			wantNotContain: []string{token},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := WrapErrorWithContext(tt.originalErr, tt.context, tt.operation, token)

			for _, shouldContain := range tt.wantContain {
				if !containsString(wrappedErr.Error(), shouldContain) {
					t.Errorf("WrapErrorWithContext() should contain %q, but got: %s", shouldContain, wrappedErr.Error())
				}
			}

			for _, shouldNotContain := range tt.wantNotContain {
				if containsString(wrappedErr.Error(), shouldNotContain) {
					t.Errorf("WrapErrorWithContext() should not contain %q, but got: %s", shouldNotContain, wrappedErr.Error())
				}
			}
		})
	}
}

// TestErrorChainUnwrapping tests that wrapped errors can be properly unwrapped
func TestErrorChainUnwrapping(t *testing.T) {
	originalErr := &TokenValidationError{Message: "invalid token", Field: "token"}
	token := "test_token_123"

	wrappedErr := WrapErrorWithContext(originalErr, "authentication", "ValidateToken", token)

	// Test that we can unwrap to get the original error
	unwrapped := wrappedErr
	for unwrapped != nil {
		if _, ok := unwrapped.(*TokenValidationError); ok {
			break
		}
		if u, ok := unwrapped.(interface{ Unwrap() error }); ok {
			unwrapped = u.Unwrap()
		} else {
			break
		}
	}

	if _, ok := unwrapped.(*TokenValidationError); !ok {
		t.Errorf("Should be able to unwrap to TokenValidationError, got %T", unwrapped)
	}
}

// TestErrorTypeAssertions tests type assertions for custom error types
func TestErrorTypeAssertions(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		assertFunc func(error) bool
		shouldPass bool
	}{
		{
			name: "TokenValidationError assertion",
			err:  &TokenValidationError{Message: "test", Field: "token"},
			assertFunc: func(e error) bool {
				_, ok := e.(*TokenValidationError)
				return ok
			},
			shouldPass: true,
		},
		{
			name: "AuthNetworkError assertion",
			err:  &AuthNetworkError{Message: "test", URL: "https://example.com"},
			assertFunc: func(e error) bool {
				_, ok := e.(*AuthNetworkError)
				return ok
			},
			shouldPass: true,
		},
		{
			name: "AuthTimeoutError assertion",
			err:  &AuthTimeoutError{Timeout: time.Second},
			assertFunc: func(e error) bool {
				_, ok := e.(*AuthTimeoutError)
				return ok
			},
			shouldPass: true,
		},
		{
			name: "AuthServerError assertion",
			err:  &AuthServerError{StatusCode: 500, Message: "test"},
			assertFunc: func(e error) bool {
				_, ok := e.(*AuthServerError)
				return ok
			},
			shouldPass: true,
		},
		{
			name: "wrong type assertion should fail",
			err:  &TokenValidationError{Message: "test", Field: "token"},
			assertFunc: func(e error) bool {
				_, ok := e.(*AuthNetworkError)
				return ok
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.assertFunc(tt.err)
			if result != tt.shouldPass {
				t.Errorf("Type assertion result = %v, want %v", result, tt.shouldPass)
			}
		})
	}
}

// TestNewAuthErrorConstructors tests the constructor functions for auth errors
func TestNewAuthErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() error
		expectedType string
		expectedMsg  string
	}{
		{
			name: "NewAuthNetworkError",
			constructor: func() error {
				return NewAuthNetworkError("connection failed", "https://example.com")
			},
			expectedType: "*auth.AuthNetworkError",
			expectedMsg:  "authentication network error: connection failed (URL: https://example.com)",
		},
		{
			name: "NewAuthNetworkError without URL",
			constructor: func() error {
				return NewAuthNetworkError("connection failed", "")
			},
			expectedType: "*auth.AuthNetworkError",
			expectedMsg:  "authentication network error: connection failed",
		},
		{
			name: "NewAuthTimeoutError",
			constructor: func() error {
				return NewAuthTimeoutError(10 * time.Second)
			},
			expectedType: "*auth.AuthTimeoutError",
			expectedMsg:  "authentication timed out after 10s",
		},
		{
			name: "NewAuthServerError",
			constructor: func() error {
				return NewAuthServerError(500, "internal server error")
			},
			expectedType: "*auth.AuthServerError",
			expectedMsg:  "authentication server error: 500 - internal server error",
		},
		{
			name: "NewAuthServerError without status code",
			constructor: func() error {
				return NewAuthServerError(0, "internal server error")
			},
			expectedType: "*auth.AuthServerError",
			expectedMsg:  "authentication server error: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()
			if fmt.Sprintf("%T", err) != tt.expectedType {
				t.Errorf("Error type = %T, want %s", err, tt.expectedType)
			}
			if err.Error() != tt.expectedMsg {
				t.Errorf("Error message = %q, want %q", err.Error(), tt.expectedMsg)
			}
		})
	}
}

// TestIsTemporaryAuthError tests temporary vs permanent error classification
func TestIsTemporaryAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "AuthNetworkError is temporary",
			err:      NewAuthNetworkError("connection failed", "https://example.com"),
			expected: true,
		},
		{
			name:     "AuthTimeoutError is temporary",
			err:      NewAuthTimeoutError(5 * time.Second),
			expected: true,
		},
		{
			name:     "TokenValidationError is not temporary",
			err:      &TokenValidationError{Message: "invalid format", Field: "token"},
			expected: false,
		},
		{
			name:     "AuthServerError 500 is not temporary",
			err:      NewAuthServerError(500, "internal error"),
			expected: false,
		},
		{
			name:     "AuthServerError 401 is not temporary",
			err:      NewAuthServerError(401, "unauthorized"),
			expected: false,
		},
		{
			name:     "wrapped temporary error",
			err:      WrapErrorWithContext(NewAuthNetworkError("connection failed", ""), "test", "operation", "token"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTemporaryAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("IsTemporaryAuthError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsPermanentAuthError tests permanent error classification
func TestIsPermanentAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "TokenValidationError is permanent",
			err:      &TokenValidationError{Message: "invalid format", Field: "token"},
			expected: true,
		},
		{
			name:     "AuthServerError 401 is permanent",
			err:      NewAuthServerError(401, "unauthorized"),
			expected: true,
		},
		{
			name:     "AuthServerError 403 is permanent",
			err:      NewAuthServerError(403, "forbidden"),
			expected: true,
		},
		{
			name:     "AuthServerError 500 is not permanent",
			err:      NewAuthServerError(500, "internal error"),
			expected: false,
		},
		{
			name:     "AuthNetworkError is not permanent",
			err:      NewAuthNetworkError("connection failed", "https://example.com"),
			expected: false,
		},
		{
			name:     "AuthTimeoutError is not permanent",
			err:      NewAuthTimeoutError(5 * time.Second),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPermanentAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("IsPermanentAuthError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetAuthErrorType tests error type identification
func TestGetAuthErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "TokenValidationError",
			err:      &TokenValidationError{Message: "test", Field: "token"},
			expected: "TokenValidationError",
		},
		{
			name:     "AuthNetworkError",
			err:      NewAuthNetworkError("test", "https://example.com"),
			expected: "AuthNetworkError",
		},
		{
			name:     "AuthTimeoutError",
			err:      NewAuthTimeoutError(time.Second),
			expected: "AuthTimeoutError",
		},
		{
			name:     "AuthServerError",
			err:      NewAuthServerError(500, "test"),
			expected: "AuthServerError",
		},
		{
			name:     "ContextWrappedError",
			err:      WrapErrorWithContext(fmt.Errorf("test"), "context", "operation", "token"),
			expected: "ContextWrappedError",
		},
		{
			name:     "unknown error type",
			err:      fmt.Errorf("generic error"),
			expected: "UnknownError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAuthErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("GetAuthErrorType() = %q, want %q", result, tt.expected)
			}
		})
	}
}
