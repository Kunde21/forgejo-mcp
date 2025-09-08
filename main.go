// Package main provides the entry point for the forgejo-mcp application.
// This application serves as an MCP (Model Context Protocol) server for
// interacting with Forgejo/Gitea repositories.
//
// Migration Note: Updated to use the official MCP SDK for improved
// protocol compliance and long-term stability.
package main

import (
	"os"

	"github.com/kunde21/forgejo-mcp/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	// Execute the root command and handle any errors
	// Cobra CLI framework manages command parsing and execution
	if err := rootCmd.Execute(); err != nil {
		// Cobra handles error output, just exit with failure code
		os.Exit(1)
	}
}
