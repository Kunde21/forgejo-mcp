package servertest

import (
	"context"
	"strings"
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
