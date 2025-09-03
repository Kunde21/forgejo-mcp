package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/Kunde21/forgejo-mcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	host      string
	port      int
	transport string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the Model Context Protocol server to handle requests from AI agents.

Examples:
  # Start server with default settings
  forgejo-mcp serve

  # Start server on custom host and port
  forgejo-mcp serve --host 0.0.0.0 --port 8080

  # Start server with debug logging
  forgejo-mcp serve --debug --log-level debug

  # Start server with SSE transport
  forgejo-mcp serve --transport sse`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate host and port
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}

		// Validate transport
		if transport != "stdio" && transport != "sse" {
			return fmt.Errorf("transport must be either 'stdio' or 'sse', got: %s", transport)
		}

		// Validate log level
		validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
		if slices.Contains(validLevels, logLevel) {
			return nil
		}
		return fmt.Errorf("invalid log level '%s', must be one of: %v", logLevel, validLevels)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func runServer() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command line flags
	cfg.Host = host
	cfg.Port = port
	cfg.LogLevel = logLevel
	cfg.Debug = debug

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.WithFields(logrus.Fields{
		"host":      cfg.Host,
		"port":      cfg.Port,
		"transport": transport,
		"log_level": cfg.LogLevel,
		"debug":     cfg.Debug,
	}).Info("Starting MCP server")

	// Create MCP server instance
	mcpSrv, err := server.NewMCPServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create transport based on command line flag
	var mcpTransport mcp.Transport
	switch transport {
	case "stdio":
		mcpTransport = mcp.NewStdioTransport()
	case "sse":
		// For SSE, we would need to set up an HTTP server
		// For now, fall back to stdio
		logger.Warn("SSE transport not yet implemented, falling back to stdio")
		mcpTransport = mcp.NewStdioTransport()
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}

	logger.Info("MCP server started successfully")

	// Start the server with the selected transport
	errCh := make(chan error, 1)
	go func() {
		errCh <- mcpSrv.Run(ctx, mcpTransport)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()

		// Give server time to shut down gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Wait for the server to finish shutting down
		select {
		case <-shutdownCtx.Done():
			logger.Warn("Server shutdown timed out")
		case err := <-errCh:
			if err != nil {
				logger.WithError(err).Error("Error during server shutdown")
				return fmt.Errorf("server shutdown error: %w", err)
			}
		}

		logger.Info("MCP server shutdown complete")
		return nil

	case err := <-errCh:
		if err != nil {
			logger.WithError(err).Error("Server error")
			return fmt.Errorf("server error: %w", err)
		}
		logger.Info("Server stopped normally")
		return nil
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Add serve command flags
	serveCmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to")
	serveCmd.Flags().IntVar(&port, "port", 3000, "Port to listen on")
	serveCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	serveCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode with verbose logging")
	serveCmd.Flags().StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")

	// Add aliases
	serveCmd.Aliases = []string{"server", "start"}
}
