package tea

import (
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestBuildPullRequestListOptions(t *testing.T) {
	tests := []struct {
		name    string
		filters *PullRequestFilters
		want    *gitea.ListPullRequestsOptions
	}{
		{
			name:    "no filters",
			filters: &PullRequestFilters{},
			want:    &gitea.ListPullRequestsOptions{},
		},
		{
			name: "with pagination",
			filters: &PullRequestFilters{
				Page:     2,
				PageSize: 25,
			},
			want: &gitea.ListPullRequestsOptions{
				ListOptions: gitea.ListOptions{
					Page:     2,
					PageSize: 25,
				},
			},
		},
		{
			name: "with state",
			filters: &PullRequestFilters{
				State: StateOpen,
			},
			want: &gitea.ListPullRequestsOptions{
				State: gitea.StateOpen,
			},
		},
		{
			name: "with sort",
			filters: &PullRequestFilters{
				Sort: "updated",
			},
			want: &gitea.ListPullRequestsOptions{
				Sort: "updated",
			},
		},
		{
			name: "with milestone",
			filters: &PullRequestFilters{
				Milestone: 123,
			},
			want: &gitea.ListPullRequestsOptions{
				Milestone: 123,
			},
		},
		{
			name: "with all filters",
			filters: &PullRequestFilters{
				Page:      1,
				PageSize:  10,
				State:     StateClosed,
				Sort:      "created",
				Milestone: 456,
			},
			want: &gitea.ListPullRequestsOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 10,
				},
				State:     gitea.StateClosed,
				Sort:      "created",
				Milestone: 456,
			},
		},
		{
			name:    "nil filters",
			filters: nil,
			want:    &gitea.ListPullRequestsOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildPullRequestListOptions(tt.filters)
			if !cmp.Equal(tt.want, opts) {
				t.Error(cmp.Diff(tt.want, opts))
			}
		})
	}
}

func TestPullRequestFilters(t *testing.T) {
	filters := &PullRequestFilters{
		Page:      1,
		PageSize:  25,
		State:     StateOpen,
		Sort:      "updated",
		Milestone: 123,
	}

	if filters.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", filters.Page)
	}
	if filters.PageSize != 25 {
		t.Errorf("Expected PageSize to be 25, got %d", filters.PageSize)
	}
	if filters.State != StateOpen {
		t.Errorf("Expected State to be %s, got %s", StateOpen, filters.State)
	}
	if filters.Sort != "updated" {
		t.Errorf("Expected Sort to be 'updated', got %s", filters.Sort)
	}
	if filters.Milestone != 123 {
		t.Errorf("Expected Milestone to be 123, got %d", filters.Milestone)
	}
}
