package cmd

import (
	"slices"
	"testing"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

func TestRegisterTools(t *testing.T) {
	// Create MCP server
	impl := &mcp.Implementation{
		Name:    "forgejo-mcp-test",
		Version: "1.1.0",
	}
	mcpServer := mcp.NewServer(impl, nil)

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test that MCP server was created successfully
	if mcpServer == nil {
		t.Fatal("MCP server should not be nil")
	}

	// Test that we can add tools to the server (basic functionality test)
	tool := &mcp.Tool{
		Name:        "test_tool",
		Description: "Test tool for registration",
	}

	// This should not panic or error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Adding tool to MCP server panicked: %v", r)
		}
	}()

	// Note: We can't actually test tool registration without a real client
	// since the handlers require a Gitea client interface
	// This test just verifies the MCP server setup works
	_ = tool // Avoid unused variable warning
}

func TestServeCmd_PreRunE(t *testing.T) {
	// Since we removed PreRunE validation for host/port,
	// this test verifies that PreRunE is nil or doesn't exist
	if serveCmd.PreRunE != nil {
		err := serveCmd.PreRunE(serveCmd, []string{})
		if err != nil {
			t.Errorf("PreRunE failed: %v", err)
		}
	}
}

func TestServeCmd_Flags(t *testing.T) {
	// Test that no host/port flags exist (since we use stdio transport)
	hostFlag := serveCmd.Flags().Lookup("host")
	if hostFlag != nil {
		t.Error("host flag should not exist for stdio transport")
	}

	portFlag := serveCmd.Flags().Lookup("port")
	if portFlag != nil {
		t.Error("port flag should not exist for stdio transport")
	}
}

func TestServeCmd_Aliases(t *testing.T) {
	// Test that aliases are properly set
	expectedAliases := []string{"server", "start"}
	if len(serveCmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(serveCmd.Aliases))
	}
	for i, alias := range expectedAliases {
		if i >= len(serveCmd.Aliases) || serveCmd.Aliases[i] != alias {
			t.Errorf("Expected alias %d to be %s", i, alias)
		}
	}
}

func TestServeCmd_Usage(t *testing.T) {
	// Test command usage information
	if serveCmd.Use != "serve" {
		t.Errorf("Expected command use 'serve', got %s", serveCmd.Use)
	}
	if serveCmd.Short == "" {
		t.Error("Command short description should not be empty")
	}
}

// Test that the command integrates with root command
func TestServeCmd_Integration(t *testing.T) {
	// Verify serve command is added to root
	found := slices.Contains(rootCmd.Commands(), serveCmd)
	if !found {
		t.Error("serve command should be added to root command")
	}
}

func TestServeCmd_ServerInitialization(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		ForgejoURL: "https://test.forgejo.com",
		AuthToken:  "test-required-auth-token",
		Host:       "localhost", // Required for config validation
		Port:       3000,        // Required for config validation
		LogLevel:   "info",
	}

	// Validate config
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Set up logging
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create MCP server using SDK
	impl := &mcp.Implementation{
		Name:    "forgejo-mcp-test",
		Version: "1.1.0",
	}
	mcpServer := mcp.NewServer(impl, nil)
	if mcpServer == nil {
		t.Fatal("NewServer returned nil")
	}

	// Test that MCP server initialization works
	// Note: We can't test registerTools without mocking the Gitea client
	// since it tries to connect to a real server
}
