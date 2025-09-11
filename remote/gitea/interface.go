package gitea

import (
	"context"
)

// Issue represents a Git repository issue
type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// IssueLister defines the interface for listing issues from a Git repository
type IssueLister interface {
	ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error)
}

// IssueComment represents a comment on a Git repository issue
type IssueComment struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
	Created string `json:"created"`
}

// IssueCommenter defines the interface for creating comments on Git repository issues
type IssueCommenter interface {
	CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error)
}

// IssueCommentList represents a collection of comments with pagination metadata
type IssueCommentList struct {
	Comments []IssueComment `json:"comments"`
	Total    int            `json:"total"`
	Limit    int            `json:"limit"`
	Offset   int            `json:"offset"`
}

// ListIssueCommentsArgs represents the arguments for listing issue comments with validation tags
type ListIssueCommentsArgs struct {
	Repository  string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	IssueNumber int    `json:"issue_number" validate:"required,min=1"`
	Limit       int    `json:"limit" validate:"min=1,max=100"`
	Offset      int    `json:"offset" validate:"min=0"`
}

// IssueCommentLister defines the interface for listing comments from Git repository issues
type IssueCommentLister interface {
	ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error)
}

// EditIssueCommentArgs represents the arguments for editing an issue comment with validation tags
type EditIssueCommentArgs struct {
	Repository  string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	IssueNumber int    `json:"issue_number" validate:"required,min=1"`
	CommentID   int    `json:"comment_id" validate:"required,min=1"`
	NewContent  string `json:"new_content" validate:"required,min=1"`
}

// IssueCommentEditor defines the interface for editing comments on Git repository issues
type IssueCommentEditor interface {
	EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*IssueComment, error)
}

// PullRequestBranch represents a branch reference in a pull request
type PullRequestBranch struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

// PullRequest represents a Git repository pull request
type PullRequest struct {
	ID        int               `json:"id"`
	Number    int               `json:"number"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	State     string            `json:"state"`
	User      string            `json:"user"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Head      PullRequestBranch `json:"head"`
	Base      PullRequestBranch `json:"base"`
}

// ListPullRequestsOptions represents the options for listing pull requests with validation tags
type ListPullRequestsOptions struct {
	State  string `json:"state" validate:"oneof=open closed all"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Offset int    `json:"offset" validate:"min=0"`
}

// PullRequestLister defines the interface for listing pull requests from a Git repository
type PullRequestLister interface {
	ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error)
}

// PullRequestComment represents a comment on a Git repository pull request
type PullRequestComment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	User      string `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PullRequestCommentList represents a collection of pull request comments with pagination metadata
type PullRequestCommentList struct {
	Comments []PullRequestComment `json:"comments"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

// ListPullRequestCommentsArgs represents the arguments for listing pull request comments with validation tags
type ListPullRequestCommentsArgs struct {
	Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	Limit             int    `json:"limit" validate:"min=1,max=100"`
	Offset            int    `json:"offset" validate:"min=0"`
}

// CreatePullRequestCommentArgs represents the arguments for creating a pull request comment
type CreatePullRequestCommentArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	Comment           string `json:"comment"`
}

// PullRequestCommentLister defines the interface for listing comments from Git repository pull requests
type PullRequestCommentLister interface {
	ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error)
}

// PullRequestCommenter defines the interface for creating comments on Git repository pull requests
type PullRequestCommenter interface {
	CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error)
}

// GiteaClientInterface combines IssueLister, IssueCommenter, IssueCommentLister, IssueCommentEditor, PullRequestLister, PullRequestCommentLister, and PullRequestCommenter for complete Gitea operations
type GiteaClientInterface interface {
	IssueLister
	IssueCommenter
	IssueCommentLister
	IssueCommentEditor
	PullRequestLister
	PullRequestCommentLister
	PullRequestCommenter
}
