package server

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockGiteaClient implements GiteaClientInterface for testing
type mockGiteaClient struct {
	listIssuesFunc func(ctx context.Context, repo string, limit, offset int) ([]gitea.Issue, error)
}

func (m *mockGiteaClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]gitea.Issue, error) {
	if m.listIssuesFunc != nil {
		return m.listIssuesFunc(ctx, repo, limit, offset)
	}
	return []gitea.Issue{}, nil
}

func (m *mockGiteaClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*gitea.IssueComment, error) {
	return &gitea.IssueComment{}, nil
}

func TestHandleListIssuesValidation(t *testing.T) {
	mockClient := &mockGiteaClient{}
	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	testCases := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input",
			args: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
			},
			expectError: false,
		},
		{
			name: "missing repository",
			args: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "empty repository",
			args: map[string]interface{}{
				"repository": "",
				"limit":      10,
				"offset":     0,
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "invalid repository format",
			args: map[string]interface{}{
				"repository": "invalid-format",
				"limit":      10,
				"offset":     0,
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "repository with spaces",
			args: map[string]interface{}{
				"repository": "own er/repo",
				"limit":      10,
				"offset":     0,
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "limit zero (defaults to 15)",
			args: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      0,
				"offset":     0,
			},
			expectError: false,
		},
		{
			name: "limit too high",
			args: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      101,
				"offset":     0,
			},
			expectError: true,
			errorMsg:    "limit",
		},
		{
			name: "negative offset",
			args: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     -1,
			},
			expectError: true,
			errorMsg:    "offset",
		},
		{
			name: "default limit",
			args: map[string]interface{}{
				"repository": "owner/repo",
				"offset":     0,
			},
			expectError: false,
		},
		{
			name: "valid with numbers in repo",
			args: map[string]interface{}{
				"repository": "user123/repo456",
				"limit":      50,
				"offset":     10,
			},
			expectError: false,
		},
		{
			name: "valid with underscores",
			args: map[string]interface{}{
				"repository": "user_name/repo_name",
				"limit":      25,
				"offset":     5,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert map to struct for handler
			args := struct {
				Repository string `json:"repository"`
				Limit      int    `json:"limit"`
				Offset     int    `json:"offset"`
			}{}

			if repo, ok := tc.args["repository"].(string); ok {
				args.Repository = repo
			}
			if limit, ok := tc.args["limit"].(int); ok {
				args.Limit = limit
			}
			if offset, ok := tc.args["offset"].(int); ok {
				args.Offset = offset
			}

			result, _, err := server.handleListIssues(context.Background(), &mcp.CallToolRequest{}, args)

			if tc.expectError {
				if err != nil {
					t.Errorf("Expected validation error but handler returned error: %v", err)
				} else if result != nil && !result.IsError {
					t.Errorf("Expected validation error but got success result")
				} else if result != nil && len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok && tc.errorMsg != "" {
						if !contains(textContent.Text, tc.errorMsg) {
							t.Errorf("Expected error message to contain '%s', got: %s", tc.errorMsg, textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("Expected success but got error result")
				}
			}
		})
	}
}

func TestHandleListIssuesSuccess(t *testing.T) {
	mockClient := &mockGiteaClient{
		listIssuesFunc: func(ctx context.Context, repo string, limit, offset int) ([]gitea.Issue, error) {
			return []gitea.Issue{
				{Number: 1, Title: "Test Issue", State: "open"},
			}, nil
		},
	}

	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	args := struct {
		Repository string `json:"repository"`
		Limit      int    `json:"limit"`
		Offset     int    `json:"offset"`
	}{
		Repository: "owner/repo",
		Limit:      10,
		Offset:     0,
	}

	result, _, err := server.handleListIssues(context.Background(), &mcp.CallToolRequest{}, args)

	if err != nil {
		t.Errorf("Expected success but got error: %v", err)
	}
	if result == nil {
		t.Error("Expected result but got nil")
	}
	if result.IsError {
		t.Error("Expected success result but got error")
	}
	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}
}

