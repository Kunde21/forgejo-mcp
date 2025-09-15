package remote

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

// Comment represents a comment on a Git repository issue or pull request
type Comment struct {
	ID      int    `json:"id"`
	Content string `json:"body"`
	Author  string `json:"user"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// IssueCommenter defines the interface for creating comments on Git repository issues
type IssueCommenter interface {
	CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*Comment, error)
}

// IssueCommentList represents a collection of comments with pagination metadata
type IssueCommentList struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// ListIssueCommentsArgs represents the arguments for listing issue comments
type ListIssueCommentsArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// IssueCommentLister defines the interface for listing comments from Git repository issues
type IssueCommentLister interface {
	ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error)
}

// EditIssueCommentArgs represents the arguments for editing an issue comment
type EditIssueCommentArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	CommentID   int    `json:"comment_id"`
	NewContent  string `json:"new_content"`
}

// IssueCommentEditor defines the interface for editing comments on Git repository issues
type IssueCommentEditor interface {
	EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*Comment, error)
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
	CreatedAt string            `json:"created"`
	UpdatedAt string            `json:"updated"`
	Head      PullRequestBranch `json:"head"`
	Base      PullRequestBranch `json:"base"`
}

// ListPullRequestsOptions represents the options for listing pull requests
type ListPullRequestsOptions struct {
	State  string `json:"state"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// PullRequestLister defines the interface for listing pull requests from a Git repository
type PullRequestLister interface {
	ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error)
}

// PullRequestCommentList represents a collection of pull request comments with pagination metadata
type PullRequestCommentList struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// ListPullRequestCommentsArgs represents the arguments for listing pull request comments
type ListPullRequestCommentsArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	Limit             int    `json:"limit"`
	Offset            int    `json:"offset"`
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
	CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*Comment, error)
}

// EditPullRequestCommentArgs represents the arguments for editing a pull request comment
type EditPullRequestCommentArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	CommentID         int    `json:"comment_id"`
	NewContent        string `json:"new_content"`
}

// PullRequestCommentEditor defines the interface for editing comments on Git repository pull requests
type PullRequestCommentEditor interface {
	EditPullRequestComment(ctx context.Context, args EditPullRequestCommentArgs) (*Comment, error)
}

// ClientInterface combines IssueLister, IssueCommenter, IssueCommentLister, IssueCommentEditor, PullRequestLister, PullRequestCommentLister, PullRequestCommenter, and PullRequestCommentEditor for complete Git operations
type ClientInterface interface {
	IssueLister
	IssueCommenter
	IssueCommentLister
	IssueCommentEditor
	PullRequestLister
	PullRequestCommentLister
	PullRequestCommenter
	PullRequestCommentEditor
}
