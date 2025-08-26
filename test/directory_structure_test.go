package test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDirectoryStructure validates that all required directories exist
func TestDirectoryStructure(t *testing.T) {
	requiredDirs := []string{
		"cmd",
		"server",
		"tea",
		"context",
		"auth",
		"config",
		"types",
		"test",
		"test/integration",
		"test/e2e",
	}

	for _, dir := range requiredDirs {
		fullPath := filepath.Join("..", dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Required directory %s does not exist", dir)
		}
	}
}

// TestPackageFiles validates that essential package files exist
func TestPackageFiles(t *testing.T) {
	requiredFiles := []string{
		"cmd/main.go",
		"cmd/root.go",
		"cmd/serve.go",
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join("..", file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Required file %s does not exist", file)
		}
	}
}
