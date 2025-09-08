package server

import (
	"context"
	"fmt"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the MCP server instance
type Server struct {
	mcpServer    *mcp.Server
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
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "forgejo-mcp",
		Version: "1.0.0",
	}, nil)

	// Add tools using the new SDK
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "hello",
		Description: "Returns a hello world message",
	}, s.handleHello)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_issues",
		Description: "List issues from a Gitea/Forgejo repository",
	}, s.handleListIssues)

	s.mcpServer = mcpServer
	return s, nil
}

// Start starts the MCP server
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	// MCP server doesn't have a direct stop method for stdio
	// It runs until the process ends
	return nil
}

func (s *Server) MCPServer() *mcp.Server { return s.mcpServer }
