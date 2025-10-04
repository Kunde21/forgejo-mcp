package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/kunde21/forgejo-mcp/remote"
)

// GiteaClient implements IssueLister using the Gitea SDK
type GiteaClient struct {
	client *gitea.Client
}

// NewGiteaClient creates a new Gitea client
func NewGiteaClient(url, token string) (*GiteaClient, error) {
	return NewGiteaClientWithHTTPClient(url, token, nil)
}

// NewGiteaClientWithHTTPClient creates a new Gitea client with a custom HTTP client
func NewGiteaClientWithHTTPClient(url, token string, httpClient *http.Client) (*GiteaClient, error) {
	var client *gitea.Client
	var err error

	if httpClient != nil {
		client, err = gitea.NewClient(url, gitea.SetToken(token), gitea.SetHTTPClient(httpClient))
	} else {
		client, err = gitea.NewClient(url, gitea.SetToken(token))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &GiteaClient{
		client: client,
	}, nil
}

// ListIssues retrieves issues from the specified repository
func (c *GiteaClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]remote.Issue, error) {
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
	issues := make([]remote.Issue, len(giteaIssues))
	for i, gi := range giteaIssues {
		issues[i] = remote.Issue{
			Number: int(gi.Index),
			Title:  gi.Title,
			State:  string(gi.State),
		}
	}

	return issues, nil
}

// CreateIssueComment creates a comment on the specified issue
func (c *GiteaClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*remote.Comment, error) {
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

	// Convert to our Comment struct
	issueComment := &remote.Comment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
		Updated: giteaComment.Updated.Format("2006-01-02T15:04:05Z"),
	}

	return issueComment, nil
}

// ListIssueComments retrieves comments from the specified issue
func (c *GiteaClient) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*remote.IssueCommentList, error) {
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

	// Convert to our Comment struct
	comments := make([]remote.Comment, len(giteaComments))
	for i, gc := range giteaComments {
		comments[i] = remote.Comment{
			ID:      int(gc.ID),
			Content: gc.Body,
			Author:  gc.Poster.UserName,
			Created: gc.Created.Format("2006-01-02T15:04:05Z"),
			Updated: gc.Updated.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Create IssueCommentList with pagination metadata
	// Note: Gitea SDK doesn't provide total count in ListIssueComments response
	// We return the actual number of comments returned as total
	commentList := &remote.IssueCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}

// EditIssueComment edits an existing comment on the specified issue
func (c *GiteaClient) EditIssueComment(ctx context.Context, args remote.EditIssueCommentArgs) (*remote.Comment, error) {
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

	// Convert to our Comment struct
	issueComment := &remote.Comment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
		Updated: giteaComment.Updated.Format("2006-01-02T15:04:05Z"),
	}

	return issueComment, nil
}

// ListPullRequests retrieves pull requests from the specified repository
func (c *GiteaClient) ListPullRequests(ctx context.Context, repo string, options remote.ListPullRequestsOptions) ([]remote.PullRequest, error) {
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
	prs := make([]remote.PullRequest, len(giteaPRs))
	for i, gpr := range giteaPRs {
		prs[i] = remote.PullRequest{
			ID:        int(gpr.ID),
			Number:    int(gpr.Index),
			Title:     gpr.Title,
			Body:      gpr.Body,
			State:     string(gpr.State),
			User:      gpr.Poster.UserName,
			CreatedAt: gpr.Created.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: gpr.Updated.Format("2006-01-02T15:04:05Z"),
			Head: remote.PullRequestBranch{
				Ref: gpr.Head.Ref,
				Sha: gpr.Head.Sha,
			},
			Base: remote.PullRequestBranch{
				Ref: gpr.Base.Ref,
				Sha: gpr.Base.Sha,
			},
		}
	}

	return prs, nil
}

// ListPullRequestComments retrieves comments from the specified pull request
func (c *GiteaClient) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*remote.PullRequestCommentList, error) {
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

	// Convert to our Comment struct
	comments := make([]remote.Comment, len(giteaComments))
	for i, gc := range giteaComments {
		comments[i] = remote.Comment{
			ID:      int(gc.ID),
			Content: gc.Body,
			Author:  gc.Poster.UserName,
			Created: gc.Created.Format("2006-01-02T15:04:05Z"),
			Updated: gc.Updated.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Create PullRequestCommentList with pagination metadata
	// Note: Gitea SDK doesn't provide total count in ListPullRequestComments response
	// We return the actual number of comments returned as total
	commentList := &remote.PullRequestCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}

// CreatePullRequestComment creates a comment on the specified pull request
func (c *GiteaClient) CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*remote.Comment, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// Create comment using Gitea SDK
	opts := gitea.CreateIssueCommentOption{
		Body: comment,
	}

	giteaComment, _, err := c.client.CreateIssueComment(owner, repoName, int64(pullRequestNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request comment: %w", err)
	}

	// Convert to our Comment struct
	prComment := &remote.Comment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
		Updated: giteaComment.Updated.Format("2006-01-02T15:04:05Z"),
	}

	return prComment, nil
}

// EditPullRequestComment edits an existing comment on the specified pull request
func (c *GiteaClient) EditPullRequestComment(ctx context.Context, args remote.EditPullRequestCommentArgs) (*remote.Comment, error) {
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
		return nil, fmt.Errorf("failed to edit pull request comment: %w", err)
	}

	// Convert to our Comment struct
	prComment := &remote.Comment{
		ID:      int(giteaComment.ID),
		Content: giteaComment.Body,
		Author:  giteaComment.Poster.UserName,
		Created: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
		Updated: giteaComment.Updated.Format("2006-01-02T15:04:05Z"),
	}

	return prComment, nil
}

