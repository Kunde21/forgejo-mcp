package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
)

// GiteaClient implements IssueLister using the Gitea SDK
type GiteaClient struct {
	client *gitea.Client
}

// NewGiteaClient creates a new Gitea client
func NewGiteaClient(url, token string) (*GiteaClient, error) {
	client, err := gitea.NewClient(url, gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &GiteaClient{
		client: client,
	}, nil
}

// ListIssues retrieves issues from the specified repository
func (c *GiteaClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
	// Parse repository string (format: "owner/repo")
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}
	owner, repoName := parts[0], parts[1]

	// List issues using Gitea SDK
	opts := gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{
			PageSize: limit,
			Page:     offset/limit + 1, // Gitea uses 1-based pagination
		},
		State: gitea.StateOpen, // Only open issues for now
	}

	giteaIssues, _, err := c.client.ListRepoIssues(owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	// Convert to our Issue struct
	issues := make([]Issue, len(giteaIssues))
	for i, gi := range giteaIssues {
		issues[i] = Issue{
			Number: int(gi.Index),
			Title:  gi.Title,
			State:  string(gi.State),
		}
	}

	return issues, nil
}
