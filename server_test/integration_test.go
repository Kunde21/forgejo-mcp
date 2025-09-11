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

// TestToolDiscovery tests tool discovery functionality
func TestToolDiscovery(t *testing.T) {
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

	// List available tools
	tools, err := client.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}
	// Check that we have the expected tools
	if len(tools.Tools) != 4 {
		t.Fatalf("Expected 4 tools, got %d", len(tools.Tools))
	}

	// Find tools
	var helloTool *mcp.Tool
	var listIssuesTool *mcp.Tool
	var createCommentTool *mcp.Tool
	var listCommentsTool *mcp.Tool
	for _, tool := range tools.Tools {
		switch tool.Name {
		case "hello":
			helloTool = tool
		case "issue_list":
			listIssuesTool = tool
		case "issue_comment_create":
			createCommentTool = tool
		case "issue_comment_list":
			listCommentsTool = tool
		}
	}

	if helloTool == nil {
		t.Fatal("hello tool not found")
	}
	if helloTool.Description != "Returns a hello world message" {
		t.Errorf("Expected hello tool description 'Returns a hello world message', got '%s'", helloTool.Description)
	}

	if listIssuesTool == nil {
		t.Fatal("issue_list tool not found")
	}
	if listIssuesTool.Description != "List issues from a Gitea/Forgejo repository" {
		t.Errorf("Expected issue_list tool description 'List issues from a Gitea/Forgejo repository', got '%s'", listIssuesTool.Description)
	}

	if createCommentTool == nil {
		t.Fatal("issue_comment_create tool not found")
	}
	if createCommentTool.Description != "Create a comment on a Forgejo/Gitea repository issue" {
		t.Errorf("Expected issue_comment_create tool description 'Create a comment on a Forgejo/Gitea repository issue', got '%s'", createCommentTool.Description)
	}

	if listCommentsTool == nil {
		t.Fatal("issue_comment_list tool not found")
	}
	if listCommentsTool.Description != "List comments from a Forgejo/Gitea repository issue with pagination support" {
		t.Errorf("Expected issue_comment_list tool description 'List comments from a Forgejo/Gitea repository issue with pagination support', got '%s'", listCommentsTool.Description)
	}

	// Verify tool has input schema
	if createCommentTool.InputSchema == nil {
		t.Error("issue_comment_create tool should have input schema")
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
	wg.Wait()

	// Check results
	for i := range numRequests {
		if errors[i] != nil {
			t.Errorf("Concurrent request %d failed: %v", i, errors[i])
		}
		if results[i] != "Hello, World!" {
			t.Errorf("Concurrent request %d got unexpected result: %s", i, results[i])
		}
	}
}

// TestCreateIssueCommentToolSuccess tests successful comment creation
func TestCreateIssueCommentToolSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock client
	mockClient := NewMockGiteaClient()

	ts := NewTestServerWithClient(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": "http://mock.localhost",
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, mockClient)

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 42,
			"comment":      "This is a test comment",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_comment_create tool: %v", err)
	}

	// Verify the result structure
	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	// Check that the response contains expected information
	expectedParts := []string{"Comment created successfully", "ID:", "Created:"}
	for _, part := range expectedParts {
		if !strings.Contains(textContent.Text, part) {
			t.Errorf("Expected response to contain '%s', got: %s", part, textContent.Text)
		}
	}
}

// TestCreateIssueCommentToolValidationErrors tests parameter validation
func TestCreateIssueCommentToolValidationErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock client
	mockClient := NewMockGiteaClient()

	ts := NewTestServerWithClient(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": "http://mock.localhost",
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, mockClient)

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	testCases := []struct {
		name        string
		arguments   map[string]any
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing repository",
			arguments: map[string]any{
				"issue_number": 42,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository":   "invalid-format",
				"issue_number": 42,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "repository",
		},
		{
			name: "negative issue number",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": -1,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue_number",
		},
		{
			name: "zero issue number",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 0,
				"comment":      "Test comment",
			},
			expectError: true,
			errorMsg:    "issue number",
		},
		{
			name: "empty comment",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 42,
				"comment":      "",
			},
			expectError: true,
			errorMsg:    "comment",
		},
		{
			name: "missing comment",
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 42,
			},
			expectError: true,
			errorMsg:    "comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "issue_comment_create",
				Arguments: tc.arguments,
			})

			if tc.expectError {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("Expected error for test case '%s', but got success", tc.name)
				}
				if result != nil && len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						if !strings.Contains(textContent.Text, tc.errorMsg) {
							t.Errorf("Expected error message to contain '%s', got: %s", tc.errorMsg, textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected success for test case '%s', but got error: %v", tc.name, err)
				}
			}
		})
	}
}

// TestCreateIssueCommentToolAPIError tests API error scenarios
func TestCreateIssueCommentToolAPIError(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock client that returns errors
	mockClient := NewMockGiteaClient()

	ts := NewTestServerWithClient(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": "http://mock.localhost",
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, mockClient)

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with repository that doesn't exist (mock will return error)
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "nonexistent/repo",
			"issue_number": 42,
			"comment":      "Test comment",
		},
	})

	if err == nil && (result == nil || !result.IsError) {
		t.Error("Expected error for nonexistent repository")
	}
}

// TestCreateIssueCommentToolCancelledContext tests context cancellation
func TestCreateIssueCommentToolCancelledContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock client
	mockClient := NewMockGiteaClient()

	ts := NewTestServerWithClient(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": "http://mock.localhost",
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, mockClient)

	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Cancel context immediately
	cancelledCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc()

	result, err := client.CallTool(cancelledCtx, &mcp.CallToolParams{
		Name: "issue_comment_create",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 42,
			"comment":      "Test comment",
		},
	})

	if err == nil {
		t.Error("Expected error when calling tool with cancelled context")
	}
	if result != nil && !result.IsError {
		t.Error("Expected error result for cancelled context")
	}
}
