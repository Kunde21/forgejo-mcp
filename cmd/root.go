package main

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "forgejo-mcp",
	Short: "A Model Context Protocol server for Forgejo repositories",
	Long: `forgejo-mcp is a server that enables AI agents to interact with Forgejo repositories
through standardized CLI tools. It provides functionality for managing pull requests
and issues in Forgejo repositories.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// TODO: Add root command initialization
}
