package forgejo

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/kunde21/forgejo-mcp/remote"
)

// ListPullRequests lists pull requests from a repository
func (c *ForgejoClient) ListPullRequests(ctx context.Context, repo string, options remote.ListPullRequestsOptions) ([]remote.PullRequest, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// Convert state to Forgejo SDK format
	var state forgejo.StateType
	switch options.State {
	case "open":
		state = forgejo.StateOpen
	case "closed":
		state = forgejo.StateClosed
	case "all":
		state = forgejo.StateAll
	default:
		state = forgejo.StateOpen // Default to open if invalid state
	}

	// List pull requests using Forgejo SDK
	pageSize := options.Limit
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	page := 1
	if options.Limit > 0 {
		page = options.Offset/options.Limit + 1 // Forgejo uses 1-based pagination
	}

	opts := forgejo.ListPullRequestsOptions{
		ListOptions: forgejo.ListOptions{
			PageSize: pageSize,
			Page:     page,
		},
		State: state,
	}

	forgejoPRs, _, err := c.client.ListRepoPullRequests(owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	// Convert to our PullRequest struct
	prs := make([]remote.PullRequest, len(forgejoPRs))
	for i, fpr := range forgejoPRs {
		user := "unknown"
		if fpr.Poster != nil {
			user = fpr.Poster.UserName
		}

		createdAt := ""
		if fpr.Created != nil {
			createdAt = fpr.Created.Format("2006-01-02T15:04:05Z")
		}

		updatedAt := ""
		if fpr.Updated != nil {
			updatedAt = fpr.Updated.Format("2006-01-02T15:04:05Z")
		}

		// Convert head branch
		var head remote.PullRequestBranch
		if fpr.Head != nil {
			head = remote.PullRequestBranch{
				Ref: fpr.Head.Ref,
				Sha: fpr.Head.Sha,
			}
		}

		// Convert base branch
		var base remote.PullRequestBranch
		if fpr.Base != nil {
			base = remote.PullRequestBranch{
				Ref: fpr.Base.Ref,
				Sha: fpr.Base.Sha,
			}
		}

		prs[i] = remote.PullRequest{
			ID:        int(fpr.ID),
			Number:    int(fpr.Index),
			Title:     fpr.Title,
			Body:      fpr.Body,
			State:     string(fpr.State),
			User:      user,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Head:      head,
			Base:      base,
		}
	}

	return prs, nil
}

// ListPullRequestComments lists comments on a pull request
func (c *ForgejoClient) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*remote.PullRequestCommentList, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	if pullRequestNumber <= 0 {
		return nil, fmt.Errorf("invalid pull request number: %d, must be positive", pullRequestNumber)
	}

	// List comments using Forgejo SDK (same method as for issues)
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	page := 1
	if limit > 0 {
		page = offset/limit + 1 // Forgejo uses 1-based pagination
	}

	opts := forgejo.ListIssueCommentOptions{
		ListOptions: forgejo.ListOptions{
			PageSize: pageSize,
			Page:     page,
		},
	}

	forgejoComments, _, err := c.client.ListIssueComments(owner, repoName, int64(pullRequestNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull request comments: %w", err)
	}

	// Convert to our Comment struct
	comments := make([]remote.Comment, len(forgejoComments))
	for i, fc := range forgejoComments {
		author := "unknown"
		if fc.Poster != nil {
			author = fc.Poster.UserName
		}

		created := ""
		if !fc.Created.IsZero() {
			created = fc.Created.Format("2006-01-02T15:04:05Z")
		}

		updated := ""
		if !fc.Updated.IsZero() {
			updated = fc.Updated.Format("2006-01-02T15:04:05Z")
		}

		comments[i] = remote.Comment{
			ID:      int(fc.ID),
			Content: fc.Body,
			Author:  author,
			Created: created,
			Updated: updated,
		}
	}

	// Create PullRequestCommentList with pagination metadata
	// Note: Forgejo SDK doesn't provide total count in ListIssueComments response
	// We return the actual number of comments returned as total
	commentList := &remote.PullRequestCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}

// CreatePullRequestComment creates a comment on a pull request
func (c *ForgejoClient) CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*remote.Comment, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	if pullRequestNumber <= 0 {
		return nil, fmt.Errorf("invalid pull request number: %d, must be positive", pullRequestNumber)
	}

	if comment == "" {
		return nil, fmt.Errorf("comment cannot be empty")
	}

	// Create comment using Forgejo SDK (same method as for issues)
	forgejoComment, _, err := c.client.CreateIssueComment(owner, repoName, int64(pullRequestNumber), forgejo.CreateIssueCommentOption{
		Body: comment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request comment: %w", err)
	}

	// Convert to our Comment struct
	author := "unknown"
	if forgejoComment.Poster != nil {
		author = forgejoComment.Poster.UserName
	}

	created := ""
	if !forgejoComment.Created.IsZero() {
		created = forgejoComment.Created.Format("2006-01-02T15:04:05Z")
	}

	updated := ""
	if !forgejoComment.Updated.IsZero() {
		updated = forgejoComment.Updated.Format("2006-01-02T15:04:05Z")
	}

	commentResult := &remote.Comment{
		ID:      int(forgejoComment.ID),
		Content: forgejoComment.Body,
		Author:  author,
		Created: created,
		Updated: updated,
	}

	return commentResult, nil
}

// EditPullRequestComment edits an existing pull request comment
func (c *ForgejoClient) EditPullRequestComment(ctx context.Context, args remote.EditPullRequestCommentArgs) (*remote.Comment, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(args.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
	}

	if args.CommentID <= 0 {
		return nil, fmt.Errorf("invalid comment ID: %d, must be positive", args.CommentID)
	}

	if args.NewContent == "" {
		return nil, fmt.Errorf("new content cannot be empty")
	}

	// Edit comment using Forgejo SDK (same method as for issues)
	opts := forgejo.EditIssueCommentOption{
		Body: args.NewContent,
	}

	forgejoComment, _, err := c.client.EditIssueComment(owner, repoName, int64(args.CommentID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to edit pull request comment: %w", err)
	}

	// Convert to our Comment struct
	author := "unknown"
	if forgejoComment.Poster != nil {
		author = forgejoComment.Poster.UserName
	}

	created := ""
	if !forgejoComment.Created.IsZero() {
		created = forgejoComment.Created.Format("2006-01-02T15:04:05Z")
	}

	updated := ""
	if !forgejoComment.Updated.IsZero() {
		updated = forgejoComment.Updated.Format("2006-01-02T15:04:05Z")
	}

	comment := &remote.Comment{
		ID:      int(forgejoComment.ID),
		Content: forgejoComment.Body,
		Author:  author,
		Created: created,
		Updated: updated,
	}

	return comment, nil
}

// EditPullRequest edits an existing pull request
func (c *ForgejoClient) EditPullRequest(ctx context.Context, args remote.EditPullRequestArgs) (*remote.PullRequest, error) {
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
	var editOptions forgejo.EditPullRequestOption
	hasChanges := false

	if args.Title != "" {
		editOptions.Title = args.Title
		hasChanges = true
	}

	if args.Body != "" {
		editOptions.Body = args.Body
		hasChanges = true
	}

	if args.State != "" {
		// Convert state to Forgejo SDK format
		var state forgejo.StateType
		switch args.State {
		case "open":
			state = forgejo.StateOpen
		case "closed":
			state = forgejo.StateClosed
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

	// Edit pull request using Forgejo SDK
	forgejoPR, _, err := c.client.EditPullRequest(owner, repoName, int64(args.PullRequestNumber), editOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to edit pull request: %w", err)
	}

	// Convert to our PullRequest struct
	user := "unknown"
	if forgejoPR.Poster != nil {
		user = forgejoPR.Poster.UserName
	}

	createdAt := ""
	if forgejoPR.Created != nil {
		createdAt = forgejoPR.Created.Format("2006-01-02T15:04:05Z")
	}

	updatedAt := ""
	if forgejoPR.Updated != nil {
		updatedAt = forgejoPR.Updated.Format("2006-01-02T15:04:05Z")
	}

	// Convert head branch
	var head remote.PullRequestBranch
	if forgejoPR.Head != nil {
		head = remote.PullRequestBranch{
			Ref: forgejoPR.Head.Ref,
			Sha: forgejoPR.Head.Sha,
		}
	}

	// Convert base branch
	var base remote.PullRequestBranch
	if forgejoPR.Base != nil {
		base = remote.PullRequestBranch{
			Ref: forgejoPR.Base.Ref,
			Sha: forgejoPR.Base.Sha,
		}
	}

	return &remote.PullRequest{
		ID:        int(forgejoPR.ID),
		Number:    int(forgejoPR.Index),
		Title:     forgejoPR.Title,
		Body:      forgejoPR.Body,
		State:     string(forgejoPR.State),
		User:      user,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Head:      head,
		Base:      base,
	}, nil
}

// CreatePullRequest creates a new pull request in the repository
func (c *ForgejoClient) CreatePullRequest(ctx context.Context, args remote.CreatePullRequestArgs) (*remote.PullRequest, error) {
	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(args.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
	}

	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Handle draft PRs with title prefix since SDK lacks draft field
	title := args.Title
	if args.Draft {
		title = "[DRAFT] " + title
	}

	opts := forgejo.CreatePullRequestOption{
		Head:     args.Head,
		Base:     args.Base,
		Title:    title,
		Body:     args.Body,
		Assignee: args.Assignee,
	}

	fpr, _, err := c.client.CreatePullRequest(owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Transform Forgejo SDK response to remote interface format
	user := "unknown"
	if fpr.Poster != nil {
		user = fpr.Poster.UserName
	}

	createdAt := ""
	if fpr.Created != nil {
		createdAt = fpr.Created.Format("2006-01-02T15:04:05Z")
	}

	updatedAt := ""
	if fpr.Updated != nil {
		updatedAt = fpr.Updated.Format("2006-01-02T15:04:05Z")
	}

	// Convert head branch
	var head remote.PullRequestBranch
	if fpr.Head != nil {
		head = remote.PullRequestBranch{
			Ref: fpr.Head.Ref,
			Sha: fpr.Head.Sha,
		}
	}

	// Convert base branch
	var base remote.PullRequestBranch
	if fpr.Base != nil {
		base = remote.PullRequestBranch{
			Ref: fpr.Base.Ref,
			Sha: fpr.Base.Sha,
		}
	}

	return &remote.PullRequest{
		ID:        int(fpr.ID),
		Number:    int(fpr.Index),
		Title:     fpr.Title,
		Body:      fpr.Body,
		State:     string(fpr.State),
		User:      user,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Head:      head,
		Base:      base,
	}, nil
}

// GetFileContent fetches file content from repository
func (c *ForgejoClient) GetFileContent(ctx context.Context, owner, repo, ref, filepath string) ([]byte, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	content, _, err := c.client.GetFile(owner, repo, ref, filepath)
	return content, err
}
