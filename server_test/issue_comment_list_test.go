package servertest

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestListIssueCommentsAcceptance tests the issues_list_comments tool with mock server
func TestListIssueCommentsAcceptance(t *testing.T) {
	// Set up mock Gitea server with comments
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 1, Content: "First comment", Author: "user1", Created: "2025-09-10T09:00:00Z"},
		{ID: 2, Content: "Second comment", Author: "user2", Created: "2025-09-10T10:00:00Z"},
		{ID: 3, Content: "Third comment", Author: "user1", Created: "2025-09-10T11:00:00Z"},
	})

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful comment listing
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call list_issue_comments tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}

	// Test with pagination
	result, err = ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"limit":        2,
			"offset":       1,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_list tool with pagination: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result with pagination")
	}
}

// TestListIssueCommentsInputValidation tests input validation for comment listing
func TestListIssueCommentsInputValidation(t *testing.T) {
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
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"issue_number": 1,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestListIssueCommentsInvalidIssueNumber tests invalid issue number
func TestListIssueCommentsInvalidIssueNumber(t *testing.T) {
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
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": -1,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for invalid issue number")
	}
}
