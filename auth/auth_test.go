package auth

import (
	"os"
	"testing"
)

func TestGetTokenFromEnv(t *testing.T) {
	// Save original value to restore later
	originalToken := os.Getenv("FORGEJO_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("FORGEJO_TOKEN", originalToken)
		} else {
			os.Unsetenv("FORGEJO_TOKEN")
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
			envValue:  "forgejo_token_12345678901234567890",
			wantToken: "forgejo_token_12345678901234567890",
			wantErr:   false,
		},
		{
			name:        "empty token",
			envValue:    "",
			wantToken:   "",
			wantErr:     true,
			errContains: "FORGEJO_TOKEN environment variable is not set",
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
				os.Setenv("FORGEJO_TOKEN", tt.envValue)
			} else {
				os.Unsetenv("FORGEJO_TOKEN")
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
			token:   "forgejo_token_12345678901234567890",
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
			token:       "forgejo token 12345678901234567890",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:        "token with special characters",
			token:       "forgejo@token#12345678901234567890",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:        "token with newlines",
			token:       "forgejo_token_12345678901234567890\n",
			wantErr:     true,
			errContains: "token contains invalid characters",
		},
		{
			name:        "token with only special chars",
			token:       "._-._-._-._-._-._-._-._-._-._-._-._-._-",
			wantErr:     false,
		},
		{
			name:        "token exactly 20 chars",
			token:       "abcdefghijklmnopqrst",
			wantErr:     false,
		},
		{
			name:        "token 19 chars",
			token:       "abcdefghijklmnopqrs",
			wantErr:     true,
			errContains: "token is too short",
		},
		{
			name:        "very long token",
			token:       "forgejo_token_1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			wantErr:     false,
		},
		{
			name:        "token with unicode characters",
			token:       "forgejo_tökén_12345678901234567890",
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
			token:    "forgejo_token_12345678901234567890",
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
			token:      "forgejo_token_12345678901234567890",
			wantMasked: "forg**************************7890",
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
			wantMasked: "abcd**************qrst",
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
	originalToken := os.Getenv("FORGEJO_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("FORGEJO_TOKEN", originalToken)
		} else {
			os.Unsetenv("FORGEJO_TOKEN")
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
			envValue:  "forgejo_token_12345678901234567890",
			wantToken: "forgejo_token_12345678901234567890",
			wantErr:   false,
		},
		{
			name:        "missing environment variable",
			envValue:    "",
			wantToken:   "",
			wantErr:     true,
			errContains: "FORGEJO_TOKEN environment variable is not set",
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
			envValue:    "forgejo token 12345678901234567890",
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
				os.Setenv("FORGEJO_TOKEN", tt.envValue)
			} else {
				os.Unsetenv("FORGEJO_TOKEN")
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
			field:    "FORGEJO_TOKEN",
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
			expected: "a**e",
		},
		{
			name:     "six characters",
			token:    "abcdef",
			expected: "a***f",
		},
		{
			name:     "seven characters",
			token:    "abcdefg",
			expected: "a****g",
		},
		{
			name:     "eight characters",
			token:    "abcdefgh",
			expected: "a*****h",
		},
		{
			name:     "nine characters",
			token:    "abcdefghi",
			expected: "abcd***hi",
		},
		{
			name:     "ten characters",
			token:    "abcdefghij",
			expected: "abcd****ij",
		},
		{
			name:     "standard token",
			token:    "forgejo_token_12345678901234567890",
			expected: "forg**************************7890",
		},
		{
			name:     "very long token",
			token:    "very_long_token_with_many_characters_1234567890123456789012345678901234567890",
			expected: "very_**************************************************************567890",
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
