package cmd

import (
	"os"
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}

	if cmd.Use != "forgejo-mcp" {
		t.Errorf("Expected command use to be 'forgejo-mcp', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestRootCmd_GlobalFlags(t *testing.T) {
	cmd := NewRootCmd()

	// Check if --config flag exists
	configFlag := cmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("Expected --config flag to be defined")
	}

	// Check if --verbose flag exists
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected --verbose flag to be defined")
	}
}

func TestRootCmd_Subcommands(t *testing.T) {
	cmd := NewRootCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Expected root command to have subcommands")
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
		t.Error("Expected 'serve' subcommand to be registered")
	}
}

func TestRootCmd_ConfigFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Test setting config flag
	err := cmd.PersistentFlags().Set("config", "/tmp/test-config.yaml")
	if err != nil {
		t.Fatalf("Failed to set config flag: %v", err)
	}

	// Verify the flag value
	configFlag := cmd.PersistentFlags().Lookup("config")
	if configFlag.Value.String() != "/tmp/test-config.yaml" {
		t.Errorf("Expected config flag value to be '/tmp/test-config.yaml', got '%s'", configFlag.Value.String())
	}
}

func TestRootCmd_VerboseFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Test setting verbose flag
	err := cmd.PersistentFlags().Set("verbose", "true")
	if err != nil {
		t.Fatalf("Failed to set verbose flag: %v", err)
	}

	// Verify the flag value
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag.Value.String() != "true" {
		t.Errorf("Expected verbose flag value to be 'true', got '%s'", verboseFlag.Value.String())
	}
}

func TestRootCmd_PersistentPreRun_ConfigFlag(t *testing.T) {
	// Clean up environment
	originalConfig := os.Getenv("FORGEJO_CONFIG_FILE")
	defer func() {
		if originalConfig != "" {
			os.Setenv("FORGEJO_CONFIG_FILE", originalConfig)
		} else {
			os.Unsetenv("FORGEJO_CONFIG_FILE")
		}
	}()

	cmd := NewRootCmd()

	// Test setting config flag and running pre-run
	err := cmd.PersistentFlags().Set("config", "/tmp/test-config.yaml")
	if err != nil {
		t.Fatalf("Failed to set config flag: %v", err)
	}

	// Execute the persistent pre-run
	err = cmd.PersistentPreRunE(cmd, []string{})
	if err != nil {
		t.Fatalf("PersistentPreRunE failed: %v", err)
	}

	// Verify environment variable was set
	configEnv := os.Getenv("FORGEJO_CONFIG_FILE")
	if configEnv != "/tmp/test-config.yaml" {
		t.Errorf("Expected FORGEJO_CONFIG_FILE to be '/tmp/test-config.yaml', got '%s'", configEnv)
	}
}

func TestRootCmd_PersistentPreRun_VerboseFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Test setting verbose flag and running pre-run
	err := cmd.PersistentFlags().Set("verbose", "true")
	if err != nil {
		t.Fatalf("Failed to set verbose flag: %v", err)
	}

	// Execute the persistent pre-run
	err = cmd.PersistentPreRunE(cmd, []string{})
	if err != nil {
		t.Fatalf("PersistentPreRunE failed: %v", err)
	}

	// Note: We can't easily test the log.SetFlags call without capturing log output
	// This is tested indirectly through integration tests
}

func TestRootCmd_PersistentPreRun_NoFlags(t *testing.T) {
	// Clean up environment
	originalConfig := os.Getenv("FORGEJO_CONFIG_FILE")
	defer func() {
		if originalConfig != "" {
			os.Setenv("FORGEJO_CONFIG_FILE", originalConfig)
		} else {
			os.Unsetenv("FORGEJO_CONFIG_FILE")
		}
	}()

	cmd := NewRootCmd()

	// Execute the persistent pre-run without setting flags
	err := cmd.PersistentPreRunE(cmd, []string{})
	if err != nil {
		t.Fatalf("PersistentPreRunE failed: %v", err)
	}

	// Verify environment variable was not set
	configEnv := os.Getenv("FORGEJO_CONFIG_FILE")
	if configEnv != "" {
		t.Errorf("Expected FORGEJO_CONFIG_FILE to be empty, got '%s'", configEnv)
	}
}
