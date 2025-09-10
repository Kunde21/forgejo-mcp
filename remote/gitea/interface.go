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

// GiteaClientInterface combines IssueLister, IssueCommenter, and IssueCommentLister for complete Gitea operations
type GiteaClientInterface interface {
	IssueLister
	IssueCommenter
	IssueCommentLister
}
