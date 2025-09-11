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
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

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

// CreateIssueComment creates a comment on the specified issue
func (c *GiteaClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// Create comment using Gitea SDK
	opts := gitea.CreateIssueCommentOption{
		Body: comment,
	}

	giteaComment, _, err := c.client.CreateIssueComment(owner, repoName, int64(issueNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue comment: %w", err)
	}

	// Convert to our IssueComment struct
	issueComment := &IssueComment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
	}

	return issueComment, nil
}

// ListIssueComments retrieves comments from the specified issue
func (c *GiteaClient) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// List comments using Gitea SDK
	opts := gitea.ListIssueCommentOptions{
		ListOptions: gitea.ListOptions{
			PageSize: limit,
			Page:     offset/limit + 1, // Gitea uses 1-based pagination
		},
	}

	giteaComments, _, err := c.client.ListIssueComments(owner, repoName, int64(issueNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issue comments: %w", err)
	}

	// Convert to our IssueComment struct
	comments := make([]IssueComment, len(giteaComments))
	for i, gc := range giteaComments {
		comments[i] = IssueComment{
			ID:      int(gc.ID),
			Content: gc.Body,
			Author:  gc.Poster.UserName,
			Created: gc.Created.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Create IssueCommentList with pagination metadata
	// Note: Gitea SDK doesn't provide total count in ListIssueComments response
	// We return the actual number of comments returned as total
	commentList := &IssueCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}

// EditIssueComment edits an existing comment on the specified issue
func (c *GiteaClient) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*IssueComment, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(args.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
	}

	// Edit comment using Gitea SDK
	opts := gitea.EditIssueCommentOption{
		Body: args.NewContent,
	}

	giteaComment, _, err := c.client.EditIssueComment(owner, repoName, int64(args.CommentID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to edit issue comment: %w", err)
	}

	// Convert to our IssueComment struct
	issueComment := &IssueComment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
	}

	return issueComment, nil
}

// ListPullRequests retrieves pull requests from the specified repository
func (c *GiteaClient) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// Convert state to Gitea SDK format
	var state gitea.StateType
	switch options.State {
	case "open":
		state = gitea.StateOpen
	case "closed":
		state = gitea.StateClosed
	case "all":
		state = gitea.StateAll
	default:
		state = gitea.StateOpen // Default to open if invalid state
	}

	// List pull requests using Gitea SDK
	opts := gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{
			PageSize: options.Limit,
			Page:     options.Offset/options.Limit + 1, // Gitea uses 1-based pagination
		},
		State: state,
	}

	giteaPRs, _, err := c.client.ListRepoPullRequests(owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	// Convert to our PullRequest struct
	prs := make([]PullRequest, len(giteaPRs))
	for i, gpr := range giteaPRs {
		prs[i] = PullRequest{
			ID:        int(gpr.ID),
			Number:    int(gpr.Index),
			Title:     gpr.Title,
			Body:      gpr.Body,
			State:     string(gpr.State),
			User:      gpr.Poster.UserName,
			CreatedAt: gpr.Created.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: gpr.Updated.Format("2006-01-02T15:04:05Z"),
			Head: PullRequestBranch{
				Ref: gpr.Head.Ref,
				Sha: gpr.Head.Sha,
			},
			Base: PullRequestBranch{
				Ref: gpr.Base.Ref,
				Sha: gpr.Base.Sha,
			},
		}
	}

	return prs, nil
}

// ListPullRequestComments retrieves comments from the specified pull request
func (c *GiteaClient) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// List comments using Gitea SDK
	opts := gitea.ListIssueCommentOptions{
		ListOptions: gitea.ListOptions{
			PageSize: limit,
			Page:     offset/limit + 1, // Gitea uses 1-based pagination
		},
	}

	giteaComments, _, err := c.client.ListIssueComments(owner, repoName, int64(pullRequestNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull request comments: %w", err)
	}

	// Convert to our PullRequestComment struct
	comments := make([]PullRequestComment, len(giteaComments))
	for i, gc := range giteaComments {
		comments[i] = PullRequestComment{
			ID:        int(gc.ID),
			Body:      gc.Body,
			User:      gc.Poster.UserName,
			CreatedAt: gc.Created.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: gc.Updated.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Create PullRequestCommentList with pagination metadata
	// Note: Gitea SDK doesn't provide total count in ListPullRequestComments response
	// We return the actual number of comments returned as total
	commentList := &PullRequestCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}
