// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
)

// MockTransport implements a simple in-memory transport for testing
type MockTransport struct {
	input  io.Reader
	output io.Writer
}

func (mt *MockTransport) Read(p []byte) (n int, err error) {
	return mt.input.Read(p)
}

func (mt *MockTransport) Write(p []byte) (n int, err error) {
	return mt.output.Write(p)
}

func (mt *MockTransport) Close() error {
	return nil
}

func TestNewStdioTransport_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() should not return error, got: %v", err)
	}
	if server == nil {
		t.Fatal("New() should return non-nil server")
	}

	// Test that we can create a transport (placeholder for now)
	// transport := NewStdioTransport()
	// TODO: Implement transport creation and test it
}

func TestTransportRequestRouting_ValidRequest(t *testing.T) {
	// Test request routing logic (placeholder implementation)
	// This will test the request dispatcher functionality

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "pr_list",
			"arguments": map[string]interface{}{
				"state": "open",
			},
		},
	}

	// TODO: Implement request routing and test it
	// For now, just test that we can marshal/unmarshal JSON
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if !cmp.Equal(request, unmarshaled) {
		t.Errorf("Request marshaling/unmarshaling failed: %v", cmp.Diff(request, unmarshaled))
	}
}

func TestTransportRequestRouting_InvalidRequest(t *testing.T) {
	// Test handling of invalid requests
	invalidRequests := []string{
		`{"invalid": "json"`,
		`{"jsonrpc": "1.0", "method": "invalid"}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "params": {}}`,
	}

	for _, req := range invalidRequests {
		t.Run("InvalidRequest_"+req[:min(20, len(req))], func(t *testing.T) {
			var request map[string]interface{}
			if err := json.Unmarshal([]byte(req), &request); err != nil {
				// This is expected for malformed JSON
				return
			}

			// TODO: Test request validation logic
			// For now, just verify the request structure
			if _, hasJsonrpc := request["jsonrpc"]; !hasJsonrpc {
				t.Log("Request missing jsonrpc field")
			}
		})
	}
}

func TestTransportConnectionLifecycle(t *testing.T) {
	// Test connection lifecycle management
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Test server start/stop with transport
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping server
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Check if Start() returned
	select {
	case err := <-startErr:
		if err != nil {
			t.Errorf("Start() error = %v, expected no error after Stop", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Start() did not return after Stop() was called")
	}
}

func TestTransportTimeoutHandling(t *testing.T) {
	// Test timeout handling for requests
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  1, // Very short timeout for testing
		WriteTimeout: 1,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// TODO: Test actual timeout behavior
	// For now, just verify server creation with timeout config
	if server.config.ReadTimeout != 1 {
		t.Errorf("Expected ReadTimeout to be 1, got %d", server.config.ReadTimeout)
	}
	if server.config.WriteTimeout != 1 {
		t.Errorf("Expected WriteTimeout to be 1, got %d", server.config.WriteTimeout)
	}
}

func TestTransportErrorResponse(t *testing.T) {
	// Test error response formatting
	testCases := []struct {
		name         string
		errorCode    int
		errorMessage string
		expectedJSON string
	}{
		{
			name:         "MethodNotFound",
			errorCode:    -32601,
			errorMessage: "Method not found",
			expectedJSON: `{"jsonrpc":"2.0","id":null,"error":{"code":-32601,"message":"Method not found"}}`,
		},
		{
			name:         "InvalidParams",
			errorCode:    -32602,
			errorMessage: "Invalid parameters",
			expectedJSON: `{"jsonrpc":"2.0","id":null,"error":{"code":-32602,"message":"Invalid parameters"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: Implement error response formatting
			// For now, just test JSON structure
			errorResponse := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      nil,
				"error": map[string]interface{}{
					"code":    tc.errorCode,
					"message": tc.errorMessage,
				},
			}

			data, err := json.Marshal(errorResponse)
			if err != nil {
				t.Fatalf("Failed to marshal error response: %v", err)
			}

			// Verify the JSON contains expected error structure
			if !strings.Contains(string(data), `"error"`) {
				t.Error("Error response should contain error field")
			}
			if !strings.Contains(string(data), `"code"`) {
				t.Error("Error response should contain error code")
			}
			if !strings.Contains(string(data), `"message"`) {
				t.Error("Error response should contain error message")
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
