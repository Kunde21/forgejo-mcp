// Package auth provides authentication functionality for Forgejo repositories
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
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

// GetTokenFromEnv reads the GITEA_TOKEN environment variable
// Returns the token value and any validation errors
func GetTokenFromEnv() (string, error) {
	token := os.Getenv("GITEA_TOKEN")
	if token == "" {
		return "", &TokenValidationError{
			Message: "GITEA_TOKEN environment variable is not set",
			Field:   "GITEA_TOKEN",
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

// TokenValidator defines the interface for token validation
type TokenValidator interface {
	ValidateToken(baseURL, token string) error
}

// ValidateTokenWithClient validates a token using a Gitea SDK client
// This function performs actual authentication against the Forgejo server
func ValidateTokenWithClient(baseURL, token string, validator TokenValidator) error {
	if baseURL == "" {
		return &TokenValidationError{
			Message: "baseURL cannot be empty",
			Field:   "baseURL",
		}
	}

	if token == "" {
		return &TokenValidationError{
			Message: "token cannot be empty",
			Field:   "token",
		}
	}

	if err := ValidateTokenFormat(token); err != nil {
		return err
	}

	return validator.ValidateToken(baseURL, token)
}

// ValidateTokenWithTimeout validates a token with a specified timeout
func ValidateTokenWithTimeout(baseURL, token string, timeout time.Duration, validator TokenValidator) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a channel to receive the validation result
	resultChan := make(chan error, 1)

	// Run validation in a goroutine
	go func() {
		resultChan <- ValidateTokenWithClient(baseURL, token, validator)
	}()

	// Wait for either the result or timeout
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return &TokenValidationError{
			Message: fmt.Sprintf("token validation timed out after %v", timeout),
			Field:   "timeout",
		}
	}
}

// DefaultValidationTimeout is the default timeout for token validation (5 seconds)
const DefaultValidationTimeout = 5 * time.Second

// ValidateTokenWithTimeoutDefault validates a token with the default 5-second timeout
func ValidateTokenWithTimeoutDefault(baseURL, token string, validator TokenValidator) error {
	return ValidateTokenWithTimeout(baseURL, token, DefaultValidationTimeout, validator)
}

// validationCache holds cached validation results
var (
	validationCache   = make(map[string]bool)
	validationCacheMu sync.RWMutex
)

// CacheKey generates a cache key for the given baseURL and token
func CacheKey(baseURL, token string) string {
	return fmt.Sprintf("%s:%s", baseURL, MaskToken(token))
}

// ValidateTokenCached validates a token with caching of successful results
// Only successful validations are cached; failures are not cached to allow retries
func ValidateTokenCached(baseURL, token string, validator TokenValidator) error {
	key := CacheKey(baseURL, token)

	// Check cache first
	validationCacheMu.RLock()
	if cached, exists := validationCache[key]; exists && cached {
		validationCacheMu.RUnlock()
		return nil
	}
	validationCacheMu.RUnlock()

	// Perform validation
	err := ValidateTokenWithTimeoutDefault(baseURL, token, validator)

	// Cache successful validation only
	if err == nil {
		validationCacheMu.Lock()
		validationCache[key] = true
		validationCacheMu.Unlock()
	}

	return err
}

// ClearValidationCache clears the validation cache (used for testing)
func ClearValidationCache() {
	validationCacheMu.Lock()
	defer validationCacheMu.Unlock()
	validationCache = make(map[string]bool)
}

// AuthNetworkError represents network-related authentication errors
type AuthNetworkError struct {
	Message string
	URL     string
}

// Error implements the error interface for AuthNetworkError
func (e *AuthNetworkError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("authentication network error: %s (URL: %s)", e.Message, e.URL)
	}
	return fmt.Sprintf("authentication network error: %s", e.Message)
}

// Unwrap returns nil as AuthNetworkError doesn't wrap another error
func (e *AuthNetworkError) Unwrap() error {
	return nil
}

// AuthTimeoutError represents timeout errors during authentication
type AuthTimeoutError struct {
	Timeout time.Duration
}

// Error implements the error interface for AuthTimeoutError
func (e *AuthTimeoutError) Error() string {
	return fmt.Sprintf("authentication timed out after %v", e.Timeout)
}

// Unwrap returns nil as AuthTimeoutError doesn't wrap another error
func (e *AuthTimeoutError) Unwrap() error {
	return nil
}

