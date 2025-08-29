// Package context provides repository context detection functionality
package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepository checks if the given path is a git repository
// Supports both regular git repositories and git worktrees
func IsGitRepository(path string) bool {
	// Check for .git directory (regular repository)
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil {
		if info.IsDir() {
			return true
		}
		// Check if .git is a file (worktree case)
		if content, err := os.ReadFile(gitDir); err == nil {
			line := strings.TrimSpace(string(content))
			if gitDirPath, found := strings.CutPrefix(line, "gitdir: "); found {
				// This is a worktree, verify the gitdir path exists
				if !filepath.IsAbs(gitDirPath) {
					gitDirPath = filepath.Join(path, gitDirPath)
				}
				if _, err := os.Stat(gitDirPath); err == nil {
					return true
				}
			}
		}
	}

	// Fallback to git command for edge cases
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// GetRemoteURL retrieves the remote URL for the specified remote name
// Defaults to "origin" if no name is provided
func GetRemoteURL(name string) (string, error) {
	return GetRemoteURLInDir(name, ".")
}

// GetRemoteURLInDir retrieves the remote URL for the specified remote name in a specific directory
// Defaults to "origin" if no name is provided
func GetRemoteURLInDir(name, dir string) (string, error) {
	if name == "" {
		name = "origin"
	}

	cmd := exec.Command("git", "remote", "get-url", name)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL for '%s': %w", name, err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("remote '%s' has no URL configured", name)
	}

	return url, nil
}
