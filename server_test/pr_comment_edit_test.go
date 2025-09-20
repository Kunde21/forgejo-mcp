package servertest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type prCommentEditTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestEditPullRequestComment(t *testing.T) {
	t.Parallel()
	testCases := []prCommentEditTestCase{
		{
			name: "successful comment edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      123,
						Content: "Original comment content",
						Author:  "testuser",
						Created: "2025-09-10T10:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated comment content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Updated comment content"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "Updated comment content",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
			},
		},
		{
			name: "invalid repository format",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "missing repository",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "missing new content",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: new_content: cannot be blank."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "empty new content",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: new_content: cannot be blank."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "whitespace only new content",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "   \n\t   ",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: new_content: cannot be blank."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "negative pull request number",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: pull_request_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "zero pull request number",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: pull_request_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "negative comment ID",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          -1,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment_id: must be no less than 1."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "zero comment ID",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          0,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment_id: must be no less than 1."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "non-existent repository",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			arguments: map[string]any{
				"repository":          "nonexistent/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to edit pull request comment: failed to edit pull request comment: unknown API error: 404\nRequest: '/api/v1/repos/nonexistent/repo/issues/comments/123' with 'PATCH' method and '404 page not found\n' body"},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "comment not found",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      999,
						Content: "Different comment",
						Author:  "testuser",
						Created: "2025-09-10T10:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to edit pull request comment: failed to edit pull request comment: unknown API error: 404\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and '404 page not found\n' body"},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "permission error",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      123,
						Content: "Original comment",
						Author:  "testuser",
						Created: "2025-09-10T10:00:00Z",
					},
				})
				mock.SetForbiddenCommentEdit(123)
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to edit pull request comment: failed to edit pull request comment: unknown API error: 403\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and 'Forbidden\n' body"},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}
			ts := NewTestServer(t, t.Context(), map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

func TestEditPullRequestCommentConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      123,
			Content: "Original comment",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
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
	for i := range numGoroutines {
		go func(commentNum int) {
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "pr_comment_edit",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": 1,
					"comment_id":          123,
					"new_content":         "Concurrent edit " + string(rune(commentNum+'0')),
				},
			})
			results <- err
		}(i)
	}
	for range numGoroutines {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}

func TestEditPullRequestCommentPermissionError(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      123,
			Content: "Original comment",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
	})
	mock.SetForbiddenCommentEdit(123)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with permission error
	result, err := client.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          123,
			"new_content":         "Updated content",
		},
	})
	// Should return an error result for permission issues
	if err != nil {
		t.Errorf("Expected successful call with error result, got client error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	} else if !result.IsError {
		t.Errorf("Expected error result for permission issue, got success result: %+v", result)
	}
}

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

// TestPullRequestCommentEditCompleteWorkflow tests the complete pull request comment edit workflow
func TestPullRequestCommentEditCompleteWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Set up mock Gitea server with existing comment
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      123,
			Content: "Original comment content",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test successful comment editing
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          123,
			"new_content":         "Updated comment content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
	}

	// Verify response structure
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Content == nil {
		t.Fatal("Expected non-nil content")
	}
	if len(result.Content) == 0 {
		t.Fatal("Expected at least one content item")
	}

	// Verify content type and message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}
	if textContent.Text == "" {
		t.Error("Expected non-empty text content")
	}

	// Should contain success message
	if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}

	// Should contain comment ID
	if !strings.Contains(textContent.Text, "ID: 123") {
		t.Errorf("Expected comment ID in message, got: %s", textContent.Text)
	}

	// Should contain updated comment body
	if !strings.Contains(textContent.Text, "Updated comment content") {
		t.Errorf("Expected updated comment body in message, got: %s", textContent.Text)
	}
}

// TestPullRequestCommentEditValidationErrors tests validation error scenarios
func TestPullRequestCommentEditValidationErrors(t *testing.T) {
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

	tests := []struct {
		name        string
		args        map[string]any
		wantError   bool
		errorSubstr string
	}{
		{
			name: "missing repository",
			args: map[string]any{
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "at least one of directory or repository must be provided",
		},
		{
			name: "invalid repository format",
			args: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "repository: repository must be in format 'owner/repo'",
		},
		{
			name: "missing new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "empty new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "",
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "whitespace only new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "   \n\t   ",
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "zero pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "negative pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "zero comment ID",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          0,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "comment_id: must be no less than 1",
		},
		{
			name: "negative comment ID",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          -1,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "comment_id: must be no less than 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tt.args,
			})

			if tt.wantError {
				if err != nil {
					t.Fatalf("Expected error in result, got call error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected error result, got nil")
				}
				if !result.IsError {
					t.Error("Expected result to be marked as error")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Fatalf("Expected TextContent, got %T", result.Content[0])
				}
				if !strings.Contains(textContent.Text, tt.errorSubstr) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorSubstr, textContent.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected call error: %v", err)
				}
				if result == nil || result.IsError {
					t.Error("Expected successful result")
				}
			}
		})
	}
}

// TestPullRequestCommentEditSuccessfulParameters tests successful comment editing with valid parameters
func TestPullRequestCommentEditSuccessfulParameters(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("owner", "repo", []MockComment{
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
		t.Fatal(err)
	}
	client := ts.Client()

	tests := []struct {
		name     string
		args     map[string]any
		expected string
	}{
		{
			name: "basic comment edit",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated basic comment.",
			},
			expected: "Updated basic comment.",
		},
		{
			name: "comment edit with special characters",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated comment with special chars: @#$%^&*()",
			},
			expected: "Updated comment with special chars: @#$%^&*()",
		},
		{
			name: "multiline comment edit",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated line 1\nUpdated line 2\nUpdated line 3",
			},
			expected: "Updated line 1\nUpdated line 2\nUpdated line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
			}

			if result == nil || result.Content == nil {
				t.Fatal("Expected non-nil result with content")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("Expected TextContent, got %T", result.Content[0])
			}

			// Verify success message
			if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
				t.Errorf("Expected success message, got: %s", textContent.Text)
			}

			// Verify updated comment body is included
			if !strings.Contains(textContent.Text, tt.expected) {
				t.Errorf("Expected updated comment body '%s' in response, got: %s", tt.expected, textContent.Text)
			}
		})
	}
}

// TestPullRequestCommentEditDirectoryParameter tests directory parameter functionality
func TestPullRequestCommentEditDirectoryParameter(t *testing.T) {
	t.Parallel()
	testCases := []prCommentEditTestCase{
		{
			name: "directory parameter with valid git repo",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      123,
						Content: "Original comment content",
						Author:  "testuser",
						Created: "2025-09-10T10:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"directory":           createTempGitRepo(t, "testuser", "testrepo"),
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated comment content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Updated comment content"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "Updated comment content",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
			},
		},
		{
			name: "directory parameter with invalid git repo",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory":           "/nonexistent/path",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "directory parameter with valid git repo and mock setup",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      123,
						Content: "Original comment content",
						Author:  "testuser",
						Created: "2025-09-10T10:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"directory":           createTempGitRepo(t, "testuser", "testrepo"),
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: test content"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "test content",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
			},
		},
		{
			name: "directory parameter with both directory and repository provided (directory takes precedence)",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"directory":           createTempGitRepo(t, "testuser", "testrepo"),
				"repository":          "different/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated comment content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Updated comment content"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "Updated comment content",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
		{
			name: "directory parameter with empty string",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory":           "",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "directory parameter with whitespace only",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory":           "   \n\t   ",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: directory must be an absolute path."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}
			ts := NewTestServer(t, t.Context(), map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}
