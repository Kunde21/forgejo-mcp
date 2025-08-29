package tea

import (
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestQueryBuilder_BuildRepositoryQuery(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		filters *RepositoryFilters
		want    string
	}{
		{
			name:    "no filters",
			filters: &RepositoryFilters{},
			want:    "",
		},
		{
			name: "with search query",
			filters: &RepositoryFilters{
				Query: "test-repo",
			},
			want: "test-repo",
		},
		{
			name: "with private filter",
			filters: &RepositoryFilters{
				IsPrivate: &trueVal,
			},
			want: "",
		},
		{
			name: "with archived filter",
			filters: &RepositoryFilters{
				IsArchived: &falseVal,
			},
			want: "",
		},
		{
			name: "with multiple filters",
			filters: &RepositoryFilters{
				Query:      "test-repo",
				IsPrivate:  &trueVal,
				IsArchived: &falseVal,
			},
			want: "test-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder()
			result := qb.BuildRepositoryQuery(tt.filters)
			if result != tt.want {
				t.Errorf("BuildRepositoryQuery() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestQueryBuilder_BuildIssueQuery(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		filters *IssueFilters
		want    string
	}{
		{
			name:    "no filters",
			filters: &IssueFilters{},
			want:    "",
		},
		{
			name: "with keyword",
			filters: &IssueFilters{
				KeyWord: "bug",
			},
			want: "bug",
		},
		{
			name: "with labels",
			filters: &IssueFilters{
				Labels: []string{"bug", "urgent"},
			},
			want: "label:bug label:urgent",
		},
		{
			name: "with author",
			filters: &IssueFilters{
				CreatedBy: "john",
			},
			want: "author:john",
		},
		{
			name: "with assignee",
			filters: &IssueFilters{
				AssignedBy: "jane",
			},
			want: "assignee:jane",
		},
		{
			name: "with milestone",
			filters: &IssueFilters{
				Milestones: []string{"v1.0"},
			},
			want: "milestone:v1.0",
		},
		{
			name: "with state",
			filters: &IssueFilters{
				State: StateOpen,
			},
			want: "state:open",
		},
		{
			name: "with since time",
			filters: &IssueFilters{
				Since: &now,
			},
			want: "updated:>=" + now.Format(time.RFC3339)[:10],
		},
		{
			name: "with before time",
			filters: &IssueFilters{
				Before: &now,
			},
			want: "updated:<=" + now.Format(time.RFC3339)[:10],
		},
		{
			name: "with multiple filters",
			filters: &IssueFilters{
				KeyWord:    "bug",
				CreatedBy:  "john",
				AssignedBy: "jane",
				State:      StateOpen,
			},
			want: "bug state:open author:john assignee:jane",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder()
			result := qb.BuildIssueQuery(tt.filters)
			if result != tt.want {
				t.Errorf("BuildIssueQuery() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestQueryBuilder_BuildPullRequestQuery(t *testing.T) {
	tests := []struct {
		name    string
		filters *PullRequestFilters
		want    string
	}{
		{
			name:    "no filters",
			filters: &PullRequestFilters{},
			want:    "",
		},
		{
			name: "with state",
			filters: &PullRequestFilters{
				State: StateOpen,
			},
			want: "state:open",
		},
		{
			name: "with milestone",
			filters: &PullRequestFilters{
				Milestone: 123,
			},
			want: "milestone:123",
		},
		{
			name: "with multiple filters",
			filters: &PullRequestFilters{
				State:     StateClosed,
				Milestone: 123,
			},
			want: "state:closed milestone:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder()
			result := qb.BuildPullRequestQuery(tt.filters)
			if result != tt.want {
				t.Errorf("BuildPullRequestQuery() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestPaginationHandler_BuildPaginationOptions(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     gitea.ListOptions
	}{
		{
			name:     "default pagination",
			page:     0,
			pageSize: 0,
			want: gitea.ListOptions{
				Page:     1,
				PageSize: 30,
			},
		},
		{
			name:     "custom pagination",
			page:     2,
			pageSize: 50,
			want: gitea.ListOptions{
				Page:     2,
				PageSize: 50,
			},
		},
		{
			name:     "only page specified",
			page:     3,
			pageSize: 0,
			want: gitea.ListOptions{
				Page:     3,
				PageSize: 30,
			},
		},
		{
			name:     "only page size specified",
			page:     0,
			pageSize: 100,
			want: gitea.ListOptions{
				Page:     1,
				PageSize: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := NewPaginationHandler()
			result := ph.BuildPaginationOptions(tt.page, tt.pageSize)
			if !cmp.Equal(tt.want, result) {
				t.Error(cmp.Diff(tt.want, result))
			}
		})
	}
}

func TestSortHandler_BuildSortOptions(t *testing.T) {
	tests := []struct {
		name  string
		sort  string
		order string
		want  struct {
			Sort  string
			Order string
		}
	}{
		{
			name:  "no sort options",
			sort:  "",
			order: "",
			want: struct {
				Sort  string
				Order string
			}{
				Sort:  "",
				Order: "",
			},
		},
		{
			name:  "sort by created",
			sort:  "created",
			order: "asc",
			want: struct {
				Sort  string
				Order string
			}{
				Sort:  "created",
				Order: "asc",
			},
		},
		{
			name:  "sort by updated with desc",
			sort:  "updated",
			order: "desc",
			want: struct {
				Sort  string
				Order string
			}{
				Sort:  "updated",
				Order: "desc",
			},
		},
		{
			name:  "sort without order",
			sort:  "name",
			order: "",
			want: struct {
				Sort  string
				Order string
			}{
				Sort:  "name",
				Order: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSortHandler()
			sort, order := sh.BuildSortOptions(tt.sort, tt.order)
			if sort != tt.want.Sort || order != tt.want.Order {
				t.Errorf("BuildSortOptions() = (%v, %v), want (%v, %v)", sort, order, tt.want.Sort, tt.want.Order)
			}
		})
	}
}

func TestCursorPagination(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		pageSize    int
		totalCount  int
		want        *CursorPagination
		testGetNext bool
		testGetPrev bool
	}{
		{
			name:       "first page with next",
			page:       1,
			pageSize:   10,
			totalCount: 50,
			want: &CursorPagination{
				CurrentPage: 1,
				PageSize:    10,
				TotalPages:  5,
				HasNext:     true,
				HasPrev:     false,
				NextCursor:  "2",
				PrevCursor:  "0",
			},
			testGetNext: true,
			testGetPrev: true,
		},
		{
			name:       "middle page",
			page:       3,
			pageSize:   10,
			totalCount: 50,
			want: &CursorPagination{
				CurrentPage: 3,
				PageSize:    10,
				TotalPages:  5,
				HasNext:     true,
				HasPrev:     true,
				NextCursor:  "4",
				PrevCursor:  "2",
			},
			testGetNext: true,
			testGetPrev: true,
		},
		{
			name:       "last page",
			page:       5,
			pageSize:   10,
			totalCount: 50,
			want: &CursorPagination{
				CurrentPage: 5,
				PageSize:    10,
				TotalPages:  5,
				HasNext:     false,
				HasPrev:     true,
				NextCursor:  "6",
				PrevCursor:  "4",
			},
			testGetNext: true,
			testGetPrev: true,
		},
		{
			name:       "default values",
			page:       0,
			pageSize:   0,
			totalCount: 50,
			want: &CursorPagination{
				CurrentPage: 1,
				PageSize:    30,
				TotalPages:  2,
				HasNext:     true,
				HasPrev:     false,
				NextCursor:  "2",
				PrevCursor:  "0",
			},
			testGetNext: true,
			testGetPrev: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewCursorPagination(tt.page, tt.pageSize, tt.totalCount)

			if cp.CurrentPage != tt.want.CurrentPage {
				t.Errorf("CurrentPage = %v, want %v", cp.CurrentPage, tt.want.CurrentPage)
			}
			if cp.PageSize != tt.want.PageSize {
				t.Errorf("PageSize = %v, want %v", cp.PageSize, tt.want.PageSize)
			}
			if cp.TotalPages != tt.want.TotalPages {
				t.Errorf("TotalPages = %v, want %v", cp.TotalPages, tt.want.TotalPages)
			}
			if cp.HasNext != tt.want.HasNext {
				t.Errorf("HasNext = %v, want %v", cp.HasNext, tt.want.HasNext)
			}
			if cp.HasPrev != tt.want.HasPrev {
				t.Errorf("HasPrev = %v, want %v", cp.HasPrev, tt.want.HasPrev)
			}
			if cp.NextCursor != tt.want.NextCursor {
				t.Errorf("NextCursor = %v, want %v", cp.NextCursor, tt.want.NextCursor)
			}
			if cp.PrevCursor != tt.want.PrevCursor {
				t.Errorf("PrevCursor = %v, want %v", cp.PrevCursor, tt.want.PrevCursor)
			}

			if tt.testGetNext {
				expectedNext := tt.want.CurrentPage + 1
				if !tt.want.HasNext {
					expectedNext = 0
				}
				if next := cp.GetNextPage(); next != expectedNext {
					t.Errorf("GetNextPage() = %v, want %v", next, expectedNext)
				}
			}

			if tt.testGetPrev {
				expectedPrev := tt.want.CurrentPage - 1
				if !tt.want.HasPrev {
					expectedPrev = 0
				}
				if prev := cp.GetPrevPage(); prev != expectedPrev {
					t.Errorf("GetPrevPage() = %v, want %v", prev, expectedPrev)
				}
			}
		})
	}
}
