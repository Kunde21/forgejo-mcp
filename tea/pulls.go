package tea

import (
	"code.gitea.io/sdk/gitea"
)

func buildPullRequestListOptions(filters *PullRequestFilters) *gitea.ListPullRequestsOptions {
	opts := &gitea.ListPullRequestsOptions{}

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

	// Handle state filter
	if filters.State != "" {
		opts.State = gitea.StateType(filters.State)
	}

	// Handle sorting
	if filters.Sort != "" {
		opts.Sort = filters.Sort
	}

	// Handle milestone filter
	opts.Milestone = filters.Milestone

	return opts
}

// PullRequestFilters holds filter parameters for pull request operations
type PullRequestFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// State filter
	State StateType `json:"state,omitempty"`

	// Sorting
	Sort string `json:"sort,omitempty"`

	// Milestone filter
	Milestone int64 `json:"milestone,omitempty"`
}
