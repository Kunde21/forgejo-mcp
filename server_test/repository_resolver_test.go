package servertest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kunde21/forgejo-mcp/server"
)

// Test cases for repository resolver functionality
type repositoryResolverTestCase struct {
	name        string
	setup       func(*testing.T) string
	expectError bool
	errorMsg    string
	expectRepo  *server.RepositoryResolution
}

func TestRepositoryResolver_ValidateDirectory(t *testing.T) {
	t.Parallel()

	testCases := []repositoryResolverTestCase{
		{
			name: "valid_git_repository",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				gitDir := filepath.Join(dir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				return dir
			},
			expectError: false,
		},
		{
			name: "nonexistent_directory",
			setup: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			expectError: true,
			errorMsg:    "directory does not exist",
		},
		{
			name: "directory_without_git",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				// Create a regular file, not .git directory
				if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return dir
			},
			expectError: true,
			errorMsg:    "not a git repository",
		},
		{
			name: "git_directory_is_file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				gitFile := filepath.Join(dir, ".git")
				if err := os.WriteFile(gitFile, []byte("not a directory"), 0644); err != nil {
					t.Fatalf("Failed to create .git file: %v", err)
				}
				return dir
			},
			expectError: true,
			errorMsg:    "not a git repository",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()
			dir := tc.setup(t)

			err := resolver.ValidateDirectory(dir)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestRepositoryResolver_ExtractRemoteInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupGit    func(*testing.T, string)
		expectError bool
		errorMsg    string
		expectRepo  string
	}{
		{
			name: "valid_https_remote",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				configContent := `[remote "origin"]
	url = https://forgejo.example.com/owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name: "valid_ssh_remote",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				configContent := `[remote "origin"]
	url = git@forgejo.example.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name: "no_remote_configured",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				// Empty config file
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(""), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: true,
			errorMsg:    "no configured remotes",
		},
		{
			name: "invalid_remote_url",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				configContent := `[remote "origin"]
	url = invalid-url
	fetch = +refs/heads/*:refs/remotes/origin/*
`
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: true,
			errorMsg:    "failed to parse remote URL",
		},
		{
			name: "non_origin_remote",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				configContent := `[remote "upstream"]
	url = https://forgejo.example.com/owner/repo.git
	fetch = +refs/heads/*:refs/remotes/upstream/*
`
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: false,
			expectRepo:  "owner/repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()
			dir := t.TempDir()
			tc.setupGit(t, dir)

			repo, err := resolver.ExtractRemoteInfo(dir)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				} else if repo != tc.expectRepo {
					t.Errorf("Expected repo '%s', got '%s'", tc.expectRepo, repo)
				}
			}
		})
	}
}

func TestRepositoryResolver_ResolveRepository(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupGit    func(*testing.T, string)
		expectError bool
		errorMsg    string
		expectRepo  *server.RepositoryResolution
	}{
		{
			name: "successful_resolution",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				configContent := `[remote "origin"]
	url = https://forgejo.example.com/owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: false,
			expectRepo: &server.RepositoryResolution{
				Directory:  "",
				Repository: "owner/repo",
				RemoteURL:  "https://forgejo.example.com/owner/repo.git",
				RemoteName: "origin",
			},
		},
		{
			name: "invalid_directory",
			setupGit: func(t *testing.T, dir string) {
				// Use a non-existent path
				dir = "/nonexistent/path/that/does/not/exist"
			},
			expectError: true,
			errorMsg:    "directory does not exist",
		},
		{
			name: "no_remote_configured",
			setupGit: func(t *testing.T, dir string) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("Failed to create .git directory: %v", err)
				}
				// Empty config file
				if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(""), 0644); err != nil {
					t.Fatalf("Failed to create git config: %v", err)
				}
			},
			expectError: true,
			errorMsg:    "no configured remotes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()
			var dir string

			if tc.name == "invalid_directory" {
				dir = "/nonexistent/path/that/does/not/exist"
			} else {
				dir = t.TempDir()
				tc.setupGit(t, dir)
			}

			result, err := resolver.ResolveRepository(dir)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				} else {
					// Update expected directory to match actual
					tc.expectRepo.Directory = dir
					if result.Directory != tc.expectRepo.Directory {
						t.Errorf("Expected directory '%s', got '%s'", tc.expectRepo.Directory, result.Directory)
					}
					if result.Repository != tc.expectRepo.Repository {
						t.Errorf("Expected repository '%s', got '%s'", tc.expectRepo.Repository, result.Repository)
					}
					if result.RemoteURL != tc.expectRepo.RemoteURL {
						t.Errorf("Expected remote URL '%s', got '%s'", tc.expectRepo.RemoteURL, result.RemoteURL)
					}
					if result.RemoteName != tc.expectRepo.RemoteName {
						t.Errorf("Expected remote name '%s', got '%s'", tc.expectRepo.RemoteName, result.RemoteName)
					}
				}
			}
		})
	}
}

func TestRepositoryResolver_ParameterValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		directory   string
		repository  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "both_parameters_empty",
			directory:   "",
			repository:  "",
			expectError: true,
			errorMsg:    "exactly one of directory or repository must be provided",
		},
		{
			name:        "both_parameters_provided",
			directory:   "/some/path",
			repository:  "owner/repo",
			expectError: true,
			errorMsg:    "exactly one of directory or repository must be provided",
		},
		{
			name:        "only_directory_provided",
			directory:   "/some/path",
			repository:  "",
			expectError: false,
		},
		{
			name:        "only_repository_provided",
			directory:   "",
			repository:  "owner/repo",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()

			err := resolver.ValidateParameters(tc.directory, tc.repository)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
