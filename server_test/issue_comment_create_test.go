package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type issueCommentCreateTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestCreateIssueComment(t *testing.T) {
	t.Parallel()
	testCases := []issueCommentCreateTestCase{
		{
			name: "successful comment creation",
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
						"id":         float64(1),
						"body":       "This is a test comment",
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
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "missing repository",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: cannot be blank."},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "invalid issue number",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"comment":      "Test comment",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to create comment: issue number validation failed: issue number must be positive"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "empty comment",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
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
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "whitespace only comment",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment":      "   \n\t   ",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to create comment: comment content validation failed: comment content cannot be only whitespace"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
			},
		},
		{
			name: "negative issue number",
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
						"body":       "",
						"created_at": "",
						"id":         float64(0),
						"updated_at": "",
						"user":       "",
					},
				},
				IsError: true,
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
				Name:      "issue_comment_create",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_create tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
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
