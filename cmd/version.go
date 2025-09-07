package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - these would typically be set during build
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// NewVersionCmd creates the version subcommand
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version information including build details and Go version",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("forgejo-mcp %s\n", Version)
			cmd.Printf("Commit: %s\n", Commit)
			cmd.Printf("Built: %s\n", BuildTime)
			cmd.Printf("Go version: %s\n", runtime.Version())
			cmd.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
