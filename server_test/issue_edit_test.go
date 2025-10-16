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

func TestEditIssue(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(*MockGiteaServer)
		arguments map[string]any
		expect    *mcp.CallToolResult
	}{
		{
			name: "successful title edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssue("testuser", "testrepo", MockIssue{
					Index:   123,
					Title:   "Original title",
					Body:    "Original body",
					State:   "open",
					Created: "2025-09-11T10:30:00Z",
					Updated: "2025-09-11T10:30:00Z",
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"title":        "Updated title",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Issue edited successfully. Number: 123, Title: Updated title, State: open, Updated: 2025-10-06T12:00:00Z\nBody: Original body\n"},
				},
				StructuredContent: map[string]any{
					"issue": map[string]any{
						"id":      float64(123),
						"number":  float64(123),
						"title":   "Updated title",
						"body":    "Original body",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-06T12:00:00Z",
					},
				},
			},
		},
		{
			name: "successful body edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssue("testuser", "testrepo", MockIssue{
					Index:   456,
					Title:   "Test Issue",
					Body:    "Original body",
					State:   "open",
					Created: "2025-09-11T10:30:00Z",
					Updated: "2025-09-11T10:30:00Z",
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 456,
				"body":         "Updated body content",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Issue edited successfully. Number: 456, Title: Test Issue, State: open, Updated: 2025-10-06T12:00:00Z\nBody: Updated body content\n"},
				},
				StructuredContent: map[string]any{
					"issue": map[string]any{
						"id":      float64(456),
						"number":  float64(456),
						"title":   "Test Issue",
						"body":    "Updated body content",
						"state":   "open",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-06T12:00:00Z",
					},
				},
			},
		},
		{
			name: "successful state edit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssue("testuser", "testrepo", MockIssue{
					Index:   789,
					Title:   "Test Issue",
					Body:    "Test body",
					State:   "open",
					Created: "2025-09-11T10:30:00Z",
					Updated: "2025-09-11T10:30:00Z",
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 789,
				"state":        "closed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Issue edited successfully. Number: 789, Title: Test Issue, State: closed, Updated: 2025-10-06T12:00:00Z\nBody: Test body\n"},
				},
				StructuredContent: map[string]any{
					"issue": map[string]any{
						"id":      float64(789),
						"number":  float64(789),
						"title":   "Test Issue",
						"body":    "Test body",
						"state":   "closed",
						"user":    "testuser",
						"created": "2025-09-11T10:30:00Z",
						"updated": "2025-10-06T12:00:00Z",
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
				"repository":   "testuser/testrepo",
				"issue_number": 123,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "At least one of title, body, or state must be provided"},
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
				"repository":   "invalid-repo",
				"issue_number": 123,
				"title":        "Updated title",
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
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"state":        "invalid",
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
				Name:      "issue_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_edit tool: %v", err)
			}

			if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
				t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
			}
		})
	}
}

// TestEditIssueConcurrent tests concurrent request handling
func TestEditIssueConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	// Add multiple issues to avoid conflicts in concurrent editing
	for i := 1; i <= 3; i++ {
		mock.AddIssue("testuser", "testrepo", MockIssue{
			Index:   i,
			Title:   fmt.Sprintf("Original Issue %d", i),
			Body:    "Original body",
			State:   "open",
			Created: "2025-09-11T10:30:00Z",
			Updated: "2025-09-11T10:30:00Z",
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
				Name: "issue_edit",
				Arguments: map[string]any{
					"repository":   "testuser/testrepo",
					"issue_number": id,
					"title":        fmt.Sprintf("Concurrent edit title for Issue %d", id),
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

// TestEditIssueValidationErrors tests validation error scenarios
func TestEditIssueValidationErrors(t *testing.T) {
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
				"issue_number": 123,
				"title":        "Updated title",
			},
			wantError:   true,
			errorSubstr: "at least one of directory or repository must be provided",
		},
		{
			name: "invalid repository format",
			args: map[string]any{
				"repository":   "invalid-repo-format",
				"issue_number": 123,
				"title":        "Updated title",
			},
			wantError:   true,
			errorSubstr: "repository: repository must be in format 'owner/repo'",
		},
		{
			name: "zero issue number",
			args: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"title":        "Updated title",
			},
			wantError:   true,
			errorSubstr: "issue_number: must be no less than 1",
		},
		{
			name: "negative issue number",
			args: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"title":        "Updated title",
			},
			wantError:   true,
			errorSubstr: "issue_number: must be no less than 1",
		},
		{
			name: "invalid state",
			args: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"state":        "invalid",
			},
			wantError:   true,
			errorSubstr: "state: state must be 'open' or 'closed'",
		},
		{
			name: "title too long",
			args: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"title":        strings.Repeat("a", 256),
			},
			wantError:   true,
			errorSubstr: "title: title must be between 1 and 255 characters",
		},
		{
			name: "body too long",
			args: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 123,
				"body":         strings.Repeat("a", 65536),
			},
			wantError:   true,
			errorSubstr: "body: body must be between 1 and 65535 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_edit",
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

// TestEditIssueCompleteWorkflow tests the complete issue edit workflow
func TestEditIssueCompleteWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Set up mock Gitea server with existing issue
	mock := NewMockGiteaServer(t)
	mock.AddIssue("testuser", "testrepo", MockIssue{
		Index:   123,
		Title:   "Original title",
		Body:    "Original body",
		State:   "open",
		Created: "2025-09-11T10:30:00Z",
		Updated: "2025-09-11T10:30:00Z",
	})

	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test successful issue editing
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 123,
			"title":        "Updated title",
			"body":         "Updated body content",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_edit tool: %v", err)
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
	if !strings.Contains(textContent.Text, "Issue edited successfully") {
		t.Errorf("Expected success message, got: %s", textContent.Text)
	}

	// Should contain issue number and updated title
	if !strings.Contains(textContent.Text, "Number: 123") {
		t.Errorf("Expected issue number in message, got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Title: Updated title") {
		t.Errorf("Expected updated title in message, got: %s", textContent.Text)
	}

	// Should contain updated body
	if !strings.Contains(textContent.Text, "Updated body content") {
		t.Errorf("Expected updated body in message, got: %s", textContent.Text)
	}
}
