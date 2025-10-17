package servertest

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type notificationListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	setupDir  func(t *testing.T) string // Optional function to set up a temporary directory
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestNotificationList(t *testing.T) {
	testCases := []notificationListTestCase{
		{
			name: "acceptance - real world scenario",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{
					{ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "New issue created", Unread: true, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
					{ID: 2, Repository: "testuser/testrepo", Type: "pull", Number: 456, Title: "PR review requested", Unread: true, Updated: "2025-10-16T11:00:00Z", URL: "https://example.com/testuser/testrepo/pulls/456"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 unread notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "New issue created", "unread": true, "updated": "2025-10-16T10:00:00Z"},
						map[string]any{"id": float64(2), "repository": "testuser/testrepo", "type": "pull", "number": float64(456), "title": "PR review requested", "unread": true, "updated": "2025-10-16T11:00:00Z"},
					},
					"total":  float64(2),
					"limit":  float64(10),
					"offset": float64(0),
				},
			},
		},
		{
			name: "status filtering - read notifications",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{
					{ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "Read issue", Unread: false, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
					{ID: 2, Repository: "testuser/testrepo", Type: "pull", Number: 456, Title: "Unread PR", Unread: true, Updated: "2025-10-16T11:00:00Z", URL: "https://example.com/testuser/testrepo/pulls/456"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"status":     "read",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 read notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Read issue", "unread": false, "updated": "2025-10-16T10:00:00Z"},
					},
					"total":  float64(1),
					"limit":  float64(10),
					"offset": float64(0),
				},
			},
		},
		{
			name: "status filtering - all notifications",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{
					{ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "Read issue", Unread: false, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
					{ID: 2, Repository: "testuser/testrepo", Type: "pull", Number: 456, Title: "Unread PR", Unread: true, Updated: "2025-10-16T11:00:00Z", URL: "https://example.com/testuser/testrepo/pulls/456"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"status":     "all",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 all notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Read issue", "unread": false, "updated": "2025-10-16T10:00:00Z"},
						map[string]any{"id": float64(2), "repository": "testuser/testrepo", "type": "pull", "number": float64(456), "title": "Unread PR", "unread": true, "updated": "2025-10-16T11:00:00Z"},
					},
					"total":  float64(2),
					"limit":  float64(10),
					"offset": float64(0),
				},
			},
		},
		{
			name: "pagination - limit and offset",
			setupMock: func(mock *MockGiteaServer) {
				var notifications []MockNotification
				for i := 1; i <= 5; i++ {
					notifications = append(notifications, MockNotification{
						ID:         i,
						Repository: "testuser/testrepo",
						Type:       "issue",
						Number:     i,
						Title:      "Issue notification",
						Unread:     true,
						Updated:    "2025-10-16T10:00:00Z",
						URL:        "https://example.com/testuser/testrepo/issues/123",
					})
				}
				mock.AddNotifications(notifications)
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      2,
				"offset":     1,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 unread notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(2), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Issue notification", "unread": true, "updated": "2025-10-16T10:00:00Z"},
						map[string]any{"id": float64(3), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Issue notification", "unread": true, "updated": "2025-10-16T10:00:00Z"},
					},
					"total":  float64(5),
					"limit":  float64(2),
					"offset": float64(1),
				},
			},
		},
		{
			name: "repository filtering - different repositories",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{
					{ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "Test repo issue", Unread: true, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
					{ID: 2, Repository: "otheruser/otherrepo", Type: "pull", Number: 456, Title: "Other repo PR", Unread: true, Updated: "2025-10-16T11:00:00Z", URL: "https://example.com/otheruser/otherrepo/pulls/456"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 unread notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Test repo issue", "unread": true, "updated": "2025-10-16T10:00:00Z"},
					},
					"total":  float64(1),
					"limit":  float64(10),
					"offset": float64(0),
				},
			},
		},
		{
			name: "empty notifications",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 unread notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{},
					"total":         float64(0),
					"limit":         float64(10),
					"offset":        float64(0),
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
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "invalid status",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"status":     "invalid",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: status: status must be 'read', 'unread', or 'all'."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
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
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "negative offset",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     -1, // Invalid: negative
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "directory parameter - valid git repository",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddNotifications([]MockNotification{
					{ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "Directory-based notification", Unread: true, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
				})
			},
			setupDir: func(t *testing.T) string {
				return createTempGitRepo(t, "testuser", "testrepo")
			},
			arguments: map[string]any{
				"directory": "", // Will be set dynamically
				"limit":     10,
				"offset":    0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 unread notifications"},
				},
				StructuredContent: map[string]any{
					"notifications": []any{
						map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "Directory-based notification", "unread": true, "updated": "2025-10-16T10:00:00Z"},
					},
					"total":  float64(1),
					"limit":  float64(10),
					"offset": float64(0),
				},
			},
		},
		{
			name: "missing repository and directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"limit":  10,
				"offset": 0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context with timeout and proper cleanup
			ctx, cancel := CreateStandardTestContext(t, 10)
			defer cancel()

			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Set up temporary directory if needed
			var tempDir string
			if tc.setupDir != nil {
				tempDir = tc.setupDir(t)
				// Update arguments with the actual temp directory path
				args := make(map[string]any)
				for k, v := range tc.arguments {
					args[k] = v
				}
				if dir, ok := args["directory"].(string); ok && dir == "" {
					args["directory"] = tempDir
				}
				tc.arguments = args
			}

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			// Use standardized tool call with validation
			result, err := ts.CallToolWithValidation(ctx, "notification_list", tc.arguments)
			if err != nil {
				t.Fatalf("Failed to call notification_list tool: %v", err)
			}

			// Use standardized validation with proper comparison options
			if !ts.ValidateToolResult(tc.expect, result, t) {
				t.Errorf("Tool result validation failed for test case: %s", tc.name)
			}
		})
	}
}
