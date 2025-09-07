package main

import (
	"os"
	"strings"
	"testing"
)

func TestGoModExists(t *testing.T) {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		t.Fatal("go.mod does not exist")
	}
}

func TestCobraDependency(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "github.com/spf13/cobra") {
		t.Error("cobra dependency not found in go.mod")
	}
}

func TestCmdDirectoryExists(t *testing.T) {
	if _, err := os.Stat("cmd"); os.IsNotExist(err) {
		t.Error("cmd directory does not exist")
	}
}
