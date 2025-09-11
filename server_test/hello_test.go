package servertest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type helloTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestHelloToolTableDriven(t *testing.T) {
	t.Parallel()
	testCases := []helloTestCase{
		{
			name: "acceptance",
			setupMock: func(mock *MockGiteaServer) {
				// No specific mock setup needed for hello tool
			},
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
			},
		},
		{
			name: "empty arguments",
			setupMock: func(mock *MockGiteaServer) {
				// No specific mock setup needed for hello tool
			},
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
			},
		},
		{
			name: "multiple calls simulation",
			setupMock: func(mock *MockGiteaServer) {
				// No specific mock setup needed for hello tool
			},
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
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}
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
