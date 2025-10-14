package servertest

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPRFetch_Success tests successful PR fetch
func TestPRFetch_Success(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Add mock pull requests
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{
			ID:        1,
			Number:    1,
			Title:     "Add new feature",
			Body:      "This PR adds a new feature to the application.",
			State:     "open",
			BaseRef:   "main",
			UpdatedAt: "2025-10-14T12:00:00Z",
		},
		{
			ID:        2,
			Number:    2,
			Title:     "Fix bug in authentication",
			Body:      "This PR fixes a critical bug in the authentication system.",
			State:     "closed",
			BaseRef:   "main",
			UpdatedAt: "2025-10-13T15:30:00Z",
		},
	})

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Test fetching PR #1
	result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 1,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool: %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected success, got error: %s", GetTextContent(result.Content))
	}

	// Validate structured content
	structured := GetStructuredContent(result)
	if structured == nil {
		t.Fatal("Expected structured content, got nil")
	}

	pr, ok := structured["pull_request"].(map[string]any)
	if !ok {
		t.Fatal("Expected pull_request in structured content")
	}

	// Validate PR details
	if pr["number"] != float64(1) {
		t.Errorf("Expected PR number 1, got %v", pr["number"])
	}
	if pr["title"] != "Add new feature" {
		t.Errorf("Expected title 'Add new feature', got %v", pr["title"])
	}
	if pr["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", pr["state"])
	}
	if pr["body"] != "This PR adds a new feature to the application." {
		t.Errorf("Expected body about new feature, got %v", pr["body"])
	}

	// Test fetching PR #2
	result2, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 2,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool for PR #2: %v", err)
	}

	if result2.IsError {
		t.Fatalf("Expected success for PR #2, got error: %s", GetTextContent(result2.Content))
	}

	structured2 := GetStructuredContent(result2)
	pr2, ok := structured2["pull_request"].(map[string]any)
	if !ok {
		t.Fatal("Expected pull_request in structured content for PR #2")
	}

	if pr2["number"] != float64(2) {
		t.Errorf("Expected PR number 2, got %v", pr2["number"])
	}
	if pr2["title"] != "Fix bug in authentication" {
		t.Errorf("Expected title 'Fix bug in authentication', got %v", pr2["title"])
	}
	if pr2["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", pr2["state"])
	}
}

// TestPRFetch_NotFound tests PR fetch when PR doesn't exist
func TestPRFetch_NotFound(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Add mock pull requests (only PR #1)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{
			ID:        1,
			Number:    1,
			Title:     "Add new feature",
			Body:      "This PR adds a new feature to the application.",
			State:     "open",
			BaseRef:   "main",
			UpdatedAt: "2025-10-14T12:00:00Z",
		},
	})

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Test fetching non-existent PR #999
	result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 999,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error for non-existent PR, got success")
	}

	errorText := GetTextContent(result.Content)
	if !containsString(errorText, "not found") && !containsString(errorText, "404") {
		t.Errorf("Expected 'not found' or '404' in error message, got: %s", errorText)
	}
}

// TestPRFetch_RepositoryNotFound tests PR fetch when repository doesn't exist
func TestPRFetch_RepositoryNotFound(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Mark repository as not found
	mock.SetNotFoundRepo("nonexistent", "repo")

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Test fetching PR from non-existent repository
	result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "nonexistent/repo",
		"pull_request_number": 1,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error for non-existent repository, got success")
	}

	errorText := GetTextContent(result.Content)
	if !containsString(errorText, "not found") && !containsString(errorText, "404") {
		t.Errorf("Expected 'not found' or '404' in error message, got: %s", errorText)
	}
}

// TestPRFetch_InvalidRepository tests PR fetch with invalid repository format
func TestPRFetch_InvalidRepository(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	testCases := []struct {
		name       string
		repository string
	}{
		{"missing_owner", "/repo"},
		{"missing_repo", "owner/"},
		{"no_slash", "ownerrepo"},
		{"too_many_slashes", "owner/repo/extra"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
				"repository":          tc.repository,
				"pull_request_number": 1,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_fetch tool: %v", err)
			}

			if !result.IsError {
				t.Errorf("Expected error for invalid repository '%s', got success", tc.repository)
			}

			errorText := GetTextContent(result.Content)
			if !containsString(errorText, "repository must be in format") {
				t.Errorf("Expected 'repository must be in format' in error message, got: %s", errorText)
			}
		})
	}
}

