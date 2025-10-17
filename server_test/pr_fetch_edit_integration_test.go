package servertest

import (
	"strings"
	"testing"
)

// TestPRFetchEdit_Integration tests that a PR can be fetched and then edited using the body
func TestPRFetchEdit_Integration(t *testing.T) {
	ctx, _ := CreateTestContext(t, 0)

	// Create mock server
	mock := NewMockGiteaServer(t)

	// Add mock pull request
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{
			ID:        100,
			Number:    100,
			Title:     "Original PR Title",
			Body:      "Original PR body content.",
			State:     "open",
			BaseRef:   "main",
			UpdatedAt: "2025-10-14T16:00:00Z",
		},
	})

	// Create test server
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})

	// Step 1: Fetch the PR
	fetchResult, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 100,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool: %v", err)
	}

	if fetchResult.IsError {
		t.Fatalf("Expected success on fetch, got error: %s", GetTextContent(fetchResult.Content))
	}

	// Verify the body is in the fetch response
	fetchStructured := GetStructuredContent(fetchResult)
	if fetchStructured == nil {
		t.Fatal("Expected structured content in fetch response, got nil")
	}

	pr, ok := fetchStructured["pull_request"].(map[string]any)
	if !ok {
		t.Fatal("Expected pull_request in structured content")
	}

	originalBody, ok := pr["body"].(string)
	if !ok {
		t.Fatal("Expected body to be a string")
	}

	if originalBody != "Original PR body content." {
		t.Errorf("Expected original body 'Original PR body content.', got '%v'", originalBody)
	}

	// Step 2: Edit the PR using the fetched body (modify it)
	newBody := originalBody + "\n\nUpdated with additional information."
	editResult, err := ts.CallToolWithValidation(ctx, "pr_edit", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 100,
		"body":                newBody,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_edit tool: %v", err)
	}

	if editResult.IsError {
		t.Fatalf("Expected success on edit, got error: %s", GetTextContent(editResult.Content))
	}

	// Step 3: Fetch the PR again to verify the edit
	fetchResult2, err := ts.CallToolWithValidation(ctx, "pr_fetch", map[string]any{
		"repository":          "testuser/testrepo",
		"pull_request_number": 100,
	})
	if err != nil {
		t.Fatalf("Failed to call pr_fetch tool after edit: %v", err)
	}

	if fetchResult2.IsError {
		t.Fatalf("Expected success on second fetch, got error: %s", GetTextContent(fetchResult2.Content))
	}

	// Verify the body was updated
	fetchStructured2 := GetStructuredContent(fetchResult2)
	if fetchStructured2 == nil {
		t.Fatal("Expected structured content in second fetch response, got nil")
	}

	pr2, ok := fetchStructured2["pull_request"].(map[string]any)
	if !ok {
		t.Fatal("Expected pull_request in structured content")
	}

	updatedBody, ok := pr2["body"].(string)
	if !ok {
		t.Fatal("Expected body to be a string")
	}

	if updatedBody != newBody {
		t.Errorf("Expected updated body '%v', got '%v'", newBody, updatedBody)
	}

	// Verify the updated body is in the text response
	textContent := GetTextContent(fetchResult2.Content)
	if !strings.Contains(textContent, "Updated with additional information.") {
		t.Errorf("Expected updated body content in text response, got: %s", textContent)
	}
}
