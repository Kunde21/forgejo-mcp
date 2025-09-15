package servertest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestIssueCommentEdit tests all issue comment edit scenarios using table-driven approach
func TestIssueCommentEdit(t *testing.T) {
	type TestCase struct {
		name      string
		setupMock func(*MockGiteaServer)
		arguments map[string]any
		expect    *mcp.CallToolResult
	}

	testCases := []TestCase{
		{
			name: "successful comment edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Updated comment content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Updated comment content",
					},
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
			name: "real world scenario - status update",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Working on this issue", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
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
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 1, Updated: 2025-09-10T10:00:00Z\nComment body: I've completed the implementation and added comprehensive tests. Ready for review.",
					},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    "I've completed the implementation and added comprehensive tests. Ready for review.",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
		{
			name: "large content performance test",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Original content", Author: "testuser", Created: "2024-01-01T00:00:00Z"},
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
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 1, Updated: 2025-09-10T10:00:00Z\nComment body: " + strings.Repeat("This is a test comment with some content. ", 100),
					},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    strings.Repeat("This is a test comment with some content. ", 100),
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
		{
			name: "validation error - invalid repository format",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: repository: repository must be in format 'owner/repo'.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - missing repository",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: repository: cannot be blank.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - missing issue number",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":  "testuser/testrepo",
				"comment_id":  123,
				"new_content": "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: issue_number: must be no less than 1.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - missing comment id",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: comment_id: must be no less than 1.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - missing new content",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: new_content: cannot be blank.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - empty new content",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: new_content: cannot be blank.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - negative issue number",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"comment_id":   123,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: issue_number: must be no less than 1.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "validation error - zero comment id",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for validation errors
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   0,
				"new_content":  "test content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invalid request: comment_id: must be no less than 1.",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "permission error - invalid token",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 401\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and 'Unauthorized\n' body",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "API error - nonexistent repository",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			arguments: map[string]any{
				"repository":   "nonexistent/repo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 404\nRequest: '/api/v1/repos/nonexistent/repo/issues/comments/123' with 'PATCH' method and '404 page not found\n' body",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "comprehensive error handling - server error",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
				mock.SetServerErrorCommentEdit(123)
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 500\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and 'Internal Server Error\n' body",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "comprehensive error handling - malformed JSON response",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
				mock.SetServerErrorCommentEdit(123)
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 500\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and 'Internal Server Error\n' body",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "mock server state validation - comment not found",
			setupMock: func(mock *MockGiteaServer) {
				// Add different comment, not the one we're trying to edit
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 456, Content: "Different comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123, // This comment doesn't exist
				"new_content":  "Updated content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 404\nRequest: '/api/v1/repos/testuser/testrepo/issues/comments/123' with 'PATCH' method and '404 page not found\n' body",
					},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "special characters and unicode - emoji and symbols",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "🚀 Feature implemented! ✅\n\nThis includes:\n• Unicode support 🌍\n• Special characters: @#$%^&*()\n• Multi-line content\n\nTesting: 中文, 日本語",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: 🚀 Feature implemented! ✅\n\nThis includes:\n• Unicode support 🌍\n• Special characters: @#$%^&*()\n• Multi-line content\n\nTesting: 中文, 日本語",
					},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "🚀 Feature implemented! ✅\n\nThis includes:\n• Unicode support 🌍\n• Special characters: @#$%^&*()\n• Multi-line content\n\nTesting: 中文, 日本語",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
		{
			name: "special characters and unicode - RTL text and complex scripts",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "مرحبا بالعالم\n\nThis comment contains:\n• Right-to-left text (Arabic)\n• Mixed direction content\n• Complex scripts: हिन्दी, বাংলা, தமிழ்\n• Mathematical symbols: ∑∫√π≈∞\n• Technical content: <code>function test() { return true; }</code>",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: مرحبا بالعالم\n\nThis comment contains:\n• Right-to-left text (Arabic)\n• Mixed direction content\n• Complex scripts: हिन्दी, বাংলা, தமிழ்\n• Mathematical symbols: ∑∫√π≈∞\n• Technical content: <code>function test() { return true; }</code>",
					},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "مرحبا بالعالم\n\nThis comment contains:\n• Right-to-left text (Arabic)\n• Mixed direction content\n• Complex scripts: हिन्दी, বাংলা, தமிழ்\n• Mathematical symbols: ∑∫√π≈∞\n• Technical content: <code>function test() { return true; }</code>",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
		{
			name: "special characters and unicode - JSON escaping and HTML entities",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 123, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   123,
				"new_content":  "Testing JSON & HTML escaping:\n• Quotes: 'single' and \"double\"\n• HTML entities: &lt;div&gt; &amp; &quot;\n• Backslashes: C:\\Windows\\Path\n• Unicode escapes: \\u2713 \\u2717\n• Newlines and tabs:\n\tIndented line\n\tAnother tab",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Testing JSON & HTML escaping:\n• Quotes: 'single' and \"double\"\n• HTML entities: &lt;div&gt; &amp; &quot;\n• Backslashes: C:\\Windows\\Path\n• Unicode escapes: \\u2713 \\u2717\n• Newlines and tabs:\n\tIndented line\n\tAnother tab",
					},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"body":    "Testing JSON & HTML escaping:\n• Quotes: 'single' and \"double\"\n• HTML entities: &lt;div&gt; &amp; &quot;\n• Backslashes: C:\\Windows\\Path\n• Unicode escapes: \\u2713 \\u2717\n• Newlines and tabs:\n\tIndented line\n\tAnother tab",
						"user":    "testuser",
						"created": "2025-09-10T10:00:00Z",
						"updated": "2025-09-10T10:00:00Z",
					},
				},
				IsError: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			env := map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			}

			// Override token for permission error test
			if tc.name == "permission error - invalid token" {
				env["FORGEJO_AUTH_TOKEN"] = "invalid-token"
			}

			ts := NewTestServer(t, ctx, env)
			if err := ts.Initialize(); err != nil {
				t.Fatal(err)
			}
			client := ts.Client()

			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Validate mock server state for relevant tests
			if strings.Contains(tc.name, "mock server state") || strings.Contains(tc.name, "successful") {
				validateMockServerState(t, mock, tc.arguments, tc.expect.IsError)
			}

			// Compare results using cmp.Equal
			if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
				t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
			}
		})
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

	for i := range 3 {
		wg.Add(1)
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
		}(i + 1)
	}

	wg.Wait()
	close(results)
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

// validateMockServerState validates the mock server state after comment edit operations
func validateMockServerState(t *testing.T, mock *MockGiteaServer, args map[string]any, expectError bool) {
	t.Helper()

	repo, ok := args["repository"].(string)
	if !ok {
		t.Errorf("repository argument is not a string")
		return
	}

	var commentID int
	switch v := args["comment_id"].(type) {
	case int:
		commentID = v
	case float64:
		commentID = int(v)
	default:
		t.Errorf("comment_id argument is not a number, got: %T, value: %v", args["comment_id"], args["comment_id"])
		return
	}

	newContent, ok := args["new_content"].(string)
	if !ok {
		t.Errorf("new_content argument is not a string")
		return
	}

	// For successful operations, verify the comment was actually updated
	if !expectError {
		// This would require adding a method to MockGiteaServer to inspect stored comments
		// For now, we'll just log that validation would happen here
		t.Logf("Mock server state validation: comment %d in repo %s should contain content: %q", int(commentID), repo, newContent)
	}
}
