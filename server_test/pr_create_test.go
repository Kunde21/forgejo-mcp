package servertest

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kunde21/forgejo-mcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPullRequestCreateBasicValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        server.PullRequestCreateArgs
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid with repository only",
			args: server.PullRequestCreateArgs{
				Repository: "owner/repo",
				Title:      "Test PR",
				Head:       "feature",
				Base:       "main",
			},
			expectError: false,
		},
		{
			name: "missing title",
			args: server.PullRequestCreateArgs{
				Repository: "owner/repo",
				Head:       "feature",
				Base:       "main",
			},
			expectError: true,
			errorMsg:    "title is required",
		},
		{
			name: "invalid repository format",
			args: server.PullRequestCreateArgs{
				Repository: "invalid-repo",
				Title:      "Test PR",
				Head:       "feature",
				Base:       "main",
			},
			expectError: true,
			errorMsg:    "repository must be in format 'owner/repo'",
		},
		{
			name: "empty title",
			args: server.PullRequestCreateArgs{
				Repository: "owner/repo",
				Title:      "",
				Head:       "feature",
				Base:       "main",
			},
			expectError: true,
			errorMsg:    "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			// Create mock server
			mock := NewMockGiteaServer(t)

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}
			client := ts.Client()

			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_create",
				Arguments: tt.args,
			})

			if tt.expectError {
				if err == nil && result != nil && !result.IsError {
					t.Errorf("Expected error but got success")
					return
				}
				if result != nil && tt.errorMsg != "" {
					text := GetTextContent(result.Content)
					if !strings.Contains(text, tt.errorMsg) {
						t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, text)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != nil && result.IsError {
					text := GetTextContent(result.Content)
					t.Errorf("Expected success but got error: %s", text)
				}
			}
		})
	}
}

func TestPullRequestCreateSuccess(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []struct {
		name      string
		setupMock func(*MockGiteaServer)
		arguments map[string]any
		// Check for success indicators rather than exact match
		expectSuccess bool
		checkContent  func(string) bool
	}{
		{
			name: "successful PR creation with all parameters",
			setupMock: func(mock *MockGiteaServer) {
				// No existing PRs needed for creation
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"title":      "Add new feature",
				"head":       "feature-branch",
				"base":       "main",
				"body":       "This PR adds a new feature that improves user experience",
				"assignee":   "reviewer",
			},
			expectSuccess: true,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Pull request created successfully")
			},
		},
		{
			name: "successful draft PR creation",
			setupMock: func(mock *MockGiteaServer) {
				// No existing PRs needed for creation
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"title":      "Draft: WIP feature",
				"head":       "wip-branch",
				"base":       "main",
				"body":       "Work in progress - not ready for review",
				"draft":      true,
			},
			expectSuccess: true,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Pull request created successfully")
			},
		},
		{
			name: "PR creation with minimal parameters",
			setupMock: func(mock *MockGiteaServer) {
				// No existing PRs needed for creation
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"title":      "Simple PR",
				"head":       "simple-branch",
				"base":       "main",
			},
			expectSuccess: true,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Pull request created successfully")
			},
		},
		{
			name: "repository not found error",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			arguments: map[string]any{
				"repository": "nonexistent/repo",
				"title":      "Test PR",
				"head":       "feature",
				"base":       "main",
			},
			expectSuccess: false,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Repository or branch not found")
			},
		},
		{
			name: "missing required title",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed - validation should catch this
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"head":       "feature",
				"base":       "main",
			},
			expectSuccess: false,
			checkContent: func(text string) bool {
				return strings.Contains(text, "title is required")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			mock := NewMockGiteaServer(t)
			tc.setupMock(mock)

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_create",
				Arguments: tc.arguments,
			})

			if err != nil {
				t.Fatalf("Unexpected error calling tool: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Check error status matches expectation
			if result.IsError && tc.expectSuccess {
				t.Errorf("Expected success but got error: %s", GetTextContent(result.Content))
			}
			if !result.IsError && !tc.expectSuccess {
				t.Errorf("Expected error but got success: %s", GetTextContent(result.Content))
			}

			// Check content
			text := GetTextContent(result.Content)
			if !tc.checkContent(text) {
				t.Errorf("Content check failed for: %s", text)
			}
		})
	}
}

