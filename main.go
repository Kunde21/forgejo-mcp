package main

import (
	"os"

	"github.com/kunde21/forgejo-mcp/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		// Cobra handles error output, just exit
		os.Exit(1)
	}
}
