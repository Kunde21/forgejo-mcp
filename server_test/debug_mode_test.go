package servertest

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestDebugModeHelloTool specifically tests that the hello tool is only available in debug mode
func TestDebugModeHelloTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)

	// Test with debug mode enabled
	tsDebug := NewTestServerWithDebug(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, true)
	if err := tsDebug.Initialize(); err != nil {
		t.Fatal(err)
	}
	clientDebug := tsDebug.Client()

	toolsDebug, err := clientDebug.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools in debug mode: %v", err)
	}

	// Count tools in debug mode
	debugToolCount := len(toolsDebug.Tools)
	hasHelloInDebug := false
	for _, tool := range toolsDebug.Tools {
		if tool.Name == "hello" {
			hasHelloInDebug = true
			break
		}
	}

	if !hasHelloInDebug {
		t.Error("Hello tool should be available in debug mode")
	}

	// Test with debug mode disabled (regular mode)
	tsRegular := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := tsRegular.Initialize(); err != nil {
		t.Fatal(err)
	}
	clientRegular := tsRegular.Client()

	toolsRegular, err := clientRegular.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools in regular mode: %v", err)
	}

	// Count tools in regular mode
	regularToolCount := len(toolsRegular.Tools)
	hasHelloInRegular := false
	for _, tool := range toolsRegular.Tools {
		if tool.Name == "hello" {
			hasHelloInRegular = true
			break
		}
	}

	if hasHelloInRegular {
		t.Error("Hello tool should NOT be available in regular mode")
	}

	// Verify that debug mode has exactly one more tool (the hello tool)
	expectedDiff := 1
	actualDiff := debugToolCount - regularToolCount
	if actualDiff != expectedDiff {
		t.Errorf("Expected %d more tools in debug mode, got %d (debug: %d, regular: %d)",
			expectedDiff, actualDiff, debugToolCount, regularToolCount)
	}

	t.Logf("Debug mode: %d tools (includes hello tool)", debugToolCount)
	t.Logf("Regular mode: %d tools (no hello tool)", regularToolCount)
}
