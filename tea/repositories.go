package tea

import (
	"code.gitea.io/sdk/gitea"
)

// RepositoryFilters holds filter parameters for repository operations
type RepositoryFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// Search
	Query string `json:"query,omitempty"`

	// Ownership and access
	OwnerID       int64  `json:"owner_id,omitempty"`
	StarredByUser int64  `json:"starred_by_user,omitempty"`
	Type          string `json:"type,omitempty"`

	// Visibility and status
	IsPrivate  *bool `json:"is_private,omitempty"`
	IsArchived *bool `json:"is_archived,omitempty"`

	// Sorting
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`

	// Additional filters
	ExcludeTemplate bool `json:"exclude_template,omitempty"`
}

func buildRepoListOptions(filters *RepositoryFilters) *gitea.ListReposOptions {
	opts := &gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 30,
		},
	}

	if filters == nil {
		return opts
	}

	if filters.Page > 0 {
		opts.Page = filters.Page
	}
	if filters.PageSize > 0 {
		opts.PageSize = filters.PageSize
	}

	return opts
}

func buildSearchRepoOptions(filters *RepositoryFilters) *gitea.SearchRepoOptions {
	opts := &gitea.SearchRepoOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 30,
		},
	}

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

	// Handle search query
	if filters.Query != "" {
		opts.Keyword = filters.Query
	}

	// Handle owner ID
	if filters.OwnerID > 0 {
		opts.OwnerID = filters.OwnerID
	}

	// Handle starred by user ID
	if filters.StarredByUser > 0 {
		opts.StarredByUserID = filters.StarredByUser
	}

	// Handle private filter
	if filters.IsPrivate != nil {
		opts.IsPrivate = filters.IsPrivate
	}

	// Handle archived filter
	if filters.IsArchived != nil {
		opts.IsArchived = filters.IsArchived
	}

	// Handle exclude template
	opts.ExcludeTemplate = filters.ExcludeTemplate

	// Handle repository type
	if filters.Type != "" {
		opts.Type = gitea.RepoType(filters.Type)
	}

	// Handle sort
	if filters.Sort != "" {
		opts.Sort = filters.Sort
	}

	// Handle order
	if filters.Order != "" {
		opts.Order = filters.Order
	}

	return opts
}
