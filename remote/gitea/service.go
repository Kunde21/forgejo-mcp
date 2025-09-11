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

// EditIssueComment edits an existing comment with validation
func (s *Service) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*IssueComment, error) {
	// Validate repository format
	if err := s.validateRepository(args.Repository); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate issue number
	if err := s.validateIssueNumber(args.IssueNumber); err != nil {
		return nil, fmt.Errorf("issue number validation failed: %w", err)
	}

	// Validate comment ID
	if err := s.validateCommentID(args.CommentID); err != nil {
		return nil, fmt.Errorf("comment ID validation failed: %w", err)
	}

	// Validate new content
	if err := s.validateCommentContent(args.NewContent); err != nil {
		return nil, fmt.Errorf("new content validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.EditIssueComment(ctx, args)
}

// ListPullRequests lists pull requests for a repository with validation
func (s *Service) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
	// Validate repository format
	if err := s.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate pull request options
	if err := s.validatePullRequestOptions(options); err != nil {
		return nil, fmt.Errorf("pull request options validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.ListPullRequests(ctx, repo, options)
}

// ListPullRequestComments lists comments for a pull request with validation
func (s *Service) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
	// Validate repository format
	if err := s.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Validate pull request number
	if err := s.validatePullRequestNumber(pullRequestNumber); err != nil {
		return nil, fmt.Errorf("pull request number validation failed: %w", err)
	}

	// Validate pagination parameters
	if err := s.validatePagination(limit, offset); err != nil {
		return nil, fmt.Errorf("pagination validation failed: %w", err)
	}

	// Call the underlying client
	return s.client.ListPullRequestComments(ctx, repo, pullRequestNumber, limit, offset)
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

// validatePullRequestNumber checks if the pull request number is valid
func (s *Service) validatePullRequestNumber(pullRequestNumber int) error {
	if pullRequestNumber < 1 {
		return fmt.Errorf("pull request number must be positive")
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

// validateCommentID checks if the comment ID is valid
func (s *Service) validateCommentID(commentID int) error {
	if commentID < 1 {
		return fmt.Errorf("comment ID must be positive")
	}
	return nil
}

// validatePullRequestOptions checks if the pull request options are valid
func (s *Service) validatePullRequestOptions(options ListPullRequestsOptions) error {
	// Validate pagination parameters
	if err := s.validatePagination(options.Limit, options.Offset); err != nil {
		return err
	}

	// Validate state parameter
	if err := s.validatePullRequestState(options.State); err != nil {
		return err
	}

	return nil
}

// validatePullRequestState checks if the pull request state is valid
func (s *Service) validatePullRequestState(state string) error {
	if state == "" {
		return fmt.Errorf("state cannot be empty")
	}

	// Valid states are "open", "closed", "all"
	switch state {
	case "open", "closed", "all":
		return nil
	default:
		return fmt.Errorf("state must be one of: open, closed, all")
	}
}
