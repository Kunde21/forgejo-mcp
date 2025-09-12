package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestIssueCommentEditSuccessful tests successful comment editing
func TestIssueCommentEditSuccessful(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 123, Content: "Test comment 1", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
		{ID: 234, Content: "Test comment 2", Author: "testuser", Created: "2025-09-09T10:31:00Z"},
	})

	// Test successful comment editing
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated comment content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	// Verify the result structure
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	// Check that the response contains expected success message
	if !strings.Contains(textContent.Text, "Comment edited successfully") {
		t.Errorf("Expected response to contain 'Comment edited successfully', got '%s'", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Updated comment content") {
		t.Errorf("Expected response to contain updated content, got '%s'", textContent.Text)
	}
}

// TestIssueCommentEditValidationError tests validation error scenarios
func TestIssueCommentEditValidationError(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	tests := []struct {
		name      string
		arguments map[string]any
		wantError bool
	}{
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository":   "invalid-repo-format",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "test content",
			},
			wantError: true,
		},
		{
			name: "missing repository",
			arguments: map[string]any{
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "test content",
			},
			wantError: true,
		},
		{
			name: "missing issue number",
			arguments: map[string]any{
				"repository":  "testuser/testrepo",
				"comment_id":  123,
				"new_content": "test content",
			},
			wantError: true,
		},
		{
			name: "missing comment id",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"new_content":  "test content",
			},
			wantError: true,
		},
		{
			name: "missing new content",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
			},
			wantError: true,
		},
		{
			name: "empty new content",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "",
			},
			wantError: true,
		},
		{
			name: "negative issue number",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"comment_id":   123,
				"new_content":  "test content",
			},
			wantError: true,
		},
		{
			name: "zero comment id",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   0,
				"new_content":  "test content",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_comment_edit",
				Arguments: tt.arguments,
			})

			if tt.wantError {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("Expected error for test case '%s', but got none. Err: %v, Result: %+v", tt.name, err, result)
					if result != nil && len(result.Content) > 0 {
						if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
							t.Logf("Response content: %s", textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test case '%s': %v", tt.name, err)
				}
				if result != nil && result.IsError {
					t.Errorf("Unexpected error result for test case '%s'", tt.name)
				}
			}
		})
	}
}

// TestIssueCommentEditPermissionError tests permission error scenarios
func TestIssueCommentEditPermissionError(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "invalid-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with invalid token (simulates permission error)
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated content",
		},
	})
	// Should return an error result for permission issues
	if err != nil {
		t.Errorf("Expected successful call with error result, got client error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	} else if !result.IsError {
		t.Errorf("Expected error result for permission issue, got success result: %+v", result)
	}
}

// TestIssueCommentEditAPIError tests API failure scenarios
func TestIssueCommentEditAPIError(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	mock.SetNotFoundRepo("nonexistent", "repo")
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with nonexistent repository (should trigger API error)
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "nonexistent/repo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated content",
		},
	})
	// Should return an error result for API failure
	if err != nil {
		t.Errorf("Expected successful call with error result, got client error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	} else if !result.IsError {
		t.Errorf("Expected error result for API failure, got success result: %+v", result)
	}
}

// Concurrent testing is now covered in acceptance tests to avoid duplication
