package gitea

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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

// CreateIssueComment creates a comment on an issue with validation
func (s *Service) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error) {
	// Validate repository format
	if err := s.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate issue number
	if err := s.validateIssueNumber(issueNumber); err != nil {
		return nil, fmt.Errorf("issue number validation failed: %w", err)
	}

	// Validate comment content
	if err := s.validateCommentContent(comment); err != nil {
		return nil, fmt.Errorf("comment content validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.CreateIssueComment(ctx, repo, issueNumber, comment)
}

// ListIssueComments lists comments for an issue with validation
func (s *Service) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
	// Validate repository format
	if err := s.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate issue number
	if err := s.validateIssueNumber(issueNumber); err != nil {
		return nil, fmt.Errorf("issue number validation failed: %w", err)
	}

	// Validate pagination parameters
	if err := s.validatePagination(limit, offset); err != nil {
		return nil, fmt.Errorf("pagination validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.ListIssueComments(ctx, repo, issueNumber, limit, offset)
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

// validateIssueNumber checks if the issue number is valid
func (s *Service) validateIssueNumber(issueNumber int) error {
	if issueNumber < 1 {
		return fmt.Errorf("issue number must be positive")
	}
	return nil
}

// validateCommentContent checks if the comment content is valid
func (s *Service) validateCommentContent(comment string) error {
	if comment == "" {
		return fmt.Errorf("comment content cannot be empty")
	}
	// Trim whitespace and check again
	if len(strings.TrimSpace(comment)) == 0 {
		return fmt.Errorf("comment content cannot be only whitespace")
	}
	return nil
}
