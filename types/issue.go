package types

import (
	"fmt"
)

// IssueState represents the state of an issue
type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

// Milestone represents a milestone for issues
type Milestone struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	DueDate     Timestamp `json:"dueDate,omitempty"`
	State       string    `json:"state"`
}

// Issue represents an issue in Forgejo
type Issue struct {
	ID           int        `json:"id"`
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Body         string     `json:"body,omitempty"`
	State        IssueState `json:"state"`
	Author       *User      `json:"author,omitempty"`
	Labels       []PRLabel  `json:"labels,omitempty"`
	Assignees    []User     `json:"assignees,omitempty"`
	Milestone    *Milestone `json:"milestone,omitempty"`
	CreatedAt    Timestamp  `json:"createdAt"`
	UpdatedAt    Timestamp  `json:"updatedAt"`
	ClosedAt     *Timestamp `json:"closedAt,omitempty"`
	CommentCount int        `json:"commentCount"`
	URL          string     `json:"url"`
}

// Validate checks if an Issue has required fields
func (i *Issue) Validate() error {
	if i.ID <= 0 {
		return fmt.Errorf("issue ID must be positive")
	}
	if i.Number <= 0 {
		return fmt.Errorf("issue number must be positive")
	}
	if i.Title == "" {
		return fmt.Errorf("issue title cannot be empty")
	}
	if i.URL == "" {
		return fmt.Errorf("issue URL cannot be empty")
	}
	if i.State != IssueStateOpen && i.State != IssueStateClosed {
		return fmt.Errorf("issue state must be 'open' or 'closed'")
	}

	// Validate timestamps
	if i.CreatedAt.Time.IsZero() {
		return fmt.Errorf("issue created time cannot be zero")
	}
	if i.UpdatedAt.Time.IsZero() {
		return fmt.Errorf("issue updated time cannot be zero")
	}

	return nil
}

// HasLabel returns true if the issue has a label with the given name
func (i *Issue) HasLabel(name string) bool {
	for _, label := range i.Labels {
		if label.Name == name {
			return true
		}
	}
	return false
}

// Validate checks if a Milestone has required fields
func (m *Milestone) Validate() error {
	if m.ID <= 0 {
		return fmt.Errorf("milestone ID must be positive")
	}
	if m.Title == "" {
		return fmt.Errorf("milestone title cannot be empty")
	}
	return nil
}
