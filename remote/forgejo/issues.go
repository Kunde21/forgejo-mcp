package forgejo

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/kunde21/forgejo-mcp/remote"
)

// ListIssues retrieves issues from the specified repository
func (c *ForgejoClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]remote.Issue, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	// List issues using Forgejo SDK
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	page := 1
	if limit > 0 {
		page = offset/limit + 1 // Forgejo uses 1-based pagination
	}

	opts := forgejo.ListIssueOption{
		ListOptions: forgejo.ListOptions{
			PageSize: pageSize,
			Page:     page,
		},
		State: forgejo.StateOpen, // Only open issues for now
	}

	forgejoIssues, _, err := c.client.ListRepoIssues(owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	// Convert to our Issue struct
	issues := make([]remote.Issue, len(forgejoIssues))
	for i, gi := range forgejoIssues {
		issues[i] = remote.Issue{
			Number: int(gi.Index),
			Title:  gi.Title,
			State:  string(gi.State),
		}
	}

	return issues, nil
}

// CreateIssueComment creates a comment on an issue
func (c *ForgejoClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*remote.Comment, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	if issueNumber <= 0 {
		return nil, fmt.Errorf("invalid issue number: %d, must be positive", issueNumber)
	}

	if comment == "" {
		return nil, fmt.Errorf("comment cannot be empty")
	}

	// Create comment using Forgejo SDK
	forgejoComment, _, err := c.client.CreateIssueComment(owner, repoName, int64(issueNumber), forgejo.CreateIssueCommentOption{
		Body: comment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create issue comment: %w", err)
	}

	// Convert to our Comment struct
	commentResult := &remote.Comment{
		ID:      int(forgejoComment.ID),
		Content: forgejoComment.Body,
		Author:  "forgejo-user", // TODO: Get actual user from SDK
		Created: "",             // TODO: Get actual created time from SDK
		Updated: "",             // TODO: Get actual updated time from SDK
	}

	return commentResult, nil
}

// ListIssueComments lists comments on an issue
func (c *ForgejoClient) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*remote.IssueCommentList, error) {
	// Check if client is initialized
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Parse repository string (format: "owner/repo")
	owner, repoName, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
	}

	if issueNumber <= 0 {
		return nil, fmt.Errorf("invalid issue number: %d, must be positive", issueNumber)
	}

	// List comments using Forgejo SDK
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

	forgejoComments, _, err := c.client.ListIssueComments(owner, repoName, int64(issueNumber), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issue comments: %w", err)
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

	// Create IssueCommentList with pagination metadata
	// Note: Forgejo SDK doesn't provide total count in ListIssueComments response
	// We return the actual number of comments returned as total
	commentList := &remote.IssueCommentList{
		Comments: comments,
		Total:    len(comments),
		Limit:    limit,
		Offset:   offset,
	}

	return commentList, nil
}

// EditIssueComment edits an existing issue comment
func (c *ForgejoClient) EditIssueComment(ctx context.Context, args remote.EditIssueCommentArgs) (*remote.Comment, error) {
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

	// Edit comment using Forgejo SDK
	opts := forgejo.EditIssueCommentOption{
		Body: args.NewContent,
	}

	forgejoComment, _, err := c.client.EditIssueComment(owner, repoName, int64(args.CommentID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to edit issue comment: %w", err)
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
