package context

import (
	"testing"
)

func TestIsForgejoRemote(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "codeberg.org HTTPS",
			url:      "https://codeberg.org/user/repo.git",
			expected: true,
		},
		{
			name:     "codeberg.org SSH",
			url:      "git@codeberg.org:user/repo.git",
			expected: true,
		},
		{
			name:     "forgejo.org HTTPS",
			url:      "https://forgejo.org/user/repo.git",
			expected: true,
		},
		{
			name:     "sourcehut HTTPS",
			url:      "https://git.sr.ht/~user/repo",
			expected: true,
		},
		{
			name:     "custom forgejo instance",
			url:      "https://git.example.com/user/repo.git",
			expected: true,
		},
		{
			name:     "github.com",
			url:      "https://github.com/user/repo.git",
			expected: false,
		},
		{
			name:     "gitlab.com",
			url:      "https://gitlab.com/user/repo.git",
			expected: false,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: false,
		},
		{
			name:     "empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "localhost",
			url:      "https://localhost/user/repo.git",
			expected: true,
		},
		{
			name:     "IP address",
			url:      "https://192.168.1.100/user/repo.git",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForgejoRemote(tt.url)
			if result != tt.expected {
				t.Errorf("IsForgejoRemote(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedOwner string
		expectedRepo  string
		expectError   bool
	}{
		{
			name:          "HTTPS URL with .git",
			url:           "https://codeberg.org/user/repo.git",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "HTTPS URL without .git",
			url:           "https://codeberg.org/user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "SSH URL with .git",
			url:           "git@codeberg.org:user/repo.git",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "SSH URL without .git",
			url:           "git@codeberg.org:user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "SSH URL with ssh:// protocol",
			url:           "ssh://git@codeberg.org/user/repo.git",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:        "invalid SSH format - no colon",
			url:         "git@codeberg.org/user/repo.git",
			expectError: true,
		},
		{
			name:        "invalid SSH format - too many parts",
			url:         "git@codeberg.org:user:repo.git",
			expectError: true,
		},
		{
			name:        "invalid HTTPS URL",
			url:         "https://invalid-url",
			expectError: true,
		},
		{
			name:        "unsupported protocol",
			url:         "ftp://example.com/user/repo.git",
			expectError: true,
		},
		{
			name:        "empty owner",
			url:         "https://codeberg.org//repo.git",
			expectError: true,
		},
		{
			name:        "empty repo",
			url:         "https://codeberg.org/user/.git",
			expectError: true,
		},
		{
			name:          "subgroup path",
			url:           "https://codeberg.org/group/subgroup/repo.git",
			expectedOwner: "group",
			expectedRepo:  "subgroup",
			expectError:   true, // This should fail as we expect owner/repo format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepository(tt.url)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseRepository(%q) expected error, got none", tt.url)
				}
			} else {
				if err != nil {
					t.Errorf("ParseRepository(%q) unexpected error: %v", tt.url, err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("ParseRepository(%q) owner = %q, expected %q", tt.url, owner, tt.expectedOwner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("ParseRepository(%q) repo = %q, expected %q", tt.url, repo, tt.expectedRepo)
				}
			}
		})
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL",
			url:      "https://codeberg.org/user/repo.git",
			expected: "codeberg.org",
		},
		{
			name:     "SSH URL",
			url:      "git@codeberg.org:user/repo.git",
			expected: "codeberg.org",
		},
		{
			name:     "SSH URL with ssh://",
			url:      "ssh://git@codeberg.org/user/repo.git",
			expected: "codeberg.org",
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHost(tt.url)
			if result != tt.expected {
				t.Errorf("extractHost(%q) = %q, expected %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestIsValidHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{
			name:     "valid domain",
			host:     "codeberg.org",
			expected: true,
		},
		{
			name:     "valid subdomain",
			host:     "git.example.com",
			expected: true,
		},
		{
			name:     "localhost",
			host:     "localhost",
			expected: true,
		},
		{
			name:     "IP address",
			host:     "192.168.1.100",
			expected: true,
		},
		{
			name:     "invalid - no dots",
			host:     "invalidhost",
			expected: false,
		},
		{
			name:     "empty string",
			host:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidHost(tt.host)
			if result != tt.expected {
				t.Errorf("isValidHost(%q) = %v, expected %v", tt.host, result, tt.expected)
			}
		})
	}
}
