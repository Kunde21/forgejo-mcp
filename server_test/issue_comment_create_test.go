package servertest

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCreateIssueCommentAcceptance tests the issue_comment_create tool with mock server
func TestCreateIssueCommentAcceptance(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful comment creation
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment":      "This is a test comment for acceptance testing",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected successful comment creation, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	// Verify response contains expected information
	expectedParts := []string{"Comment created successfully", "ID:", "Created:"}
	for _, part := range expectedParts {
		if !strings.Contains(textContent.Text, part) {
			t.Errorf("Expected response to contain '%s', got: %s", part, textContent.Text)
		}
	}
}

// TestCreateIssueCommentInputValidation tests input validation for comment creation
func TestCreateIssueCommentInputValidation(t *testing.T) {
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
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"issue_number": 1,
			"comment":      "Test comment",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestCreateIssueCommentEmptyComment tests empty comment validation
func TestCreateIssueCommentEmptyComment(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with empty comment
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment":      "",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for empty comment")
	}
}

// TestCreateIssueCommentInvalidIssueNumber tests invalid issue number
func TestCreateIssueCommentInvalidIssueNumber(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with invalid issue number (negative)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": -1,
			"comment":      "Test comment",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for invalid issue number")
	}
}

// TestCreateIssueCommentInvalidRepository tests invalid repository format
func TestCreateIssueCommentInvalidRepository(t *testing.T) {
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
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "invalid-repo-format",
			"issue_number": 1,
			"comment":      "Test comment",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for invalid repository format")
	}
}

// TestCreateIssueCommentLongComment tests long comment handling
func TestCreateIssueCommentLongComment(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Create a long comment (over 1000 characters)
	longComment := strings.Repeat("This is a very long comment that should test the handling of large comment bodies. ", 50)
	longComment += "End of long comment."

	// Test successful long comment creation
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment":      longComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool with long comment: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected successful long comment creation, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	// Verify response contains expected information
	expectedParts := []string{"Comment created successfully", "ID:", "Created:"}
	for _, part := range expectedParts {
		if !strings.Contains(textContent.Text, part) {
			t.Errorf("Expected response to contain '%s', got: %s", part, textContent.Text)
		}
	}

	// Verify the comment content is in the response
	if !strings.Contains(textContent.Text, "Comment body:") {
		t.Error("Expected response to contain comment body")
	}
}
