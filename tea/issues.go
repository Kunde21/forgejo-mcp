package tea

import (
	"time"

	"code.gitea.io/sdk/gitea"
)

// StateType represents the state of an issue or pull request
type StateType string

const (
	StateOpen   StateType = "open"
	StateClosed StateType = "closed"
	StateAll    StateType = "all"
)

func buildIssueListOptions(filters *IssueFilters) *gitea.ListIssueOption {
	opts := &gitea.ListIssueOption{}

	if filters == nil {
		return opts
	}

	// Handle pagination
	if filters.Page > 0 {
		opts.Page = filters.Page
	}
	if filters.PageSize > 0 {
		opts.PageSize = filters.PageSize
	}

	// Handle state and type filters
	if filters.State != "" {
		opts.State = gitea.StateType(filters.State)
	}
	if filters.Type != "" {
		opts.Type = gitea.IssueType(filters.Type)
	}

	// Handle content filters
	if len(filters.Labels) > 0 {
		opts.Labels = filters.Labels
	}
	if len(filters.Milestones) > 0 {
		opts.Milestones = filters.Milestones
	}
	if filters.KeyWord != "" {
		opts.KeyWord = filters.KeyWord
	}

	// Handle time filters
	if filters.Since != nil {
		opts.Since = *filters.Since
	}
	if filters.Before != nil {
		opts.Before = *filters.Before
	}

	// Handle user filters
	if filters.CreatedBy != "" {
		opts.CreatedBy = filters.CreatedBy
	}
	if filters.AssignedBy != "" {
		opts.AssignedBy = filters.AssignedBy
	}
	if filters.MentionedBy != "" {
		opts.MentionedBy = filters.MentionedBy
	}
	if filters.Owner != "" {
		opts.Owner = filters.Owner
	}
	if filters.Team != "" {
		opts.Team = filters.Team
	}

	return opts
}

// IssueFilters holds filter parameters for issue operations
type IssueFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// State and type filters
	State StateType `json:"state,omitempty"`
	Type  string    `json:"type,omitempty"`

	// Content filters
	Labels     []string `json:"labels,omitempty"`
	Milestones []string `json:"milestones,omitempty"`
	KeyWord    string   `json:"keyword,omitempty"`

	// Time filters
	Since  *time.Time `json:"since,omitempty"`
	Before *time.Time `json:"before,omitempty"`

	// User filters
	CreatedBy   string `json:"created_by,omitempty"`
	AssignedBy  string `json:"assigned_by,omitempty"`
	MentionedBy string `json:"mentioned_by,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Team        string `json:"team,omitempty"`
}
