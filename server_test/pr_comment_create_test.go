package servertest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

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
			name: "real-world code review comment",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{}) // Start with no comments
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 42,
				"comment": `I've reviewed your changes and have a few suggestions:

1. Consider adding error handling for the edge case where the input is empty
2. The function could benefit from more descriptive variable names
3. Add unit tests for the new functionality

Overall, great work on this feature!`,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request comment created successfully. ID: 1, Created: 2024-01-01T00:00:00Z\nComment body: I've reviewed your changes and have a few suggestions:\n\n1. Consider adding error handling for the edge case where the input is empty\n2. The function could benefit from more descriptive variable names\n3. Add unit tests for the new functionality\n\nOverall, great work on this feature!"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":         float64(1),
						"body":       "I've reviewed your changes and have a few suggestions:\n\n1. Consider adding error handling for the edge case where the input is empty\n2. The function could benefit from more descriptive variable names\n3. Add unit tests for the new functionality\n\nOverall, great work on this feature!",
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
				"comment":             "Test comment",
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
				StructuredContent: map[string]any{},
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
				StructuredContent: map[string]any{},
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
				StructuredContent: map[string]any{},
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

// TestPullRequestCommentLifecycle tests the complete PR comment lifecycle: create, list
// This test is kept as a multi-step acceptance test since it involves multiple tool calls
func TestPullRequestCommentLifecycle(t *testing.T) {
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
	pullRequestNumber := 1
	testComment := "This is a test comment on a pull request for acceptance testing."

	// Step 1: Create a PR comment
	t.Log("Step 1: Creating PR comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"comment":             testComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create PR comment: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("PR comment creation failed: %s", getTextContent(createResult.Content))
	}

	// Verify creation response
	createText := getTextContent(createResult.Content)
	if !strings.Contains(createText, "Pull request comment created successfully") {
		t.Errorf("Expected successful creation message, got: %s", createText)
	}
	if !strings.Contains(createText, testComment) {
		t.Errorf("Expected comment in response, got: %s", createText)
	}

	// Step 2: List PR comments to verify creation
	t.Log("Step 2: Listing PR comments to verify creation")
	listResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_list",
		Arguments: map[string]any{
			"repository":          repo,
			"pull_request_number": pullRequestNumber,
			"limit":               10,
			"offset":              0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to list PR comments: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("PR comment listing failed: %s", getTextContent(listResult.Content))
	}

	// Verify the comment appears in the list
	listText := getTextContent(listResult.Content)
	if !strings.Contains(listText, testComment) {
		t.Errorf("Expected comment in list, got: %s", listText)
	}

	t.Log("✅ PR comment lifecycle test completed successfully")
}

// TestPullRequestCommentCreationPerformance tests performance and edge cases
// This acceptance test focuses on end-to-end performance scenarios
func TestPullRequestCommentCreationPerformance(t *testing.T) {
	t.Parallel()
	t.Skip()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test large content scenario - should handle efficiently
	largeComment := strings.Repeat("This is a detailed code review comment with comprehensive feedback. ", 200) // ~10KB
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment":             largeComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_create tool with large content: %v", err)
	}
	opt := cmp.FilterPath(func(p cmp.Path) bool {
		return p.Last().String() == ".Text"
	}, cmp.Comparer(func(a, b string) bool {
		prefix, test := a, b
		if len(b) < len(a) {
			prefix, test = b, a
		}
		return strings.HasPrefix(test, prefix)
	}))
	want := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Pull request comment created successfully"},
		},
		StructuredContent: map[string]any{
			"body":       largeComment,
			"created_at": "2025-09-10T10:00:00Z",
			"id":         float64(123),
			"updated_at": "2025-09-10T10:00:00Z",
			"user":       "testuser",
		},
		IsError: false,
	}
	if !cmp.Equal(want, result, opt) {
		t.Error(cmp.Diff(want, result, opt))
	}

	if result == nil || result.IsError {
		t.Fatal("Expected successful result with large content")
	}

	// Verify the result contains the large content
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
		t.Error("Expected success message for large content")
	}
}

