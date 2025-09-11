package servertest

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type issueListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestListIssues(t *testing.T) {
	t.Parallel()
	testCases := []issueListTestCase{
		{
			name: "acceptance",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Bug: Login fails", State: "open"},
					{Index: 2, Title: "Feature: Add dark mode", State: "open"},
					{Index: 3, Title: "Fix: Memory leak", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 3 issues"},
				},
				StructuredContent: map[string]any{
					"issues": []any{
						map[string]any{"number": float64(0), "title": "Bug: Login fails", "state": "open"},
						map[string]any{"number": float64(0), "title": "Feature: Add dark mode", "state": "open"},
						map[string]any{"number": float64(0), "title": "Fix: Memory leak", "state": "closed"},
					},
				},
			},
		},
		{
			name: "pagination",
			setupMock: func(mock *MockGiteaServer) {
				var issues []MockIssue
				for i := 1; i <= 25; i++ {
					issues = append(issues, MockIssue{
						Index: i,
						Title: fmt.Sprintf("Issue %d", i),
						State: "open",
					})
				}
				mock.AddIssues("testuser", "testrepo", issues)
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 25 issues"},
				},
				StructuredContent: map[string]any{
					"issues": func() []any {
						var issues []any
						for i := 1; i <= 25; i++ {
							issues = append(issues, map[string]any{
								"number": float64(0),
								"title":  fmt.Sprintf("Issue %d", i),
								"state":  "open",
							})
						}
						return issues
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
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{"issues": nil},
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
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: cannot be blank."},
				},
				StructuredContent: map[string]any{"issues": nil},
				IsError:           true,
			},
		},
		{
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Test Issue 1", State: "open"},
					{Index: 2, Title: "Test Issue 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      200, // Invalid: > 100
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: map[string]any{"issues": nil},
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
				Name:      "issue_list",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_list tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestListIssuesConcurrent tests concurrent request handling
func TestListIssuesConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Concurrent Issue 1", State: "open"},
		{Index: 2, Title: "Concurrent Issue 2", State: "open"},
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
				Name: "issue_list",
				Arguments: map[string]any{
					"repository": "testuser/testrepo",
					"limit":      10,
					"offset":     0,
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