// AuthServerError represents server-side authentication errors
type AuthServerError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface for AuthServerError
func (e *AuthServerError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("authentication server error: %d - %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("authentication server error: %s", e.Message)
}

// Unwrap returns nil as AuthServerError doesn't wrap another error
func (e *AuthServerError) Unwrap() error {
	return nil
}

// SecureErrorMessage creates a secure error message that never exposes sensitive tokens
// It replaces any occurrence of the token with its masked version
func SecureErrorMessage(err error, token string) string {
	if err == nil {
		return ""
	}

	message := err.Error()

	// Early return if token is empty or not in message
	if token == "" || !strings.Contains(message, token) {
		return message
	}

	// Replace all occurrences of the token with its masked version
	maskedToken := MaskToken(token)
	return strings.ReplaceAll(message, token, maskedToken)
}

// WrapErrorWithContext wraps an error with additional context while maintaining security
func WrapErrorWithContext(err error, context, operation, token string) error {
	if err == nil {
		return nil
	}

	// Create a secure version of the original error message
	secureMessage := SecureErrorMessage(err, token)

	// Create a new error that wraps the original with context
	wrappedMessage := fmt.Sprintf("%s failed during %s: %s", operation, context, secureMessage)

	// Return a wrapped error that implements the Unwrap interface
	return &ContextWrappedError{
		Message:   wrappedMessage,
		Context:   context,
		Operation: operation,
		Cause:     err,
	}
}

// ContextWrappedError represents an error wrapped with additional context
type ContextWrappedError struct {
	Message   string
	Context   string
	Operation string
	Cause     error
}

func (e *ContextWrappedError) Error() string {
	return e.Message
}

func (e *ContextWrappedError) Unwrap() error {
	return e.Cause
}

// IsAuthError checks if an error is an authentication-related error
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	var authErr *TokenValidationError
	var networkErr *AuthNetworkError
	var timeoutErr *AuthTimeoutError
	var serverErr *AuthServerError

	return errors.As(err, &authErr) ||
		errors.As(err, &networkErr) ||
		errors.As(err, &timeoutErr) ||
		errors.As(err, &serverErr)
}

// GetAuthErrorType returns the type of authentication error
func GetAuthErrorType(err error) string {
	if err == nil {
		return ""
	}

	switch err.(type) {
	case *TokenValidationError:
		return "TokenValidationError"
	case *AuthNetworkError:
		return "AuthNetworkError"
	case *AuthTimeoutError:
		return "AuthTimeoutError"
	case *AuthServerError:
		return "AuthServerError"
	case *ContextWrappedError:
		return "ContextWrappedError"
	default:
		return "UnknownError"
	}
}

// NewAuthNetworkError creates a new AuthNetworkError with the given message and URL
func NewAuthNetworkError(message, url string) *AuthNetworkError {
	return &AuthNetworkError{
		Message: message,
		URL:     url,
	}
}

// NewAuthTimeoutError creates a new AuthTimeoutError with the given timeout duration
func NewAuthTimeoutError(timeout time.Duration) *AuthTimeoutError {
	return &AuthTimeoutError{
		Timeout: timeout,
	}
}

// NewAuthServerError creates a new AuthServerError with the given status code and message
func NewAuthServerError(statusCode int, message string) *AuthServerError {
	return &AuthServerError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// IsTemporaryAuthError checks if an authentication error is temporary and could be retried
func IsTemporaryAuthError(err error) bool {
	if err == nil {
		return false
	}

	var networkErr *AuthNetworkError
	var timeoutErr *AuthTimeoutError

	// Network errors and timeouts are typically temporary
	return errors.As(err, &networkErr) || errors.As(err, &timeoutErr)
}

// IsPermanentAuthError checks if an authentication error is permanent and should not be retried
func IsPermanentAuthError(err error) bool {
	if err == nil {
		return false
	}

	var validationErr *TokenValidationError
	var serverErr *AuthServerError

	// Validation errors and server errors (like 401/403) are typically permanent
	if errors.As(err, &validationErr) {
		return true
	}

	if errors.As(err, &serverErr) {
		// 401 Unauthorized and 403 Forbidden are permanent
		return serverErr.StatusCode == 401 || serverErr.StatusCode == 403
	}

	return false
}
