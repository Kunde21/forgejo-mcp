package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()

	if cmd == nil {
		t.Fatal("NewConfigCmd() returned nil")
	}

	if cmd.Use != "config" {
		t.Errorf("Expected command use to be 'config', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestConfigCmd_Run(t *testing.T) {
	originalURL := os.Getenv("FORGEJO_REMOTE_URL")
	originalToken := os.Getenv("FORGEJO_AUTH_TOKEN")
	defer func() {
		os.Setenv("FORGEJO_REMOTE_URL", originalURL)
		os.Setenv("FORGEJO_AUTH_TOKEN", originalToken)
	}()

	os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")

	cmd := NewConfigCmd()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Config command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Config command should produce output")
	}
}

func TestConfigCmd_Validation(t *testing.T) {
	originalURL := os.Getenv("FORGEJO_REMOTE_URL")
	originalToken := os.Getenv("FORGEJO_AUTH_TOKEN")
	defer func() {
		os.Setenv("FORGEJO_REMOTE_URL", originalURL)
		os.Setenv("FORGEJO_AUTH_TOKEN", originalToken)
	}()

	os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")

	cmd := NewConfigCmd()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Config command failed: %v", err)
	}

	output := buf.String()

	expectedStrings := []string{
		"Configuration",
		"Remote URL",
		"Auth Token",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output)
		}
	}
}

func TestConfigCmd_ConnectivityTest(t *testing.T) {
	originalURL := os.Getenv("FORGEJO_REMOTE_URL")
	originalToken := os.Getenv("FORGEJO_AUTH_TOKEN")
	defer func() {
		os.Setenv("FORGEJO_REMOTE_URL", originalURL)
		os.Setenv("FORGEJO_AUTH_TOKEN", originalToken)
	}()

	os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")

	cmd := NewConfigCmd()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Config command failed: %v", err)
	}

	output := buf.String()

	connectivityIndicators := []string{
		"Skipping connectivity test",
		"Configuration is valid",
	}

	foundConnectivity := false
	for _, indicator := range connectivityIndicators {
		if bytes.Contains([]byte(output), []byte(indicator)) {
			foundConnectivity = true
			break
		}
	}

	if !foundConnectivity {
		t.Errorf("Expected output to contain connectivity information, but it didn't. Output: %s", output)
	}
}

func TestConfigCmd_ErrorReporting(t *testing.T) {
	// Set up minimal environment for testing
	originalURL := os.Getenv("FORGEJO_REMOTE_URL")
	originalToken := os.Getenv("FORGEJO_AUTH_TOKEN")
	defer func() {
		os.Setenv("FORGEJO_REMOTE_URL", originalURL)
		os.Setenv("FORGEJO_AUTH_TOKEN", originalToken)
	}()

	// Set test values
	os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")

	cmd := NewConfigCmd()

	// Test with invalid configuration
	// This would require setting up invalid environment variables
	// For now, we test the basic error handling structure

	// Capture output
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run the command
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Config command failed: %v", err)
	}

	// The command should handle errors gracefully
	// We can't easily test error conditions without mocking
}

func TestConfigCmd_Flags(t *testing.T) {
	cmd := NewConfigCmd()

	// Config command might have flags for different operations
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command should have a flags object")
	}
}

func TestConfigCmd_Subcommands(t *testing.T) {
	cmd := NewConfigCmd()

	// Config command might have subcommands like 'validate', 'test', etc.
	subcommands := cmd.Commands()

	// For now, we just check that the command structure is valid
	// The actual subcommands would be tested separately
	// Commands() returns nil when there are no subcommands, which is fine
	if len(subcommands) > 0 {
		t.Logf("Config command has %d subcommands", len(subcommands))
	}
}
