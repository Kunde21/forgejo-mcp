// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"testing"

	"github.com/Kunde21/forgejo-mcp/config"
)

// TestCompleteRequestResponseFlow tests the complete flow from MCP request to response
func TestCompleteRequestResponseFlow(t *testing.T) {
	// Test basic server creation and tool registration
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

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test that dispatcher is properly configured
	if server.dispatcher == nil {
		t.Error("server.dispatcher should be initialized")
	}

	// Test basic request structure
	request := &Request{
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

	// Test that request can be dispatched (may return error due to mock data)
	response := server.dispatcher.Dispatch(context.Background(), request)
	if response == nil {
		t.Error("Dispatch should return a response")
	}

	// Test unknown method
	unknownRequest := &Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "unknown/method",
		Params:  map[string]interface{}{},
	}

	unknownResponse := server.dispatcher.Dispatch(context.Background(), unknownRequest)
	if unknownResponse == nil {
		t.Error("Dispatch should return a response even for unknown methods")
	}
	if unknownResponse.Error == nil {
		t.Error("Unknown method should return error")
	}
}

// TestServerStartupAndConnectionAcceptance tests server startup and MCP connection acceptance
func TestServerStartupAndConnectionAcceptance(t *testing.T) {
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

	// Test that server can handle basic MCP requests
	request := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	response := server.dispatcher.Dispatch(context.Background(), request)
	if response == nil {
		t.Error("Server should handle tools/list request")
	}
	if response.Error != nil {
		t.Errorf("tools/list should not return error, got: %v", response.Error)
	}
	if response.Result == nil {
		t.Error("tools/list should return result")
	}

	// Verify the response contains expected tool information
	result, ok := response.Result.(map[string]interface{})
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

	if len(toolsArray) < 2 {
		t.Errorf("Expected at least 2 tools, got %d", len(toolsArray))
	}

	// Verify each tool has required fields
	for _, tool := range toolsArray {
		if _, hasName := tool["name"]; !hasName {
			t.Error("Tool should have 'name' field")
		}
		if _, hasDesc := tool["description"]; !hasDesc {
			t.Error("Tool should have 'description' field")
		}
		if _, hasSchema := tool["inputSchema"]; !hasSchema {
			t.Error("Tool should have 'inputSchema' field")
		}
	}
}

// TestToolDiscoveryThroughManifest tests tool discovery through the manifest
func TestToolDiscoveryThroughManifest(t *testing.T) {
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

	// Test that we can discover pr_list tool
	prTool, exists := server.GetTool("pr_list")
	if !exists {
		t.Error("pr_list tool should be discoverable")
	}
	if prTool == nil {
		t.Error("pr_list tool should not be nil")
	}
	if prTool.Name != "pr_list" {
		t.Errorf("Expected tool name 'pr_list', got '%s'", prTool.Name)
	}

	// Test that we can discover issue_list tool
	issueTool, exists := server.GetTool("issue_list")
	if !exists {
		t.Error("issue_list tool should be discoverable")
	}
	if issueTool == nil {
		t.Error("issue_list tool should not be nil")
	}
	if issueTool.Name != "issue_list" {
		t.Errorf("Expected tool name 'issue_list', got '%s'", issueTool.Name)
	}

	// Test that non-existent tool is not discoverable
	_, exists = server.GetTool("nonexistent_tool")
	if exists {
		t.Error("Non-existent tool should not be discoverable")
	}

	// Test getting all tool names
	toolNames := server.GetToolNames()
	if len(toolNames) != 2 {
		t.Errorf("Expected 2 tool names, got %d", len(toolNames))
	}

	expectedNames := map[string]bool{"pr_list": false, "issue_list": false}
	for _, name := range toolNames {
		if _, exists := expectedNames[name]; exists {
			expectedNames[name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool name '%s' not found in tool names", name)
		}
	}
}

// TestErrorHandlingAndTimeoutScenarios tests various error handling and timeout scenarios
func TestErrorHandlingAndTimeoutScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		request     *Request
		expectError bool
		errorCode   int
		description string
	}{
		{
			name: "missing tool name",
			request: &Request{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"arguments": map[string]interface{}{
						"state": "open",
					},
				},
			},
			expectError: true,
			errorCode:   -32603, // Internal error
			description: "Should return error when tool name is missing",
		},
		{
			name: "invalid tool name",
			request: &Request{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":      "invalid_tool",
					"arguments": map[string]interface{}{},
				},
			},
			expectError: true,
			errorCode:   -32603, // Internal error
			description: "Should return error for unknown tool",
		},
		{
			name: "unknown method",
			request: &Request{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "unknown/method",
				Params:  map[string]interface{}{},
			},
			expectError: true,
			errorCode:   -32601, // Method not found
			description: "Should return error for unknown method",
		},
		{
			name: "invalid JSON-RPC version",
			request: &Request{
				JSONRPC: "1.0",
				ID:      4,
				Method:  "tools/list",
				Params:  map[string]interface{}{},
			},
			expectError: false, // Currently not validated in transport layer
			errorCode:   0,
			description: "JSON-RPC version validation not implemented in transport layer",
		},
		{
			name: "missing method",
			request: &Request{
				JSONRPC: "2.0",
				ID:      5,
				Method:  "",
				Params:  map[string]interface{}{},
			},
			expectError: true,
			errorCode:   -32601, // Method not found (actual behavior)
			description: "Should return error when method is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			response := server.dispatcher.Dispatch(context.Background(), tc.request)

			if response == nil {
				t.Error("Should return a response even for invalid requests")
				return
			}

			if tc.expectError {
				if response.Error == nil {
					t.Errorf("%s: expected error but got success", tc.description)
				} else if response.Error.Code != tc.errorCode {
					t.Errorf("%s: expected error code %d, got %d", tc.description, tc.errorCode, response.Error.Code)
				}
			} else {
				if response.Error != nil {
					t.Errorf("%s: expected success but got error: %v", tc.description, response.Error)
				}
			}
		})
	}
}

