package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCreateIssueCommentToolSuccessHTTP tests successful comment creation with HTTP-based MockGiteaServer
func TestCreateIssueCommentToolSuccessHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create HTTP-based mock server
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_issue_comment",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 42,
			"comment":      "This is a test comment via HTTP",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call create_issue_comment tool: %v", err)
	}

	// Verify the result structure
	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	// Check that the response contains expected information
	expectedParts := []string{"Comment created successfully", "ID:", "Created:"}
	for _, part := range expectedParts {
		if !strings.Contains(textContent.Text, part) {
			t.Errorf("Expected response to contain '%s', got: %s", part, textContent.Text)
		}
	}
}

// TestCreateIssueCommentToolValidationErrorsHTTP tests parameter validation with HTTP-based MockGiteaServer
func TestCreateIssueCommentToolValidationErrorsHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create HTTP-based mock server
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	testCases := []struct {
		name        string
		arguments   map[string]any
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing repository",
			arguments: map[string]any{
				"issue_number": 42,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 42,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "negative issue number",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue_number",
		},
		{
			name: "zero issue number",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue_number",
		},
		{
			name: "empty comment",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 42,
				"comment":      "",
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "missing comment",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 42,
			},
			expectError: true,
			errorMsg:    "comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "create_issue_comment",
				Arguments: tc.arguments,
			})

			if tc.expectError {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("Expected error for test case '%s', but got success", tc.name)
				}
				if result != nil && len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						if !strings.Contains(textContent.Text, tc.errorMsg) {
							t.Errorf("Expected error message to contain '%s', got: %s", tc.errorMsg, textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected success for test case '%s', but got error: %v", tc.name, err)
				}
			}
		})
	}
}

// TestCreateIssueCommentToolAPIErrorHTTP tests API error scenarios with HTTP-based MockGiteaServer
func TestCreateIssueCommentToolAPIErrorHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create HTTP-based mock server
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with repository that doesn't exist (mock will return 404)
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_issue_comment",
		Arguments: map[string]any{
			"repository":   "nonexistent/repo",
			"issue_number": 42,
			"comment":      "Test comment",
		},
	})

	if err == nil && (result == nil || !result.IsError) {
		t.Error("Expected error for nonexistent repository")
	}
}

// TestCreateIssueCommentToolCancelledContextHTTP tests context cancellation with HTTP-based MockGiteaServer
func TestCreateIssueCommentToolCancelledContextHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create HTTP-based mock server
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Cancel context immediately
	cancelledCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc()

	result, err := client.CallTool(cancelledCtx, &mcp.CallToolParams{
		Name: "create_issue_comment",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 42,
			"comment":      "Test comment",
		},
	})

	if err == nil {
		t.Error("Expected error when calling tool with cancelled context")
	}
	if result != nil && !result.IsError {
		t.Error("Expected error result for cancelled context")
	}
}
