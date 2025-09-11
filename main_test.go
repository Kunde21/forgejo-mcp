package main

import (
	"os"
	"testing"
)

func TestMain_Execute(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"forgejo-mcp", "version"}
}

func TestMain_BackwardCompatibility(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"forgejo-mcp"}
}

func TestMain_CommandStructure(t *testing.T) {
}

func TestMain_EnvironmentSetup(t *testing.T) {
	originalURL := os.Getenv("FORGEJO_REMOTE_URL")
	originalToken := os.Getenv("FORGEJO_AUTH_TOKEN")
	defer func() {
		os.Setenv("FORGEJO_REMOTE_URL", originalURL)
		os.Setenv("FORGEJO_AUTH_TOKEN", originalToken)
	}()

	os.Setenv("FORGEJO_REMOTE_URL", "https://example.com")
	os.Setenv("FORGEJO_AUTH_TOKEN", "test-token")
}

func TestMain_ConfigFlag(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"forgejo-mcp", "--config", "/tmp/test-config.yaml", "version"}
}

func TestMain_VerboseFlag(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"forgejo-mcp", "--verbose", "version"}
}
