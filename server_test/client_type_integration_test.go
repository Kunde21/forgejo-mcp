package servertest

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestClientTypeGiteaIntegration tests that Gitea client type works correctly
func TestClientTypeGiteaIntegration(t *testing.T) {
	// Create a mock server that responds as Gitea
	mock := NewMockGiteaServer(t)
	mock.server.Config.Handler = mock.createGiteaHandler()

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL":  mock.URL(),
		"FORGEJO_AUTH_TOKEN":  "mock-token",
		"FORGEJO_CLIENT_TYPE": "gitea",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that we can list issues with Gitea client
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
	}
}

// TestClientTypeForgejoIntegration tests that Forgejo client type works correctly
func TestClientTypeForgejoIntegration(t *testing.T) {
	// Create a mock server that responds as Forgejo
	mock := NewMockGiteaServer(t)
	mock.server.Config.Handler = mock.createForgejoHandler()

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL":  mock.URL(),
		"FORGEJO_AUTH_TOKEN":  "mock-token",
		"FORGEJO_CLIENT_TYPE": "forgejo",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that we can list issues with Forgejo client
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
	}
}

// TestClientTypeAutoDetectionGitea tests auto-detection with Gitea version
func TestClientTypeAutoDetectionGitea(t *testing.T) {
	// Create a mock server that responds as Gitea
	mock := NewMockGiteaServer(t)
	mock.server.Config.Handler = mock.createGiteaHandler()

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL":  mock.URL(),
		"FORGEJO_AUTH_TOKEN":  "mock-token",
		"FORGEJO_CLIENT_TYPE": "auto",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that we can list issues with auto-detected Gitea client
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
	}
}

// TestClientTypeAutoDetectionForgejo tests auto-detection with Forgejo version
func TestClientTypeAutoDetectionForgejo(t *testing.T) {
	// Create a mock server that responds as Forgejo
	mock := NewMockGiteaServer(t)
	mock.server.Config.Handler = mock.createForgejoHandler()

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL":  mock.URL(),
		"FORGEJO_AUTH_TOKEN":  "mock-token",
		"FORGEJO_CLIENT_TYPE": "auto",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that we can list issues with auto-detected Forgejo client
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
	}
}

// TestClientTypeInvalid tests error handling for invalid client types
func TestClientTypeInvalid(t *testing.T) {
	// Set all required environment variables
	t.Setenv("FORGEJO_REMOTE_URL", "http://example.com")
	t.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
	t.Setenv("FORGEJO_CLIENT_TYPE", "invalid")

	// Load config to see what happens
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Config load failed: %v", err)
	}

	// Validate the config
	err = cfg.Validate()
	if err == nil {
		t.Fatal("Expected validation error for invalid client type")
	}

	if !contains(err.Error(), "ClientType must be one of") {
		t.Errorf("Expected validation error, got: %v", err)
	}

	t.Logf("Config ClientType: '%s'", cfg.ClientType)
}

// TestClientTypeEmptyDefaultsToAuto tests that empty client type defaults to auto
func TestClientTypeEmptyDefaultsToAuto(t *testing.T) {
	// Create a mock server that responds as Gitea
	mock := NewMockGiteaServer(t)
	mock.server.Config.Handler = mock.createGiteaHandler()

	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
		// No FORGEJO_CLIENT_TYPE set - should default to auto
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	// Test that we can list issues with auto-detected client
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
			"limit":      10,
			"offset":     0,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list tool: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", GetTextContent(result.Content))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
