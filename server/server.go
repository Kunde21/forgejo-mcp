package server

import (
	"context"
	"fmt"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the MCP server instance using the official MCP SDK.
// This server provides tools for interacting with Forgejo/Gitea repositories
// through the Model Context Protocol.
//
// Migration Note: Updated from mark3labs/mcp-go to github.com/modelcontextprotocol/go-sdk/mcp v0.4.0
// for official protocol compliance and long-term stability.
type Server struct {
	mcpServer *mcp.Server
	config    *config.Config
	remote    gitea.GiteaClientInterface
}

// New creates a new MCP server instance with default configuration.
// It initializes the server with the official MCP SDK and registers all available tools.
//
// Migration Note: Constructor updated to use the official MCP SDK's server initialization
// pattern instead of the third-party library's approach.
func New() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewFromConfig(cfg)
}

// NewFromConfig creates a new MCP server instance with the provided configuration.
// This allows for custom server setup while maintaining the official SDK integration.
//
// Migration Note: Tool registration updated to use mcp.AddTool() instead of the
// previous SDK's tool registration methods.
func NewFromConfig(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	giteaClient, err := gitea.NewGiteaClient(cfg.RemoteURL, cfg.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}
	return NewFromService(giteaClient, cfg)
}

// NewFromService creates a new MCP server instance with the provided service.
// This allows for dependency injection, particularly useful for testing with mock services.
func NewFromService(service gitea.GiteaClientInterface, cfg *config.Config) (*Server, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	s := &Server{
		config: cfg,
		remote: service,
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
		Name:        "issue_list",
		Description: "List issues from a Gitea/Forgejo repository",
	}, s.handleIssueList)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "issue_comment_create",
		Description: "Create a comment on a Forgejo/Gitea repository issue",
	}, s.handleIssueCommentCreate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "issue_comment_list",
		Description: "List comments from a Forgejo/Gitea repository issue with pagination support",
	}, s.handleIssueCommentList)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "issue_comment_edit",
		Description: "Edit an existing comment on a Forgejo/Gitea repository issue",
	}, s.handleIssueCommentEdit)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_list",
		Description: "List pull requests from a Forgejo/Gitea repository with pagination and state filtering",
	}, s.handlePullRequestList)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_comment_list",
		Description: "List comments from a Forgejo/Gitea repository pull request with pagination support",
	}, s.handlePullRequestCommentList)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_comment_create",
		Description: "Create a comment on a Forgejo/Gitea repository pull request",
	}, s.handlePullRequestCommentCreate)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pr_comment_edit",
		Description: "Edit an existing comment on a Forgejo/Gitea repository pull request",
	}, s.handlePullRequestCommentEdit)

	s.mcpServer = mcpServer
	return s, nil
}

// Start starts the MCP server using stdio transport.
// The server will listen for MCP protocol messages on stdin/stdout.
//
// Migration Note: Updated to use the official SDK's Run method with StdioTransport
// instead of the previous SDK's server start pattern.
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}

// Stop stops the MCP server gracefully.
// Note: For stdio transport, the server runs until the process ends,
// so this method primarily handles cleanup of resources.
//
// Migration Note: The official SDK handles server lifecycle differently;
// stdio servers run until process termination rather than explicit stopping.
func (s *Server) Stop() error {
	// MCP server doesn't have a direct stop method for stdio
	// It runs until the process ends
	return nil
}

// MCPServer returns the underlying MCP server instance.
// This provides access to the official SDK server for advanced use cases.
//
// Migration Note: Returns the official SDK's *mcp.Server instead of the
// previous third-party library's server type.
func (s *Server) MCPServer() *mcp.Server { return s.mcpServer }
