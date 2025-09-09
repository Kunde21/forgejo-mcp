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

// GiteaClientInterface combines IssueLister and IssueCommenter for complete Gitea operations
type GiteaClientInterface interface {
	IssueLister
	IssueCommenter
}
