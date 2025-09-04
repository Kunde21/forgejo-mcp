// Package server implements the MCP server functionality using the MCP Go SDK
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/config"
	ctxt "github.com/Kunde21/forgejo-mcp/context"
	"github.com/Kunde21/forgejo-mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// parseTime is a helper function to parse time strings
func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}

// PRListArgs represents the arguments for the pr_list tool
type PRListArgs struct {
	State  string `json:"state,omitempty"`
	Author string `json:"author,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// IssueListArgs represents the arguments for the issue_list tool
type IssueListArgs struct {
	State  string   `json:"state,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Author string   `json:"author,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}

// ContextDetectArgs represents the arguments for the context_detect tool
type ContextDetectArgs struct {
	Path string `json:"path,omitempty"`
}

// MCPServer wraps the MCP SDK server with Forgejo-specific functionality
type MCPServer struct {
	mcpServer   *mcp.Server
	config      *config.Config
	logger      *logrus.Logger
	authState   *AuthState
	giteaClient *client.ForgejoClient
}

// NewMCPServer creates a new MCP server using the MCP SDK
func NewMCPServer(cfg *config.Config) (*MCPServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", cfg.LogLevel, err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})
	if cfg.Debug {
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05Z07:00"})
	}

	// Create MCP server with implementation info
	impl := &mcp.Implementation{
		Name:    "forgejo-mcp",
		Version: "1.0.0",
	}
	mcpServer := mcp.NewServer(impl, &mcp.ServerOptions{})

	// Initialize authentication state
	authState := NewAuthState(nil, logger)

	// Try to create Gitea client for authentication validation
	var giteaClient *client.ForgejoClient
	if cfg.ForgejoURL != "" && cfg.AuthToken != "" {
		giteaClient, err = client.New(cfg.ForgejoURL, cfg.AuthToken)
		if err != nil {
			logger.WithError(err).Warn("Failed to create Gitea client, authentication may not work")
		} else {
			logger.Info("Gitea client initialized for authentication")
		}
	}

	server := &MCPServer{
		mcpServer:   mcpServer,
		config:      cfg,
		logger:      logger,
		authState:   authState,
		giteaClient: giteaClient,
	}

	// Register MCP tools
	if err := server.registerMCPTools(); err != nil {
		return nil, fmt.Errorf("failed to register MCP tools: %w", err)
	}

	logger.Info("MCP server created successfully")
	return server, nil
}

// registerMCPTools registers all MCP tools with the server
func (s *MCPServer) registerMCPTools() error {
	s.logger.Info("Registering MCP tools...")

	// Register pr_list tool with automatic schema generation
	prListTool := &mcp.Tool{
		Name:        "pr_list",
		Description: "List pull requests from the Forgejo repository",
	}
	mcp.AddTool(s.mcpServer, prListTool, s.handlePRList)

	// Register issue_list tool with automatic schema generation
	issueListTool := &mcp.Tool{
		Name:        "issue_list",
		Description: "List issues from the Forgejo repository",
	}
	mcp.AddTool(s.mcpServer, issueListTool, s.handleIssueList)

	// Register context_detect tool with automatic schema generation
	contextDetectTool := &mcp.Tool{
		Name:        "context_detect",
		Description: "Detect repository context from the current git environment",
	}
	mcp.AddTool(s.mcpServer, contextDetectTool, s.handleContextDetect)

	s.logger.Info("MCP tools registered successfully")
	return nil
}

