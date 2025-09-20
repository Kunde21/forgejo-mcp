package servertest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type issueListTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	setupDir  func(t *testing.T) string // Optional function to set up a temporary directory
	arguments map[string]any
	expect    *mcp.CallToolResult
}

// createTempGitRepo creates a temporary directory with a mock .git structure
// that points to testuser/testrepo for testing directory parameter functionality
func createTempGitRepo(t *testing.T, owner, repo string) string {
	tempDir, err := os.MkdirTemp("", "forgejo-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create .git/config with a mock remote
	configPath := filepath.Join(gitDir, "config")
	configContent := fmt.Sprintf(`[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://example.com/%s/%s.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`, owner, repo)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Clean up after test
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

func TestListIssues(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness
	testCases := []issueListTestCase{
		{
			name: "acceptance - real world scenario",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Bug: Login fails", State: "open"},
					{Index: 2, Title: "Feature: Add dark mode", State: "open"},
					{Index: 3, Title: "Fix: Memory leak", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 issues"}, // Only open issues are returned by default
				},
				StructuredContent: map[string]any{
					"issues": []any{
						map[string]any{"number": float64(1), "title": "Bug: Login fails", "state": "open"},
						map[string]any{"number": float64(2), "title": "Feature: Add dark mode", "state": "open"},
					},
				},
			},
		},
		{
			name: "pagination - large dataset handling",
			setupMock: func(mock *MockGiteaServer) {
				var issues []MockIssue
				for i := 1; i <= 25; i++ {
					issues = append(issues, MockIssue{
						Index: i,
						Title: fmt.Sprintf("Issue %d", i),
						State: "open",
					})
				}
				mock.AddIssues("testuser", "testrepo", issues)
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 issues"}, // Only first 10 due to limit
				},
				StructuredContent: map[string]any{
					"issues": func() []any {
						var issues []any
						for i := 1; i <= 10; i++ { // Only first 10 due to limit
							issues = append(issues, map[string]any{
								"number": float64(i),
								"title":  fmt.Sprintf("Issue %d", i),
								"state":  "open",
							})
						}
						return issues
					}(),
				},
			},
		},
		{
			name: "invalid repository format",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"repository": "invalid-repo-format",
				"limit":      10,
				"offset":     0,
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
			name: "missing repository and directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"limit":  10,
				"offset": 0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Test Issue 1", State: "open"},
					{Index: 2, Title: "Test Issue 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      200, // Invalid: > 100
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "negative offset",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Test Issue 1", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     -1, // Invalid: negative
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "empty repository",
			setupMock: func(mock *MockGiteaServer) {
				// Add empty issues list
				mock.AddIssues("testuser", "emptyrepo", []MockIssue{})
			},
			arguments: map[string]any{
				"repository": "testuser/emptyrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 issues"},
				},
				StructuredContent: map[string]any{},
			},
		},
		{
			name: "directory parameter - valid git repository",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Directory-based issue", State: "open"},
				})
			},
			setupDir: func(t *testing.T) string {
				return createTempGitRepo(t, "testuser", "testrepo")
			},
			arguments: map[string]any{
				"directory": "", // Will be set dynamically
				"limit":     10,
				"offset":    0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 issues"},
				},
				StructuredContent: map[string]any{
					"issues": []any{
						map[string]any{"number": float64(1), "title": "Directory-based issue", "state": "open"},
					},
				},
			},
		},
		{
			name: "directory parameter - non-existent directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory": "/non/existent/directory",
				"limit":     10,
				"offset":    0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "directory parameter - directory without git",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			setupDir: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "forgejo-test-no-git-*")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}
				// Don't create .git directory - this should fail validation
				t.Cleanup(func() {
					os.RemoveAll(tempDir)
				})
				return tempDir
			},
			arguments: map[string]any{
				"directory": "", // Will be set dynamically
				"limit":     10,
				"offset":    0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to resolve directory: not a git repository: "}, // Will be followed by actual path
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "directory takes precedence - both directory and repository provided",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"directory":  "/tmp/test-repo",
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to resolve directory: repository validate failed for /tmp/test-repo: directory does not exist"},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "missing repository and directory",
			setupMock: func(mock *MockGiteaServer) {
				// No mock setup needed for error case
			},
			arguments: map[string]any{
				"limit":  10,
				"offset": 0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "repository parameter - deprecation warning",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Legacy repository issue", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 issues"},
				},
				StructuredContent: map[string]any{
					"issues": []any{
						map[string]any{"number": float64(1), "title": "Legacy repository issue", "state": "open"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context with timeout and proper cleanup
			ctx, cancel := CreateStandardTestContext(t, 10)
			defer cancel()

			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Set up temporary directory if needed
			var tempDir string
			if tc.setupDir != nil {
				tempDir = tc.setupDir(t)
				// Update arguments with the actual temp directory path
				args := make(map[string]any)
				for k, v := range tc.arguments {
					args[k] = v
				}
				if dir, ok := args["directory"].(string); ok && dir == "" {
					args["directory"] = tempDir
				}
				tc.arguments = args
			}

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			// Use standardized tool call with validation
			result, err := ts.CallToolWithValidation(ctx, "issue_list", tc.arguments)
			if err != nil {
				t.Fatalf("Failed to call issue_list tool: %v", err)
			}

			// Special handling for directory without git test (dynamic path)
			if tc.name == "directory parameter - directory without git" {
				if !result.IsError {
					t.Errorf("Expected error result for test case: %s", tc.name)
				}
				textContent := GetTextContent(result.Content)
				if !strings.Contains(textContent, "Failed to resolve directory: not a git repository:") {
					t.Errorf("Expected error message containing 'Failed to resolve directory: not a git repository:', got: %s", textContent)
				}
			} else {
				// Use standardized validation with proper comparison options
				if !ts.ValidateToolResult(tc.expect, result, t) {
					t.Errorf("Tool result validation failed for test case: %s", tc.name)
				}
			}
		})
	}
}

