package servertest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type helloTestCase struct {
	name      string
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestHelloToolTableDriven(t *testing.T) {
	t.Parallel()
	testCases := []helloTestCase{
		{
			name:      "acceptance",
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
			},
		},
		{
			name:      "empty arguments",
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
			},
		},
		{
			name:      "multiple calls simulation",
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockGiteaServer(t)
			ts := NewTestServer(t, t.Context(), map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "hello",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call hello tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}

// TestHelloToolConcurrent tests concurrent hello tool calls
func TestHelloToolConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	results := make(chan error, numGoroutines)
	for range numGoroutines {
		go func() {
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "hello",
				Arguments: map[string]any{},
			})
			results <- err
		}()
	}
	for range numGoroutines {
		if err := <-results; err != nil {
			t.Errorf("Concurrent hello tool call failed: %v", err)
		}
	}
}

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

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	// Verify all requests succeeded
	for i, result := range results {
		if result == "" {
			t.Errorf("Request %d returned empty result", i)
		}
	}
}
