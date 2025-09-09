package servertest

import (
	"context"
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
	if len(tools.Tools) != 3 {
		t.Fatalf("Expected 3 tools, got %d", len(tools.Tools))
	}

	// Find tools
	var helloTool *mcp.Tool
	var listIssuesTool *mcp.Tool
	var createCommentTool *mcp.Tool
	for _, tool := range tools.Tools {
		switch tool.Name {
		case "hello":
			helloTool = tool
		case "list_issues":
			listIssuesTool = tool
		case "create_issue_comment":
			createCommentTool = tool
		}
	}

	if helloTool == nil {
		t.Fatal("hello tool not found")
	}
	if helloTool.Description != "Returns a hello world message" {
		t.Errorf("Expected hello tool description 'Returns a hello world message', got '%s'", helloTool.Description)
	}

	if listIssuesTool == nil {
		t.Fatal("list_issues tool not found")
	}
	if listIssuesTool.Description != "List issues from a Gitea/Forgejo repository" {
		t.Errorf("Expected list_issues tool description 'List issues from a Gitea/Forgejo repository', got '%s'", listIssuesTool.Description)
	}

	if createCommentTool == nil {
		t.Fatal("create_issue_comment tool not found")
	}
	if createCommentTool.Description != "Create a comment on a Forgejo/Gitea repository issue" {
		t.Errorf("Expected create_issue_comment tool description 'Create a comment on a Forgejo/Gitea repository issue', got '%s'", createCommentTool.Description)
	}

	// Verify tool has input schema
	if createCommentTool.InputSchema == nil {
		t.Error("create_issue_comment tool should have input schema")
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
