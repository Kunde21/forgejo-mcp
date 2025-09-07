package main

import (
	"context"
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer    *server.MCPServer
	config       *config.Config
	giteaService *gitea.Service
}

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
