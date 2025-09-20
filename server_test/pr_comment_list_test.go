package servertest

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type prCommentListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestListPullRequestComments(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []prCommentListTestCase{
		{
			name: "acceptance",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "This is a great PR!",
						Author:  "reviewer1",
						Created: "2024-01-01T00:00:00Z",
						Updated: "2024-01-01T00:00:00Z",
					},
					{
						ID:      2,
						Content: "I agree, well done!",
						Author:  "reviewer2",
						Created: "2024-01-02T00:00:00Z",
						Updated: "2024-01-02T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"limit":               10,
				"offset":              0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 comments (showing 1-2):\nComment 1 (ID: 1): This is a great PR!\nComment 2 (ID: 2): I agree, well done!\n"},
				},
				StructuredContent: map[string]any{
					"pull_request_comments": []any{
						map[string]any{
							"id":      float64(1),
							"body":    "This is a great PR!",
							"user":    "reviewer1",
							"created": "2024-01-01T00:00:00Z",
							"updated": "2024-01-01T00:00:00Z",
						},
						map[string]any{
							"id":      float64(2),
							"body":    "I agree, well done!",
							"user":    "reviewer2",
							"created": "2024-01-02T00:00:00Z",
							"updated": "2024-01-02T00:00:00Z",
						},
					},
				},
			},
		},
		{
			name: "pagination",
			setupMock: func(mock *MockGiteaServer) {
				var comments []MockComment
				for i := 1; i <= 25; i++ {
					comments = append(comments, MockComment{
						ID:      i,
						Content: "Comment " + string(rune(i+'0')),
						Author:  "user" + string(rune(i+'0')),
						Created: "2024-01-01T00:00:00Z",
						Updated: "2024-01-01T00:00:00Z",
					})
				}
				mock.AddComments("testuser", "testrepo", comments)
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"limit":               10,
				"offset":              0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 comments (showing 1-10):\nComment 1 (ID: 1): Comment 1\nComment 2 (ID: 2): Comment 2\nComment 3 (ID: 3): Comment 3\nComment 4 (ID: 4): Comment 4\nComment 5 (ID: 5): Comment 5\nComment 6 (ID: 6): Comment 6\nComment 7 (ID: 7): Comment 7\nComment 8 (ID: 8): Comment 8\nComment 9 (ID: 9): Comment 9\nComment 10 (ID: 10): Comment :\n"},
				},
				StructuredContent: map[string]any{
					"pull_request_comments": func() []any {
						var comments []any
						for i := 1; i <= 10; i++ {
							comments = append(comments, map[string]any{
								"id":      float64(i),
								"body":    "Comment " + string(rune(i+'0')),
								"user":    "user" + string(rune(i+'0')),
								"created": "2024-01-01T00:00:00Z",
								"updated": "2024-01-01T00:00:00Z",
							})
						}
						return comments
					}(),
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
				"limit":               10,
				"offset":              0,
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
			name: "missing repository and directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"pull_request_number": 1,
				"limit":               10,
				"offset":              0,
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
			name: "invalid pull request number",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"limit":               10,
				"offset":              0,
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
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"limit":               200, // Invalid: > 100
				"offset":              0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "default values",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Default test comment", Author: "testuser", Created: "2024-01-01T00:00:00Z", Updated: "2024-01-01T00:00:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 comments (showing 1-1):\nComment 1 (ID: 1): Default test comment\n"},
				},
				StructuredContent: map[string]any{
					"pull_request_comments": []any{
						map[string]any{
							"id":      float64(1),
							"body":    "Default test comment",
							"user":    "testuser",
							"created": "2024-01-01T00:00:00Z",
							"updated": "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		},
		{
			name: "invalid offset",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"limit":               10,
				"offset":              -1, // Invalid: negative
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
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

			result, err := ts.Client().CallTool(t.Context(), &mcp.CallToolParams{
				Name:      "pr_comment_list",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_list tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
			if t.Failed() {
				fmt.Println("FAILED")
			}
		})
	}
}

func TestListPullRequestCommentsConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 1, Content: "Concurrent comment 1", Author: "user1", Created: "2024-01-01T00:00:00Z"},
		{ID: 2, Content: "Concurrent comment 2", Author: "user2", Created: "2024-01-02T00:00:00Z"},
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
				Name: "pr_comment_list",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": 1,
					"limit":               10,
					"offset":              0,
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
