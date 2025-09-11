package servertest

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestEditIssueCommentAcceptance tests the issue_comment_edit tool with mock server
func TestEditIssueCommentAcceptance(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful comment editing
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated comment content for acceptance testing",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected successful comment editing, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	// Verify response contains expected information
	expectedParts := []string{"Comment edited successfully", "ID:", "Updated:"}
	for _, part := range expectedParts {
		if !strings.Contains(textContent.Text, part) {
			t.Errorf("Expected response to contain '%s', got: %s", part, textContent.Text)
		}
	}
}

// TestEditIssueCommentInputValidation tests input validation for comment editing
func TestEditIssueCommentInputValidation(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with missing repository
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestEditIssueCommentEmptyContent tests empty content validation
func TestEditIssueCommentEmptyContent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with empty new content
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for empty new content")
	}
}

// TestEditIssueCommentInvalidCommentID tests invalid comment ID
func TestEditIssueCommentInvalidCommentID(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with invalid comment ID (zero)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   0,
			"new_content":  "Updated content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for invalid comment ID")
	}
}

// TestEditIssueCommentInvalidRepository tests invalid repository format
func TestEditIssueCommentInvalidRepository(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with invalid repository format
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "invalid-repo-format",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for invalid repository format")
	}
}

// TestEditIssueCommentToolRegistration tests that the tool is properly registered
func TestEditIssueCommentToolRegistration(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that the tool is available
	tools, err := ts.Client().ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	// Find the issue_comment_edit tool
	var foundTool *mcp.Tool
	for _, tool := range tools.Tools {
		if tool.Name == "issue_comment_edit" {
			foundTool = tool
			break
		}
	}

	if foundTool == nil {
		t.Error("issue_comment_edit tool not found in tool list")
	}

	// Verify tool metadata
	if foundTool != nil {
		expectedDescription := "Edit an existing comment on a Forgejo/Gitea repository issue"
		if foundTool.Description != expectedDescription {
			t.Errorf("Expected tool description '%s', got '%s'", expectedDescription, foundTool.Description)
		}
	}
}
