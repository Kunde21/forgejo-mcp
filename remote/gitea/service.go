package gitea

import (
	"context"
)

// Service provides business logic for Gitea operations
type Service struct {
	client GiteaClientInterface
}

// NewService creates a new Gitea service
func NewService(client GiteaClientInterface) *Service {
	return &Service{
		client: client,
	}
}

// ListIssues lists issues for a repository
func (s *Service) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
	// Call the underlying client
	return s.client.ListIssues(ctx, repo, limit, offset)
}

// CreateIssueComment creates a comment on an issue
func (s *Service) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error) {
	// Call the underlying client
	return s.client.CreateIssueComment(ctx, repo, issueNumber, comment)
}
