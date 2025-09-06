package servertest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestTestServerStartAndIsRunning(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if ts.IsRunning() {
		t.Error("IsRunning should be false before start")
	}
	if err := ts.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	if !ts.IsRunning() {
		t.Error("IsRunning should be true after start")
	}
}

// TestServerLifecycle tests the complete server lifecycle: start, run, and stop
func TestServerLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Start(); err != nil {
		t.Fatal("Failed to start server:", err)
	}
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()
	result, err := client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "hello"}})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("Expected tool result to have content")
	}
}

// TestMCPInitialization tests MCP protocol initialization
func TestMCPInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Start(); err != nil {
		t.Fatal("Failed to start server:", err)
	}
	client := ts.Client()
	result, err := client.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize MCP protocol: %v", err)
	}
	if result.ServerInfo.Name != "forgejo-mcp" {
		t.Errorf("Expected server name 'forgejo-mcp', got '%s'", result.ServerInfo.Name)
	}
}

// TestToolDiscovery tests tool discovery functionality
func TestToolDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// List available tools
	tools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}
	want := &mcp.ListToolsResult{
		Tools: []mcp.Tool{{
			Name: "hello", Description: "Returns a hello world message",
			InputSchema: mcp.ToolInputSchema{Type: "object"},
		}},
	}
	opts := cmpopts.IgnoreFields(mcp.Tool{}, "Annotations")
	if !cmp.Equal(want, tools, opts) {
		t.Error(cmp.Diff(want, tools, opts))
	}
}

// TestServerStart tests that server can start without error
func TestServerStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	if !ts.IsRunning() {
		t.Error("Server should be running after start")
	}
}

// TestHelloTool tests that the hello tool returns correct response
func TestHelloTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "hello"},
	})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}
	want := &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.NewTextContent("Hello, World!")},
		StructuredContent: nil,
	}
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestHelloToolWithNilContext tests error handling with nil context
func TestHelloToolWithNilContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with cancelled context should return error
	cancelledCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc() // Cancel immediately

	result, err := client.CallTool(cancelledCtx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "hello"},
	})
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
	ts := NewTestServer(t, ctx)
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test calling the "hello" tool
	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "hello"},
	})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}
	want := &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.NewTextContent("Hello, World!")},
		StructuredContent: nil,
	}
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test calling a non-existent tool
	_, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "nonexistent_tool"},
	})
	if err == nil {
		t.Error("Expected error when calling non-existent tool")
	}

	// Test calling tool with invalid parameters (if applicable)
	// For hello tool, no params needed, so this should still work
	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "hello",
			Arguments: map[string]any{"invalid": "param"},
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error with extra params: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("Expected tool result to have content even with extra params")
	}
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	ts := NewTestServer(t, ctx)
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
			result, err := client.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{Name: "hello"},
			})
			if err != nil {
				errors[index] = err
				return
			}
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
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
