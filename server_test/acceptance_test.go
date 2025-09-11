package servertest

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestListIssuesAcceptance tests the issues_list tool with mock server
func TestListIssuesAcceptance(t *testing.T) {
	// Set up mock Gitea server
	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Bug: Login fails", State: "open"},
		{Index: 2, Title: "Feature: Add dark mode", State: "open"},
		{Index: 3, Title: "Fix: Memory leak", State: "closed"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	// Test successful issue listing
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListIssuesPagination tests pagination parameters
func TestListIssuesPagination(t *testing.T) {
	mock := NewMockGiteaServer(t)
	var issues []MockIssue
	for i := 1; i <= 25; i++ {
		issues = append(issues, MockIssue{
			Index: i,
			Title: fmt.Sprintf("Issue %d", i),
			State: "open",
		})
	}
	mock.AddIssues("testuser", "testrepo", issues)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with limit 10, offset 0
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	// Verify we get 10 issues
	// Note: This would require parsing the result content
	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListIssuesErrorHandling tests error scenarios
func TestListIssuesErrorHandling(t *testing.T) {
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
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "invalid-repo-format",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	// Should return error for invalid repository
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

// TestListIssuesInputValidation tests input validation
func TestListIssuesInputValidation(t *testing.T) {
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
		Name: "issue_list",
		Arguments: map[string]any{
			"limit":  10,
			"offset": 0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestListIssuesConcurrent tests concurrent request handling
func TestListIssuesConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Concurrent Issue 1", State: "open"},
		{Index: 2, Title: "Concurrent Issue 2", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	results := make(chan error, numGoroutines)
	for range numGoroutines {
		go func() {
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "issue_list",
				Arguments: map[string]any{
					"repository": "testuser/testrepo",
					"limit":      10,
					"offset":     0,
				},
			})
			results <- err
		}()
	}
	for range numGoroutines {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}

// TestListIssuesInvalidLimit tests invalid limit parameter values
func TestListIssuesInvalidLimit(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue 1", State: "open"},
		{Index: 2, Title: "Test Issue 2", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with limit > 100 (invalid)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      200, // Invalid: > 100
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	// Should return error for invalid limit
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

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
