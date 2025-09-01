package cmd

import (
	"testing"
)

func TestExecute(t *testing.T) {
	// Test that Execute function exists
	// We can't easily test Execute() without mocking since it calls os.Exit
	// But we can verify it exists and has the right signature
	_ = Execute // Just reference it to ensure it exists
}

func TestRootCommandStructure(t *testing.T) {
	// Test that rootCmd is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd should be initialized")
	}

	if rootCmd.Use != "forgejo-mcp" {
		t.Errorf("Expected command use to be 'forgejo-mcp', got %s", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rootCmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestRootCommandFlags(t *testing.T) {
	// Test global flags
	configFlag := rootCmd.Flag("config")
	if configFlag == nil {
		t.Error("Root command should have 'config' flag")
	}

	debugFlag := rootCmd.Flag("debug")
	if debugFlag == nil {
		t.Error("Root command should have 'debug' flag")
	}

	logLevelFlag := rootCmd.Flag("log-level")
	if logLevelFlag == nil {
		t.Error("Root command should have 'log-level' flag")
	}
}

func TestServeCommandStructure(t *testing.T) {
	// Test that serveCmd is properly initialized
	if serveCmd == nil {
		t.Fatal("serveCmd should be initialized")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("Expected serve command use to be 'serve', got %s", serveCmd.Use)
	}

	if serveCmd.Short == "" {
		t.Error("Serve command should have a short description")
	}

	// Test serve command flags
	hostFlag := serveCmd.Flag("host")
	if hostFlag == nil {
		t.Error("Serve command should have 'host' flag")
	}

	portFlag := serveCmd.Flag("port")
	if portFlag == nil {
		t.Error("Serve command should have 'port' flag")
	}
}

func TestCommandRegistration(t *testing.T) {
	// Test that subcommands are registered
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Root command should have subcommands")
	}

	// Check for serve subcommand
	foundServe := false
	for _, subcmd := range subcommands {
		if subcmd.Use == "serve" {
			foundServe = true
			break
		}
	}
	if !foundServe {
		t.Error("Root command should have 'serve' subcommand")
	}
}
