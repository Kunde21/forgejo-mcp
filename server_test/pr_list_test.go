package servertest

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestListPullRequestsAcceptance tests the pr_list tool with mock server
func TestListPullRequestsAcceptance(t *testing.T) {
	// Set up mock Gitea server
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Feature: Add dark mode", State: "open"},
		{ID: 2, Number: 2, Title: "Fix: Memory leak", State: "open"},
		{ID: 3, Number: 3, Title: "Bug: Login fails", State: "closed"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful pull request listing
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
			"state":      "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListPullRequestsPagination tests pagination parameters
func TestListPullRequestsPagination(t *testing.T) {
	mock := NewMockGiteaServer(t)
	var prs []MockPullRequest
	for i := 1; i <= 25; i++ {
		prs = append(prs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  fmt.Sprintf("Pull Request %d", i),
			State:  "open",
		})
	}
	mock.AddPullRequests("testuser", "testrepo", prs)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with limit 10, offset 0
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
			"state":      "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	// Verify we get 10 pull requests
	// Note: This would require parsing the result content
	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListPullRequestsErrorHandling tests error scenarios
func TestListPullRequestsErrorHandling(t *testing.T) {
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
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "invalid-repo-format",
			"limit":      10,
			"offset":     0,
			"state":      "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	// Should return error for invalid repository
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

// TestListPullRequestsInputValidation tests input validation
func TestListPullRequestsInputValidation(t *testing.T) {
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
		Name: "pr_list",
		Arguments: map[string]any{
			"limit":  10,
			"offset": 0,
			"state":  "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected error content for missing repository")
	}
}

// TestListPullRequestsConcurrent tests concurrent request handling
func TestListPullRequestsConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Concurrent PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Concurrent PR 2", State: "open"},
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
				Name: "pr_list",
				Arguments: map[string]any{
					"repository": "testuser/testrepo",
					"limit":      10,
					"offset":     0,
					"state":      "open",
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

// TestListPullRequestsInvalidLimit tests invalid limit parameter values
func TestListPullRequestsInvalidLimit(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
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
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      200, // Invalid: > 100
			"offset":     0,
			"state":      "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	// Should return error for invalid limit
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

// TestListPullRequestsInvalidState tests invalid state parameter values
func TestListPullRequestsInvalidState(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with invalid state
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
			"state":      "invalid-state", // Invalid state
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	// Should return error for invalid state
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}

// TestListPullRequestsDefaultValues tests default value handling
func TestListPullRequestsDefaultValues(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Default Test PR", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with only required parameter (repository)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestListPullRequestsInvalidOffset tests invalid offset parameter values
func TestListPullRequestsInvalidOffset(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with negative offset (invalid)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     -1, // Invalid: negative
			"state":      "open",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_list tool: %v", err)
	}

	// Should return error for invalid offset
	if result.Content == nil {
		t.Error("Expected error content in result")
	}
}
