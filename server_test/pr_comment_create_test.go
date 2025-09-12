package servertest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type prCommentCreateTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestCreatePullRequestComment(t *testing.T) {
	t.Parallel()
	testCases := []prCommentCreateTestCase{
		{
			name: "successful comment creation",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{}) // Start with no comments
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment":             "This is a helpful comment on the pull request.",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: This is a helpful comment on the pull request."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":         float64(1),
						"body":       "This is a helpful comment on the pull request.",
						"user":       "testuser",
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z",
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
				"comment":             "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{"comment": nil},
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
				"comment":             "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: cannot be blank."},
				},
				StructuredContent: map[string]any{"comment": nil},
				IsError:           true,
			},
		},
		{
			name: "invalid pull request number",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"comment":             "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: pull_request_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{"comment": nil},
				IsError:           true,
			},
		},
		{
			name: "empty comment",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment":             "",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment: cannot be blank."},
				},
				StructuredContent: map[string]any{"comment": nil},
				IsError:           true,
			},
		},
		{
			name: "whitespace only comment",
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment":             "   \n\t   ",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment: cannot be blank."},
				},
				StructuredContent: map[string]any{"comment": nil},
				IsError:           true,
			},
		},
		{
			name: "negative pull request number",
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment":             "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: pull_request_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{"comment": nil},
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
				Name:      "pr_comment_create",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_create tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

func TestCreatePullRequestCommentConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{}) // Start with no comments
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
				Name: "pr_comment_create",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": 1,
					"comment":             "Concurrent comment " + string(rune(commentNum+'0')),
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
