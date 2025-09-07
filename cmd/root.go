package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for forgejo-mcp
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "forgejo-mcp",
		Short: "MCP server for Forgejo integration",
		Long: `A Model Context Protocol (MCP) server that provides integration
with Forgejo (Gitea fork) for repository management and issue tracking.

This server enables AI assistants to interact with Forgejo repositories
through standardized MCP tools and resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, default to serve command
			// This maintains backward compatibility with go run main.go
			return NewServeCmd().RunE(cmd, args)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Handle config flag
			if configFlag := cmd.PersistentFlags().Lookup("config"); configFlag != nil {
				configPath, err := cmd.PersistentFlags().GetString("config")
				if err != nil {
					return fmt.Errorf("failed to get config flag: %w", err)
				}

				if configPath != "" {
					// Set FORGEJO_CONFIG_FILE environment variable if config path is provided
					if err := os.Setenv("FORGEJO_CONFIG_FILE", configPath); err != nil {
						return fmt.Errorf("failed to set config environment variable: %w", err)
					}
				}
			}

			// Handle verbose flag
			if verboseFlag := cmd.PersistentFlags().Lookup("verbose"); verboseFlag != nil {
				verbose, err := cmd.PersistentFlags().GetBool("verbose")
				if err != nil {
					return fmt.Errorf("failed to get verbose flag: %w", err)
				}

				if verbose {
					// Enable verbose logging
					log.SetFlags(log.LstdFlags | log.Lshortfile)
					log.Println("Verbose logging enabled")
				}
			}

			return nil
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().String("config", "", "Path to configuration file")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Add subcommands
	rootCmd.AddCommand(NewServeCmd())
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewConfigCmd())

	return rootCmd
}
