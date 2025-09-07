package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Integration_VersionCommand(t *testing.T) {
	// Test version command via CLI
	cmd := exec.Command("go", "run", "main.go", "version")
	cmd.Dir = "../" // Run from project root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "forgejo-mcp") {
		t.Errorf("Expected version output to contain 'forgejo-mcp', got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Go version") {
		t.Errorf("Expected version output to contain 'Go version', got: %s", outputStr)
	}
}

func TestCLI_Integration_ConfigCommand(t *testing.T) {
	// Set up environment for config command
	env := os.Environ()
	env = append(env, "FORGEJO_REMOTE_URL=https://example.com")
	env = append(env, "FORGEJO_AUTH_TOKEN=test-token")

	cmd := exec.Command("go", "run", "main.go", "config")
	cmd.Dir = "../" // Run from project root
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Config command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Configuration validation passed") {
		t.Errorf("Expected config output to contain validation success, got: %s", outputStr)
	}
}

func TestCLI_Integration_HelpCommand(t *testing.T) {
	// Test help command
	cmd := exec.Command("go", "run", "main.go", "--help")
	cmd.Dir = "../" // Run from project root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Available Commands") {
		t.Errorf("Expected help output to contain 'Available Commands', got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "serve") {
		t.Errorf("Expected help output to contain 'serve' command, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "version") {
		t.Errorf("Expected help output to contain 'version' command, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "config") {
		t.Errorf("Expected help output to contain 'config' command, got: %s", outputStr)
	}
}

func TestCLI_Integration_GlobalFlags(t *testing.T) {
	// Test global flags with version command
	cmd := exec.Command("go", "run", "main.go", "--verbose", "version")
	cmd.Dir = "../" // Run from project root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command with verbose flag failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "forgejo-mcp") {
		t.Errorf("Expected version output, got: %s", outputStr)
	}
}

func TestCLI_Integration_BackwardCompatibility(t *testing.T) {
	// Test that running without arguments shows help (backward compatibility)
	env := os.Environ()
	env = append(env, "FORGEJO_REMOTE_URL=https://example.com")
	env = append(env, "FORGEJO_AUTH_TOKEN=test-token")

	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = "../" // Run from project root
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's the expected error (help shown)
		outputStr := string(output)
		if strings.Contains(outputStr, "Available Commands") {
			// This is expected - help is shown when no command is provided
			return
		}
		t.Fatalf("Unexpected error: %v, output: %s", err, string(output))
	}
}

func TestCLI_Integration_CommandHelp(t *testing.T) {
	// Test individual command help
	cmd := exec.Command("go", "run", "main.go", "serve", "--help")
	cmd.Dir = "../" // Run from project root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Serve help command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Start the Model Context Protocol server") {
		t.Errorf("Expected serve help to contain description, got: %s", outputStr)
	}
}

func TestCLI_Integration_ConfigFlag(t *testing.T) {
	// Test --config flag
	cmd := exec.Command("go", "run", "main.go", "--config", "/tmp/nonexistent.yaml", "version")
	cmd.Dir = "../" // Run from project root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command with config flag failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "forgejo-mcp") {
		t.Errorf("Expected version output, got: %s", outputStr)
	}
}