// TestPullRequestCommentCreationConcurrentDifferentPRs tests concurrent request handling on different PRs
// This acceptance test focuses on end-to-end concurrent behavior across different pull requests
func TestPullRequestCommentCreationConcurrentDifferentPRs(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		prNumber := i + 1 // Each goroutine comments on a different PR
		go func(prNum int) {
			defer wg.Done()
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "pr_comment_create",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": prNum,
					"comment":             fmt.Sprintf("Concurrent comment on PR %d", prNum),
				},
			})
			results <- err
		}(prNumber)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check results
	for err := range results {
		if err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
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

// TestPullRequestCommentCreateCompleteWorkflow tests the complete pull request comment create workflow
func TestPullRequestCommentCreateCompleteWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Set up mock Gitea server
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{}) // Start with no comments

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test successful comment creation
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment":             "This is a helpful comment on the pull request.",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_create tool: %v", err)
	}

	// Verify response structure
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Content == nil {
		t.Fatal("Expected non-nil content")
	}
	if len(result.Content) == 0 {
		t.Fatal("Expected at least one content item")
	}

	// Verify content type and message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}
	if textContent.Text == "" {
		t.Error("Expected non-empty text content")
	}

	// Should contain success message
	if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}

	// Should contain comment ID
	if !strings.Contains(textContent.Text, "ID: 1") {
		t.Errorf("Expected comment ID in message, got: %s", textContent.Text)
	}

	// Should contain comment body
	if !strings.Contains(textContent.Text, "This is a helpful comment on the pull request.") {
		t.Errorf("Expected comment body in message, got: %s", textContent.Text)
	}
}

// TestPullRequestCommentCreateValidationErrors tests validation error scenarios
func TestPullRequestCommentCreateValidationErrors(t *testing.T) {
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

	tests := []struct {
		name        string
		args        map[string]any
		wantError   bool
		errorSubstr string
	}{
		{
			name: "missing repository",
			args: map[string]any{
				"pull_request_number": 1,
				"comment":             "Test comment",
			},
			wantError:   true,
			errorSubstr: "repository: cannot be blank",
		},
		{
			name: "invalid repository format",
			args: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 1,
				"comment":             "Test comment",
			},
			wantError:   true,
			errorSubstr: "repository must be in format 'owner/repo'",
		},
		{
			name: "invalid pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"comment":             "Test comment",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "empty comment",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment":             "",
			},
			wantError:   true,
			errorSubstr: "comment: cannot be blank",
		},
		{
			name: "whitespace only comment",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment":             "   \n\t   ",
			},
			wantError:   true,
			errorSubstr: "comment: cannot be blank",
		},
		{
			name: "negative pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment":             "Test comment",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_create",
				Arguments: tt.args,
			})

			if tt.wantError {
				if err != nil {
					t.Fatalf("Expected error in result, got call error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected error result, got nil")
				}
				if !result.IsError {
					t.Error("Expected result to be marked as error")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Fatalf("Expected TextContent, got %T", result.Content[0])
				}
				if !strings.Contains(textContent.Text, tt.errorSubstr) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorSubstr, textContent.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected call error: %v", err)
				}
				if result == nil || result.IsError {
					t.Error("Expected successful result")
				}
			}
		})
	}
}

// TestPullRequestCommentCreateSuccessfulParameters tests successful comment creation with valid parameters
func TestPullRequestCommentCreateSuccessfulParameters(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("owner", "repo", []MockComment{}) // Start with no comments

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	tests := []struct {
		name     string
		args     map[string]any
		expected string
	}{
		{
			name: "basic comment creation",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment":             "This is a basic comment.",
			},
			expected: "This is a basic comment.",
		},
		{
			name: "comment with special characters",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 2,
				"comment":             "Comment with special chars: @#$%^&*()",
			},
			expected: "Comment with special chars: @#$%^&*()",
		},
		{
			name: "multiline comment",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 3,
				"comment":             "Line 1\nLine 2\nLine 3",
			},
			expected: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_create",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_create tool: %v", err)
			}

			if result == nil || result.Content == nil {
				t.Fatal("Expected non-nil result with content")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("Expected TextContent, got %T", result.Content[0])
			}

			// Verify success message
			if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
				t.Errorf("Expected success message, got: %s", textContent.Text)
			}

			// Verify comment body is included
			if !strings.Contains(textContent.Text, tt.expected) {
				t.Errorf("Expected comment body '%s' in response, got: %s", tt.expected, textContent.Text)
			}
		})
	}
}
