package gitea

import (
	"fmt"
	"strings"
)

// SDKError represents an error from the Gitea SDK with additional context
type SDKError struct {
	Operation string // The operation that failed (e.g., "ListRepoPullRequests")
	Cause     error  // The original SDK error
	Context   string // Additional context about the operation
}

func (e *SDKError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("Gitea SDK %s failed (%s): %v", e.Operation, e.Context, e.Cause)
	}
	return fmt.Sprintf("Gitea SDK %s failed: %v", e.Operation, e.Cause)
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

// NewSDKError creates a new SDK error with context
func NewSDKError(operation string, cause error, context ...string) *SDKError {
	ctx := ""
	if len(context) > 0 {
		ctx = strings.Join(context, ", ")
	}
	return &SDKError{
		Operation: operation,
		Cause:     cause,
		Context:   ctx,
	}
}
