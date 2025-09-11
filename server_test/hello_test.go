package servertest

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHelloToolAcceptance tests the hello tool with mock server
func TestHelloToolAcceptance(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test successful hello tool execution
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result")
	}
}

// TestHelloToolEmptyArguments tests hello tool with empty arguments map
func TestHelloToolEmptyArguments(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test hello tool with empty arguments (should work)
	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Failed to call hello tool with empty arguments: %v", err)
	}

	if result.Content == nil {
		t.Error("Expected content in result with empty arguments")
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

// TestHelloToolMultipleCalls tests multiple sequential hello tool calls
func TestHelloToolMultipleCalls(t *testing.T) {
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Make multiple sequential calls
	for i := 0; i < 10; i++ {
		result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "hello",
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatalf("Failed to call hello tool on attempt %d: %v", i+1, err)
		}

		if result.Content == nil {
			t.Errorf("Expected content in result on attempt %d", i+1)
		}
	}
}
