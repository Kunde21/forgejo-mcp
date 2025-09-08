package server

import (
	"fmt"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