func TestListIssuesPerformance(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	// Create test context with timeout and proper cleanup
	ctx, cancel := CreateStandardTestContext(t, 30)
	defer cancel()

	// Create a large dataset for performance testing
	var issues []MockIssue
	for i := 1; i <= 100; i++ {
		issues = append(issues, MockIssue{
			Index: i,
			Title: fmt.Sprintf("Performance Test Issue %d - This is a longer title to test content handling", i),
			State: "open",
		})
	}

	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", issues)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with maximum limit to validate performance with large datasets
	result, err := ts.CallToolWithValidation(ctx, "issue_list", map[string]any{
		"repository": "testuser/testrepo",
		"limit":      100,
		"offset":     0,
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool for performance test: %v", err)
	}

	// Validate that we got all 100 issues
	structured := GetStructuredContent(result)
	if issuesList, ok := structured["issues"].([]any); ok {
		if len(issuesList) != 100 {
			t.Errorf("Expected 100 issues in performance test, got %d", len(issuesList))
		}
	} else {
		t.Error("Expected structured content to contain issues array")
	}

	// Validate success message
	if !ts.ValidateSuccessResult(result, "Found 100 issues", t) {
		t.Error("Performance test failed to validate success result")
	}
}

func TestListIssuesConcurrent(t *testing.T) {
	// Note: t.Parallel() is not compatible with t.Setenv() used in NewTestServer
	// Concurrent testing can still be done without parallel execution

	// Create test context with timeout and proper cleanup
	ctx, cancel := CreateStandardTestContext(t, 15)
	defer cancel()

	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Concurrent Issue 1", State: "open"},
		{Index: 2, Title: "Concurrent Issue 2", State: "open"},
	})
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5

	// Use standardized concurrent test helper
	RunConcurrentTest(t, numGoroutines, func(id int) error {
		_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
			Name: "issue_list",
			Arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
			},
		})
		return err
	})
}
