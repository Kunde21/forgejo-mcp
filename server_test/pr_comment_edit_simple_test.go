package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPRCommentEditValidation tests input validation for pr_comment_edit
func TestPRCommentEditValidation(t *testing.T) {
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
		errorMsg  string
	}{
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError: true,
			errorMsg:  "repository must be in format 'owner/repo'",
		},
		{
			name: "missing repository",
			arguments: map[string]any{
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError: true,
			errorMsg:  "repository: cannot be blank",
		},
		{
			name: "missing new content",
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
			},
			wantError: true,
			errorMsg:  "new_content: cannot be blank",
		},
		{
			name: "empty new content",
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "",
			},
			wantError: true,
			errorMsg:  "new_content: cannot be blank",
		},
		{
			name: "negative pull request number",
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError: true,
			errorMsg:  "pull_request_number: must be no less than 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tt.arguments,
			})

			if tt.wantError {
				if err != nil {
					t.Errorf("Expected successful call with error result, got client error: %v", err)
				} else if result == nil {
					t.Error("Expected result, got nil")
				} else if !result.IsError {
					t.Errorf("Expected error result for test case '%s', got success result: %+v", tt.name, result)
				} else {
					// Check if error message contains expected text
					errorText := getTextContent(result.Content)
					if !strings.Contains(errorText, tt.errorMsg) {
						t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, errorText)
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

// TestPRCommentEditPermissionError tests permission error scenarios
func TestPRCommentEditPermissionError(t *testing.T) {
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
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          123,
			"new_content":         "Updated content",
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
