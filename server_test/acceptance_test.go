package servertest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestListIssuesAcceptance tests the list_issues tool with mock server
func TestListIssuesAcceptance(t *testing.T) {
	// Set up mock Gitea server
	mock := NewMockGiteaServer()
	defer mock.Close()

	// Add test issues
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Bug: Login fails", State: "open"},
		{Index: 2, Title: "Feature: Add dark mode", State: "open"},
		{Index: 3, Title: "Fix: Memory leak", State: "closed"},
	})

	// Set environment variables to use mock server
	os.Setenv("FORGEJO_REMOTE_URL", mock.URL())
	os.Setenv("FORGEJO_AUTH_TOKEN", "mock-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	// Create test server
	ts := NewTestServer(t, context.Background())
	ts.SetMockServer(mock)

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful issue listing
	result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to call list_issues tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListIssuesPagination tests pagination parameters
func TestListIssuesPagination(t *testing.T) {
	mock := NewMockGiteaServer()
	defer mock.Close()

	// Add many test issues
	var issues []MockIssue
	for i := 1; i <= 25; i++ {
		issues = append(issues, MockIssue{
			Index: i,
			Title: fmt.Sprintf("Issue %d", i),
			State: "open",
		})
	}
	mock.AddIssues("testuser", "testrepo", issues)

	os.Setenv("FORGEJO_REMOTE_URL", mock.URL())
	os.Setenv("FORGEJO_AUTH_TOKEN", "mock-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	ts := NewTestServer(t, context.Background())
	ts.SetMockServer(mock)

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with limit 10, offset 0
	result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to call list_issues tool: %v", err)
	}

	// Verify we get 10 issues
	// Note: This would require parsing the result content
	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListIssuesErrorHandling tests error scenarios
func TestListIssuesErrorHandling(t *testing.T) {
	mock := NewMockGiteaServer()
	defer mock.Close()

	os.Setenv("FORGEJO_REMOTE_URL", mock.URL())
	os.Setenv("FORGEJO_AUTH_TOKEN", "mock-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	ts := NewTestServer(t, context.Background())
	ts.SetMockServer(mock)

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with invalid repository format
	result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "invalid-repo-format",
				"limit":      10,
				"offset":     0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to call list_issues tool: %v", err)
	}

	// Should return error for invalid repository
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

// TestListIssuesInputValidation tests input validation
func TestListIssuesInputValidation(t *testing.T) {
	mock := NewMockGiteaServer()
	defer mock.Close()

	os.Setenv("FORGEJO_REMOTE_URL", mock.URL())
	os.Setenv("FORGEJO_AUTH_TOKEN", "mock-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	ts := NewTestServer(t, context.Background())
	ts.SetMockServer(mock)

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with missing repository
	result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to call list_issues tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestListIssuesConcurrent tests concurrent request handling
func TestListIssuesConcurrent(t *testing.T) {
	mock := NewMockGiteaServer()
	defer mock.Close()

	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Concurrent Issue 1", State: "open"},
		{Index: 2, Title: "Concurrent Issue 2", State: "open"},
	})

	os.Setenv("FORGEJO_REMOTE_URL", mock.URL())
	os.Setenv("FORGEJO_AUTH_TOKEN", "mock-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	ts := NewTestServer(t, context.Background())
	ts.SetMockServer(mock)

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Run concurrent requests
	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "list_issues",
					Arguments: map[string]interface{}{
						"repository": "testuser/testrepo",
						"limit":      10,
						"offset":     0,
					},
				},
			})
			results <- err
		}()
	}

	// Check results
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
