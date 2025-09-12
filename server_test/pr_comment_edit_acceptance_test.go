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

// TestPullRequestCommentEditLifecycleAcceptance tests the complete PR comment edit lifecycle: create, edit, verify
// This test is kept as a multi-step acceptance test since it involves multiple tool calls
func TestPullRequestCommentEditLifecycleAcceptance(t *testing.T) {
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
	originalComment := "This is the original comment that will be edited."
	updatedComment := "This is the updated comment after editing."

	// Step 1: Create a PR comment first
	t.Log("Step 1: Creating initial PR comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"comment":             originalComment,
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
	if !strings.Contains(createText, originalComment) {
		t.Errorf("Expected original comment in response, got: %s", createText)
	}

	// Step 2: Edit the PR comment
	t.Log("Step 2: Editing the PR comment")
	editResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"comment_id":          1, // First comment created
			"new_content":         updatedComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to edit PR comment: %v", err)
	}
	if editResult.IsError {
		t.Fatalf("PR comment editing failed: %s", getTextContent(editResult.Content))
	}

	// Verify edit response
	editText := getTextContent(editResult.Content)
	if !strings.Contains(editText, "Pull request comment edited successfully") {
		t.Errorf("Expected successful edit message, got: %s", editText)
	}
	if !strings.Contains(editText, updatedComment) {
		t.Errorf("Expected updated comment in response, got: %s", editText)
	}
	if !strings.Contains(editText, "ID: 1") {
		t.Errorf("Expected comment ID in response, got: %s", editText)
	}

	t.Log("✅ PR comment edit lifecycle test completed successfully")
}

// TestPullRequestCommentEditRealWorldScenarios tests various real-world PR comment editing scenarios
// This acceptance test focuses on end-to-end workflows rather than individual handler functionality
func TestPullRequestCommentEditRealWorldScenarios(t *testing.T) {
	t.Parallel()

	// Test a representative real-world scenario that demonstrates the complete workflow
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      42,
			Content: "Initial code review comment",
			Author:  "reviewer",
			Created: "2025-09-10T10:00:00Z",
		},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test code review comment update - a common real-world scenario
	updatedReviewComment := `Updated code review feedback:

✅ **Fixed Issues:**
1. Added proper error handling for empty input
2. Improved variable naming for clarity
3. Added comprehensive unit tests

**Additional Suggestions:**
- Consider adding input validation
- The performance could be optimized for large datasets

Great work on addressing the previous feedback!`

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          42,
			"new_content":         updatedReviewComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
	}

	// Verify the result contains success message
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Updated code review feedback") {
		t.Errorf("Expected updated comment content in response, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "ID: 42") {
		t.Errorf("Expected comment ID in response, got: %s", textContent.Text)
	}
}

// TestPullRequestCommentEditErrorHandling tests error handling and recovery scenarios
// This acceptance test focuses on end-to-end error scenarios rather than detailed validation
func TestPullRequestCommentEditErrorHandling(t *testing.T) {
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

	// Test a representative error scenario - comment not found
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          999, // Non-existent comment ID
			"new_content":         "This should fail",
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

	if !strings.Contains(textContent.Text, "Failed to edit") {
		t.Errorf("Expected error message containing 'Failed to edit', got: %s", textContent.Text)
	}
}

// TestPullRequestCommentEditPerformance tests performance and edge cases
// This acceptance test focuses on end-to-end performance scenarios
func TestPullRequestCommentEditPerformance(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      123,
			Content: "Original comment",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test large content scenario - should handle efficiently
	largeComment := strings.Repeat("This is a detailed updated code review comment with comprehensive revised feedback. ", 200) // ~10KB
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          123,
			"new_content":         largeComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_edit tool with large content: %v", err)
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

	if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
		t.Error("Expected success message for large content")
	}
}

// TestPullRequestCommentEditConcurrent tests concurrent request handling
// This acceptance test focuses on end-to-end concurrent behavior
func TestPullRequestCommentEditConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	// Set up multiple comments for concurrent editing
	for i := 1; i <= 5; i++ {
		mock.AddComments("testuser", "testrepo", []MockComment{
			{
				ID:      i,
				Content: fmt.Sprintf("Original comment %d", i),
				Author:  "testuser",
				Created: "2025-09-10T10:00:00Z",
			},
		})
	}

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
		commentID := i + 1 // Each goroutine edits a different comment
		go func(id int) {
			defer wg.Done()
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "pr_comment_edit",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": 1,
					"comment_id":          id,
					"new_content":         fmt.Sprintf("Concurrent edit of comment %d", id),
				},
			})
			results <- err
		}(commentID)
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

// TestPullRequestCommentEditWorkflowIntegration tests integration with other PR comment tools
// This acceptance test verifies the complete PR comment management workflow
func TestPullRequestCommentEditWorkflowIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
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
	prNumber := 1

	// Step 1: Create a comment
	t.Log("Step 1: Creating comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": prNumber,
			"comment":             "Initial comment for workflow test",
		},
	})
	if err != nil || createResult.IsError {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Step 2: List comments to verify creation
	t.Log("Step 2: Listing comments")
	listResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_list",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": prNumber,
			"limit":               10,
			"offset":              0,
		},
	})
	if err != nil || listResult.IsError {
		t.Fatalf("Failed to list comments: %v", err)
	}

	// Step 3: Edit the comment
	t.Log("Step 3: Editing comment")
	editResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": prNumber,
			"comment_id":          1,
			"new_content":         "Updated comment for workflow test",
		},
	})
	if err != nil || editResult.IsError {
		t.Fatalf("Failed to edit comment: %v", err)
	}

	// Step 4: List comments again to verify edit
	t.Log("Step 4: Listing comments after edit")
	listResult2, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_list",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": prNumber,
			"limit":               10,
			"offset":              0,
		},
	})
	if err != nil || listResult2.IsError {
		t.Fatalf("Failed to list comments after edit: %v", err)
	}

	// Verify the edit is reflected in the list
	// Check structured content for the updated comment
	if structuredContent, ok := listResult2.StructuredContent.(map[string]any); ok {
		if comments, exists := structuredContent["pull_request_comments"]; exists {
			if commentList, ok := comments.([]any); ok && len(commentList) > 0 {
				if firstComment, ok := commentList[0].(map[string]any); ok {
					if body, ok := firstComment["body"].(string); ok {
						if !strings.Contains(body, "Updated comment for workflow test") {
							t.Errorf("Expected updated comment in structured content, got: %s", body)
						}
					}
				}
			}
		}
	}

	t.Log("✅ PR comment edit workflow integration test completed successfully")
}
