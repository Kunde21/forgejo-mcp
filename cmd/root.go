package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	debug    bool
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "forgejo-mcp",
	Short: "A Model Context Protocol server for Forgejo repositories",
	Long: `forgejo-mcp is a server that enables AI agents to interact with Forgejo repositories
through standardized CLI tools. It provides functionality for managing pull requests
and issues in Forgejo repositories.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Any global pre-run logic can go here
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "set log level (trace, debug, info, warn, error, fatal, panic)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// If a config file is found, read it in.
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "config" (without extension)
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("FORGEJO_MCP")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there's no config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}
