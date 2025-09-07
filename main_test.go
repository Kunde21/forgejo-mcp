package main

import (
	"context"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleListIssues(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("FORGEJO_REMOTE_URL", "https://nonexistent-domain-12345.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test with valid request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
			},
		},
	}

	result, err := server.handleListIssues(context.Background(), request)
	if err != nil {
		t.Errorf("handleListIssues failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestHandleListIssues_InvalidRepository(t *testing.T) {
	os.Setenv("FORGEJO_REMOTE_URL", "https://nonexistent-domain-12345.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "invalid-repo",
				"limit":      10,
				"offset":     0,
			},
		},
	}

	_, err = server.handleListIssues(context.Background(), request)
	if err == nil {
		t.Error("Expected error for invalid repository format")
	}
}

func TestHandleListIssues_InvalidLimit(t *testing.T) {
	os.Setenv("FORGEJO_REMOTE_URL", "https://nonexistent-domain-12345.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("FORGEJO_REMOTE_URL")
		os.Unsetenv("FORGEJO_AUTH_TOKEN")
	}()

	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_issues",
			Arguments: map[string]interface{}{
				"repository": "owner/repo",
				"limit":      200, // Invalid: > 100
				"offset":     0,
			},
		},
	}

	_, err = server.handleListIssues(context.Background(), request)
	if err == nil {
		t.Error("Expected error for invalid limit")
	}
}
