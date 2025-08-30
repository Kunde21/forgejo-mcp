// Package auth provides authentication functionality for Forgejo repositories
package auth

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TokenSource represents the source of authentication tokens
type TokenSource string

const (
	// TokenSourceEnv represents tokens sourced from environment variables
	TokenSourceEnv TokenSource = "environment"
	// TokenSourceConfig represents tokens sourced from configuration files
	TokenSourceConfig TokenSource = "config"
)

// TokenValidationError represents an error during token validation
type TokenValidationError struct {
	Message string
	Field   string
}

func (e *TokenValidationError) Error() string {
	return fmt.Sprintf("token validation failed: %s", e.Message)
}

// GetTokenFromEnv reads the FORGEJO_TOKEN environment variable
// Returns the token value and any validation errors
func GetTokenFromEnv() (string, error) {
	token := os.Getenv("FORGEJO_TOKEN")
	if token == "" {
		return "", &TokenValidationError{
			Message: "FORGEJO_TOKEN environment variable is not set",
			Field:   "FORGEJO_TOKEN",
		}
	}
	return token, nil
}

// ValidateTokenFormat validates the format and presence of a token
// Basic validation ensures the token is not empty and contains only valid characters
func ValidateTokenFormat(token string) error {
	if token == "" {
		return &TokenValidationError{
			Message: "token cannot be empty",
			Field:   "token",
		}
	}

	// Basic token format validation - tokens should be alphanumeric with some special chars
	// This is a basic check; actual validation would depend on Forgejo's token format
	validTokenPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	if !validTokenPattern.MatchString(token) {
		return &TokenValidationError{
			Message: "token contains invalid characters",
			Field:   "token",
		}
	}

	// Check minimum length (typical tokens are at least 20 characters)
	if len(token) < 20 {
		return &TokenValidationError{
			Message: "token is too short (minimum 20 characters)",
			Field:   "token",
		}
	}

	return nil
}

// MaskToken securely masks a token for logging purposes
// Replaces all but the first and last 4 characters with asterisks
func MaskToken(token string) string {
	if token == "" {
		return ""
	}

	tokenLen := len(token)
	if tokenLen <= 4 {
		// For very short tokens, mask everything
		return strings.Repeat("*", tokenLen)
	}

	if tokenLen <= 8 {
		// For tokens 5-8 chars, show first 2 and last 2, mask the middle
		return token[:2] + strings.Repeat("*", tokenLen-4) + token[tokenLen-2:]
	}

	// Show first 4 and last 4 characters, mask the middle
	return token[:4] + strings.Repeat("*", tokenLen-8) + token[tokenLen-4:]
}

// ValidateAndMaskToken validates a token and returns a masked version for logging
// This is a convenience function that combines validation and masking
func ValidateAndMaskToken(token string) (string, error) {
	if err := ValidateTokenFormat(token); err != nil {
		return "", err
	}
	return MaskToken(token), nil
}

// GetValidatedToken attempts to get a token from environment and validate it
// Returns the original token (not masked) and any errors
func GetValidatedToken() (string, error) {
	token, err := GetTokenFromEnv()
	if err != nil {
		return "", err
	}

	if err := ValidateTokenFormat(token); err != nil {
		return "", err
	}

	return token, nil
}
