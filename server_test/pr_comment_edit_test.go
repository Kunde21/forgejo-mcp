package servertest

import (
	"context"
	"testing"

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
						"id":         float64(123),
						"body":       "Updated comment content",
						"user":       "testuser",
						"created_at": "2025-09-10T10:00:00Z",
						"updated_at": "2025-09-10T10:00:00Z",
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
					&mcp.TextContent{Text: "Invalid request: repository: cannot be blank."},
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
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "invalid-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with invalid token (simulates permission error)
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
