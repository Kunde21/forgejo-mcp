package cmd

import (
	"context"
	"fmt"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// registerTools registers MCP tools with the server
func registerTools(server *mcp.Server, cfg *config.Config, logger *logrus.Logger) error {
	// Register PR list tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pr_list",
		Description: "List pull requests from the Forgejo repository",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		State  string `json:"state,omitempty"`
		Author string `json:"author,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: Implement actual PR listing using tea CLI
		// For now, return mock data
		mockPRs := []map[string]interface{}{
			{
				"number":    42,
				"title":     "Add dark mode support",
				"author":    "developer1",
				"state":     "open",
				"createdAt": "2025-08-26T10:00:00Z",
				"updatedAt": "2025-08-26T15:30:00Z",
			},
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Found %d pull requests", len(mockPRs)),
				},
			},
		}, mockPRs, nil
	})

	// Register issue list tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "issue_list",
		Description: "List issues from the Forgejo repository",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		State  string   `json:"state,omitempty"`
		Author string   `json:"author,omitempty"`
		Labels []string `json:"labels,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: Implement actual issue listing using tea CLI
		// For now, return mock data
		mockIssues := []map[string]interface{}{
			{
				"number":    123,
				"title":     "UI responsiveness issue on mobile",
				"author":    "user1",
				"state":     "open",
				"labels":    []string{"bug", "ui", "mobile"},
				"createdAt": "2025-08-24T08:30:00Z",
			},
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Found %d issues", len(mockIssues)),
				},
			},
		}, mockIssues, nil
	})

	logger.Info("Registered MCP tools")
	return nil
}

var (
	host string
	port int
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the Model Context Protocol server to handle requests from AI agents.
	
Examples:
  # Start server with default settings
  forgejo-mcp serve

  # Start server on custom host and port
  forgejo-mcp serve --host 0.0.0.0 --port 8080`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate host and port
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Override config with command line flags
		cfg.Host = host
		cfg.Port = port

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
		fmt.Printf("Starting MCP server on %s:%d\n", host, port)

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

	// Add serve command flags
	serveCmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to")
	serveCmd.Flags().IntVar(&port, "port", 3000, "Port to listen on")

	// Add aliases
	serveCmd.Aliases = []string{"server", "start"}
}
