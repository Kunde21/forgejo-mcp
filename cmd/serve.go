package cmd

import (
	"context"
	"fmt"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/Kunde21/forgejo-mcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// registerTools registers MCP tools with the server
func registerTools(mcpServer *mcp.Server, cfg *config.Config, logger *logrus.Logger) error {
	// Create Gitea SDK client
	giteaClient, err := cfg.CreateGiteaClient()
	if err != nil {
		logger.Errorf("Failed to create Gitea SDK client: %v", err)
		return fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Create handlers with SDK client
	prHandler := server.NewTeaPRListHandler(logger, giteaClient)
	issueHandler := server.NewTeaIssueListHandler(logger, giteaClient)

	// Register PR list tool
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_list",
		Description: "List pull requests from the Forgejo repository",
	}, prHandler.HandlePRListRequest)

	// Register issue list tool
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "issue_list",
		Description: "List issues from the Forgejo repository",
	}, issueHandler.HandleIssueListRequest)

	logger.Info("Registered MCP tools with tea CLI integration")
	return nil
}

// Note: MCP SDK uses stdio transport, so host/port configuration is not needed

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the Model Context Protocol server to handle requests from AI agents.

The server uses stdio transport for MCP communication and requires proper
Forgejo configuration to be set up before starting.

Examples:
  # Start server with default settings
  forgejo-mcp serve`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Set up logging
		logger := logrus.New()
		level, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			return fmt.Errorf("invalid log level '%s': %w", cfg.LogLevel, err)
		}
		logger.SetLevel(level)
		logger.SetFormatter(&logrus.JSONFormatter{})

		// Create MCP server using SDK
		impl := &mcp.Implementation{
			Name:    "forgejo-mcp",
			Version: "1.0.0",
		}
		mcpServer := mcp.NewServer(impl, nil)

		// Register tools
		if err := registerTools(mcpServer, cfg, logger); err != nil {
			return fmt.Errorf("failed to register tools: %w", err)
		}

		// Start the server
		logger.Info("Starting MCP server with stdio transport")

		// Use stdio transport (MCP SDK default)
		transport := mcp.NewStdioTransport()
		if err := mcpServer.Run(context.Background(), transport); err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Add aliases
	serveCmd.Aliases = []string{"server", "start"}
}
