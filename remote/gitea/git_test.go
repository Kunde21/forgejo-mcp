package gitea

import (
	"testing"
)

func TestResolveCWDToRepository(t *testing.T) {
	tests := []struct {
		name        string
		cwd         string
		expectError bool
		expectRepo  string
	}{
		{
			name:        "empty CWD",
			cwd:         "",
			expectError: true,
		},
		{
			name:        "valid CWD with git remote",
			cwd:         "/path/to/project",
			expectError: false,
			expectRepo:  "to/project",
		},
		{
			name:        "CWD without git repository",
			cwd:         "/non/git/path",
			expectError: false,
			expectRepo:  "git/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveCWDToRepository(tt.cwd)
			if tt.expectError && err == nil {
				t.Errorf("resolveCWDToRepository(%q) expected error but got none", tt.cwd)
			}
			if !tt.expectError && err != nil {
				t.Errorf("resolveCWDToRepository(%q) unexpected error: %v", tt.cwd, err)
			}
			if !tt.expectError && result != tt.expectRepo {
				t.Errorf("resolveCWDToRepository(%q) = %q, want %q", tt.cwd, result, tt.expectRepo)
			}
		})
	}
}

func TestParseGitRemoteOutput(t *testing.T) {
	tests := []struct {
		name            string
		gitRemoteOutput string
		expectError     bool
		expectRepo      string
	}{
		{
			name:            "empty output",
			gitRemoteOutput: "",
			expectError:     true,
		},
		{
			name:            "no fetch URLs",
			gitRemoteOutput: "origin\thttps://github.com/owner/repo (push)\n",
			expectError:     true,
		},
		{
			name:            "HTTPS URL",
			gitRemoteOutput: "origin\thttps://github.com/owner/repo.git (fetch)\norigin\thttps://github.com/owner/repo.git (push)\n",
			expectError:     false,
			expectRepo:      "owner/repo",
		},
		{
			name:            "SSH URL with git@",
			gitRemoteOutput: "origin\tgit@github.com:owner/repo.git (fetch)\n",
			expectError:     false,
			expectRepo:      "owner/repo",
		},
		{
			name:            "SSH URL with ssh://",
			gitRemoteOutput: "origin\tssh://git@github.com/owner/repo (fetch)\n",
			expectError:     false,
			expectRepo:      "owner/repo",
		},
		{
			name:            "multiple remotes, first valid",
			gitRemoteOutput: "origin\thttps://github.com/owner/repo.git (fetch)\nupstream\thttps://github.com/upstream/repo.git (fetch)\n",
			expectError:     false,
			expectRepo:      "owner/repo",
		},
		{
			name:            "invalid format",
			gitRemoteOutput: "origin\tinvalid-url (fetch)\n",
			expectError:     true,
		},
		{
			name:            "no valid repository identifier",
			gitRemoteOutput: "origin\thttps://github.com/invalid_format (fetch)\n",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitRemoteOutput(tt.gitRemoteOutput)
			if tt.expectError && err == nil {
				t.Errorf("parseGitRemoteOutput(%q) expected error but got none", tt.gitRemoteOutput)
			}
			if !tt.expectError && err != nil {
				t.Errorf("parseGitRemoteOutput(%q) unexpected error: %v", tt.gitRemoteOutput, err)
			}
			if !tt.expectError && result != tt.expectRepo {
				t.Errorf("parseGitRemoteOutput(%q) = %q, want %q", tt.gitRemoteOutput, result, tt.expectRepo)
			}
		})
	}
}

func TestResolveCWDFromPath(t *testing.T) {
	tests := []struct {
		name        string
		cwd         string
		expectError bool
		expectRepo  string
	}{
		{
			name:        "empty CWD",
			cwd:         "",
			expectError: true,
		},
		{
			name:        "simple owner/repo path",
			cwd:         "/home/user/projects/owner/repo",
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name:        "path with trailing slash",
			cwd:         "/home/user/projects/owner/repo/",
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name:        "deep nested path",
			cwd:         "/very/deep/nested/path/owner/repo",
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name:        "path with only one part",
			cwd:         "/single",
			expectError: true,
		},
		{
			name:        "path with empty parts",
			cwd:         "/home//user/projects/owner/repo",
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name:        "path with spaces in names",
			cwd:         "/home/user/my project/owner/repo",
			expectError: false,
			expectRepo:  "owner/repo",
		},
		{
			name:        "path with spaces in later segments",
			cwd:         "/home/user/projects/owner/repo with spaces",
			expectError: false,
			expectRepo:  "projects/owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveCWDFromPath(tt.cwd)
			if tt.expectError && err == nil {
				t.Errorf("resolveCWDFromPath(%q) expected error but got none", tt.cwd)
			}
			if !tt.expectError && err != nil {
				t.Errorf("resolveCWDFromPath(%q) unexpected error: %v", tt.cwd, err)
			}
			if !tt.expectError && result != tt.expectRepo {
				t.Errorf("resolveCWDFromPath(%q) = %q, want %q", tt.cwd, result, tt.expectRepo)
			}
		})
	}
}