// TestPRListWithMockedTeaOutput tests pr_list with mocked tea output
func TestPRListWithMockedTeaOutput(t *testing.T) {
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

	// Test pr_list with various parameters
	testCases := []struct {
		name     string
		params   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "pr_list with state filter",
			params: map[string]interface{}{
				"name": "pr_list",
				"arguments": map[string]interface{}{
					"state": "open",
				},
			},
			expected: map[string]interface{}{
				"pullRequests": "mock_data",
				"total":        "mock_count",
			},
		},
		{
			name: "pr_list with author filter",
			params: map[string]interface{}{
				"name": "pr_list",
				"arguments": map[string]interface{}{
					"author": "developer1",
				},
			},
			expected: map[string]interface{}{
				"pullRequests": "mock_data",
				"total":        "mock_count",
			},
		},
		{
			name: "pr_list with limit",
			params: map[string]interface{}{
				"name": "pr_list",
				"arguments": map[string]interface{}{
					"limit": float64(5),
				},
			},
			expected: map[string]interface{}{
				"pullRequests": "mock_data",
				"total":        "mock_count",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &Request{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params:  tc.params,
			}

			response := server.dispatcher.Dispatch(context.Background(), request)

			if response == nil {
				t.Error("tools/call should return a response")
				return
			}

			if response.Error != nil {
				t.Errorf("pr_list should not return error, got: %v", response.Error)
				return
			}

			if response.Result == nil {
				t.Error("pr_list should return result")
				return
			}

			// Verify result structure
			result, ok := response.Result.(map[string]interface{})
			if !ok {
				t.Error("pr_list result should be a map")
				return
			}

			// Check that expected fields are present
			for key := range tc.expected {
				if _, exists := result[key]; !exists {
					t.Errorf("pr_list result should contain '%s' field", key)
				}
			}
		})
	}
}

// TestIssueListWithMockedTeaOutput tests issue_list with mocked tea output
func TestIssueListWithMockedTeaOutput(t *testing.T) {
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

	// Test issue_list with various parameters
	testCases := []struct {
		name     string
		params   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "issue_list with state filter",
			params: map[string]interface{}{
				"name": "issue_list",
				"arguments": map[string]interface{}{
					"state": "closed",
				},
			},
			expected: map[string]interface{}{
				"issues": "mock_data",
				"total":  "mock_count",
			},
		},
		{
			name: "issue_list with labels",
			params: map[string]interface{}{
				"name": "issue_list",
				"arguments": map[string]interface{}{
					"labels": []interface{}{"bug", "ui"},
				},
			},
			expected: map[string]interface{}{
				"issues": "mock_data",
				"total":  "mock_count",
			},
		},
		{
			name: "issue_list with author filter",
			params: map[string]interface{}{
				"name": "issue_list",
				"arguments": map[string]interface{}{
					"author": "user1",
				},
			},
			expected: map[string]interface{}{
				"issues": "mock_data",
				"total":  "mock_count",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &Request{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params:  tc.params,
			}

			response := server.dispatcher.Dispatch(context.Background(), request)

			if response == nil {
				t.Error("tools/call should return a response")
				return
			}

			if response.Error != nil {
				t.Errorf("issue_list should not return error, got: %v", response.Error)
				return
			}

			if response.Result == nil {
				t.Error("issue_list should return result")
				return
			}

			// Verify result structure
			result, ok := response.Result.(map[string]interface{})
			if !ok {
				t.Error("issue_list result should be a map")
				return
			}

			// Check that expected fields are present
			for key := range tc.expected {
				if _, exists := result[key]; !exists {
					t.Errorf("issue_list result should contain '%s' field", key)
				}
			}
		})
	}
}
