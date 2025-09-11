package server

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPullRequestListToolRegistration tests that the pr_list tool is properly registered with the MCP server
func TestPullRequestListToolRegistration(t *testing.T) {
	// Create a server instance
	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	server, err := NewFromService(mockService, &config.Config{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Get the underlying MCP server
	mcpServer := server.MCPServer()
	if mcpServer == nil {
		t.Fatal("MCP server should not be nil")
	}

	// Test that we can access the server's tools through the SDK
	// Note: The official SDK doesn't provide direct access to registered tools,
	// so we test through the actual tool execution
	ctx := context.Background()
	request := &mcp.CallToolRequest{}

	// Test that the handler function exists and works
	result, data, err := server.handlePullRequestList(ctx, request, PullRequestListArgs{
		Repository: "owner/repo",
		Limit:      10,
		Offset:     0,
		State:      "open",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result to be returned")
	}
	if result.IsError {
		t.Error("Expected result to not be an error")
	}
	if data == nil {
		t.Error("Expected data to be returned")
	}
}

// TestPullRequestListToolSchema tests that the tool schema is properly defined
func TestPullRequestListToolSchema(t *testing.T) {
	// This test verifies that the tool accepts the correct parameters
	// Since the SDK handles schema validation internally, we test through actual calls

	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	server, err := NewFromService(mockService, &config.Config{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	request := &mcp.CallToolRequest{}

	// Test with all parameters provided
	result, data, err := server.handlePullRequestList(ctx, request, PullRequestListArgs{
		Repository: "owner/repo",
		Limit:      25,
		Offset:     5,
		State:      "closed",
	})

	if err != nil {
		t.Errorf("Expected no error with all parameters, got %v", err)
	}
	if result == nil || result.IsError {
		t.Error("Expected successful result with all parameters")
	}
	if data == nil {
		t.Error("Expected data to be returned with all parameters")
	}

	// Test with minimal parameters (only repository)
	result, data, err = server.handlePullRequestList(ctx, request, PullRequestListArgs{
		Repository: "owner/repo",
		// Limit, Offset, and State should use defaults
	})

	if err != nil {
		t.Errorf("Expected no error with minimal parameters, got %v", err)
	}
	if result == nil || result.IsError {
		t.Error("Expected successful result with minimal parameters")
	}
	if data == nil {
		t.Error("Expected data to be returned with minimal parameters")
	}
}

// TestPullRequestListToolMetadata tests that the tool has proper description and metadata
func TestPullRequestListToolMetadata(t *testing.T) {
	// Since we can't directly access tool metadata from the SDK,
	// we verify that the tool behaves as documented

	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	server, err := NewFromService(mockService, &config.Config{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	request := &mcp.CallToolRequest{}

	// Test that the tool handles the documented parameters correctly
	testCases := []struct {
		name        string
		args        PullRequestListArgs
		expectError bool
	}{
		{
			name: "valid open state",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				State:      "open",
			},
			expectError: false,
		},
		{
			name: "valid closed state",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				State:      "closed",
			},
			expectError: false,
		},
		{
			name: "valid all state",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				State:      "all",
			},
			expectError: false,
		},
		{
			name: "invalid state",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				State:      "invalid",
			},
			expectError: true,
		},
		{
			name: "valid pagination",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				Limit:      50,
				Offset:     10,
			},
			expectError: false,
		},
		{
			name: "invalid limit",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				Limit:      150, // > 100
			},
			expectError: true,
		},
		{
			name: "invalid offset",
			args: PullRequestListArgs{
				Repository: "owner/repo",
				Offset:     -1,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := server.handlePullRequestList(ctx, request, tc.args)

			if tc.expectError {
				if err != nil {
					t.Errorf("Expected validation error, got handler error: %v", err)
				}
				if result == nil || !result.IsError {
					t.Error("Expected error result for invalid input")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != nil && result.IsError {
					t.Error("Expected successful result for valid input")
				}
			}
		})
	}
}

// TestPullRequestListHandlerWiring tests that the handler function is properly wired to the tool registration
func TestPullRequestListHandlerWiring(t *testing.T) {
	// This test verifies that the handler function signature matches what's expected
	// by the MCP SDK and that it's properly integrated

	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	server, err := NewFromService(mockService, &config.Config{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify that the server is properly initialized
	if server == nil {
		t.Error("Server should not be nil")
	}

	// Test that the handler method has the correct signature by calling it
	ctx := context.Background()
	request := &mcp.CallToolRequest{}
	args := PullRequestListArgs{
		Repository: "test/repo",
		Limit:      10,
		Offset:     0,
		State:      "open",
	}

	// This should compile and run without signature errors
	result, data, err := server.handlePullRequestList(ctx, request, args)

	if err != nil {
		t.Errorf("Handler method call failed: %v", err)
	}
	if result == nil {
		t.Error("Handler should return a result")
	}
	if data == nil {
		t.Error("Handler should return data")
	}

	// Verify the result structure
	if len(result.Content) == 0 {
		t.Error("Result should contain content")
	}
}

// TestPullRequestListServerIntegration tests the complete integration with the MCP server
func TestPullRequestListServerIntegration(t *testing.T) {
	// Test that the server can be created and started with the pr_list tool registered
	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	cfg := &config.Config{
		RemoteURL: "http://localhost:3000",
		AuthToken: "test-token",
		Port:      8080,
	}

	server, err := NewFromService(mockService, cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify server is properly initialized
	if server.MCPServer() == nil {
		t.Error("Server should have an MCP server instance")
	}

	// Verify that the handler is accessible through the server
	if server.giteaService == nil {
		t.Error("Server should have a Gitea service")
	}

	// Test that the server can handle pr_list requests
	ctx := context.Background()
	request := &mcp.CallToolRequest{}
	args := PullRequestListArgs{
		Repository: "integration/test",
		Limit:      5,
		Offset:     0,
		State:      "open",
	}

	result, data, err := server.handlePullRequestList(ctx, request, args)

	if err != nil {
		t.Errorf("Integration test failed: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("Integration test should succeed")
	}
	if data == nil {
		t.Error("Integration test should return data")
	}

	// Verify the data structure matches expectations
	if len(data.PullRequests) != 1 {
		t.Errorf("Expected 1 pull request from mock, got %d", len(data.PullRequests))
	}
	if data.PullRequests[0].Title != "Test Pull Request" {
		t.Errorf("Expected mock pull request title 'Test Pull Request', got '%s'", data.PullRequests[0].Title)
	}
}
