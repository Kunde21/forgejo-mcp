package servertest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// getRepoKeyFromRequest extracts the repository key from path values using modern Go 1.22+ routing
func getRepoKeyFromRequest(r *http.Request) (string, error) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	if owner == "" || repo == "" {
		return "", http.ErrMissingFile
	}

	return owner + "/" + repo, nil
}

// parsePagination extracts limit and offset from query parameters
func parsePagination(r *http.Request) (limit, offset int) {
	// Set default values to match server implementation
	limit = 15 // Default limit matches server default
	offset = 0

	// Parse limit from query parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Check if page parameter exists (Gitea SDK uses page-based pagination)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			// Convert page to offset: offset = (page - 1) * limit
			offset = (parsedPage - 1) * limit
		}
	} else {
		// Parse offset from query parameters (fallback for direct offset usage)
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}
	}

	return limit, offset
}

// validateAuthToken validates authentication token from request headers
func validateAuthToken(r *http.Request) bool {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")

	// Empty authorization header should be rejected
	if authHeader == "" {
		// Check if there's a token in query parameters
		queryToken := r.URL.Query().Get("token")
		if queryToken == "" {
			// No authentication at all - acceptable for backward compatibility
			return true
		}
		// There's a query token, validate it
		return queryToken == "mock-token" || queryToken == "test-token"
	}

	var headerToken string
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		headerToken = after
	} else if after, ok := strings.CutPrefix(authHeader, "token "); ok {
		headerToken = after
	} else {
		// Malformed authorization header (not Bearer or token format)
		return false
	}

	// Check token in query parameters
	queryToken := r.URL.Query().Get("token")

	// Reject invalid tokens
	if headerToken == "invalid-token" || queryToken == "invalid-token" {
		return false
	}

	// If there's a header token, it must be valid
	if headerToken != "mock-token" && headerToken != "test-token" {
		return false
	}

	// If there's a query token, it must be valid
	if queryToken != "" && queryToken != "mock-token" && queryToken != "test-token" {
		return false
	}

	return true
}

// writeJSONResponse writes standardized JSON responses
func writeJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ValidateToolCall executes a tool call with standardized validation and error handling
//
// Parameters:
//   - t: *testing.T for test context and error reporting
//   - client: *mcp.ClientSession the MCP client session
//   - ctx: context.Context for the tool call
//   - toolName: string name of the tool to call
//   - arguments: map[string]any arguments for the tool call
//   - expectedError: string expected error text (empty for success case)
//
// Returns:
//   - *mcp.CallToolResult: the result of the tool call (nil on error)
//
// Example usage:
//
//	result := ValidateToolCall(t, client, ctx, "issue_list", map[string]any{
//	    "repository": "testuser/testrepo",
//	}, "")
//	if result != nil {
//	    t.Log("Tool call succeeded")
//	}
func ValidateToolCall(t *testing.T, client *mcp.ClientSession, ctx context.Context, toolName string, arguments map[string]any, expectedError string) *mcp.CallToolResult {
	t.Helper()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		if expectedError != "" {
			if strings.Contains(err.Error(), expectedError) {
				return nil
			}
			t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
		} else {
			t.Errorf("Tool call failed unexpectedly: %v", err)
		}
		return nil
	}

	// Check if we expected an error but got success
	if expectedError != "" {
		t.Errorf("Expected error containing '%s', but got success result", expectedError)
		return nil
	}

	return result
}

// AssertToolResultEqual compares two tool results for equality with detailed error reporting
//
// Parameters:
//   - t: *testing.T for test context and error reporting
//   - expected: *mcp.CallToolResult the expected result
//   - actual: *mcp.CallToolResult the actual result
//
// Example usage:
//
//	AssertToolResultEqual(t, tc.expect, result)
func AssertToolResultEqual(t *testing.T, expected, actual *mcp.CallToolResult) {
	t.Helper()

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
		t.Errorf("Tool result mismatch (-expected +actual):\n%s",
			cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(mcp.TextContent{})))
	}
}

// AssertToolResultContains validates that a tool result contains expected text
//
// Parameters:
//   - t: *testing.T for test context and error reporting
//   - result: *mcp.CallToolResult the result to validate
//   - expectedText: string the expected text to contain
//   - expectError: bool whether to expect an error result
//
// Example usage:
//
//	AssertToolResultContains(t, result, "Comment created successfully", false)
func AssertToolResultContains(t *testing.T, result *mcp.CallToolResult, expectedText string, expectError bool) {
	t.Helper()

	if result == nil {
		t.Fatal("Tool result is nil")
	}

	if result.IsError != expectError {
		if expectError {
			t.Errorf("Expected error result, but got success")
		} else {
			t.Errorf("Expected success result, but got error")
		}
		return
	}

	actualText := GetTextContent(result.Content)
	if !strings.Contains(actualText, expectedText) {
		t.Errorf("Expected text '%s' not found in result: '%s'", expectedText, actualText)
	}
}

// CreateStandardTestContext creates a standardized test context with proper timeout and cleanup
//
// Parameters:
//   - t: *testing.T for test context
//   - timeoutSeconds: int timeout in seconds (defaults to 5 if 0)
//
// Returns:
//   - context.Context: the created context
//   - context.CancelFunc: the cancel function for cleanup
//
// Example usage:
//
//	ctx, cancel := CreateStandardTestContext(t, 10)
//	defer cancel()
func CreateStandardTestContext(t *testing.T, timeoutSeconds int) (context.Context, context.CancelFunc) {
	t.Helper()

	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(timeoutSeconds)*time.Second)
	t.Cleanup(cancel)

	return ctx, cancel
}

// RunConcurrentTest executes a function concurrently with proper synchronization and error handling
//
// Parameters:
//   - t: *testing.T for test context and error reporting
//   - numGoroutines: int number of goroutines to run
//   - testFunc: func(int) error function to execute in each goroutine
//
// Example usage:
//
//	RunConcurrentTest(t, 3, func(id int) error {
//	    _, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
//	        Name: "tool_name",
//	        Arguments: map[string]any{
//	            "id": id,
//	        },
//	    })
//	    return err
//	})
func RunConcurrentTest(t *testing.T, numGoroutines int, testFunc func(int) error) {
	t.Helper()

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if err := testFunc(id); err != nil {
				errChan <- fmt.Errorf("goroutine %d failed: %w", id, err)
			}
		}(i + 1)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Error(err)
	}
}
