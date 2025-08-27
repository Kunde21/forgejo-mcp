package client

import (
	"errors"
	"testing"
)

func TestAPIError(t *testing.T) {
	err := &APIError{
		Message:    "test error",
		StatusCode: 404,
	}

	if err.Error() != "API error (404): test error" {
		t.Errorf("Expected 'API error (404): test error', got '%s'", err.Error())
	}
}

func TestAuthError(t *testing.T) {
	err := &AuthError{
		Message: "authentication failed",
	}

	if err.Error() != "authentication error: authentication failed" {
		t.Errorf("Expected 'authentication error: authentication failed', got '%s'", err.Error())
	}
}

func TestNetworkError(t *testing.T) {
	underlyingErr := errors.New("connection refused")
	err := &NetworkError{
		Message: "network request failed",
		Cause:   underlyingErr,
	}

	expected := "network error: network request failed: connection refused"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Message: "invalid repository name",
		Field:   "repo",
	}

	expected := "validation error for field 'repo': invalid repository name"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}
