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

type issueCommentEditTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

// TestCommentLifecycleAcceptance tests the complete comment lifecycle: create, list, edit
// This test is kept as a multi-step acceptance test since it involves multiple tool calls
func TestCommentLifecycleAcceptance(t *testing.T) {
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
	issueNumber := 1
	originalComment := "This is the original comment for testing."
	updatedComment := "This is the updated comment with new information."

	// Step 1: Create a comment
	t.Log("Step 1: Creating comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"comment":      originalComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("Comment creation failed: %s", getTextContent(createResult.Content))
	}

	// Verify creation response
	createText := getTextContent(createResult.Content)
	if !strings.Contains(createText, "Comment created successfully") {
		t.Errorf("Expected successful creation message, got: %s", createText)
	}
	if !strings.Contains(createText, originalComment) {
		t.Errorf("Expected original comment in response, got: %s", createText)
	}

	// Step 2: List comments to verify creation
	t.Log("Step 2: Listing comments to verify creation")
	listResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to list comments: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("Comment listing failed: %s", getTextContent(listResult.Content))
	}

	// Verify the comment appears in the list
	listText := getTextContent(listResult.Content)
	if !strings.Contains(listText, originalComment) {
		t.Errorf("Expected original comment in list, got: %s", listText)
	}

	// Extract comment ID from the list (assuming it's the first comment)
	// For this test, we'll use a fixed comment ID since our mock server returns predictable IDs
	commentID := 1

	// Step 3: Edit the comment
	t.Log("Step 3: Editing comment")
	editResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"comment_id":   commentID,
			"new_content":  updatedComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to edit comment: %v", err)
	}
	if editResult.IsError {
		t.Fatalf("Comment editing failed: %s", getTextContent(editResult.Content))
	}

	// Verify edit response
	editText := getTextContent(editResult.Content)
	if !strings.Contains(editText, "Comment edited successfully") {
		t.Errorf("Expected successful edit message, got: %s", editText)
	}
	if !strings.Contains(editText, updatedComment) {
		t.Errorf("Expected updated comment in response, got: %s", editText)
	}

	// Step 4: List comments again to verify the edit
	t.Log("Step 4: Listing comments to verify edit")
	listResult2, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to list comments after edit: %v", err)
	}
	if listResult2.IsError {
		t.Fatalf("Comment listing failed after edit: %s", getTextContent(listResult2.Content))
	}

	// Verify the updated comment appears in the list
	listText2 := getTextContent(listResult2.Content)
	if !strings.Contains(listText2, updatedComment) {
		t.Errorf("Expected updated comment in list, got: %s", listText2)
	}
	if strings.Contains(listText2, originalComment) && !strings.Contains(listText2, updatedComment) {
		t.Errorf("Original comment should be replaced by updated comment, got: %s", listText2)
	}

	t.Log("✅ Comment lifecycle test completed successfully")
}

// TestCommentEditingRealWorldScenarios tests various real-world editing scenarios
// This acceptance test focuses on end-to-end workflows rather than individual handler functionality
func TestCommentEditingRealWorldScenarios(t *testing.T) {
	t.Parallel()

	// Test a representative real-world scenario that demonstrates the complete workflow
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      1,
			Content: "Working on this issue",
			Author:  "testuser",
			Created: "2024-01-01T00:00:00Z",
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

	// Test updating status - a common real-world scenario
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   1,
			"new_content":  "I've completed the implementation and added comprehensive tests. Ready for review.",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
	}

	// Verify the result contains success message
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Comment edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Ready for review") {
		t.Errorf("Expected updated content in response, got: %s", textContent.Text)
	}
}

// TestCommentEditingErrorHandling tests error handling and recovery scenarios
// This acceptance test focuses on end-to-end error scenarios rather than detailed validation
func TestCommentEditingErrorHandling(t *testing.T) {
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
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "invalid-format",
			"issue_number": 1,
			"comment_id":   1,
			"new_content":  "test content",
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

// TestCommentEditingPerformance tests performance and edge cases
// This acceptance test focuses on end-to-end performance scenarios
func TestCommentEditingPerformance(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      1,
			Content: "Original content",
			Author:  "testuser",
			Created: "2024-01-01T00:00:00Z",
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
	largeContent := strings.Repeat("This is a test comment with some content. ", 100) // ~4KB
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   1,
			"new_content":  largeContent,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_edit tool with large content: %v", err)
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

	if !strings.Contains(textContent.Text, largeContent) {
		t.Error("Expected large content to be present in response")
	}
}

// TestCommentEditingConcurrent tests concurrent request handling
// This acceptance test focuses on end-to-end concurrent behavior
func TestCommentEditingConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	// Add multiple comments to avoid conflicts in concurrent editing
	for i := 1; i <= 3; i++ {
		mock.AddComments("testuser", "testrepo", []MockComment{
			{
				ID:      i,
				Content: fmt.Sprintf("Original comment %d", i),
				Author:  "testuser",
				Created: "2024-01-01T00:00:00Z",
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

	const numGoroutines = 3
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		commentID := i + 1 // Each goroutine edits a different comment
		go func(id int) {
			defer wg.Done()
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "issue_comment_edit",
				Arguments: map[string]any{
					"repository":   "testuser/testrepo",
					"issue_number": 1,
					"comment_id":   id,
					"new_content":  fmt.Sprintf("Concurrent edit content for comment %d", id),
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

// getTextContent extracts text content from MCP result
func getTextContent(content []mcp.Content) string {
	for _, c := range content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}
