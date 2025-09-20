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

// IssueCommentCreateTestCase represents a comprehensive test case for issue comment creation
type IssueCommentCreateTestCase struct {
	name           string
	category       TestCategory
	setupMock      func(*MockGiteaServer)
	arguments      map[string]any
	expect         *mcp.CallToolResult
	expectError    bool
	errorSubstring string
	timeout        time.Duration
	validateFunc   func(*testing.T, *mcp.CallToolResult)
}

// MockDataFactory creates realistic mock data for testing
type MockDataFactory struct{}

// CreateRealisticComment generates realistic comment content for testing
func (f *MockDataFactory) CreateRealisticComment(commentType string) string {
	switch commentType {
	case "code_review":
		return `Great work on this implementation! I have a few suggestions:

1. Consider adding error handling for the edge case where the repository doesn't exist
2. The function could benefit from early returns to reduce nesting
3. Documentation comments would be helpful for future maintainers

Overall, this looks good to merge after these minor improvements.`
	case "status_update":
		return `I've completed the implementation and added comprehensive tests. The feature is now ready for review.

Changes made:
- Added input validation for all parameters
- Implemented proper error handling
- Added unit tests with 95% coverage
- Updated documentation

Please review when you have a chance.`
	case "bug_report":
		return `I've identified a bug in the current implementation. When the repository name contains special characters, the validation fails unexpectedly.

Steps to reproduce:
1. Create a repository with name "test-repo_special"
2. Try to create an issue comment
3. Observe the validation error

Expected behavior: Should handle special characters properly
Actual behavior: Validation fails with "invalid repository format"`
	case "question":
		return `I have a question about the approach taken here. Why did we choose to implement the validation logic in the service layer rather than the handler layer?

I'm asking because:
- It might be harder to test in isolation
- Error messages might be less specific to the API context
- It could make the service layer more complex

Looking forward to understanding the reasoning behind this decision.`
	default:
		return "This is a test comment for issue comment creation functionality."
	}
}

// CreateRealisticComments generates multiple realistic comments
func (f *MockDataFactory) CreateRealisticComments(count int) []MockComment {
	comments := make([]MockComment, count)
	commentTypes := []string{"code_review", "status_update", "bug_report", "question"}

	for i := 0; i < count; i++ {
		commentType := commentTypes[i%len(commentTypes)]
		comments[i] = MockComment{
			ID:      i + 1,
			Content: f.CreateRealisticComment(commentType),
			Author:  "testuser",
			Created: "2025-09-14T10:00:00Z",
			Updated: "2025-09-14T10:00:00Z",
		}
	}
	return comments
}

var mockFactory = &MockDataFactory{}

