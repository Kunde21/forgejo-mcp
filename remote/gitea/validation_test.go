package gitea

import (
	"testing"
)

func TestValidateRepositoryFormat(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		expectValid bool
		expectError string
	}{
		{
			name:        "valid simple format",
			repoParam:   "owner/repo",
			expectValid: true,
		},
		{
			name:        "valid with numbers",
			repoParam:   "owner123/repo456",
			expectValid: true,
		},
		{
			name:        "valid with underscores",
			repoParam:   "owner_name/repo_name",
			expectValid: true,
		},
		{
			name:        "valid with hyphens",
			repoParam:   "owner-name/repo-name",
			expectValid: true,
		},
		{
			name:        "valid with dots",
			repoParam:   "owner.name/repo.name",
			expectValid: true,
		},
		{
			name:        "valid with special chars",
			repoParam:   "owner@domain.com/repo",
			expectValid: true,
		},
		{
			name:        "empty string",
			repoParam:   "",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "no slash",
			repoParam:   "ownerrepo",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "empty owner",
			repoParam:   "/repo",
			expectValid: false,
			expectError: "invalid repository format: owner cannot be empty",
		},
		{
			name:        "empty repo",
			repoParam:   "owner/",
			expectValid: false,
			expectError: "invalid repository format: repository name cannot be empty",
		},
		{
			name:        "invalid characters",
			repoParam:   "owner/repo!",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "spaces in names",
			repoParam:   "owner name/repo name",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateRepositoryFormat(tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("ValidateRepositoryFormat(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("ValidateRepositoryFormat(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}
