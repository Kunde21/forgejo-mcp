package servertest

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestDirectoryParameterConsistency tests that directory parameter resolution
// works consistently across all tools that support it
func TestDirectoryParameterConsistency(t *testing.T) {
	t.Parallel()

	// Create a temporary git repository for testing
	tempDir := createTempGitRepo(t, "testuser", "testrepo")
	defer os.RemoveAll(tempDir)

	mock := NewMockGiteaServer(t)

	// Set up mock data for all tools
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 1, Content: "Test comment", Author: "testuser", Created: "2025-09-11T10:30:00Z", Updated: "2025-09-11T10:30:00Z"},
	})

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	tools := []string{
		"issue_list",
		"pr_list",
		"issue_comment_create",
		"issue_comment_list",
		"issue_comment_edit",
		"pr_comment_create",
		"pr_comment_list",
		"pr_comment_edit",
	}

	// Test that all tools accept directory parameter without error
	for _, tool := range tools {
		t.Run(tool+"_directory_resolution", func(t *testing.T) {
			args := map[string]any{
				"directory": tempDir,
			}

			// Add tool-specific required parameters
			switch tool {
			case "issue_list":
				args["limit"] = 10
				args["offset"] = 0
			case "pr_list":
				args["limit"] = 10
				args["offset"] = 0
				args["state"] = "open"
			case "issue_comment_create":
				args["issue_number"] = 1
				args["comment"] = "Test comment"
			case "issue_comment_list":
				args["issue_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "issue_comment_edit":
				args["issue_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated comment"
			case "pr_comment_create":
				args["pull_request_number"] = 1
				args["comment"] = "Test PR comment"
			case "pr_comment_list":
				args["pull_request_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "pr_comment_edit":
				args["pull_request_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated PR comment"
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      tool,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("Tool %s failed with directory parameter: %v", tool, err)
			}

			// Should not be an error (directory resolution should succeed)
			if result.IsError {
				t.Errorf("Tool %s returned error with valid directory: %s", tool, GetTextContent(result.Content))
			}
		})
	}
}

// TestDirectoryParameterErrorConsistency tests that error handling is consistent
// across all tools when directory resolution fails
func TestDirectoryParameterErrorConsistency(t *testing.T) {
	t.Parallel()

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	tools := []string{
		"issue_list",
		"pr_list",
		"issue_comment_create",
		"issue_comment_list",
		"issue_comment_edit",
		"pr_comment_create",
		"pr_comment_list",
		"pr_comment_edit",
	}

	errorScenarios := []struct {
		name        string
		directory   string
		expectError string
	}{
		{
			name:        "non_existent_directory",
			directory:   "/non/existent/directory",
			expectError: "Invalid request: directory: invalid directory.",
		},
		{
			name:        "empty_directory_string",
			directory:   "",
			expectError: "Invalid request: directory: at least one of directory or repository must be provided",
		},
	}

	for _, scenario := range errorScenarios {
		for _, tool := range tools {
			t.Run(tool+"_"+scenario.name, func(t *testing.T) {
				args := map[string]any{
					"directory": scenario.directory,
				}

				// Add tool-specific required parameters
				switch tool {
				case "issue_list":
					args["limit"] = 10
					args["offset"] = 0
				case "pr_list":
					args["limit"] = 10
					args["offset"] = 0
					args["state"] = "open"
				case "issue_comment_create":
					args["issue_number"] = 1
					args["comment"] = "Test comment"
				case "issue_comment_list":
					args["issue_number"] = 1
					args["limit"] = 10
					args["offset"] = 0
				case "issue_comment_edit":
					args["issue_number"] = 1
					args["comment_id"] = 1
					args["new_content"] = "Updated comment"
				case "pr_comment_create":
					args["pull_request_number"] = 1
					args["comment"] = "Test PR comment"
				case "pr_comment_list":
					args["pull_request_number"] = 1
					args["limit"] = 10
					args["offset"] = 0
				case "pr_comment_edit":
					args["pull_request_number"] = 1
					args["comment_id"] = 1
					args["new_content"] = "Updated PR comment"
				}

				result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
					Name:      tool,
					Arguments: args,
				})
				if err != nil {
					t.Fatalf("Tool %s failed: %v", tool, err)
				}

				if !result.IsError {
					t.Errorf("Tool %s should have returned error for %s", tool, scenario.name)
				}

				textContent := GetTextContent(result.Content)
				if !strings.Contains(textContent, scenario.expectError) {
					t.Errorf("Tool %s error message doesn't match expected pattern. Got: %s, Expected to contain: %s",
						tool, textContent, scenario.expectError)
				}
			})
		}
	}
}

// TestRepositoryParameterBackwardCompatibility tests that repository parameter
// still works for backward compatibility
func TestRepositoryParameterBackwardCompatibility(t *testing.T) {
	t.Parallel()

	mock := NewMockGiteaServer(t)

	// Set up mock data for all tools
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 1, Content: "Test comment", Author: "testuser", Created: "2025-09-11T10:30:00Z", Updated: "2025-09-11T10:30:00Z"},
	})

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	tools := []string{
		"issue_list",
		"pr_list",
		"issue_comment_create",
		"issue_comment_list",
		"issue_comment_edit",
		"pr_comment_create",
		"pr_comment_list",
		"pr_comment_edit",
	}

	// Test that all tools still accept repository parameter
	for _, tool := range tools {
		t.Run(tool+"_repository_backward_compat", func(t *testing.T) {
			args := map[string]any{
				"repository": "testuser/testrepo",
			}

			// Add tool-specific required parameters
			switch tool {
			case "issue_list":
				args["limit"] = 10
				args["offset"] = 0
			case "pr_list":
				args["limit"] = 10
				args["offset"] = 0
				args["state"] = "open"
			case "issue_comment_create":
				args["issue_number"] = 1
				args["comment"] = "Test comment"
			case "issue_comment_list":
				args["issue_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "issue_comment_edit":
				args["issue_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated comment"
			case "pr_comment_create":
				args["pull_request_number"] = 1
				args["comment"] = "Test PR comment"
			case "pr_comment_list":
				args["pull_request_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "pr_comment_edit":
				args["pull_request_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated PR comment"
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      tool,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("Tool %s failed with repository parameter: %v", tool, err)
			}

			// Should not be an error (repository parameter should work)
			if result.IsError {
				t.Errorf("Tool %s returned error with repository parameter: %s", tool, GetTextContent(result.Content))
			}
		})
	}
}

// TestDirectoryParameterEdgeCases tests edge cases for directory parameter
func TestDirectoryParameterEdgeCases(t *testing.T) {
	t.Parallel()

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	edgeCases := []struct {
		name        string
		directory   string
		expectError string
	}{
		{
			name:        "non_existent_absolute_directory",
			directory:   "/definitely/does/not/exist",
			expectError: "Invalid request: directory: invalid directory.",
		},
		{
			name:        "directory_with_spaces",
			directory:   "/tmp/path with spaces",
			expectError: "Invalid request: directory: invalid directory.",
		},
		{
			name:        "very_long_directory_path",
			directory:   "/very/long/path/that/might/exceed/limits/and/cause/issues/with/file/system/operations/test",
			expectError: "Invalid request: directory: invalid directory.",
		},
		{
			name:        "relative_directory",
			directory:   "./relative/path",
			expectError: "Invalid request: directory: directory must be an absolute path.",
		},
		{
			name:        "empty_directory_after_trim",
			directory:   "   ",
			expectError: "Invalid request: directory: directory must be an absolute path.",
		},
	}

	for _, ec := range edgeCases {
		t.Run("issue_list_"+ec.name, func(t *testing.T) {
			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "issue_list",
				Arguments: map[string]any{
					"directory": ec.directory,
					"limit":     10,
					"offset":    0,
				},
			})
			if err != nil {
				t.Fatalf("Tool issue_list failed: %v", err)
			}

			if !result.IsError {
				t.Errorf("Expected error for %s", ec.name)
			}

			textContent := GetTextContent(result.Content)
			if !strings.Contains(textContent, ec.expectError) {
				t.Errorf("Error message doesn't match expected for %s. Got: %s, Expected to contain: %s",
					ec.name, textContent, ec.expectError)
			}
		})
	}
}

// TestDirectoryVsRepositoryParameter tests that both parameters cannot be provided simultaneously
func TestDirectoryVsRepositoryParameter(t *testing.T) {
	t.Parallel()

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	tools := []string{
		"issue_list",
		"pr_list",
		"issue_comment_create",
		"issue_comment_list",
		"issue_comment_edit",
		"pr_comment_create",
		"pr_comment_list",
		"pr_comment_edit",
	}

	for _, tool := range tools {
		t.Run(tool+"_both_parameters_error", func(t *testing.T) {
			args := map[string]any{
				"directory":  "/some/directory",
				"repository": "testuser/testrepo",
			}

			// Add tool-specific required parameters
			switch tool {
			case "issue_list":
				args["limit"] = 10
				args["offset"] = 0
			case "pr_list":
				args["limit"] = 10
				args["offset"] = 0
				args["state"] = "open"
			case "issue_comment_create":
				args["issue_number"] = 1
				args["comment"] = "Test comment"
			case "issue_comment_list":
				args["issue_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "issue_comment_edit":
				args["issue_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated comment"
			case "pr_comment_create":
				args["pull_request_number"] = 1
				args["comment"] = "Test PR comment"
			case "pr_comment_list":
				args["pull_request_number"] = 1
				args["limit"] = 10
				args["offset"] = 0
			case "pr_comment_edit":
				args["pull_request_number"] = 1
				args["comment_id"] = 1
				args["new_content"] = "Updated PR comment"
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      tool,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("Tool %s failed: %v", tool, err)
			}

			if !result.IsError {
				t.Errorf("Tool %s should have returned error when both directory and repository provided", tool)
			}

			textContent := GetTextContent(result.Content)
			if !strings.Contains(textContent, "Invalid request: directory:") &&
				!strings.Contains(textContent, "Failed to resolve directory:") {
				t.Errorf("Tool %s error message doesn't indicate parameter conflict. Got: %s", tool, textContent)
			}
		})
	}
}
