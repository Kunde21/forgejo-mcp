package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestServerInitialization(t *testing.T) {
	// Test that server can be initialized
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	if server == nil {
		t.Fatal("Server is nil")
	}
}

func TestServerStart(t *testing.T) {
	// Test that server can start without error
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
}

func TestServerStop(t *testing.T) {
	// Test that server can stop gracefully
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

func TestHelloTool(t *testing.T) {
	// Test that the hello tool returns correct response
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test the tool handler directly with valid context
	result, err := server.handleHello(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("Failed to call hello tool: %v", err)
	}

	if result == nil {
		t.Fatal("Tool result is nil")
	}

	// Check that we get the expected text content
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent in result")
	}

	expected := "Hello, World!"
	if textContent.Text != expected {
		t.Errorf("Expected %q, got %q", expected, textContent.Text)
	}
}

func TestHelloToolWithNilContext(t *testing.T) {
	// Test error handling with nil context
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test with nil context should return error
	result, err := server.handleHello(nil, mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Tool result is nil")
	}

	// Should return an error result
	if !result.IsError {
		t.Error("Expected error result for nil context")
	}
}
