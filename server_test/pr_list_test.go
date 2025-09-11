package servertest

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type prListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestListPullRequests(t *testing.T) {
	t.Parallel()
	testCases := []prListTestCase{
		{
			name: "acceptance",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Feature: Add dark mode", State: "open"},
					{ID: 2, Number: 2, Title: "Fix: Memory leak", State: "open"},
					{ID: 3, Number: 3, Title: "Bug: Login fails", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":         float64(1),
							"number":     float64(1),
							"title":      "Feature: Add dark mode",
							"body":       "",
							"state":      "open",
							"user":       "testuser",
							"created_at": "2025-09-11T10:30:00Z",
							"updated_at": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":         float64(2),
							"number":     float64(2),
							"title":      "Fix: Memory leak",
							"body":       "",
							"state":      "open",
							"user":       "testuser",
							"created_at": "2025-09-11T10:30:00Z",
							"updated_at": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "pagination",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("testuser", "testrepo", prs)
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 1; i <= 10; i++ {
							prs = append(prs, map[string]any{
								"id":         float64(i),
								"number":     float64(i),
								"title":      fmt.Sprintf("Pull Request %d", i),
								"body":       "",
								"state":      "open",
								"user":       "testuser",
								"created_at": "2025-09-11T10:30:00Z",
								"updated_at": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
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
				"repository": "invalid-repo-format",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{"pull_requests": nil},
				IsError:           true,
			},
		},
		{
			name: "missing repository",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"limit":  10,
				"offset": 0,
				"state":  "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: cannot be blank."},
				},
				StructuredContent: map[string]any{"pull_requests": nil},
				IsError:           true,
			},
		},
		{
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      200, // Invalid: > 100
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: map[string]any{"pull_requests": nil},
				IsError:           true,
			},
		},
		{
			name: "invalid state",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "invalid-state", // Invalid state
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: state: state must be one of: open, closed, all."},
				},
				StructuredContent: map[string]any{"pull_requests": nil},
				IsError:           true,
			},
		},
		{
			name: "default values",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Default Test PR", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":         float64(1),
							"number":     float64(1),
							"title":      "Default Test PR",
							"body":       "",
							"state":      "open",
							"user":       "testuser",
							"created_at": "2025-09-11T10:30:00Z",
							"updated_at": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "invalid offset",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     -1, // Invalid: negative
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
				},
				StructuredContent: map[string]any{"pull_requests": nil},
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
				Name:      "pr_list",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_list tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

func TestListPullRequestsConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Concurrent PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Concurrent PR 2", State: "open"},
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
				Name: "pr_list",
				Arguments: map[string]any{
					"repository": "testuser/testrepo",
					"limit":      10,
					"offset":     0,
					"state":      "open",
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
