package types

import (
	"fmt"
)

// PRState represents the state of a pull request
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateMerged PRState = "merged"
)

// PRAuthor represents the author of a pull request
type PRAuthor struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
	URL       string `json:"url"`
}

// PRLabel represents a label on a pull request
type PRLabel struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// PullRequest represents a pull request in Forgejo
type PullRequest struct {
	ID         int        `json:"id"`
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	Body       string     `json:"body,omitempty"`
	State      PRState    `json:"state"`
	Author     *PRAuthor  `json:"author,omitempty"`
	HeadBranch string     `json:"headBranch"`
	BaseBranch string     `json:"baseBranch"`
	CreatedAt  Timestamp  `json:"createdAt"`
	UpdatedAt  Timestamp  `json:"updatedAt"`
	ClosedAt   *Timestamp `json:"closedAt,omitempty"`
	MergedAt   *Timestamp `json:"mergedAt,omitempty"`
	Draft      bool       `json:"draft"`
	Labels     []PRLabel  `json:"labels,omitempty"`
	Assignees  []PRAuthor `json:"assignees,omitempty"`
	Reviewers  []PRAuthor `json:"reviewers,omitempty"`
	URL        string     `json:"url"`
	DiffURL    string     `json:"diffUrl"`
}

// Validate checks if a PullRequest has required fields
func (pr *PullRequest) Validate() error {
	if pr.ID <= 0 {
		return fmt.Errorf("pull request ID must be positive")
	}
	if pr.Number <= 0 {
		return fmt.Errorf("pull request number must be positive")
	}
	if pr.Title == "" {
		return fmt.Errorf("pull request title cannot be empty")
	}
	if pr.HeadBranch == "" {
		return fmt.Errorf("pull request head branch cannot be empty")
	}
	if pr.BaseBranch == "" {
		return fmt.Errorf("pull request base branch cannot be empty")
	}
	if pr.URL == "" {
		return fmt.Errorf("pull request URL cannot be empty")
	}
	if pr.DiffURL == "" {
		return fmt.Errorf("pull request diff URL cannot be empty")
	}
	if pr.State != PRStateOpen && pr.State != PRStateClosed && pr.State != PRStateMerged {
		return fmt.Errorf("pull request state must be 'open', 'closed', or 'merged'")
	}

	// Validate timestamps
	if pr.CreatedAt.Time.IsZero() {
		return fmt.Errorf("pull request created time cannot be zero")
	}
	if pr.UpdatedAt.Time.IsZero() {
		return fmt.Errorf("pull request updated time cannot be zero")
	}

	return nil
}

// IsOpen returns true if the pull request is open
func (pr *PullRequest) IsOpen() bool {
	return pr.State == PRStateOpen
}

// IsClosed returns true if the pull request is closed
func (pr *PullRequest) IsClosed() bool {
	return pr.State == PRStateClosed
}

// IsMerged returns true if the pull request is merged
func (pr *PullRequest) IsMerged() bool {
	return pr.State == PRStateMerged
}

// Validate checks if a PRAuthor has required fields
func (a *PRAuthor) Validate() error {
	if a.Username == "" {
		return fmt.Errorf("author username cannot be empty")
	}
	return nil
}

// Validate checks if a PRLabel has required fields
func (l *PRLabel) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("label name cannot be empty")
	}
	return nil
}
