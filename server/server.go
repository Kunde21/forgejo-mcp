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

// New creates a new MCP server instance
func New() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewFromConfig(cfg)
}

func NewFromConfig(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	giteaClient, err := gitea.NewGiteaClient(cfg.RemoteURL, cfg.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}
	s := &Server{
		config:       cfg,
		giteaService: gitea.NewService(giteaClient),
	}
	mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")

	mcpServer.AddTool(mcp.NewTool("hello",
		mcp.WithDescription("Returns a hello world message"),
	), s.handleHello)

	mcpServer.AddTool(mcp.NewTool("list_issues",
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
	), s.handleListIssues)
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

func (s *Server) MCPServer() *server.MCPServer { return s.mcpServer }
