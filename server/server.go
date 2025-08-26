// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"fmt"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/sirupsen/logrus"
)

// Server represents the MCP server with all its dependencies
type Server struct {
	config *config.Config
	logger *logrus.Logger
	cancel context.CancelFunc
	stopCh chan struct{}
	// mcpServer will be added when we implement the MCP integration
	// mcpServer mcp.Server
}

// New creates a new MCP server instance with the provided configuration
func New(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger
	logger := logrus.New()

	// Set log level based on configuration
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", cfg.LogLevel, err)
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Debug {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	server := &Server{
		config: cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	logger.Info("MCP server created successfully")
	return server, nil
}

// Start begins the MCP server and blocks until stopped or an error occurs
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server...")

	// Create a cancellable context for internal use
	serverCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer cancel()

	// TODO: Implement MCP server initialization and transport setup
	// For now, we'll just wait for either context cancellation or stop signal
	select {
	case <-serverCtx.Done():
		s.logger.Info("MCP server stopped via context cancellation")
	case <-s.stopCh:
		s.logger.Info("MCP server stopped via Stop() method")
	}

	s.logger.Info("MCP server shutdown complete")
	return nil
}

// Stop gracefully shuts down the MCP server
func (s *Server) Stop() error {
	s.logger.Info("Stopping MCP server...")

	// Signal the stop channel to trigger shutdown
	select {
	case s.stopCh <- struct{}{}:
		// Successfully sent stop signal
	case <-context.Background().Done():
		// Context was cancelled, server might already be stopping
	}

	// If we have a cancel function, call it to ensure cleanup
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("MCP server stopped successfully")
	return nil
}