func TestPullRequestCreateDirectoryParameter(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []struct {
		name         string
		setupMock    func(*MockGiteaServer)
		createDir    func(t *testing.T) string
		arguments    map[string]any
		expectError  bool
		checkContent func(string) bool
	}{
		{
			name: "directory does not exist",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed - directory validation should catch this
			},
			createDir: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			arguments: map[string]any{
				"title": "Test PR",
				"head":  "feature",
				"base":  "main",
			},
			expectError: true,
			checkContent: func(text string) bool {
				return strings.Contains(text, "invalid directory") || strings.Contains(text, "does not exist")
			},
		},
		{
			name: "directory not a git repository",
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed - git validation should catch this
			},
			createDir: func(t *testing.T) string {
				return createTempNonGitDir(t)
			},
			arguments: map[string]any{
				"title": "Test PR",
				"head":  "feature",
				"base":  "main",
			},
			expectError: true,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Not a git repository")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			mock := NewMockGiteaServer(t)
			tc.setupMock(mock)

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			// Create directory and add to arguments
			directory := tc.createDir(t)
			tc.arguments["directory"] = directory

			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_create",
				Arguments: tc.arguments,
			})

			if err != nil {
				t.Fatalf("Unexpected error calling tool: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Check error status
			if result.IsError != tc.expectError {
				t.Errorf("Expected IsError=%v, got %v", tc.expectError, result.IsError)
			}

			// Check content
			text := GetTextContent(result.Content)
			if !tc.checkContent(text) {
				t.Errorf("Content check failed for: %s", text)
			}
		})
	}
}

func TestPullRequestCreateTemplateLoading(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []struct {
		name         string
		setupMock    func(*MockGiteaServer)
		arguments    map[string]any
		expectError  bool
		checkContent func(string) bool
	}{
		{
			name: "template loading when no body provided",
			setupMock: func(mock *MockGiteaServer) {
				// Add template file to mock server
				mock.AddFile("testuser", "testrepo", "main", ".gitea/PULL_REQUEST_TEMPLATE.md", []byte(`# Pull Request Template

## Description
Please describe your changes here.
`))
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"title":      "Feature with template",
				"head":       "feature-branch",
				"base":       "main",
				// No body provided - should load template
			},
			expectError: false,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Pull request created successfully")
			},
		},
		{
			name: "user body overrides template",
			setupMock: func(mock *MockGiteaServer) {
				// Add template file to mock server
				mock.AddFile("testuser", "testrepo", "main", ".gitea/PULL_REQUEST_TEMPLATE.md", []byte(`# Template
This should be overridden`))
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"title":      "Feature with custom body",
				"head":       "feature-branch",
				"base":       "main",
				"body":       "This is my custom description that should override the template",
			},
			expectError: false,
			checkContent: func(text string) bool {
				return strings.Contains(text, "Pull request created successfully") &&
					strings.Contains(text, "Template: Merged repository template")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			mock := NewMockGiteaServer(t)
			tc.setupMock(mock)

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "pr_create",
				Arguments: tc.arguments,
			})

			if err != nil {
				t.Fatalf("Unexpected error calling tool: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			if result.IsError != tc.expectError {
				t.Errorf("Expected IsError=%v, got %v", tc.expectError, result.IsError)
			}

			// Check content
			text := GetTextContent(result.Content)
			if !tc.checkContent(text) {
				t.Errorf("Content check failed for: %s", text)
			}
		})
	}
}

// Helper functions

// createTempNonGitDir creates a temporary directory that is not a git repository
func createTempNonGitDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "forgejo-test-non-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Clean up after test
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}
