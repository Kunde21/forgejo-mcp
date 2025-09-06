package gitea

import (
	"fmt"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
)

func TestExtractRepositoryMetadata(t *testing.T) {
	// Create a mock client
	testRepo := &gitea.Repository{
		ID:          123,
		Name:        "repo",
		FullName:    "owner/repo",
		Description: "Test repository",
		Private:     false,
		Fork:        false,
		Archived:    false,
		Stars:       42,
		Forks:       10,
		Size:        1024,
		HTMLURL:     "https://gitea.example.com/owner/repo",
		SSHURL:      "git@gitea.example.com:owner/repo.git",
		CloneURL:    "https://gitea.example.com/owner/repo.git",
		Owner: &gitea.User{
			ID:       456,
			UserName: "owner",
			FullName: "Test Owner",
			Email:    "owner@example.com",
		},
		Created: time.Now(),
		Updated: time.Now(),
	}

	mockClient := &mockGiteaClient{
		getRepoFunc: func(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
			if owner == "owner" && repo == "repo" {
				return testRepo, nil, nil
			}
			return nil, nil, fmt.Errorf("repository not found")
		},
	}

	tests := []struct {
		name        string
		repoParam   string
		expectError bool
		expectKeys  []string
	}{
		{
			name:        "valid repository",
			repoParam:   "owner/repo",
			expectError: false,
			expectKeys:  []string{"id", "name", "fullName", "description", "private", "fork", "archived", "stars", "forks", "size", "url", "sshUrl", "cloneUrl", "owner"},
		},
		{
			name:        "invalid repository format",
			repoParam:   "invalid",
			expectError: true,
		},
		{
			name:        "repository not found",
			repoParam:   "owner/missing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ExtractRepositoryMetadata(mockClient, tt.repoParam)
			if tt.expectError && err == nil {
				t.Errorf("extractRepositoryMetadata(%q) expected error but got none", tt.repoParam)
			}
			if !tt.expectError && err != nil {
				t.Errorf("extractRepositoryMetadata(%q) unexpected error: %v", tt.repoParam, err)
			}
			if !tt.expectError {
				for _, key := range tt.expectKeys {
					if _, exists := metadata[key]; !exists {
						t.Errorf("extractRepositoryMetadata(%q) missing expected key: %s", tt.repoParam, key)
					}
				}
			}
		})
	}
}
