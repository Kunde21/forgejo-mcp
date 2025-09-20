package servertest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCase represents a test case for the hello tool following the style guide pattern
type TestCase struct {
	name      string
	setupMock func(*MockGiteaServer) // Optional: for tests that need mock setup
	arguments map[string]any
	expect    *mcp.CallToolResult
}

// TestHelloToolTableDriven tests the hello tool using table-driven pattern
// This consolidates all basic functionality tests into a single, comprehensive test
func TestHelloToolTableDriven(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	testCases := []TestCase{
		{
			name:      "basic functionality - returns hello world message",
			setupMock: nil, // Hello tool doesn't need mock setup
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
				IsError: false,
			},
		},
		{
			name:      "empty arguments - handles gracefully",
			setupMock: nil,
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
				IsError: false,
			},
		},
		{
			name:      "nil arguments - handles gracefully",
			setupMock: nil,
			arguments: nil,
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello, World!"},
				},
				IsError: false,
			},
		},
		{
			name:      "schema validation - rejects extra arguments",
			setupMock: nil,
			arguments: map[string]any{"extra": "ignored", "another": 123},
			expect:    nil, // This should fail at the client level, so we handle it specially
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			// Create mock server even though hello tool doesn't use it
			// The test harness requires a valid server URL
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}
			client := ts.Client()

			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      "hello",
				Arguments: tc.arguments,
			})

			// Special handling for schema validation test
			if tc.name == "schema validation - rejects extra arguments" {
				if err == nil {
					t.Error("Expected error when calling tool with extra arguments")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to call hello tool: %v", err)
			}

			if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
				t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
			}
		})
	}
}

// TestHelloToolErrorHandling tests error scenarios and edge cases
func TestHelloToolErrorHandling(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	testCases := []TestCase{
		{
			name:      "non-existent tool - returns error",
			setupMock: nil,
			arguments: map[string]any{},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "tool not found: nonexistent_tool"},
				},
				IsError: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			// Create mock server even though hello tool doesn't use it
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			ts := NewTestServer(t, ctx, map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}
			client := ts.Client()

			// For non-existent tool test, we call a different tool name
			toolName := "hello"
			if tc.name == "non-existent tool - returns error" {
				toolName = "nonexistent_tool"
			}

			result, err := client.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: tc.arguments,
			})

			if tc.name == "non-existent tool - returns error" {
				if err == nil {
					t.Error("Expected error when calling non-existent tool")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to call hello tool: %v", err)
			}

			if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
				t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
			}
		})
	}
}

// TestHelloToolConcurrent tests concurrent hello tool calls following the style guide pattern
func TestHelloToolConcurrent(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Create mock server even though hello tool doesn't use it
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	// Launch concurrent requests
	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "hello",
				Arguments: map[string]any{},
			})
			results <- err
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check for errors
	errorCount := 0
	for err := range results {
		if err != nil {
			t.Errorf("Concurrent hello tool call failed: %v", err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("%d out of %d concurrent requests failed", errorCount, numGoroutines)
	}
}

// TestHelloToolContextCancellation tests context cancellation behavior
func TestHelloToolContextCancellation(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock server even though hello tool doesn't use it
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test with cancelled context
	cancelledCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc() // Cancel immediately

	_, err := ts.Client().CallTool(cancelledCtx, &mcp.CallToolParams{
		Name:      "hello",
		Arguments: map[string]any{},
	})

	if err == nil {
		t.Error("Expected error when calling tool with cancelled context")
	}
}

// TestHelloToolPerformance tests performance with multiple rapid calls
func TestHelloToolPerformance(t *testing.T) {
	// Note: t.Parallel() is not compatible with t.Setenv() used in NewTestServer
	// Performance testing can still be done without parallel execution

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	// Create mock server even though hello tool doesn't use it
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numRequests = 50
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	start := time.Now()

	for range numRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name:      "hello",
				Arguments: map[string]any{},
			})

			if err == nil && result != nil && !result.IsError {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// Verify all requests succeeded
	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}

	// Log performance metrics (not a failure condition, just for monitoring)
	t.Logf("Completed %d requests in %v (%.2f req/sec)",
		numRequests, duration, float64(numRequests)/duration.Seconds())
}

// TestMCPInitialization tests MCP protocol initialization
func TestMCPInitialization(t *testing.T) {
	// Note: t.Parallel() disabled due to incompatibility with t.Setenv() used in test harness

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	// Create mock server even though hello tool doesn't use it
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
