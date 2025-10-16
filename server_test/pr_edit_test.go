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

type prEditTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestEditPullRequest(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []prEditTestCase{
		{
			name: "successful title edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        1,
						Number:    123,
						Title:     "Original title",
						Body:      "Original body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
				"title":               "Updated title",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 123, Title: Updated title, State: open, Updated: 2025-10-04T12:00:00Z\nBody: Original body\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(1),
						"number":  float64(123),
						"title":   "Updated title",
						"body":    "Original body",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
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
		{
			name: "successful body edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        2,
						Number:    456,
						Title:     "Test PR",
						Body:      "Original body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 456,
				"body":                "Updated body content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 456, Title: Test PR, State: open, Updated: 2025-10-04T12:00:00Z\nBody: Updated body content\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(2),
						"number":  float64(456),
						"title":   "Test PR",
						"body":    "Updated body content",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
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
		{
			name: "successful state edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        3,
						Number:    789,
						Title:     "Test PR",
						Body:      "Test body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 789,
				"state":               "closed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 789, Title: Test PR, State: closed, Updated: 2025-10-04T12:00:00Z\nBody: Test body\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(3),
						"number":  float64(789),
						"title":   "Test PR",
						"body":    "Test body",
						"state":   "closed",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
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
		{
			name: "successful base branch edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        4,
						Number:    101,
						Title:     "Test PR",
						Body:      "Test body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 101,
				"base_branch":         "develop",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 101, Title: Test PR, State: open, Updated: 2025-10-04T12:00:00Z\nBody: Test body\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(4),
						"number":  float64(101),
						"title":   "Test PR",
						"body":    "Test body",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
						"head": map[string]any{
							"ref": "feature-branch",
							"sha": "abc123",
						},
						"base": map[string]any{
							"ref": "develop",
							"sha": "def456",
						},
					},
				},
			},
		},
		{
			name: "error: no changes provided",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "At least one of title, body, state, or base_branch must be provided"},
				},
				StructuredContent: nil,
				IsError:           true,
			},
		},
		{
			name: "error: invalid repository",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "invalid-repo",
				"pull_request_number": 123,
				"title":               "Updated title",
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
			name: "error: invalid state",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for error case
			},
			arguments: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
				"state":               "invalid",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: state: state must be 'open' or 'closed'."},
				},
				StructuredContent: nil,
				IsError:           true,
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

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_edit tool: %v", err)
			}

			if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
				t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
			}
		})
	}
}

// TestEditPullRequestConcurrent tests concurrent request handling
func TestEditPullRequestConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	// Add multiple PRs to avoid conflicts in concurrent editing
	for i := 1; i <= 3; i++ {
		mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
			{
				ID:        i,
				Number:    i,
				Title:     fmt.Sprintf("Original PR %d", i),
				Body:      "Original body",
				State:     "open",
				BaseRef:   "main",
				UpdatedAt: "2025-09-11T10:30:00Z",
			},
		})
	}

	ts := NewTestServer(t, t.Context(), map[string]string{
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
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "pr_edit",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": id,
					"title":               fmt.Sprintf("Concurrent edit title for PR %d", id),
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

// TestEditPullRequestDirectoryParameter tests directory parameter functionality
func TestEditPullRequestDirectoryParameter(t *testing.T) {
	testCases := []prEditTestCase{
		{
			name: "directory parameter with valid git repo",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        1,
						Number:    123,
						Title:     "Original title",
						Body:      "Original body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"directory":           createTempGitRepo(t, "testuser", "testrepo"),
				"pull_request_number": 123,
				"title":               "Updated title",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 123, Title: Updated title, State: open, Updated: 2025-10-04T12:00:00Z\nBody: Original body\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(1),
						"number":  float64(123),
						"title":   "Updated title",
						"body":    "Original body",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
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
		{
			name: "directory parameter with invalid git repo",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory":           "/nonexistent/path",
				"pull_request_number": 123,
				"title":               "Updated title",
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
			name: "directory parameter with both directory and repository provided (directory takes precedence)",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{
						ID:        1,
						Number:    123,
						Title:     "Original title",
						Body:      "Original body",
						State:     "open",
						BaseRef:   "main",
						UpdatedAt: "2025-09-11T10:30:00Z",
					},
				})
			},
			arguments: map[string]any{
				"directory":           createTempGitRepo(t, "testuser", "testrepo"),
				"repository":          "different/repo",
				"pull_request_number": 123,
				"title":               "Updated title",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Pull request edited successfully. Number: 123, Title: Updated title, State: open, Updated: 2025-10-04T12:00:00Z\nBody: Original body\n"},
				},
				StructuredContent: map[string]any{
					"pull_request": map[string]any{
						"id":      float64(1),
						"number":  float64(123),
						"title":   "Updated title",
						"body":    "Original body",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-04T12:00:00Z",
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
				Name:      "pr_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestEditPullRequestValidationErrors tests validation error scenarios
func TestEditPullRequestValidationErrors(t *testing.T) {
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
			name: "missing repository and directory",
			args: map[string]any{
				"pull_request_number": 123,
				"title":               "Updated title",
			},
			wantError:   true,
			errorSubstr: "at least one of directory or repository must be provided",
		},
		{
			name: "invalid repository format",
			args: map[string]any{
				"repository":          "invalid-repo-format",
				"pull_request_number": 123,
				"title":               "Updated title",
			},
			wantError:   true,
			errorSubstr: "repository: repository must be in format 'owner/repo'",
		},
		{
			name: "zero pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 0,
				"title":               "Updated title",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "negative pull request number",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": -1,
				"title":               "Updated title",
			},
			wantError:   true,
			errorSubstr: "pull_request_number: must be no less than 1",
		},
		{
			name: "invalid state",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
				"state":               "invalid",
			},
			wantError:   true,
			errorSubstr: "state: state must be 'open' or 'closed'",
		},
		{
			name: "title too long",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
				"title":               strings.Repeat("a", 256),
			},
			wantError:   true,
			errorSubstr: "title: title must be between 1 and 255 characters",
		},
		{
			name: "body too long",
			args: map[string]any{
				"repository":          "testuser/testrepo",
				"pull_request_number": 123,
				"body":                strings.Repeat("a", 65536),
			},
			wantError:   true,
			errorSubstr: "body: body must be between 1 and 65535 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_edit",
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

// TestEditPullRequestCompleteWorkflow tests the complete pull request edit workflow
func TestEditPullRequestCompleteWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Set up mock Gitea server with existing PR
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{
			ID:        1,
			Number:    123,
			Title:     "Original title",
			Body:      "Original body",
			State:     "open",
			BaseRef:   "main",
			UpdatedAt: "2025-09-11T10:30:00Z",
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

	// Test successful PR editing
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_edit",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 123,
			"title":               "Updated title",
			"body":                "Updated body content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_edit tool: %v", err)
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
	if !strings.Contains(textContent.Text, "Pull request edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}

	// Should contain PR number and updated title
	if !strings.Contains(textContent.Text, "Number: 123") {
		t.Errorf("Expected PR number in message, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Title: Updated title") {
		t.Errorf("Expected updated title in message, got: %s", textContent.Text)
	}

	// Should contain updated body
	if !strings.Contains(textContent.Text, "Updated body content") {
		t.Errorf("Expected updated body in message, got: %s", textContent.Text)
	}
}
