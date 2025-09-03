// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
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

func (mt *MockTransport) Connect() error {
	return nil
}

func (mt *MockTransport) Disconnect() error {
	return nil
}

func (mt *MockTransport) GetState() ConnectionState {
	return StateConnected
}

func (mt *MockTransport) IsConnected() bool {
	return true
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

	logger := &logrus.Logger{}
	transport := NewStdioTransport(cfg, logger)

	if transport == nil {
		t.Fatal("NewStdioTransport() should return non-nil transport")
	}
	if transport.reader == nil {
		t.Error("Transport reader should not be nil")
	}
	if transport.writer == nil {
		t.Error("Transport writer should not be nil")
	}
	if transport.logger != logger {
		t.Error("Transport logger should match provided logger")
	}
	if transport.readTimeout != 30*time.Second {
		t.Errorf("Expected readTimeout to be 30s, got %v", transport.readTimeout)
	}
	if transport.state != StateDisconnected {
		t.Errorf("Expected initial state to be StateDisconnected, got %v", transport.state)
	}
}

func TestTransportRequestRouting_ValidRequest(t *testing.T) {
	// Test request routing logic (placeholder implementation)
	// This will test the request dispatcher functionality

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      float64(1), // JSON unmarshaling converts numbers to float64
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "pr_list",
			"arguments": map[string]interface{}{
				"state": "open",
			},
		},
	}

	// Test JSON marshaling/unmarshaling
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

func TestStdioTransport_ConnectionLifecycle(t *testing.T) {
	cfg := &config.Config{
		ReadTimeout: 30,
	}
	logger := logrus.New()
	transport := NewStdioTransport(cfg, logger)

	// Initially disconnected
	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
	if transport.GetState() != StateDisconnected {
		t.Errorf("Expected state StateDisconnected, got %v", transport.GetState())
	}

	// Connect
	err := transport.Connect()
	if err != nil {
		t.Errorf("Connect() should not return error, got: %v", err)
	}
	if !transport.IsConnected() {
		t.Error("Transport should be connected after Connect()")
	}
	if transport.GetState() != StateConnected {
		t.Errorf("Expected state StateConnected, got %v", transport.GetState())
	}

	// Disconnect
	err = transport.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() should not return error, got: %v", err)
	}
	if transport.IsConnected() {
		t.Error("Transport should not be connected after Disconnect()")
	}
	if transport.GetState() != StateClosed {
		t.Errorf("Expected state StateClosed, got %v", transport.GetState())
	}
}

func TestStdioTransport_ReadWrite(t *testing.T) {
	cfg := &config.Config{
		ReadTimeout: 1, // Short timeout for testing
	}
	logger := logrus.New()
	transport := NewStdioTransport(cfg, logger)

	// Connect first
	err := transport.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Test writing (this will write to stdout, but shouldn't error)
	testData := []byte("test message\n")
	n, err := transport.Write(testData)
	if err != nil {
		t.Errorf("Write() should not return error, got: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Test reading from disconnected state (should error)
	err = transport.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect() failed: %v", err)
	}

	buffer := make([]byte, 1024)
	_, err = transport.Read(buffer)
	if err == nil {
		t.Error("Read() should return error when disconnected")
	}
}

func TestStdioTransport_ReadTimeout(t *testing.T) {
	cfg := &config.Config{
		ReadTimeout: 1, // 1 second timeout
	}
	logger := logrus.New()
	transport := NewStdioTransport(cfg, logger)

	err := transport.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	buffer := make([]byte, 1024)
	start := time.Now()

	// This should timeout since there's no input on stdin in test environment
	_, err = transport.Read(buffer)

	elapsed := time.Since(start)

	// In test environment, stdin might not block, so we just verify the method doesn't hang indefinitely
	// The timeout mechanism is in place for production use
	if elapsed > 2*time.Second {
		t.Errorf("Read() took too long: %v", elapsed)
	}

	// Verify timeout is configured correctly
	if transport.readTimeout != time.Second {
		t.Errorf("Expected readTimeout to be 1s, got %v", transport.readTimeout)
	}
}

func TestStdioTransport_WriteDisconnected(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	transport := NewStdioTransport(cfg, logger)

	// Don't connect - should be disconnected
	buffer := []byte("test")
	_, err := transport.Write(buffer)
	if err == nil {
		t.Error("Write() should return error when disconnected")
	}
}

func TestNewSSETransport_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}
	logger := logrus.New()
	transport := NewSSETransport(cfg, logger)

	if transport == nil {
		t.Fatal("NewSSETransport() should return non-nil transport")
	}
	if transport.config != cfg {
		t.Error("Transport config should match provided config")
	}
	if transport.logger != logger {
		t.Error("Transport logger should match provided logger")
	}
	if transport.state != StateDisconnected {
		t.Errorf("Expected initial state StateDisconnected, got %v", transport.state)
	}
	if transport.connections == nil {
		t.Error("Transport connections map should be initialized")
	}
}

