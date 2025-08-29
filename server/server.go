// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/sirupsen/logrus"
)

// Server represents the MCP server with all its dependencies
type Server struct {
	config       *config.Config
	logger       *logrus.Logger
	cancel       context.CancelFunc
	stopCh       chan struct{}
	transport    Transport
	dispatcher   *RequestDispatcher
	processor    *MessageProcessor
	toolRegistry *ToolRegistry
}

// New creates a new MCP server instance with the provided configuration
func New(cfg *config.Config) (*Server, error) {
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
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	}

	transport := NewStdioTransport(cfg, logger)
	dispatcher := NewRequestDispatcher(logger)
	server := &Server{
		config:     cfg,
		logger:     logger,
		stopCh:     make(chan struct{}),
		transport:  transport,
		dispatcher: dispatcher,
		processor:  NewMessageProcessor(dispatcher, transport, logger),
	}
	// Initialize tool system
	if err := server.InitializeToolSystem(); err != nil {
		return nil, fmt.Errorf("failed to initialize tool system: %w", err)
	}

	logger.Info("MCP server created successfully")
	return server, nil
}

// ReplaceTeaExecutor replaces the tea executor in all handlers (for testing)
func (s *Server) ReplaceTeaExecutor(executor interface{}) {
	// This is a testing utility method to replace the tea executor
	// In a real implementation, this would be done through dependency injection
	// For now, this is a placeholder that would need to be implemented
	// based on the actual handler structure
	s.logger.Warn("ReplaceTeaExecutor not implemented - this is a testing utility")
}

// Start begins the MCP server and blocks until stopped or an error occurs
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server...")
	if err := s.transport.Connect(); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}
	serverCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.processor.ProcessMessages(serverCtx)
	}()
	select {
	case <-serverCtx.Done():
		s.logger.Info("MCP server stopped via context cancellation")
	case <-s.stopCh:
		s.logger.Info("MCP server stopped via Stop() method")
	case err := <-errCh:
		if err != nil {
			s.logger.Errorf("Message processing error: %v", err)
			return fmt.Errorf("message processing error: %w", err)
		}
	}
	s.logger.Info("MCP server shutdown complete")
	return nil
}

// Stop gracefully shuts down the MCP server
func (s *Server) Stop() error {
	s.logger.Info("Stopping MCP server...")
	select {
	case s.stopCh <- struct{}{}:
	case <-context.Background().Done():
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.logger.Info("MCP server stopped successfully")
	return nil
}

// RegisterGiteaSDKHandlers registers handlers that use the Gitea SDK client
func (s *Server) RegisterGiteaSDKHandlers() error {
	s.logger.Info("Registering Gitea SDK handlers...")

	// Create Gitea client from configuration
	giteaClient, err := client.New(s.config.ForgejoURL, s.config.AuthToken)
	if err != nil {
		return fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Create SDK-based handlers
	sdkPRHandler := NewGiteaSDKPRListHandler(s.logger, giteaClient)
	sdkIssueHandler := NewGiteaSDKIssueListHandler(s.logger, giteaClient)

	// Create a new tool system handler with SDK handlers
	toolHandler := &GiteaSDKToolSystemHandler{
		registry:     s.toolRegistry,
		validator:    NewToolValidator(s.logger),
		prHandler:    sdkPRHandler,
		issueHandler: sdkIssueHandler,
		logger:       s.logger,
	}

	// Replace the existing tool handler
	s.dispatcher.RegisterHandler("tools/call", toolHandler)

	s.logger.Info("Gitea SDK handlers registered successfully")
	return nil
}

// GiteaSDKToolSystemHandler handles tool calls using the Gitea SDK client
type GiteaSDKToolSystemHandler struct {
	registry     *ToolRegistry
	validator    *ToolValidator
	prHandler    *GiteaSDKPRListHandler
	issueHandler *GiteaSDKIssueListHandler
	logger       *logrus.Logger
}

// HandleRequest handles a tool call request using the Gitea SDK
func (gtsh *GiteaSDKToolSystemHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	gtsh.logger.Debugf("Gitea SDK tool system handling request: %s", method)

	// Extract tool name from params
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required and must be a string")
	}

	// Extract tool arguments
	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Validate tool exists
	_, exists := gtsh.registry.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Validate parameters
	if err := gtsh.validator.ValidateParameters(toolName, arguments); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Route to appropriate SDK handler
	switch toolName {
	case "pr_list":
		return gtsh.prHandler.HandleRequest(ctx, toolName, arguments)
	case "issue_list":
		return gtsh.issueHandler.HandleRequest(ctx, toolName, arguments)
	default:
		return nil, fmt.Errorf("tool handler not implemented: %s", toolName)
	}
}
