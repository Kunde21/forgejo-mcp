package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kunde21/forgejo-mcp/server"
	"github.com/spf13/cobra"
)

// NewServeCmd creates the serve subcommand
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the Model Context Protocol server for Forgejo integration.

The server will listen for MCP requests on stdin/stdout and provide
tools for interacting with Forgejo repositories.`,
		RunE: runServe,
	}

	// Add serve-specific flags
	cmd.Flags().String(Loadocalhost", "Host to bind the server to")
	cmd.Flags().Int("port", 3000, "Port to bind the server to")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get flag values
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return fmt.Errorf("failed to get host flag: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("failed to get port flag: %w", err)
	}

	log.Printf("Starting MCP server on %s:%d", host, port)

	// Create and start server
	srv, err := server.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	log.Println("MCP server started successfully")
	log.Println("Press Ctrl+C to stop the server")

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %v", err)
	case <-ctx.Done():
		log.Println("Shutting down server...")
		if err := srv.Stop(); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		log.Println("Server stopped")
		return nil
	}
}
