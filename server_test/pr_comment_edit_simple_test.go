package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPRCommentEditToolDiscovery tests that the pr_comment_edit tool is properly registered
func TestPRCommentEditToolDiscovery(t *testing.T) {
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

	// List available tools
	tools, err := client.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	// Check that we have expected tools (should now be 9 including pr_comment_edit)
	if len(tools.Tools) != 9 {
		t.Fatalf("Expected 9 tools, got %d", len(tools.Tools))
	}

	// Find the pr_comment_edit tool
	var editTool *mcp.Tool
	for _, tool := range tools.Tools {
		if tool.Name == "pr_comment_edit" {
			editTool = tool
			break
		}
	}

	if editTool == nil {
		t.Fatal("pr_comment_edit tool not found")
	}
	if editTool.Description != "Edit an existing comment on a Forgejo/Gitea repository pull request" {
		t.Errorf("Expected pr_comment_edit tool description 'Edit an existing comment on a Forgejo/Gitea repository pull request', got '%s'", editTool.Description)
	}

	// Verify tool has input schema
	if editTool.InputSchema == nil {
		t.Error("pr_comment_edit tool should have input schema")
	}
}

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
