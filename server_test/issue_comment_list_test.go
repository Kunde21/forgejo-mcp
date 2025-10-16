package servertest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type issueCommentListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestListIssueComments(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []issueCommentListTestCase{
		{
			name: "successful comment listing",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment 1", Author: "testuser", Created: "2025-09-09T10:30:00Z", Updated: "2025-09-09T10:30:00Z"},
					{ID: 2, Content: "Test comment 2", Author: "testuser", Created: "2025-09-09T10:31:00Z", Updated: "2025-09-09T10:31:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 comments (showing 1-2):\nComment 1 (ID: 1): Test comment 1\nComment 2 (ID: 2): Test comment 2\n"},
				},
				StructuredContent: map[string]any{
					"comments": []any{
						map[string]any{
							"id":      float64(1),
							"body":    "Test comment 1",
							"user":    "testuser",
							"created": "2025-09-09T10:30:00Z",
							"updated": "2025-09-09T10:30:00Z",
						},
						map[string]any{
							"id":      float64(2),
							"body":    "Test comment 2",
							"user":    "testuser",
							"created": "2025-09-09T10:31:00Z",
							"updated": "2025-09-09T10:31:00Z",
						},
					},
					"total": float64(2),
					"limit": float64(10),
				},
			},
		},
		{
			name: "empty comments list",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 comments"},
				},
				StructuredContent: nil,
			},
		},
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository":   "invalid-repo-format",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "missing repository and directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "invalid issue number",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: issue_number: must be no less than 1."},
				},
				StructuredContent: nil,
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
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"limit":        200, // Invalid: > 100
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: nil,
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
				"repository":   "testuser/testrepo",
				"issue_number": 1,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 comments (showing 1-1):\nComment 1 (ID: 1): Default test comment\n"},
				},
				StructuredContent: map[string]any{
					"comments": []any{
						map[string]any{
							"id":      float64(1),
							"body":    "Default test comment",
							"user":    "testuser",
							"created": "2024-01-01T00:00:00Z",
							"updated": "2024-01-01T00:00:00Z",
						},
					},
					"total": float64(1),
					"limit": float64(15),
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
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       -1, // Invalid: negative
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		// Directory parameter tests
		{
			name: "directory parameter - directory does not exist",
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - SSH directory does not exist",
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo-ssh",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - invalid directory path",
			arguments: map[string]any{
				"directory":    "/nonexistent/path",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - missing directory",
			arguments: map[string]any{
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - both directory and repository provided",
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to resolve directory: repository validate failed for /home/user/projects/testrepo: directory does not exist"},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - empty comments list",
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"issue_number": 1,
				"limit":        10,
				"offset":       0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: nil,
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
				Name:      "issue_comment_list",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_list tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}
