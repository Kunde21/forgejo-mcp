package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsGitRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		testDir  func(string) string // Function to get the directory to test
		setup    func(string) error
		expected bool
	}{
		{
			name:     "non-git directory",
			testDir:  func(dir string) string { return dir },
			setup:    func(dir string) error { return nil },
			expected: false,
		},
		{
			name:    "regular git repository",
			testDir: func(dir string) string { return dir },
			setup: func(dir string) error {
				gitDir := filepath.Join(dir, ".git")
				return os.MkdirAll(gitDir, 0755)
			},
			expected: true,
		},
		{
			name:    "git worktree",
			testDir: func(dir string) string { return filepath.Join(dir, "worktree") },
			setup: func(dir string) error {
				// Create main repo
				mainRepo := filepath.Join(dir, "main")
				if err := os.MkdirAll(filepath.Join(mainRepo, ".git"), 0755); err != nil {
					return err
				}

				// Create worktree directory
				worktreeDir := filepath.Join(dir, "worktree")
				if err := os.MkdirAll(worktreeDir, 0755); err != nil {
					return err
				}

				// Create .git file pointing to main repo
				gitFile := filepath.Join(worktreeDir, ".git")
				content := "gitdir: ../main/.git\n"
				return os.WriteFile(gitFile, []byte(content), 0644)
			},
			expected: true,
		},
		{
			name:    "worktree with absolute path",
			testDir: func(dir string) string { return filepath.Join(dir, "worktree") },
			setup: func(dir string) error {
				// Create main repo
				mainRepo := filepath.Join(dir, "main")
				if err := os.MkdirAll(filepath.Join(mainRepo, ".git"), 0755); err != nil {
					return err
				}

				// Create worktree
				worktreeDir := filepath.Join(dir, "worktree")
				if err := os.MkdirAll(worktreeDir, 0755); err != nil {
					return err
				}

				// Create .git file with absolute path
				gitFile := filepath.Join(worktreeDir, ".git")
				absPath := filepath.Join(mainRepo, ".git")
				content := "gitdir: " + absPath + "\n"
				return os.WriteFile(gitFile, []byte(content), 0644)
			},
			expected: true,
		},
		{
			name:    "invalid worktree gitdir",
			testDir: func(dir string) string { return filepath.Join(dir, "worktree") },
			setup: func(dir string) error {
				worktreeDir := filepath.Join(dir, "worktree")
				if err := os.MkdirAll(worktreeDir, 0755); err != nil {
					return err
				}

				// Create .git file pointing to non-existent directory
				gitFile := filepath.Join(worktreeDir, ".git")
				content := "gitdir: /non/existent/path\n"
				return os.WriteFile(gitFile, []byte(content), 0644)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			if tt.setup != nil {
				if err := tt.setup(testDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			actualTestDir := tt.testDir(testDir)
			result := IsGitRepository(actualTestDir)
			if result != tt.expected {
				t.Errorf("IsGitRepository(%s) = %v, expected %v", actualTestDir, result, tt.expected)
			}
		})
	}
}

func TestGetRemoteURL(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git_remote_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock git repository
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repository
	initCmd := exec.Command("git", "init")
	initCmd.Dir = repoDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	tests := []struct {
		name        string
		remoteName  string
		setupCmd    string
		expectedURL string
		expectError bool
	}{
		{
			name:        "default origin remote",
			remoteName:  "",
			setupCmd:    "git remote add origin https://codeberg.org/user/repo.git",
			expectedURL: "https://codeberg.org/user/repo.git",
			expectError: false,
		},
		{
			name:        "named remote",
			remoteName:  "upstream",
			setupCmd:    "git remote add upstream https://github.com/user/repo.git",
			expectedURL: "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name:        "SSH remote",
			remoteName:  "",
			setupCmd:    "git remote remove origin 2>/dev/null || true && git remote add origin git@codeberg.org:user/repo.git",
			expectedURL: "git@codeberg.org:user/repo.git",
			expectError: false,
		},
		{
			name:        "non-existent remote",
			remoteName:  "nonexistent",
			setupCmd:    "",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing remotes first
			cleanupCmd := exec.Command("git", "remote", "remove", "origin")
			cleanupCmd.Dir = repoDir
			cleanupCmd.Run() // Ignore errors, remote might not exist

			// Clean up upstream remote if it exists
			cleanupCmd = exec.Command("git", "remote", "remove", "upstream")
			cleanupCmd.Dir = repoDir
			cleanupCmd.Run() // Ignore errors

			// Run setup command if provided
			if tt.setupCmd != "" {
				cmd := exec.Command("bash", "-c", tt.setupCmd)
				cmd.Dir = repoDir
				if err := cmd.Run(); err != nil {
					t.Logf("Setup command failed (expected for some tests): %v", err)
				}
			}

			url, err := GetRemoteURLInDir(tt.remoteName, repoDir)
			if tt.expectError {
				if err == nil {
					t.Errorf("GetRemoteURLInDir(%s, %s) expected error, got none", tt.remoteName, repoDir)
				}
			} else {
				if err != nil {
					t.Errorf("GetRemoteURLInDir(%s, %s) unexpected error: %v", tt.remoteName, repoDir, err)
				}
				if url != tt.expectedURL {
					t.Errorf("GetRemoteURLInDir(%s, %s) = %q, expected %q", tt.remoteName, repoDir, url, tt.expectedURL)
				}
			}
		})
	}
}