// EditPullRequest edits an existing pull request
func (c *GiteaClient) EditPullRequest(ctx context.Context, args remote.EditPullRequestArgs) (*remote.PullRequest, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(args.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
	}

	if args.PullRequestNumber <= 0 {
		return nil, fmt.Errorf("invalid pull request number: %d, must be positive", args.PullRequestNumber)
	}

	// Prepare edit options - only include fields that are provided
	var editOptions gitea.EditPullRequestOption
	hasChanges := false

	if args.Title != "" {
		editOptions.Title = args.Title
		hasChanges = true
	}

	if args.Body != "" {
		editOptions.Body = &args.Body
		hasChanges = true
	}

	if args.State != "" {
		// Convert state to Gitea SDK format
		var state gitea.StateType
		switch args.State {
		case "open":
			state = gitea.StateOpen
		case "closed":
			state = gitea.StateClosed
		default:
			return nil, fmt.Errorf("invalid state: %s, must be 'open' or 'closed'", args.State)
		}
		editOptions.State = &state
		hasChanges = true
	}

	if args.BaseBranch != "" {
		editOptions.Base = args.BaseBranch
		hasChanges = true
	}

	if !hasChanges {
		return nil, fmt.Errorf("no changes specified for pull request edit")
	}

	// Edit pull request using Gitea SDK
	giteaPR, _, err := c.client.EditPullRequest(owner, repoName, int64(args.PullRequestNumber), editOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to edit pull request: %w", err)
	}

	// Convert to our PullRequest struct
	user := "unknown"
	if giteaPR.Poster != nil {
		user = giteaPR.Poster.UserName
	}

	createdAt := ""
	if !giteaPR.Created.IsZero() {
		createdAt = giteaPR.Created.Format("2006-01-02T15:04:05Z")
	}

	updatedAt := ""
	if !giteaPR.Updated.IsZero() {
		updatedAt = giteaPR.Updated.Format("2006-01-02T15:04:05Z")
	}

	// Convert head branch
	var head remote.PullRequestBranch
	if giteaPR.Head != nil {
		head = remote.PullRequestBranch{
			Ref: giteaPR.Head.Ref,
			Sha: giteaPR.Head.Sha,
		}
	}

	// Convert base branch
	var base remote.PullRequestBranch
	if giteaPR.Base != nil {
		base = remote.PullRequestBranch{
			Ref: giteaPR.Base.Ref,
			Sha: giteaPR.Base.Sha,
		}
	}

	pr := &remote.PullRequest{
		ID:        int(giteaPR.ID),
		Number:    int(giteaPR.Index),
		Title:     giteaPR.Title,
		Body:      giteaPR.Body,
		State:     string(giteaPR.State),
		User:      user,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Head:      head,
		Base:      base,
	}

	return pr, nil
}
