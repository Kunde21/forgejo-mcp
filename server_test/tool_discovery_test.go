package servertest

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestToolDiscovery validates the entire tool discovery response payload
// This test merges and consolidates all individual ToolDiscovery tests to ensure
// comprehensive validation of the MCP server's tool registration and metadata
func TestToolDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// List available tools
	tools, err := client.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	// Validate total tool count
	expectedToolCount := 13
	if len(tools.Tools) != expectedToolCount {
		t.Fatalf("Expected %d tools, got %d", expectedToolCount, len(tools.Tools))
	}

	// Define expected tools with their descriptions
	expectedTools := map[string]string{
		"hello":                "Returns a hello world message",
		"issue_list":           "List issues from a Gitea/Forgejo repository",
		"issue_create":         "Create a new issue on a Forgejo/Gitea repository",
		"issue_comment_create": "Create a comment on a Forgejo/Gitea repository issue",
		"issue_comment_list":   "List comments from a Forgejo/Gitea repository issue with pagination support",
		"issue_comment_edit":   "Edit an existing comment on a Forgejo/Gitea repository issue",
		"issue_edit":           "Edit an existing issue in a Forgejo/Gitea repository",
		"pr_list":              "List pull requests from a Forgejo/Gitea repository with pagination and state filtering",
		"pr_comment_list":      "List comments from a Forgejo/Gitea repository pull request with pagination support",
		"pr_comment_create":    "Create a comment on a Forgejo/Gitea repository pull request",
		"pr_comment_edit":      "Edit an existing comment on a Forgejo/Gitea repository pull request",
		"pr_edit":              "Edit an existing pull request in a Forgejo/Gitea repository",
		"pr_create":            "Create a new pull request in a Forgejo/Gitea repository",
	}

	// Track found tools for validation
	foundTools := make(map[string]*mcp.Tool)

	// Validate each tool in the response
	for _, tool := range tools.Tools {
		expectedDesc, exists := expectedTools[tool.Name]
		if !exists {
			t.Errorf("Unexpected tool found: %s", tool.Name)
			continue
		}

		// Validate description
		if tool.Description != expectedDesc {
			t.Errorf("Tool '%s' description mismatch. Expected: '%s', Got: '%s'",
				tool.Name, expectedDesc, tool.Description)
		}

		// Validate input schema exists
		if tool.InputSchema == nil {
			t.Errorf("Tool '%s' should have input schema", tool.Name)
		}

		// Track found tool
		foundTools[tool.Name] = tool
	}

	// Validate all expected tools were found
	for expectedTool := range expectedTools {
		if _, found := foundTools[expectedTool]; !found {
			t.Errorf("Expected tool '%s' not found in discovery response", expectedTool)
		}
	}

	// Additional validation: Ensure no duplicate tools
	toolNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		if toolNames[tool.Name] {
			t.Errorf("Duplicate tool found: %s", tool.Name)
		}
		toolNames[tool.Name] = true
	}

	// Validate specific tool properties for critical tools
	criticalTools := []string{"hello", "issue_list", "pr_list"}
	for _, toolName := range criticalTools {
		tool, exists := foundTools[toolName]
		if !exists {
			continue // Already reported above
		}

		// Validate tool has required properties
		if tool.Name == "" {
			t.Errorf("Tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("Tool '%s' has empty description", toolName)
		}
	}

	// Validate response structure
	if tools.Tools == nil {
		t.Error("Tools slice should not be nil")
	}

	// Log success for debugging
	t.Logf("Successfully validated %d tools in discovery response", len(tools.Tools))
	for _, tool := range tools.Tools {
		t.Logf("  - %s: %s", tool.Name, tool.Description)
	}
}
