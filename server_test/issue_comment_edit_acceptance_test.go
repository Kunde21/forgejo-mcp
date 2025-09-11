package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
func TestCommentEditingRealWorldScenarios(t *testing.T) {
	t.Parallel()
	testCases := []issueCommentEditTestCase{
		{
			name: "fix_typo",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "This is a commment with a typo",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "This is a comment with the typo fixed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: This is a comment with the typo fixed"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "This is a comment with the typo fixed",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name: "add_information",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "I found an issue",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "I found an issue with the login functionality. The error occurs when users try to authenticate with invalid credentials.",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: I found an issue with the login functionality. The error occurs when users try to authenticate with invalid credentials."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "I found an issue with the login functionality. The error occurs when users try to authenticate with invalid credentials.",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name: "correct_misinformation",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "The bug is in the frontend code",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "After further investigation, the bug is actually in the backend authentication service",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: After further investigation, the bug is actually in the backend authentication service"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "After further investigation, the bug is actually in the backend authentication service",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name: "update_status",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "Working on this issue",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "I've completed the implementation and added comprehensive tests. Ready for review.",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: I've completed the implementation and added comprehensive tests. Ready for review."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "I've completed the implementation and added comprehensive tests. Ready for review.",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
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
				Name:      "issue_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestCommentEditingErrorHandling tests error handling and recovery scenarios
func TestCommentEditingErrorHandling(t *testing.T) {
	t.Parallel()
	testCases := []issueCommentEditTestCase{
		{
			name: "nonexistent_repository",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":   "nonexistent/repo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 404\nRequest: '/api/v1/repos/nonexistent/repo/issues/comments/1' with 'PATCH' method and '404 page not found\n' body"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(0),
						"content": "",
						"author":  "",
						"created": "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "invalid_repository_format",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(0),
						"content": "",
						"author":  "",
						"created": "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "missing_required_fields",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				// missing issue_number, comment_id, new_content
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: new_content: cannot be blank."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(0),
						"content": "",
						"author":  "",
						"created": "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "empty_content",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: new_content: cannot be blank."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(0),
						"content": "",
						"author":  "",
						"created": "",
					},
				},
				IsError: true,
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
				Name:      "issue_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestCommentEditingPerformance tests performance and edge cases
func TestCommentEditingPerformance(t *testing.T) {
	t.Parallel()
	testCases := []issueCommentEditTestCase{
		{
			name: "large_content",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "Original content",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  strings.Repeat("This is a test comment with some content. ", 100), // ~4KB
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: " + strings.Repeat("This is a test comment with some content. ", 100)},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": strings.Repeat("This is a test comment with some content. ", 100),
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name: "minimal_content",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "Original content",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "x", // Minimal valid content
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: x"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "x",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
					},
				},
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
				Name:      "issue_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestCommentEditingConcurrent tests concurrent request handling
func TestCommentEditingConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
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

	const numGoroutines = 3
	results := make(chan error, numGoroutines)
	for range numGoroutines {
		go func() {
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "issue_comment_edit",
				Arguments: map[string]any{
					"repository":   "testuser/testrepo",
					"issue_number": 1,
					"comment_id":   1,
					"new_content":  "Concurrent edit content",
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

// getTextContent extracts text content from MCP result
func getTextContent(content []mcp.Content) string {
	for _, c := range content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}
