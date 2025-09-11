package servertest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPullRequestCommentLifecycleAcceptance tests the complete PR comment lifecycle: create, list
// This test is kept as a multi-step acceptance test since it involves multiple tool calls
func TestPullRequestCommentLifecycleAcceptance(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
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

	repo := "testuser/testrepo"
	pullRequestNumber := 1
	testComment := "This is a test comment on a pull request for acceptance testing."

	// Step 1: Create a PR comment
	t.Log("Step 1: Creating PR comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"comment":             testComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create PR comment: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("PR comment creation failed: %s", getTextContent(createResult.Content))
	}

	// Verify creation response
	createText := getTextContent(createResult.Content)
	if !strings.Contains(createText, "Pull request comment created successfully") {
		t.Errorf("Expected successful creation message, got: %s", createText)
	}
	if !strings.Contains(createText, testComment) {
		t.Errorf("Expected comment in response, got: %s", createText)
	}

	// Step 2: List PR comments to verify creation
	t.Log("Step 2: Listing PR comments to verify creation")
	listResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_list",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"limit":               10,
			"offset":              0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to list PR comments: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("PR comment listing failed: %s", getTextContent(listResult.Content))
	}

	// Verify the comment appears in the list
	listText := getTextContent(listResult.Content)
	if !strings.Contains(listText, testComment) {
		t.Errorf("Expected comment in list, got: %s", listText)
	}

	t.Log("✅ PR comment lifecycle test completed successfully")
}

// TestPullRequestCommentCreationRealWorldScenarios tests various real-world PR comment creation scenarios
// This acceptance test focuses on end-to-end workflows rather than individual handler functionality
func TestPullRequestCommentCreationRealWorldScenarios(t *testing.T) {
	t.Parallel()

	// Test a representative real-world scenario that demonstrates the complete workflow
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test code review comment - a common real-world scenario
	codeReviewComment := `I've reviewed your changes and have a few suggestions:

1. Consider adding error handling for the edge case where the input is empty
2. The function could benefit from more descriptive variable names
3. Add unit tests for the new functionality

Overall, great work on this feature!`

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 42,
			"comment":             codeReviewComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_create tool: %v", err)
	}

	// Verify the result contains success message
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "I've reviewed your changes") {
		t.Errorf("Expected comment content in response, got: %s", textContent.Text)
	}
}

// TestPullRequestCommentCreationErrorHandling tests error handling and recovery scenarios
// This acceptance test focuses on end-to-end error scenarios rather than detailed validation
func TestPullRequestCommentCreationErrorHandling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test a representative error scenario - invalid repository format
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "invalid-format",
			"pull_request_number": 1,
			"comment":             "test comment",
		},
	})
	// Should return an error result
	if err != nil {
		t.Fatalf("Expected error in result, got call error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	if !result.IsError {
		t.Error("Expected result to be marked as error")
	}

	// Verify error message
	if len(result.Content) == 0 {
		t.Fatal("Expected error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Invalid request") {
		t.Errorf("Expected error message containing 'Invalid request', got: %s", textContent.Text)
	}
}

// TestPullRequestCommentCreationPerformance tests performance and edge cases
// This acceptance test focuses on end-to-end performance scenarios
func TestPullRequestCommentCreationPerformance(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test large content scenario - should handle efficiently
	largeComment := strings.Repeat("This is a detailed code review comment with comprehensive feedback. ", 200) // ~10KB
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment":             largeComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_create tool with large content: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatal("Expected successful result with large content")
	}

	// Verify the result contains the large content
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
		t.Error("Expected success message for large content")
	}
}

// TestPullRequestCommentCreationConcurrent tests concurrent request handling
// This acceptance test focuses on end-to-end concurrent behavior
func TestPullRequestCommentCreationConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		prNumber := i + 1 // Each goroutine comments on a different PR
		go func(prNum int) {
			defer wg.Done()
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "pr_comment_create",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": prNum,
					"comment":             fmt.Sprintf("Concurrent comment on PR %d", prNum),
				},
			})
			results <- err
		}(prNumber)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check results
	for err := range results {
		if err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
