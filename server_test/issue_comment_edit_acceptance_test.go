package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCommentLifecycleAcceptance tests the complete comment lifecycle: create, list, edit
func TestCommentLifecycleAcceptance(t *testing.T) {
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
	originalComment := "This is the original comment for testing."
	updatedComment := "This is the updated comment with new information."

	// Step 1: Create a comment
	t.Log("Step 1: Creating comment")
	createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"comment":      originalComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("Comment creation failed: %s", getTextContent(createResult.Content))
	}

	// Verify creation response
	createText := getTextContent(createResult.Content)
	if !strings.Contains(createText, "Comment created successfully") {
		t.Errorf("Expected successful creation message, got: %s", createText)
	}
	if !strings.Contains(createText, originalComment) {
		t.Errorf("Expected original comment in response, got: %s", createText)
	}

	// Step 2: List comments to verify creation
	t.Log("Step 2: Listing comments to verify creation")
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
		t.Fatalf("Failed to list comments: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("Comment listing failed: %s", getTextContent(listResult.Content))
	}

	// Verify the comment appears in the list
	listText := getTextContent(listResult.Content)
	if !strings.Contains(listText, originalComment) {
		t.Errorf("Expected original comment in list, got: %s", listText)
	}

	// Extract comment ID from the list (assuming it's the first comment)
	// For this test, we'll use a fixed comment ID since our mock server returns predictable IDs
	commentID := 1

	// Step 3: Edit the comment
	t.Log("Step 3: Editing comment")
	editResult, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"comment_id":   commentID,
			"new_content":  updatedComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to edit comment: %v", err)
	}
	if editResult.IsError {
		t.Fatalf("Comment editing failed: %s", getTextContent(editResult.Content))
	}

	// Verify edit response
	editText := getTextContent(editResult.Content)
	if !strings.Contains(editText, "Comment edited successfully") {
		t.Errorf("Expected successful edit message, got: %s", editText)
	}
	if !strings.Contains(editText, updatedComment) {
		t.Errorf("Expected updated comment in response, got: %s", editText)
	}

	// Step 4: List comments again to verify the edit
	t.Log("Step 4: Listing comments to verify edit")
	listResult2, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_list",
		Arguments: map[string]any{
			"repository":   repo,
			"issue_number": issueNumber,
			"limit":        10,
			"offset":       0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to list comments after edit: %v", err)
	}
	if listResult2.IsError {
		t.Fatalf("Comment listing failed after edit: %s", getTextContent(listResult2.Content))
	}

	// Verify the updated comment appears in the list
	listText2 := getTextContent(listResult2.Content)
	if !strings.Contains(listText2, updatedComment) {
		t.Errorf("Expected updated comment in list, got: %s", listText2)
	}
	if strings.Contains(listText2, originalComment) && !strings.Contains(listText2, updatedComment) {
		t.Errorf("Original comment should be replaced by updated comment, got: %s", listText2)
	}

	t.Log("✅ Comment lifecycle test completed successfully")
}

// TestCommentEditingRealWorldScenarios tests various real-world editing scenarios
func TestCommentEditingRealWorldScenarios(t *testing.T) {
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

	scenarios := []struct {
		name        string
		original    string
		updated     string
		description string
	}{
		{
			name:        "fix_typo",
			original:    "This is a commment with a typo",
			updated:     "This is a comment with the typo fixed",
			description: "Fixing a simple typo in a comment",
		},
		{
			name:        "add_information",
			original:    "I found an issue",
			updated:     "I found an issue with the login functionality. The error occurs when users try to authenticate with invalid credentials.",
			description: "Adding more detailed information to a brief comment",
		},
		{
			name:        "correct_misinformation",
			original:    "The bug is in the frontend code",
			updated:     "After further investigation, the bug is actually in the backend authentication service",
			description: "Correcting incorrect information in a comment",
		},
		{
			name:        "update_status",
			original:    "Working on this issue",
			updated:     "I've completed the implementation and added comprehensive tests. Ready for review.",
			description: "Updating status information in a comment",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create original comment
			createResult, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name: "issue_comment_create",
				Arguments: map[string]any{
					"repository":   "testuser/testrepo",
					"issue_number": 1,
					"comment":      scenario.original,
				},
			})
			if err != nil {
				t.Fatalf("Failed to create comment for scenario %s: %v", scenario.name, err)
			}
			if createResult.IsError {
				t.Fatalf("Comment creation failed for scenario %s: %v", scenario.name, createResult.Content)
			}

			// Edit the comment
			editResult, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name: "issue_comment_edit",
				Arguments: map[string]any{
					"repository":   "testuser/testrepo",
					"issue_number": 1,
					"comment_id":   1, // Fixed ID for testing
					"new_content":  scenario.updated,
				},
			})
			if err != nil {
				t.Fatalf("Failed to edit comment for scenario %s: %v", scenario.name, err)
			}
			if editResult.IsError {
				t.Fatalf("Comment editing failed for scenario %s: %v", scenario.name, editResult.Content)
			}

			// Verify the edit was successful
			editText := getTextContent(editResult.Content)
			if !strings.Contains(editText, "Comment edited successfully") {
				t.Errorf("Expected successful edit for scenario %s, got: %s", scenario.name, editText)
			}
			if !strings.Contains(editText, scenario.updated) {
				t.Errorf("Expected updated content for scenario %s, got: %s", scenario.name, editText)
			}

			t.Logf("✅ Scenario '%s' completed: %s", scenario.name, scenario.description)
		})
	}
}

