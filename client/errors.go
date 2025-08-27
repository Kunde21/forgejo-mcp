// Package client provides error types for Gitea SDK client operations
package client

import "fmt"

// APIError represents an error from the Gitea API
type APIError struct {
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Message)
}

// NetworkError represents a network-related error
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}
