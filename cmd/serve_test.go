package cmd

import (
	"testing"
)

func TestNewServeCmd(t *testing.T) {
	cmd := NewServeCmd()

	if cmd == nil {
		t.Fatal("NewServeCmd() returned nil")
	}

	if cmd.Use != "serve" {
		t.Errorf("Expected command use to be 'serve', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestServeCmd_Flags(t *testing.T) {
	cmd := NewServeCmd()

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Error("Expected --host flag to be defined")
	}

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("Expected --port flag to be defined")
	}
}

func TestServeCmd_HostFlag(t *testing.T) {
	cmd := NewServeCmd()

	err := cmd.Flags().Set("host", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to set host flag: %v", err)
	}

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag.Value.String() != "127.0.0.1" {
		t.Errorf("Expected host flag value to be '127.0.0.1', got '%s'", hostFlag.Value.String())
	}
}

func TestServeCmd_PortFlag(t *testing.T) {
	cmd := NewServeCmd()

	// Test setting port flag
	err := cmd.Flags().Set("port", "8080")
	if err != nil {
		t.Fatalf("Failed to set port flag: %v", err)
	}

	// Verify the flag value
	portFlag := cmd.Flags().Lookup("port")
	if portFlag.Value.String() != "8080" {
		t.Errorf("Expected port flag value to be '8080', got '%s'", portFlag.Value.String())
	}
}

func TestServeCmd_DefaultValues(t *testing.T) {
	cmd := NewServeCmd()

	// Check default host value
	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag.DefValue != "localhost" {
		t.Errorf("Expected default host to be 'localhost', got '%s'", hostFlag.DefValue)
	}

	// Check default port value
	portFlag := cmd.Flags().Lookup("port")
	if portFlag.DefValue != "3000" {
		t.Errorf("Expected default port to be '3000', got '%s'", portFlag.DefValue)
	}
}

func TestServeCmd_RunE(t *testing.T) {
	cmd := NewServeCmd()

	// Test that RunE is set
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set for serve command")
	}

	// Note: We can't easily test the actual server startup without mocking
	// This would be tested in integration tests
	// For now, we verify the command structure is correct
}

func TestServeCmd_ServerLifecycle(t *testing.T) {
	// This test would verify server startup and shutdown
	// For now, we test the command structure
	cmd := NewServeCmd()

	// Verify command has proper structure for server lifecycle
	if cmd.Use != "serve" {
		t.Errorf("Expected serve command, got '%s'", cmd.Use)
	}

	// Test flag parsing
	err := cmd.Flags().Set("host", "0.0.0.0")
	if err != nil {
		t.Fatalf("Failed to set host flag: %v", err)
	}

	err = cmd.Flags().Set("port", "9090")
	if err != nil {
		t.Fatalf("Failed to set port flag: %v", err)
	}

	hostFlag := cmd.Flags().Lookup("host")
	portFlag := cmd.Flags().Lookup("port")

	if hostFlag.Value.String() != "0.0.0.0" {
		t.Errorf("Expected host to be '0.0.0.0', got '%s'", hostFlag.Value.String())
	}

	if portFlag.Value.String() != "9090" {
		t.Errorf("Expected port to be '9090', got '%s'", portFlag.Value.String())
	}
}
