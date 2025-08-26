package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	host string
	port int
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
  forgejo-mcp serve --host 0.0.0.0 --port 8080`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate host and port
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server start logic
		fmt.Printf("Starting server on %s:%d\n", host, port)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Add serve command flags
	serveCmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to")
	serveCmd.Flags().IntVar(&port, "port", 3000, "Port to listen on")

	// Add aliases
	serveCmd.Aliases = []string{"server", "start"}
}
