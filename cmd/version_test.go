package cmd

import (
	"bytes"
	"testing"
)

func TestNewVersionCmd(t *testing.T) {
	cmd := NewVersionCmd()

	if cmd == nil {
		t.Fatal("NewVersionCmd() returned nil")
	}

	if cmd.Use != "version" {
		t.Errorf("Expected command use to be 'version', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestVersionCmd_Run(t *testing.T) {
	cmd := NewVersionCmd()

	// Capture output
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run the command
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	// Check that output contains expected information
	output := buf.String()
	if output == "" {
		t.Error("Version command should produce output")
	}

	// Check for version information
	expectedStrings := []string{
		"forgejo-mcp",
		"version",
		"Go version",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output)
		}
	}
}

func TestVersionCmd_OutputFormat(t *testing.T) {
	cmd := NewVersionCmd()

	// Capture output
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	// Run the command
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	output := buf.String()

	// Check that output contains structured information
	// We expect at least these fields
	if !bytes.Contains([]byte(output), []byte("Version:")) &&
		!bytes.Contains([]byte(output), []byte("version")) {
		t.Error("Output should contain version information")
	}

	if !bytes.Contains([]byte(output), []byte("Go version")) {
		t.Error("Output should contain Go version information")
	}
}

func TestVersionCmd_NoArgs(t *testing.T) {
	cmd := NewVersionCmd()

	// Test with no arguments
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Version command should not fail with no args: %v", err)
	}
}

func TestVersionCmd_WithArgs(t *testing.T) {
	cmd := NewVersionCmd()

	// Test with arguments (should be ignored)
	err := cmd.RunE(cmd, []string{"extra", "args"})
	if err != nil {
		t.Fatalf("Version command should not fail with extra args: %v", err)
	}
}

func TestVersionCmd_Flags(t *testing.T) {
	cmd := NewVersionCmd()

	// Version command typically doesn't have flags, but let's verify
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command should have a flags object")
	}

	// Check that no unexpected flags are defined
	// (This is more of a sanity check)
	if cmd.Flags().NFlag() != 0 {
		t.Errorf("Expected no flags to be set by default, but found %d", cmd.Flags().NFlag())
	}
}