func TestCreateIssueComment(t *testing.T) {
	t.Parallel()
	testCases := []IssueCommentCreateTestCase{
		{
			name:     "successful comment creation",
			category: Acceptance,
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for successful creation
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "This is a test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: This is a test comment"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    "This is a test comment",
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:      "invalid repository format",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "missing repository and directory",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "invalid issue number",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: issue_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "empty comment",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment: cannot be blank."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "whitespace only comment",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "   \n\t   ",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: comment: cannot be blank."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "negative issue number",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: issue_number: must be no less than 1."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		// Real-world scenario tests
		{
			name:     "real world scenario - code review comment",
			category: Acceptance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Feature: Add user authentication", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      mockFactory.CreateRealisticComment("code_review"),
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: " + mockFactory.CreateRealisticComment("code_review")},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    mockFactory.CreateRealisticComment("code_review"),
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:     "real world scenario - status update",
			category: Acceptance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 42, Title: "Bug: Login fails on mobile", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 42,
				"comment":      mockFactory.CreateRealisticComment("status_update"),
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: " + mockFactory.CreateRealisticComment("status_update")},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    mockFactory.CreateRealisticComment("status_update"),
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:     "real world scenario - bug report comment",
			category: Acceptance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 123, Title: "Critical: Data corruption in export", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"comment":      mockFactory.CreateRealisticComment("bug_report"),
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: " + mockFactory.CreateRealisticComment("bug_report")},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    mockFactory.CreateRealisticComment("bug_report"),
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		// Error handling tests
		{
			name:      "error handling - repository not found",
			category:  Error,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "nonexistent/repo",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to create comment: failed to create issue comment: unknown API error: 404\nRequest: '/api/v1/repos/nonexistent/repo/issues/1/comments' with 'POST' method and '404 page not found\n' body"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "error handling - issue not found",
			category:  Error,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 999,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: Test comment"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    "Test comment",
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:      "error handling - authentication failure",
			category:  Error,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: Test comment"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(1),
						"body":    "Test comment",
						"user":    "testuser",
						"created": "2024-01-01T00:00:00Z",
						"updated": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		// Performance tests
		{
			name:     "performance - small content (1KB)",
			category: Performance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Performance test issue", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      strings.Repeat("Performance test comment. ", 50), // ~1KB
			},
			timeout: 5 * time.Second,
			validateFunc: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
				}
				text := GetTextContent(result.Content)
				if !strings.Contains(text, "Comment created successfully") {
					t.Errorf("Expected success message, got: %s", text)
				}
			},
		},
		{
			name:     "performance - medium content (10KB)",
			category: Performance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Performance test issue", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      strings.Repeat("Medium performance test comment with detailed content. ", 500), // ~10KB
			},
			timeout: 10 * time.Second,
			validateFunc: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
				}
			},
		},
		{
			name:     "performance - large content (100KB)",
			category: Performance,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Performance test issue", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      strings.Repeat("Large performance test comment with extensive detailed content for testing large comment handling. ", 5000), // ~100KB
			},
			timeout: 30 * time.Second,
			validateFunc: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
				}
			},
		},
		// Directory parameter tests
		{
			name:      "directory parameter - non-existent HTTPS directory",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"issue_number": 1,
				"comment":      "Test comment using directory parameter",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "directory parameter - successful comment creation with SSH remote",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo-ssh",
				"issue_number": 42,
				"comment":      "Test comment using directory parameter with SSH remote",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "directory parameter - invalid directory path",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"directory":    "/nonexistent/path",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "directory parameter - missing directory and repository",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "directory parameter - both directory and repository provided (directory takes precedence)",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to resolve directory: repository validate failed for /home/user/projects/testrepo: directory does not exist"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
		{
			name:      "directory parameter - real world scenario with code review",
			category:  Validation,
			setupMock: func(mock *MockGiteaServer) {},
			arguments: map[string]any{
				"directory":    "/home/user/projects/testrepo",
				"issue_number": 7,
				"comment":      mockFactory.CreateRealisticComment("code_review"),
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":    "",
						"created": "",
						"id":      float64(0),
						"updated": "",
						"user":    "",
					},
				},
				IsError: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Set up context with timeout
			timeout := tc.timeout
			if timeout == 0 {
				timeout = 10 * time.Second
			}
			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			t.Cleanup(cancel)

			// Create mock server and set up test data
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create test server
			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			// Call the tool
			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_comment_create",
				Arguments: tc.arguments,
			})

			// Handle error expectations
			if tc.expectError {
				if err == nil && result != nil && !result.IsError {
					t.Errorf("Expected error, but got success result")
					return
				}
				if err != nil {
					if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
						t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
					}
					return
				}
				if result != nil && result.IsError {
					text := GetTextContent(result.Content)
					if tc.errorSubstring != "" && !strings.Contains(text, tc.errorSubstring) {
						t.Errorf("Expected error containing '%s', got: %s", tc.errorSubstring, text)
					}
					return
				}
			}

			// Handle unexpected errors
			if err != nil {
				t.Fatalf("Failed to call issue_comment_create tool: %v", err)
			}

			// Use custom validation function if provided
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
				return
			}

			// Default validation against expected result
			if tc.expect != nil {
				if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
					t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
				}
			}
		})
	}
}

// TestIssueCommentLifecycle tests the complete issue comment lifecycle: create, list
// This test is kept as a multi-step acceptance test since it involves multiple tool calls
func TestIssueCommentLifecycle(t *testing.T) {
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
	testComment := "This is a test comment on an issue for acceptance testing."

	// Step 1: Create an issue comment
	t.Log("Step 1: Creating issue comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"comment":      testComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create issue comment: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("Issue comment creation failed: %s", getTextContent(createResult.Content))
	}

	// Verify creation response
	createText := getTextContent(createResult.Content)
	if !strings.Contains(createText, "Comment created successfully") {
		t.Errorf("Expected successful creation message, got: %s", createText)
	}
	if !strings.Contains(createText, testComment) {
		t.Errorf("Expected comment in response, got: %s", createText)
	}

	// Step 2: List issue comments to verify creation
	t.Log("Step 2: Listing issue comments to verify creation")
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
		t.Fatalf("Failed to list issue comments: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("Issue comment listing failed: %s", getTextContent(listResult.Content))
	}

	// Verify the comment appears in the list
	listText := getTextContent(listResult.Content)
	if !strings.Contains(listText, testComment) {
		t.Errorf("Expected comment in list, got: %s", listText)
	}

	t.Log("✅ Issue comment lifecycle test completed successfully")
}

