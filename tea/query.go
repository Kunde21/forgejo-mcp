package tea

import (
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
)

// QueryBuilder provides methods to build complex queries for different resource types
type QueryBuilder struct{}

// NewQueryBuilder creates a new QueryBuilder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// BuildRepositoryQuery builds a query string for repository searches
func (qb *QueryBuilder) BuildRepositoryQuery(filters *RepositoryFilters) string {
	if filters == nil {
		return ""
	}

	// For repositories, the query is primarily handled by the SearchRepoOptions.Keyword field
	return filters.Query
}

// BuildIssueQuery builds a query string for issue searches
func (qb *QueryBuilder) BuildIssueQuery(filters *IssueFilters) string {
	if filters == nil {
		return ""
	}

	var parts []string

	// Add keyword search
	if filters.KeyWord != "" {
		parts = append(parts, filters.KeyWord)
	}

	// Add state filter
	if filters.State != "" {
		parts = append(parts, fmt.Sprintf("state:%s", string(filters.State)))
	}

	// Add labels
	for _, label := range filters.Labels {
		parts = append(parts, fmt.Sprintf("label:%s", label))
	}

	// Add milestones
	for _, milestone := range filters.Milestones {
		parts = append(parts, fmt.Sprintf("milestone:%s", milestone))
	}

	// Add author filter
	if filters.CreatedBy != "" {
		parts = append(parts, fmt.Sprintf("author:%s", filters.CreatedBy))
	}

	// Add assignee filter
	if filters.AssignedBy != "" {
		parts = append(parts, fmt.Sprintf("assignee:%s", filters.AssignedBy))
	}

	// Add mentioned filter
	if filters.MentionedBy != "" {
		parts = append(parts, fmt.Sprintf("mentions:%s", filters.MentionedBy))
	}

	// Add since filter
	if filters.Since != nil {
		parts = append(parts, fmt.Sprintf("updated:>=%s", filters.Since.Format("2006-01-02")))
	}

	// Add before filter
	if filters.Before != nil {
		parts = append(parts, fmt.Sprintf("updated:<=%s", filters.Before.Format("2006-01-02")))
	}

	return strings.Join(parts, " ")
}

// BuildPullRequestQuery builds a query string for pull request searches
func (qb *QueryBuilder) BuildPullRequestQuery(filters *PullRequestFilters) string {
	if filters == nil {
		return ""
	}

	var parts []string

	// Add state filter
	if filters.State != "" {
		parts = append(parts, fmt.Sprintf("state:%s", string(filters.State)))
	}

	// Add milestone filter
	if filters.Milestone > 0 {
		parts = append(parts, fmt.Sprintf("milestone:%d", filters.Milestone))
	}

	return strings.Join(parts, " ")
}

// PaginationHandler provides methods to handle pagination
type PaginationHandler struct{}

// NewPaginationHandler creates a new PaginationHandler instance
func NewPaginationHandler() *PaginationHandler { return &PaginationHandler{} }

func (ph *PaginationHandler) BuildPaginationOptions(page, pageSize int) gitea.ListOptions {
	opts := gitea.ListOptions{Page: 1, PageSize: 30}
	if page > 0 {
		opts.Page = page
	}
	if pageSize > 0 {
		opts.PageSize = pageSize
	}
	return opts
}

type SortHandler struct{}

func NewSortHandler() *SortHandler                                           { return &SortHandler{} }
func (sh *SortHandler) BuildSortOptions(sort, order string) (string, string) { return sort, order }

type CursorPagination struct {
	CurrentPage int
	PageSize    int
	TotalPages  int
	HasNext     bool
	HasPrev     bool
	NextCursor  string
	PrevCursor  string
}

func NewCursorPagination(page, pageSize, totalCount int) *CursorPagination {
	if pageSize <= 0 {
		pageSize = 30
	}
	if page <= 0 {
		page = 1
	}
	totalPages := (totalCount + pageSize - 1) / pageSize
	return &CursorPagination{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextCursor:  strconv.Itoa(page + 1),
		PrevCursor:  strconv.Itoa(page - 1),
	}
}
func (cp *CursorPagination) GetNextPage() int {
	if cp.HasNext {
		return cp.CurrentPage + 1
	}
	return 0
}
func (cp *CursorPagination) GetPrevPage() int { return max(0, cp.CurrentPage-1) }