// TestPRFetch_InvalidNumber tests PR fetch with invalid PR number
func TestPRFetch_InvalidNumber(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	testCases := []struct {
		name   string
		number any
	}{
		{"zero", 0},
		{"negative", -1},
		{"string", "invalid"},
		{"float", 1.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": tc.number,
			})
			// For JSON unmarshaling errors, err will be non-nil
			if err != nil {
				if !containsString(err.Error(), "unmarshal") {
					t.Fatalf("Expected unmarshaling error for invalid PR number '%v', got: %v", tc.number, err)
				}
				return
			}

			if !result.IsError {
				t.Errorf("Expected error for invalid PR number '%v', got success", tc.number)
			}

			errorText := GetTextContent(result.Content)
			if !containsString(errorText, "pull request number") && !containsString(errorText, "no less than 1") {
				t.Errorf("Expected 'pull request number' or 'no less than 1' in error message, got: %s", errorText)
			}
		})
	}
}

// TestPRFetch_MissingParameters tests PR fetch with missing required parameters
func TestPRFetch_MissingParameters(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	testCases := []struct {
		name       string
		parameters map[string]any
		errorText  string
	}{
		{
			name:       "missing_repository",
			parameters: map[string]any{"pull_request_number": 1},
			errorText:  "repository",
		},
		{
			name:       "missing_number",
			parameters: map[string]any{"repository": "testuser/testrepo"},
			errorText:  "pull request number",
		},
		{
			name:       "empty_parameters",
			parameters: map[string]any{},
			errorText:  "repository",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ts.CallToolWithValidation(ctx, "pr_fetch", tc.parameters)
			if err != nil {
				t.Fatalf("Failed to call pr_fetch tool: %v", err)
			}

			if !result.IsError {
				t.Error("Expected error for missing parameters, got success")
			}

			errorText := GetTextContent(result.Content)
			if !containsString(errorText, tc.errorText) {
				t.Errorf("Expected '%s' in error message, got: %s", tc.errorText, errorText)
			}
		})
	}
}

// TestPRFetch_ToolDiscovery tests that pr_fetch tool is properly discovered
func TestPRFetch_ToolDiscovery(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Get tools list
	tools, err := ts.Client().ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	// Check that pr_fetch tool is in the list
	found := false
	for _, tool := range tools.Tools {
		if tool.Name == "pr_fetch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("pr_fetch tool not found in tools list")
	}
}

// TestPRFetch_DetailedResponse tests that the PR fetch response includes all expected fields
func TestPRFetch_DetailedResponse(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Add mock pull request with detailed information
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{
			ID:        42,
			Number:    42,
			Title:     "Detailed PR Test",
			Body:      "This is a detailed PR for testing all response fields.",
			State:     "open",
			BaseRef:   "develop",
			UpdatedAt: "2025-10-14T14:30:00Z",
		},
	})

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Fetch the PR
	result, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 42,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool: %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected success, got error: %s", GetTextContent(result.Content))
	}

	// Validate structured content
	structured := GetStructuredContent(result)
	if structured == nil {
		t.Fatal("Expected structured content, got nil")
	}

	pr, ok := structured["pull_request"].(map[string]any)
	if !ok {
		t.Fatal("Expected pull_request in structured content")
	}

	// Check for expected fields in the response
	expectedFields := []string{
		"number", "title", "body", "state", "user",
		"base", "head", "html_url", "diff_url", "patch_url",
		"created", "updated", "mergeable", "has_merged", "comments",
	}

	for _, field := range expectedFields {
		if _, exists := pr[field]; !exists {
			t.Errorf("Expected field '%s' not found in PR response", field)
		}
	}

	// Validate specific values
	if pr["number"] != float64(42) {
		t.Errorf("Expected PR number 42, got %v", pr["number"])
	}
	if pr["title"] != "Detailed PR Test" {
		t.Errorf("Expected title 'Detailed PR Test', got %v", pr["title"])
	}

	// Check base branch
	if base, ok := pr["base"].(map[string]any); ok {
		if base["ref"] != "main" { // Mock returns main, not develop
			t.Errorf("Expected base ref 'main', got %v", base["ref"])
		}
	} else {
		t.Errorf("Expected base to be a map, got %v", pr["base"])
	}

	// Check user/author - it's a string in PullRequestDetails, not a map
	if user, ok := pr["user"].(string); ok {
		if user != "testuser" {
			t.Errorf("Expected user 'testuser', got %v", user)
		}
	} else {
		t.Errorf("Expected user to be a string, got %v", pr["user"])
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsStringRecursive(s, substr)))
}

func containsStringRecursive(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsStringRecursive(s[1:], substr)
}