func TestHandleCreateIssueCommentValidation(t *testing.T) {
	mockClient := &mockGiteaClient{}
	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	testCases := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 1,
				"comment":      "Valid comment",
			},
			expectError: false,
		},
		{
			name: "missing repository",
			args: map[string]interface{}{
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "empty repository",
			args: map[string]interface{}{
				"repository":   "",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "invalid repository format",
			args: map[string]interface{}{
				"repository":   "invalid-format",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "repository with spaces",
			args: map[string]interface{}{
				"repository":   "own er/repo",
				"issue_number": 1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "zero issue number",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 0,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue_number",
		},
		{
			name: "negative issue number",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": -1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue_number",
		},
		{
			name: "missing comment",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 1,
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "empty comment",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 1,
				"comment":      "",
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "whitespace comment",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 1,
				"comment":      "   ",
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "tab and newline comment",
			args: map[string]interface{}{
				"repository":   "owner/repo",
				"issue_number": 1,
				"comment":      "\t\n",
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "valid with numbers in repo",
			args: map[string]interface{}{
				"repository":   "user123/repo456",
				"issue_number": 42,
				"comment":      "Valid comment with numbers",
			},
			expectError: false,
		},
		{
			name: "valid with underscores",
			args: map[string]interface{}{
				"repository":   "user_name/repo_name",
				"issue_number": 100,
				"comment":      "Valid comment with underscores",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert map to struct for handler
			args := struct {
				Repository  string `json:"repository"`
				IssueNumber int    `json:"issue_number"`
				Comment     string `json:"comment"`
			}{}

			if repo, ok := tc.args["repository"].(string); ok {
				args.Repository = repo
			}
			if issueNum, ok := tc.args["issue_number"].(int); ok {
				args.IssueNumber = issueNum
			}
			if comment, ok := tc.args["comment"].(string); ok {
				args.Comment = comment
			}

			result, _, err := server.handleCreateIssueComment(context.Background(), &mcp.CallToolRequest{}, args)

			if tc.expectError {
				if err != nil {
					t.Errorf("Expected validation error but handler returned error: %v", err)
				} else if result != nil && !result.IsError {
					t.Errorf("Expected validation error but got success result %+v", result.Content[0])
				} else if result != nil && len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok && tc.errorMsg != "" {
						if !contains(textContent.Text, tc.errorMsg) {
							t.Errorf("Expected error message to contain '%s', got: %s", tc.errorMsg, textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("Expected success but got error result")
				}
			}
		})
	}
}

func TestHandleCreateIssueCommentSuccess(t *testing.T) {
	mockClient := &mockGiteaClient{}
	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	args := struct {
		Repository  string `json:"repository"`
		IssueNumber int    `json:"issue_number"`
		Comment     string `json:"comment"`
	}{
		Repository:  "owner/repo",
		IssueNumber: 1,
		Comment:     "Test comment",
	}

	result, _, err := server.handleCreateIssueComment(context.Background(), &mcp.CallToolRequest{}, args)

	if err != nil {
		t.Errorf("Expected success but got error: %v", err)
	}
	if result == nil {
		t.Error("Expected result but got nil")
	}
	if result.IsError {
		t.Error("Expected success result but got error")
	}
	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}
}

func TestValidationConsistencyBetweenHandlers(t *testing.T) {
	mockClient := &mockGiteaClient{}
	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	// Test repository validation consistency
	t.Run("repository validation consistency", func(t *testing.T) {
		invalidRepos := []string{"", "invalid-format", "own er/repo", "owner/repo/extra"}

		for _, repo := range invalidRepos {
			// Test listIssues handler
			listArgs := struct {
				Repository string `json:"repository"`
				Limit      int    `json:"limit"`
				Offset     int    `json:"offset"`
			}{
				Repository: repo,
				Limit:      10,
				Offset:     0,
			}

			listResult, _, _ := server.handleListIssues(context.Background(), &mcp.CallToolRequest{}, listArgs)

			// Test createIssueComment handler
			commentArgs := struct {
				Repository  string `json:"repository"`
				IssueNumber int    `json:"issue_number"`
				Comment     string `json:"comment"`
			}{
				Repository:  repo,
				IssueNumber: 1,
				Comment:     "Test comment",
			}

			commentResult, _, _ := server.handleCreateIssueComment(context.Background(), &mcp.CallToolRequest{}, commentArgs)

			// Both should either both succeed or both fail for the same invalid input
			listIsError := listResult != nil && listResult.IsError
			commentIsError := commentResult != nil && commentResult.IsError

			if listIsError != commentIsError {
				t.Errorf("Inconsistent validation for repository '%s': listIssues=%v, createIssueComment=%v",
					repo, listIsError, commentIsError)
			}

			// Both should contain repository-related error messages
			if listIsError && listResult.Content != nil && len(listResult.Content) > 0 {
				if textContent, ok := listResult.Content[0].(*mcp.TextContent); ok {
					if !contains(textContent.Text, "repository") {
						t.Errorf("listIssues error should mention repository, got: %s", textContent.Text)
					}
				}
			}

			if commentIsError && commentResult.Content != nil && len(commentResult.Content) > 0 {
				if textContent, ok := commentResult.Content[0].(*mcp.TextContent); ok {
					if !contains(textContent.Text, "repository") {
						t.Errorf("createIssueComment error should mention repository, got: %s", textContent.Text)
					}
				}
			}
		}
	})

	// Test valid repository consistency
	t.Run("valid repository consistency", func(t *testing.T) {
		validRepos := []string{"owner/repo", "user123/repo456", "user_name/repo_name"}

		for _, repo := range validRepos {
			// Test listIssues handler
			listArgs := struct {
				Repository string `json:"repository"`
				Limit      int    `json:"limit"`
				Offset     int    `json:"offset"`
			}{
				Repository: repo,
				Limit:      10,
				Offset:     0,
			}

			listResult, _, listErr := server.handleListIssues(context.Background(), &mcp.CallToolRequest{}, listArgs)

			// Test createIssueComment handler
			commentArgs := struct {
				Repository  string `json:"repository"`
				IssueNumber int    `json:"issue_number"`
				Comment     string `json:"comment"`
			}{
				Repository:  repo,
				IssueNumber: 1,
				Comment:     "Test comment",
			}

			commentResult, _, commentErr := server.handleCreateIssueComment(context.Background(), &mcp.CallToolRequest{}, commentArgs)

			// Both should succeed for valid repository
			if listErr != nil {
				t.Errorf("listIssues should succeed for valid repo '%s', got error: %v", repo, listErr)
			}
			if listResult != nil && listResult.IsError {
				t.Errorf("listIssues should succeed for valid repo '%s', got error result", repo)
			}

			if commentErr != nil {
				t.Errorf("createIssueComment should succeed for valid repo '%s', got error: %v", repo, commentErr)
			}
			if commentResult != nil && commentResult.IsError {
				t.Errorf("createIssueComment should succeed for valid repo '%s', got error result", repo)
			}
		}
	})
}

func TestErrorMessageConsistency(t *testing.T) {
	mockClient := &mockGiteaClient{}
	service := gitea.NewService(mockClient)
	server := &Server{
		giteaService: service,
	}

	// Test that error messages follow consistent patterns
	testCases := []struct {
		name             string
		testFunc         func() (*mcp.CallToolResult, any, error)
		expectedPatterns []string
	}{
		{
			name: "listIssues empty repository",
			testFunc: func() (*mcp.CallToolResult, any, error) {
				args := struct {
					Repository string `json:"repository"`
					Limit      int    `json:"limit"`
					Offset     int    `json:"offset"`
				}{
					Repository: "",
					Limit:      10,
					Offset:     0,
				}
				return server.handleListIssues(context.Background(), &mcp.CallToolRequest{}, args)
			},
			expectedPatterns: []string{"Validation failed", "repository"},
		},
		{
			name: "createIssueComment empty repository",
			testFunc: func() (*mcp.CallToolResult, any, error) {
				args := struct {
					Repository  string `json:"repository"`
					IssueNumber int    `json:"issue_number"`
					Comment     string `json:"comment"`
				}{
					Repository:  "",
					IssueNumber: 1,
					Comment:     "Test",
				}
				return server.handleCreateIssueComment(context.Background(), &mcp.CallToolRequest{}, args)
			},
			expectedPatterns: []string{"Validation failed", "repository"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := tc.testFunc()

			if err != nil {
				t.Errorf("Expected validation error but handler returned error: %v", err)
			} else if result == nil || !result.IsError {
				t.Errorf("Expected validation error but got success")
			} else if len(result.Content) == 0 {
				t.Error("Expected error content")
			} else {
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Error("Expected text content")
				} else {
					for _, pattern := range tc.expectedPatterns {
						if !contains(textContent.Text, pattern) {
							t.Errorf("Error message should contain '%s', got: %s", pattern, textContent.Text)
						}
					}
				}
			}
		})
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