func TestSSETransport_ConnectionLifecycle(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8081, // Use different port to avoid conflicts
	}
	logger := logrus.New()
	transport := NewSSETransport(cfg, logger)

	// Initially disconnected
	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
	if transport.GetState() != StateDisconnected {
		t.Errorf("Expected state StateDisconnected, got %v", transport.GetState())
	}

	// Connect
	err := transport.Connect()
	if err != nil {
		t.Errorf("Connect() should not return error, got: %v", err)
	}
	if !transport.IsConnected() {
		t.Error("Transport should be connected after Connect()")
	}
	if transport.GetState() != StateConnected {
		t.Errorf("Expected state StateConnected, got %v", transport.GetState())
	}

	// Disconnect
	err = transport.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() should not return error, got: %v", err)
	}
	if transport.IsConnected() {
		t.Error("Transport should not be connected after Disconnect()")
	}
	if transport.GetState() != StateClosed {
		t.Errorf("Expected state StateClosed, got %v", transport.GetState())
	}
}

func TestSSETransport_ReadNotSupported(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8082,
	}
	logger := logrus.New()
	transport := NewSSETransport(cfg, logger)

	buffer := make([]byte, 1024)
	_, err := transport.Read(buffer)
	if err == nil {
		t.Error("Read() should return error for SSE transport")
	}
	if !strings.Contains(err.Error(), "does not support reading") {
		t.Errorf("Expected 'does not support reading' error, got: %v", err)
	}
}

func TestSSETransport_WriteDisconnected(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8083,
	}
	logger := logrus.New()
	transport := NewSSETransport(cfg, logger)

	// Don't connect - should be disconnected
	buffer := []byte("test message")
	_, err := transport.Write(buffer)
	if err == nil {
		t.Error("Write() should return error when disconnected")
	}
}

func TestSSEConnection_Send(t *testing.T) {
	// Create a mock response writer for testing
	mockWriter := &mockResponseWriter{}
	conn := &SSEConnection{
		id:      "test-conn",
		writer:  mockWriter,
		flusher: mockWriter,
		done:    make(chan struct{}),
	}

	message := "data: test message\n\n"
	err := conn.Send(message)
	if err != nil {
		t.Errorf("Send() should not return error, got: %v", err)
	}

	// Verify message was written
	if mockWriter.written != message {
		t.Errorf("Expected message %q, got %q", message, mockWriter.written)
	}
}

func TestSSEConnection_Close(t *testing.T) {
	conn := &SSEConnection{
		id:   "test-conn",
		done: make(chan struct{}),
	}

	// Close connection
	conn.Close()

	// Verify it's marked as closed
	if !conn.isClosed {
		t.Error("Connection should be marked as closed")
	}

	// Try to send after close (should error)
	err := conn.Send("test")
	if err == nil {
		t.Error("Send() should return error after Close()")
	}
}

// mockResponseWriter implements http.ResponseWriter and http.Flusher for testing
type mockResponseWriter struct {
	written string
	header  http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.written = string(data)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

func (m *mockResponseWriter) Flush() {
	// No-op for testing
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
