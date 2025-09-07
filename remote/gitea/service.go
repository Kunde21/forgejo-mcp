package gitea

import (
	"context"
	"fmt"
	"regexp"
)

// Service provides business logic for Gitea operations
type Service struct {
	client IssueLister
}

// NewService creates a new Gitea service
func NewService(client IssueLister) *Service {
	return &Service{
		client: client,
	}
}

// ListIssues lists issues for a repository with validation
func (s *Service) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
	// Validate repository format
	if err := s.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate pagination parameters
	if err := s.validatePagination(limit, offset); err != nil {
		return nil, fmt.Errorf("pagination validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.ListIssues(ctx, repo, limit, offset)
}

// validateRepository checks if the repository string is in the correct format
func (s *Service) validateRepository(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	// Basic validation: owner/repo format
	re := regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
	if !re.MatchString(repo) {
		return fmt.Errorf("repository must be in format 'owner/repo'")
	}

	return nil
}

// validatePagination checks pagination parameters
func (s *Service) validatePagination(limit, offset int) error {
	if limit < 1 || limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}
