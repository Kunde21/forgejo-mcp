// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"testing"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/sirupsen/logrus"
)

func TestToolRegistration_BasicRegistration(t *testing.T) {
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

	// Test that dispatcher is initialized
	if server.dispatcher == nil {
		t.Error("server.dispatcher should be initialized")
	}

	// Test registering a handler
	testHandler := &TestHandler{logger: server.logger}
	server.dispatcher.RegisterHandler("test/method", testHandler)

	// Test dispatching to the handler
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "test/method",
		Params:  map[string]interface{}{"test": "data"},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("Dispatch should return a response")
	}
	if resp.Error != nil {
		t.Errorf("Dispatch should not return error for registered handler, got: %v", resp.Error)
	}
}

func TestToolRegistration_UnregisteredMethod(t *testing.T) {
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

	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "nonexistent/method",
		Params:  map[string]interface{}{},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("Dispatch should return a response even for unregistered methods")
	}
	if resp.Error == nil {
		t.Error("Dispatch should return error for unregistered method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601 (Method not found), got %d", resp.Error.Code)
	}
}

func TestToolManifest_Generation(t *testing.T) {
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test tools/list method
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/list should return a response")
	}
	if resp.Error != nil {
		t.Errorf("tools/list should not return error, got: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("tools/list should return result")
	}

	// Verify result structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Error("tools/list result should be a map")
	}

	tools, exists := result["tools"]
	if !exists {
		t.Error("tools/list result should contain 'tools' field")
	}

	toolsArray, ok := tools.([]map[string]interface{})
	if !ok {
		t.Error("tools should be an array of tool definitions")
	}

	if len(toolsArray) == 0 {
		t.Error("tools array should not be empty")
	}

	// Verify each tool has required fields
	for _, tool := range toolsArray {
		if _, exists := tool["name"]; !exists {
			t.Error("tool should have 'name' field")
		}
		if _, exists := tool["description"]; !exists {
			t.Error("tool should have 'description' field")
		}
		if _, exists := tool["inputSchema"]; !exists {
			t.Error("tool should have 'inputSchema' field")
		}
	}
}

func TestToolCall_PRList(t *testing.T) {
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test pr_list tool call
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "pr_list",
			"arguments": map[string]interface{}{
				"state": "open",
			},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response")
	}
	if resp.Error != nil {
		t.Errorf("tools/call should not return error, got: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("tools/call should return result")
	}

	// Verify result structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Error("tools/call result should be a map")
	}

	if _, exists := result["pullRequests"]; !exists {
		t.Error("pr_list result should contain 'pullRequests' field")
	}
	if _, exists := result["total"]; !exists {
		t.Error("pr_list result should contain 'total' field")
	}
}

func TestToolCall_IssueList(t *testing.T) {
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test issue_list tool call
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "issue_list",
			"arguments": map[string]interface{}{
				"state": "open",
			},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response")
	}
	if resp.Error != nil {
		t.Errorf("tools/call should not return error, got: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("tools/call should return result")
	}

	// Verify result structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Error("tools/call result should be a map")
	}

	if _, exists := result["issues"]; !exists {
		t.Error("issue_list result should contain 'issues' field")
	}
	if _, exists := result["total"]; !exists {
		t.Error("issue_list result should contain 'total' field")
	}
}

func TestToolCall_UnknownTool(t *testing.T) {
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test unknown tool call
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "unknown_tool",
			"arguments": map[string]interface{}{},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response even for unknown tools")
	}
	if resp.Error == nil {
		t.Error("tools/call should return error for unknown tool")
	}
	if resp.Error.Code != -32603 {
		t.Errorf("Expected error code -32603 (Internal error), got %d", resp.Error.Code)
	}
}

func TestToolCall_MissingToolName(t *testing.T) {
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test missing tool name
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"arguments": map[string]interface{}{},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response")
	}
	if resp.Error == nil {
		t.Error("tools/call should return error for missing tool name")
	}
}

// TestHandler is a simple test handler for testing registration
type TestHandler struct {
	logger *logrus.Logger
}

func (h *TestHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Test handler called with method: %s, params: %v", method, params)
	return map[string]interface{}{"test": "response"}, nil
}