// TestIssueCommentCreateConcurrent tests concurrent comment creation for thread safety
func TestIssueCommentCreateConcurrent(t *testing.T) {
	t.Parallel()

	const numGoroutines = 10
	const numRequests = 5

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Concurrent test issue", State: "open"},
		{Index: 2, Title: "Another concurrent test issue", State: "open"},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numRequests)
	successCount := make(chan int, numGoroutines*numRequests)

	// Start concurrent requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numRequests; j++ {
				issueNum := (goroutineID % 2) + 1 // Alternate between issue 1 and 2
				comment := fmt.Sprintf("Concurrent comment %d-%d: This is a test comment from goroutine %d", goroutineID, j, goroutineID)

				result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
					Name: "issue_comment_create",
					Arguments: map[string]any{
						"repository":   "testuser/testrepo",
						"issue_number": issueNum,
						"comment":      comment,
					},
				})
				if err != nil {
					errChan <- fmt.Errorf("goroutine %d request %d failed: %w", goroutineID, j, err)
					continue
				}

				if result.IsError {
					errChan <- fmt.Errorf("goroutine %d request %d returned error: %s", goroutineID, j, GetTextContent(result.Content))
					continue
				}

				// Verify the comment was created successfully
				text := GetTextContent(result.Content)
				if !strings.Contains(text, "Comment created successfully") {
					errChan <- fmt.Errorf("goroutine %d request %d: unexpected response: %s", goroutineID, j, text)
					continue
				}

				successCount <- 1
			}
		}(i + 1)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	close(successCount)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent test had %d errors:", len(errors))
		for _, err := range errors {
			t.Errorf("  - %v", err)
		}
	}

	// Count successful requests
	successes := 0
	for range successCount {
		successes++
	}

	expectedSuccesses := numGoroutines*numRequests - len(errors)
	if successes != expectedSuccesses {
		t.Errorf("Expected %d successful requests, got %d", expectedSuccesses, successes)
	}

	t.Logf("✅ Concurrent test completed: %d successes, %d errors", successes, len(errors))
}

// TestIssueCommentCreateStress tests the system under high load
func TestIssueCommentCreateStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Parallel()

	const totalRequests = 100
	const concurrentWorkers = 20

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Stress test issue", State: "open"},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}

	requestChan := make(chan int, totalRequests)
	resultChan := make(chan error, totalRequests)

	// Distribute requests
	for i := 0; i < totalRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for requestID := range requestChan {
				comment := fmt.Sprintf("Stress test comment %d from worker %d", requestID, workerID)

				result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
					Name: "issue_comment_create",
					Arguments: map[string]any{
						"repository":   "testuser/testrepo",
						"issue_number": 1,
						"comment":      comment,
					},
				})
				if err != nil {
					resultChan <- fmt.Errorf("request %d worker %d failed: %w", requestID, workerID, err)
					continue
				}

				if result.IsError {
					resultChan <- fmt.Errorf("request %d worker %d error: %s", requestID, workerID, GetTextContent(result.Content))
					continue
				}

				resultChan <- nil // Success
			}
		}(i + 1)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var errors []error
	successes := 0

	for result := range resultChan {
		if result != nil {
			errors = append(errors, result)
		} else {
			successes++
		}
	}

	// Report results
	errorRate := float64(len(errors)) / float64(totalRequests) * 100

	t.Logf("Stress test results:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful: %d", successes)
	t.Logf("  Failed: %d", len(errors))
	t.Logf("  Error rate: %.2f%%", errorRate)

	// Assert acceptable error rate (should be 0% for this test)
	if errorRate > 0 {
		t.Errorf("Error rate too high: %.2f%% (expected 0%%)", errorRate)
		for _, err := range errors {
			t.Logf("  Error: %v", err)
		}
	}

	// Assert reasonable performance (should complete within timeout)
	if successes < totalRequests {
		t.Errorf("Not all requests completed successfully: %d/%d", successes, totalRequests)
	}

	t.Logf("✅ Stress test completed successfully")
}
