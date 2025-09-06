package cmd

import (
	"fmt"
	"os"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/Kunde21/forgejo-mcp/server"
	"github.com/google/jsonschema-go/jsonschema"
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
	prHandler := server.NewSDKPRListHandler(logger, giteaClient)
	issueHandler := server.NewSDKIssueListHandler(logger, giteaClient)
	repoHandler := server.NewSDKRepositoryHandler(logger, giteaClient)

	// Register PR list tool
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_list",
		Description: "List pull requests from a specific Forgejo repository",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"repository": {
					Type:        "string",
					Description: "Repository identifier in 'owner/repo' format",
				},
				"cwd": {
					Type:        "string",
					Description: `Current working directory to resolve repository from. Required if "repository" is not provided`,
				},
				"state": {
					Type:        "string",
					Enum:        []any{"open", "closed", "all"},
					Description: "Filter by PR state",
				},
				"author": {
					Type:        "string",
					Description: "Filter by PR author username",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of PRs to return",
				},
			},
		},
	}, prHandler.HandlePRListRequest)

	// Register issue list tool
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "issue_list",
		Description: "List issues from a specific Forgejo repository",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"repository": {
					Type:        "string",
					Description: "Repository identifier in 'owner/repo' format",
				},
				"cwd": {
					Type:        "string",
					Description: `Current working directory to resolve repository from. Required if "repository" is not provided`,
				},
				"state": {
					Type:        "string",
					Enum:        []any{"open", "closed", "all"},
					Description: "Filter by issue state",
				},
				"author": {
					Type:        "string",
					Description: "Filter by issue author username",
				},
				"labels": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Filter by issue labels",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of issues to return",
				},
			},
		},
	}, issueHandler.HandleIssueListRequest)

	// Register repository list tool
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "repo_list",
		Description: "List repositories from the Forgejo instance",
	}, repoHandler.ListRepositories)

	logger.Info("Registered MCP tools with Gitea SDK integration")
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
		f, err := os.CreateTemp("/tmp/tmp.pK4iWXcwYw", "fj-mcp-*.logs")
		if err != nil {
			return fmt.Errorf("log file: %w", err)
		}
		defer f.Close()
		logger.SetOutput(f)

		// Create MCP server using SDK
		impl := &mcp.Implementation{
			Name:    "forgejo-mcp",
			Version: "1.1.0",
		}
		mcpServer := mcp.NewServer(impl, nil)

		// Register tools
		if err := registerTools(mcpServer, cfg, logger); err != nil {
			return fmt.Errorf("failed to register tools: %w", err)
		}

		// Start the server
		logger.Info("Starting MCP server with stdio transport")
		// Use stdio transport (MCP SDK default)
		if err := mcpServer.Run(cmd.Context(), &mcp.StdioTransport{}); err != nil {
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
