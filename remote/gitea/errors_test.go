package gitea

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSDKError_Error(t *testing.T) {
	tests := []struct {
		name     string
		sdkErr   *SDKError
		expected string
	}{
		{
			name: "error with context",
			sdkErr: &SDKError{
				Operation: "ListRepoPullRequests",
				Cause:     errors.New("connection failed"),
				Context:   "owner=testuser, repo=testrepo",
			},
			expected: "Gitea SDK ListRepoPullRequests failed (owner=testuser, repo=testrepo): connection failed",
		},
		{
			name: "error without context",
			sdkErr: &SDKError{
				Operation: "GetRepo",
				Cause:     errors.New("not found"),
				Context:   "",
			},
			expected: "Gitea SDK GetRepo failed: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sdkErr.Error()
			if result != tt.expected {
				t.Errorf("SDKError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSDKError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	sdkErr := &SDKError{
		Operation: "TestOperation",
		Cause:     originalErr,
		Context:   "test context",
	}

	unwrapped := sdkErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("SDKError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewSDKError(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		cause       error
		context     []string
		expectedErr *SDKError
	}{
		{
			name:      "error with single context",
			operation: "ListMyRepos",
			cause:     errors.New("unauthorized"),
			context:   []string{"limit=10"},
			expectedErr: &SDKError{
				Operation: "ListMyRepos",
				Cause:     errors.New("unauthorized"),
				Context:   "limit=10",
			},
		},
		{
			name:      "error with multiple context items",
			operation: "GetIssue",
			cause:     errors.New("not found"),
			context:   []string{"owner=test", "repo=test", "index=1"},
			expectedErr: &SDKError{
				Operation: "GetIssue",
				Cause:     errors.New("not found"),
				Context:   "owner=test, repo=test, index=1",
			},
		},
		{
			name:      "error without context",
			operation: "GetUserInfo",
			cause:     errors.New("network error"),
			context:   nil,
			expectedErr: &SDKError{
				Operation: "GetUserInfo",
				Cause:     errors.New("network error"),
				Context:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSDKError(tt.operation, tt.cause, tt.context...)

			if diff := cmp.Diff(tt.expectedErr, result, cmp.Comparer(func(a, b error) bool {
				return a.Error() == b.Error()
			})); diff != "" {
				t.Errorf("NewSDKError() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
