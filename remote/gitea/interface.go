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
