package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// Server represents the MCP server instance
type Server struct {
	mcpServer    *server.MCPServer
	config       *config.Config
	giteaService *gitea.Service
}

// NewServer creates a new MCP server instance
func NewServer() (*Server, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create Gitea client
	giteaClient, err := gitea.NewGiteaClient(cfg.RemoteURL, cfg.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Create Gitea service
	giteaService := gitea.NewService(giteaClient)

	s := &Server{
		config:       cfg,
		giteaService: giteaService,
	}

	// Initialize MCP server
	mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")

	// Create and register the hello tool
	helloTool := mcp.NewTool("hello",
		mcp.WithDescription("Returns a hello world message"),
	)

	mcpServer.AddTool(helloTool, s.handleHello)

	// Create and register the list_issues tool
	listIssuesTool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues from a Gitea/Forgejo repository"),
		mcp.WithString("repository",
			mcp.Required(),
			mcp.Description("Repository in format 'owner/repo'"),
		),
		mcp.WithNumber("limit",
			mcp.DefaultNumber(15),
			mcp.Description("Maximum number of issues to return (1-100)"),
		),
		mcp.WithNumber("offset",
			mcp.DefaultNumber(0),
			mcp.Description("Number of issues to skip (0-based)"),
		),
	)

	mcpServer.AddTool(listIssuesTool, s.handleListIssues)

	s.mcpServer = mcpServer
	return s, nil
}

// Start starts the MCP server
func (s *Server) Start() error {
	return server.ServeStdio(s.mcpServer)
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	// MCP server doesn't have a direct stop method for stdio
	// It runs until the process ends
	return nil
}

func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Basic validation - check if request is valid
	if ctx == nil {
		return mcp.NewToolResultError("Context is required"), nil
	}

	// Return the hello world message
	return mcp.NewToolResultText("Hello, World!"), nil
}

func (s *Server) handleListIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse arguments
	type ListArgs struct {
		Repo   string
		Limit  int
		Offset int
	}
	args := ListArgs{
		Repo:   mcp.ParseString(request, "repository", ""),
		Limit:  mcp.ParseInt(request, "limit", 15),
		Offset: mcp.ParseInt(request, "offset", 0),
	}
	if err := validation.ValidateStruct(&args,
		validation.Field(&args.Repo, validation.Required),
		validation.Field(&args.Limit, validation.Min(1), validation.Max(100)),
		validation.Field(&args.Offset, validation.Min(0)),
	); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid request", err), nil
	}

	// Call the Gitea service
	issues, err := s.giteaService.ListIssues(ctx, args.Repo, args.Limit, args.Offset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list issues: %v", err), nil
	}

	// Format response
	type issueList struct {
		Issues []gitea.Issue `json:"issues"`
	}
	return mcp.NewToolResultStructured(issueList{Issues: issues}, fmt.Sprintf("Found %d issues", len(issues))), nil
}

// NewServeCmd creates the serve subcommand
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the Model Context Protocol server for Forgejo integration.

The server will listen for MCP requests on stdin/stdout and provide
tools for interacting with Forgejo repositories.`,
		RunE: runServe,
	}

	// Add serve-specific flags
	cmd.Flags().String("host", "localhost", "Host to bind the server to")
	cmd.Flags().Int("port", 3000, "Port to bind the server to")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get flag values
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return fmt.Errorf("failed to get host flag: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("failed to get port flag: %w", err)
	}

	log.Printf("Starting MCP server on %s:%d", host, port)

	// Create and start server
	server, err := NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	log.Println("MCP server started successfully")
	log.Println("Press Ctrl+C to stop the server")

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %v", err)
	case <-ctx.Done():
		log.Println("Shutting down server...")
		if err := server.Stop(); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		log.Println("Server stopped")
		return nil
	}
}
