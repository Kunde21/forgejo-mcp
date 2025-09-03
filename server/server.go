// Package server implements the MCP server functionality for Forgejo repositories
//
// Deprecated: This package contains the legacy MCP server implementation.
// Use NewMCPServer() for new MCP SDK-based server implementation.
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/Kunde21/forgejo-mcp/auth"
	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/sirupsen/logrus"
)

// AuthState manages authentication state and validation for the MCP server
type AuthState struct {
	validator auth.TokenValidator
	cache     map[string]bool
	cacheMu   sync.RWMutex
	logger    *logrus.Logger
}

// NewAuthState creates a new authentication state manager
func NewAuthState(validator auth.TokenValidator, logger *logrus.Logger) *AuthState {
	return &AuthState{
		validator: validator,
		cache:     make(map[string]bool),
		logger:    logger,
	}
}

// ValidateToken validates a token with caching
func (as *AuthState) ValidateToken(ctx context.Context, baseURL, token string) error {
	cacheKey := auth.CacheKey(baseURL, token)

	// Check cache first
	as.cacheMu.RLock()
	if cached, exists := as.cache[cacheKey]; exists && cached {
		as.cacheMu.RUnlock()
		as.logger.Debug("Authentication cache hit for token")
		return nil
	}
	as.cacheMu.RUnlock()

	// Perform validation
	var err error
	if as.validator != nil {
		// For testing, call validator directly to avoid auth package validation
		err = as.validator.ValidateToken(baseURL, token)
	} else {
		err = auth.ValidateTokenWithTimeoutDefault(baseURL, token, as.validator)
	}

	// Cache successful validation only
	if err == nil {
		as.cacheMu.Lock()
		as.cache[cacheKey] = true
		as.cacheMu.Unlock()
		as.logger.Debug("Authentication successful, cached result")
	} else {
		as.logger.WithError(err).Debug("Authentication failed")
	}

	return err
}

// GiteaTokenValidator implements auth.TokenValidator using Gitea SDK client
type GiteaTokenValidator struct {
	client *gitea.Client
}

// ValidateToken validates a token using the Gitea SDK client
func (gtv *GiteaTokenValidator) ValidateToken(baseURL, token string) error {
	if gtv.client == nil {
		return fmt.Errorf("Gitea client not initialized")
	}

	// Try to make a simple API call to validate the token
	// We'll use ListMyRepos as it's lightweight and doesn't require specific repo access
	_, _, err := gtv.client.ListMyRepos(gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 1, // Just get one repo to validate token
		},
	})
	if err != nil {
		// Convert error to appropriate auth error type
		return auth.WrapErrorWithContext(err, "token validation", "Gitea API call", token)
	}

	return nil
}

// ClearCache clears the authentication cache
func (as *AuthState) ClearCache() {
	as.cacheMu.Lock()
	defer as.cacheMu.Unlock()
	as.cache = make(map[string]bool)
	as.logger.Debug("Authentication cache cleared")
}

// AuthenticatedToolHandler handles tool calls with authentication validation
type AuthenticatedToolHandler struct {
	registry     *ToolRegistry
	authState    *AuthState
	server       *Server
	innerHandler RequestHandler
	logger       *logrus.Logger
}

// NewAuthenticatedToolHandler creates a new authenticated tool handler
func NewAuthenticatedToolHandler(registry *ToolRegistry, authState *AuthState, server *Server, innerHandler RequestHandler, logger *logrus.Logger) *AuthenticatedToolHandler {
	return &AuthenticatedToolHandler{
		registry:     registry,
		authState:    authState,
		server:       server,
		innerHandler: innerHandler,
		logger:       logger,
	}
}

// HandleRequest handles a tool call request with authentication validation
func (ath *AuthenticatedToolHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	ath.logger.Debugf("Authenticated tool handler processing request: %s", method)

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
	_, exists := ath.registry.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Validate authentication
	if err := ath.authState.ValidateToken(ctx, ath.server.config.ForgejoURL, ath.server.config.AuthToken); err != nil {
		ath.logger.WithError(err).Error("Authentication validation failed")
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Delegate to inner handler for actual tool execution
	if ath.innerHandler != nil {
		return ath.innerHandler.HandleRequest(ctx, method, params)
	}

	// Fallback to placeholder response
	ath.logger.Debugf("Executing tool: %s", toolName)
	return map[string]interface{}{
		"tool":      toolName,
		"status":    "executed",
		"arguments": arguments,
	}, nil
}

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
	authState    *AuthState
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

	// Initialize authentication state
	authState := NewAuthState(nil, logger) // Start with nil validator

	// Try to create Gitea client for authentication validation
	giteaClient, err := client.New(cfg.ForgejoURL, cfg.AuthToken)
	if err != nil {
		logger.WithError(err).Warn("Failed to create Gitea client, authentication may not work")
	} else {
		// Get the underlying gitea client
		underlyingClient := giteaClient.GetGiteaClient()
		if underlyingClient != nil {
			validator := &GiteaTokenValidator{client: underlyingClient}
			authState.validator = validator
			logger.Info("Gitea client initialized for authentication")
		}
	}

	server := &Server{
		config:     cfg,
		logger:     logger,
		stopCh:     make(chan struct{}),
		transport:  transport,
		dispatcher: dispatcher,
		processor:  NewMessageProcessor(dispatcher, transport, logger),
		authState:  authState,
	}

	// Initialize tool system
	if err := server.InitializeToolSystem(); err != nil {
		return nil, fmt.Errorf("failed to initialize tool system: %w", err)
	}

	// Register default handlers
	server.RegisterDefaultHandlers()

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

	// Create the inner tool handler with SDK handlers
	innerHandler := &GiteaSDKToolSystemHandler{
		registry:     s.toolRegistry,
		validator:    NewToolValidator(s.logger),
		prHandler:    sdkPRHandler,
		issueHandler: sdkIssueHandler,
		logger:       s.logger,
	}

	// Wrap with authentication validation
	authenticatedHandler := NewAuthenticatedToolHandler(s.toolRegistry, s.authState, s, innerHandler, s.logger)

	// Replace the existing tool handler with authentication
	s.dispatcher.RegisterHandler("tools/call", authenticatedHandler)

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
func (gtsh *GiteaSDKToolSystemHandler) HandleRequest(ctx context.Context, method string, params map[string]any) (any, error) {
	gtsh.logger.Debugf("Gitea SDK tool system handling request: %s", method)

	// Extract tool name from params
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required and must be a string")
	}

	// Extract tool arguments
	arguments, ok := params["arguments"].(map[string]any)
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
