package servertest

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPInitialization tests MCP protocol initialization
func TestMCPInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Start(); err != nil {
		t.Fatal("Failed to start server:", err)
	}

	// In the new SDK, initialization happens automatically during connection
	// The test server is properly initialized if no errors occurred
	if !ts.IsRunning() {
		t.Error("Test server should be running after initialization")
	}
}

// TestHelloTool tests that the hello tool returns correct response
func TestHelloTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

	result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: "hello"})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}
	want := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Hello, World!"}},
	}
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestHelloToolWithNilContext tests error handling with nil context
func TestHelloToolWithNilContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

	// Test with cancelled context should return error
	cancelledCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc() // Cancel immediately

	result, err := client.CallTool(cancelledCtx, &mcp.CallToolParams{Name: "hello"})
	if err == nil {
		t.Error("Expected error when calling tool with cancelled context")
	}
	if result != nil && !result.IsError {
		t.Error("Expected error result for cancelled context")
	}
}

// TestToolExecution tests actual tool execution with the "hello" tool
func TestToolExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

	// Test calling the "hello" tool
	result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: "hello"})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}
	want := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Hello, World!"}},
	}
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

	// Test calling a non-existent tool
	_, err := client.CallTool(ctx, &mcp.CallToolParams{Name: "nonexistent_tool"})
	if err == nil {
		t.Error("Expected error when calling non-existent tool")
	}

	// Test calling tool with invalid parameters - new SDK validates schema strictly
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name:      "hello",
		Arguments: map[string]any{"invalid": "param"},
	})
	if err == nil {
		t.Error("Expected error when calling tool with invalid parameters")
	}
	if result != nil && !result.IsError {
		t.Error("Expected error result for invalid parameters")
	}
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

	// Number of concurrent requests
	numRequests := 10
	var wg sync.WaitGroup
	results := make([]string, numRequests)
	errors := make([]error, numRequests)

	for i := range numRequests {
		index := i
		wg.Go(func() {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: "hello"})
			if err != nil {
				errors[index] = err
				return
			}
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					results[index] = textContent.Text
				}
			}
		})
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	// Verify all requests succeeded
	for i, result := range results {
		if result == "" {
			t.Errorf("Request %d returned empty result", i)
		}
	}
}

// ===== PULL REQUEST COMMENT CREATE INTEGRATION TESTS =====

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

// ===== PULL REQUEST COMMENT EDIT INTEGRATION TESTS =====

// TestPullRequestCommentEditCompleteWorkflow tests the complete pull request comment edit workflow
func TestPullRequestCommentEditCompleteWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Set up mock Gitea server with existing comment
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{
		{
			ID:      123,
			Content: "Original comment content",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test successful comment editing
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment_id":          123,
			"new_content":         "Updated comment content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
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
	if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}

	// Should contain comment ID
	if !strings.Contains(textContent.Text, "ID: 123") {
		t.Errorf("Expected comment ID in message, got: %s", textContent.Text)
	}

	// Should contain updated comment body
	if !strings.Contains(textContent.Text, "Updated comment content") {
		t.Errorf("Expected updated comment body in message, got: %s", textContent.Text)
	}
}

// TestPullRequestCommentEditValidationErrors tests validation error scenarios
func TestPullRequestCommentEditValidationErrors(t *testing.T) {
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
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "repository: cannot be blank",
		},
		{
			name: "invalid repository format",
			args: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "repository must be in format 'owner/repo'",
		},
		{
			name: "missing new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "empty new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "",
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "whitespace only new content",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "   \n\t   ",
			},
			wantError:   true,
			errorSubstr: "new_content: cannot be blank",
		},
		{
			name: "zero pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "negative pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"comment_id":          123,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "zero comment ID",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          0,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "comment_id: must be no less than 1",
		},
		{
			name: "negative comment ID",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 1,
				"comment_id":          -1,
				"new_content":         "test content",
			},
			wantError:   true,
			errorSubstr: "comment_id: must be no less than 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_edit",
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

// TestPullRequestCommentEditSuccessfulParameters tests successful comment editing with valid parameters
func TestPullRequestCommentEditSuccessfulParameters(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddComments("owner", "repo", []MockComment{
		{
			ID:      123,
			Content: "Original comment",
			Author:  "testuser",
			Created: "2025-09-10T10:00:00Z",
		},
	})

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
			name: "basic comment edit",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated basic comment.",
			},
			expected: "Updated basic comment.",
		},
		{
			name: "comment edit with special characters",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated comment with special chars: @#$%^&*()",
			},
			expected: "Updated comment with special chars: @#$%^&*()",
		},
		{
			name: "multiline comment edit",
			args: map[string]any{
				"repository":          "owner/repo",
				"pull_request_number": 1,
				"comment_id":          123,
				"new_content":         "Updated line 1\nUpdated line 2\nUpdated line 3",
			},
			expected: "Updated line 1\nUpdated line 2\nUpdated line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_comment_edit",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_comment_edit tool: %v", err)
			}

			if result == nil || result.Content == nil {
				t.Fatal("Expected non-nil result with content")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("Expected TextContent, got %T", result.Content[0])
			}

			// Verify success message
			if !strings.Contains(textContent.Text, "Pull request comment edited successfully") {
				t.Errorf("Expected success message, got: %s", textContent.Text)
			}

			// Verify updated comment body is included
			if !strings.Contains(textContent.Text, tt.expected) {
				t.Errorf("Expected updated comment body '%s' in response, got: %s", tt.expected, textContent.Text)
			}
		})
	}
}