// handlePRList handles the pr_list tool call
func (s *MCPServer) handlePRList(ctx context.Context, request *mcp.CallToolRequest, args PRListArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Infof("Handling pr_list request with params: %+v", args)

	// Validate authentication
	if err := s.authState.ValidateToken(ctx, s.config.ForgejoURL, s.config.AuthToken); err != nil {
		s.logger.WithError(err).Error("Authentication validation failed")
		return nil, nil, fmt.Errorf("authentication failed: %w", err)
	}

	// For now, return mock data - this will be replaced with actual Gitea API calls
	mockPRs := []types.PullRequest{
		{
			ID:         42,
			Number:     42,
			Title:      "Add dark mode support",
			State:      types.PRStateOpen,
			Author:     &types.PRAuthor{Username: "developer1", AvatarURL: "https://example.com/avatar1.jpg", URL: "https://example.com/user/developer1"},
			HeadBranch: "feature/dark-mode",
			BaseBranch: "main",
			CreatedAt:  types.Timestamp{Time: parseTime("2025-08-26T10:00:00Z")},
			UpdatedAt:  types.Timestamp{Time: parseTime("2025-08-26T15:30:00Z")},
			URL:        "https://example.com/pr/42",
			DiffURL:    "https://example.com/pr/42.diff",
		},
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Found %d pull requests", len(mockPRs))},
		},
	}, nil, nil
}

// handleIssueList handles the issue_list tool call
func (s *MCPServer) handleIssueList(ctx context.Context, request *mcp.CallToolRequest, args IssueListArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Infof("Handling issue_list request with params: %+v", args)

	// Validate authentication
	if err := s.authState.ValidateToken(ctx, s.config.ForgejoURL, s.config.AuthToken); err != nil {
		s.logger.WithError(err).Error("Authentication validation failed")
		return nil, nil, fmt.Errorf("authentication failed: %w", err)
	}

	// For now, return mock data - this will be replaced with actual Gitea API calls
	mockIssues := []types.Issue{
		{
			ID:        123,
			Number:    123,
			Title:     "UI responsiveness issue on mobile",
			State:     types.IssueStateOpen,
			Author:    &types.User{ID: 1, Username: "user1", Email: "user1@example.com"},
			Labels:    []types.PRLabel{{Name: "bug"}, {Name: "ui"}, {Name: "mobile"}},
			CreatedAt: types.Timestamp{Time: parseTime("2025-08-24T08:30:00Z")},
			UpdatedAt: types.Timestamp{Time: parseTime("2025-08-24T10:15:00Z")},
			URL:       "https://example.com/issue/123",
		},
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Found %d issues", len(mockIssues))},
		},
	}, nil, nil
}

// handleContextDetect handles the context_detect tool call
func (s *MCPServer) handleContextDetect(ctx context.Context, request *mcp.CallToolRequest, args ContextDetectArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Infof("Handling context_detect request with params: %+v", args)

	// Extract parameters
	path := "."
	if args.Path != "" {
		path = args.Path
	}

	// Detect repository context
	repoCtx, err := ctxt.DetectContext(path)
	if err != nil {
		s.logger.Errorf("Context detection failed for path %s: %v", path, err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to detect repository context: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	s.logger.Infof("Successfully detected context: %s", repoCtx.String())

	// Return repository information
	repository := types.Repository{
		Owner:    repoCtx.Owner,
		Name:     repoCtx.Repository,
		FullName: repoCtx.String(),
		URL:      repoCtx.RemoteURL,
	}
	if s.config.ForgejoURL != repository.URL {
		s.config.ForgejoURL = repository.URL
		s.giteaClient, err = client.New(s.config.ForgejoURL, s.config.AuthToken)
		if err != nil {
			return nil, nil, err
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Repository: %s", repository.FullName)},
		},
	}, nil, nil
}

// GetMCPServer returns the underlying MCP server
func (s *MCPServer) GetMCPServer() *mcp.Server {
	return s.mcpServer
}

// GetLogger returns the logger
func (s *MCPServer) GetLogger() *logrus.Logger {
	return s.logger
}

// GetConfig returns the configuration
func (s *MCPServer) GetConfig() *config.Config {
	return s.config
}

// Run starts the MCP server with the specified transport
func (s *MCPServer) Run(ctx context.Context, transport mcp.Transport) error {
	s.logger.Info("Starting MCP server...")
	return s.mcpServer.Run(ctx, transport)
}
