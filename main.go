package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer    *server.MCPServer
	config       *Config
	giteaService *gitea.Service
}

func NewServer() (*Server, error) {
	config := LoadConfig()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create Gitea client
	giteaClient, err := gitea.NewGiteaClient(config.RemoteURL, config.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Create Gitea service
	giteaService := gitea.NewService(giteaClient)

	s := &Server{
		config:       config,
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

func (s *Server) Start() error {
	// Start the MCP server using stdio transport
	return server.ServeStdio(s.mcpServer)
}

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
	repo := mcp.ParseString(request, "repository", "")
	limit := mcp.ParseInt(request, "limit", 15)
	offset := mcp.ParseInt(request, "offset", 0)

	// Validate inputs
	if repo == "" {
		return mcp.NewToolResultError("repository parameter is required"), nil
	}
	if limit < 1 || limit > 100 {
		return mcp.NewToolResultError("limit must be between 1 and 100"), nil
	}
	if offset < 0 {
		return mcp.NewToolResultError("offset must be non-negative"), nil
	}

	// Call the Gitea service
	issues, err := s.giteaService.ListIssues(ctx, repo, limit, offset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list issues: %v", err), nil
	}

	// Format response
	return mcp.NewToolResultStructured(issues, fmt.Sprintf("Found %d issues", len(issues))), nil
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Println("Starting MCP server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