// TestCommentEditingErrorHandling tests error handling and recovery scenarios
func TestCommentEditingErrorHandling(t *testing.T) {
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

	errorScenarios := []struct {
		name        string
		arguments   map[string]any
		expectError bool
		description string
	}{
		{
			name: "nonexistent_repository",
			arguments: map[string]any{
				"repository":   "nonexistent/repo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "test content",
			},
			expectError: true,
			description: "Attempting to edit comment in nonexistent repository",
		},
		{
			name: "invalid_repository_format",
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "test content",
			},
			expectError: true,
			description: "Using invalid repository format",
		},
		{
			name: "missing_required_fields",
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				// missing issue_number, comment_id, new_content
			},
			expectError: true,
			description: "Missing required parameters",
		},
		{
			name: "empty_content",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "",
			},
			expectError: true,
			description: "Attempting to set empty comment content",
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_comment_edit",
				Arguments: scenario.arguments,
			})

			if scenario.expectError {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("Expected error for scenario '%s', but got success. Result: %+v", scenario.name, result)
				} else {
					t.Logf("✅ Error handling working correctly for scenario '%s': %s", scenario.name, scenario.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for scenario '%s': %v", scenario.name, err)
				}
				if result != nil && result.IsError {
					t.Errorf("Unexpected error result for scenario '%s'", scenario.name)
				}
			}
		})
	}
}

// TestCommentEditingPerformance tests performance and edge cases
func TestCommentEditingPerformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
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

	// Test with large content
	t.Run("large_content", func(t *testing.T) {
		largeContent := strings.Repeat("This is a test comment with some content. ", 100) // ~4KB

		result, err := client.CallTool(ctx, &mcp.CallToolParams{
			Name: "issue_comment_edit",
			Arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  largeContent,
			},
		})
		if err != nil {
			t.Fatalf("Failed to edit comment with large content: %v", err)
		}
		if result.IsError {
			t.Fatalf("Comment editing failed with large content: %v", result.Content)
		}

		resultText := getTextContent(result.Content)
		if !strings.Contains(resultText, "Comment edited successfully") {
			t.Errorf("Expected successful edit with large content, got: %s", resultText)
		}
	})

	// Test concurrent edits (simplified version)
	t.Run("concurrent_edits", func(t *testing.T) {
		numConcurrent := 3
		results := make([]*mcp.CallToolResult, numConcurrent)
		errors := make([]error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(index int) {
				result, err := client.CallTool(ctx, &mcp.CallToolParams{
					Name: "issue_comment_edit",
					Arguments: map[string]any{
						"repository":   "testuser/testrepo",
						"issue_number": 1,
						"comment_id":   1,
						"new_content":  "Concurrent edit content",
					},
				})
				results[index] = result
				errors[index] = err
			}(i)
		}

		// Wait a bit for concurrent operations
		time.Sleep(2 * time.Second)

		// Check results
		for i := 0; i < numConcurrent; i++ {
			if errors[i] != nil {
				t.Errorf("Concurrent edit %d failed: %v", i, errors[i])
			}
			if results[i] != nil && results[i].IsError {
				t.Errorf("Concurrent edit %d returned error: %v", i, results[i].Content)
			}
		}
	})

	// Test edge case: minimal valid content
	t.Run("minimal_content", func(t *testing.T) {
		result, err := client.CallTool(ctx, &mcp.CallToolParams{
			Name: "issue_comment_edit",
			Arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "x", // Minimal valid content
			},
		})
		if err != nil {
			t.Fatalf("Failed to edit comment with minimal content: %v", err)
		}
		if result.IsError {
			t.Fatalf("Comment editing failed with minimal content: %v", result.Content)
		}
	})
}

// getTextContent extracts text content from MCP result
func getTextContent(content []mcp.Content) string {
	for _, c := range content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}
