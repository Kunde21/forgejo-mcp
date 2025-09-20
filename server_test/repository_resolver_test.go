package servertest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kunde21/forgejo-mcp/server"
)

// Test cases for repository resolver functionality
type repositoryResolverTestCase struct {
	name       string
	setup      func(*testing.T) string
	wantError  error
	expectRepo *server.RepositoryResolution
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
			wantError: nil,
		},
		{
			name: "nonexistent_directory",
			setup: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			wantError: server.NewDirectoryNotFoundError("/nonexistent/directory"),
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
			wantError: &server.NotGitRepositoryError{},
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
			wantError: &server.NotGitRepositoryError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()
			dir := tc.setup(t)
			err := resolver.ValidateDirectory(dir)

			if !cmp.Equal(tc.wantError, err, cmpopts.EquateErrors()) {
				t.Error(cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()))
			}
		})
	}
}

func TestRepositoryResolver_ExtractRemoteInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		setupGit   func(*testing.T, string)
		wantError  error
		expectRepo string
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
			wantError:  nil,
			expectRepo: "owner/repo",
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
			wantError:  nil,
			expectRepo: "owner/repo",
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
			wantError: &server.NoRemotesConfiguredError{},
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
			wantError: &server.InvalidRemoteURLError{},
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
			wantError:  nil,
			expectRepo: "owner/repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := server.NewRepositoryResolver()
			dir := t.TempDir()
			tc.setupGit(t, dir)

			repo, err := resolver.ExtractRemoteInfo(dir)
			if !cmp.Equal(tc.wantError, err, cmpopts.EquateErrors()) {
				t.Error(cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()))
			}
			if !cmp.Equal(tc.expectRepo, repo) {
				t.Error(cmp.Diff(tc.expectRepo, repo))
			}
		})
	}
}

func TestRepositoryResolver_ResolveRepository(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		setupGit   func(*testing.T, string)
		wantError  error
		expectRepo func(string) *server.RepositoryResolution
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
			wantError: nil,
			expectRepo: func(dir string) *server.RepositoryResolution {
				return &server.RepositoryResolution{
					Directory:  dir,
					Repository: "owner/repo",
					RemoteURL:  "https://forgejo.example.com/owner/repo.git",
					RemoteName: "origin",
				}
			},
		},
		{
			name: "invalid_directory",
			setupGit: func(t *testing.T, dir string) {
				// Use a non-existent path
				dir = "/nonexistent/path/that/does/not/exist"
			},
			wantError: &server.DirectoryNotFoundError{},
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
			wantError: &server.NoRemotesConfiguredError{},
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
			if !cmp.Equal(tc.wantError, err, cmpopts.EquateErrors()) {
				t.Error(cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()))
			}
			var expectedRepo *server.RepositoryResolution
			if tc.expectRepo != nil {
				expectedRepo = tc.expectRepo(dir)
			}
			if !cmp.Equal(expectedRepo, result) {
				t.Error(cmp.Diff(expectedRepo, result))
			}
		})
	}
}
