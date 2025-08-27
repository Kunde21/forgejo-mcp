package tea

import (
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestBuildRepoListOptions(t *testing.T) {
	tests := []struct {
		name    string
		filters *RepositoryFilters
		want    *gitea.ListReposOptions
	}{
		{
			name:    "no filters",
			filters: &RepositoryFilters{},
			want: &gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
			},
		},
		{
			name: "with page and limit",
			filters: &RepositoryFilters{
				Page:     2,
				PageSize: 50,
			},
			want: &gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					Page:     2,
					PageSize: 50,
				},
			},
		},
		{
			name:    "nil filters",
			filters: nil,
			want: &gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildRepoListOptions(tt.filters)
			if !cmp.Equal(tt.want, opts) {
				t.Error(cmp.Diff(tt.want, opts))
			}
		})
	}
}

func TestBuildSearchRepoOptions(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		filters *RepositoryFilters
		want    *gitea.SearchRepoOptions
	}{
		{
			name:    "no filters",
			filters: &RepositoryFilters{},
			want: &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
			},
		},
		{
			name: "with search query",
			filters: &RepositoryFilters{
				Query: "test-repo",
			},
			want: &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
				Keyword: "test-repo",
			},
		},
		{
			name: "with private filter",
			filters: &RepositoryFilters{
				IsPrivate: &trueVal,
			},
			want: &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
				IsPrivate: &trueVal,
			},
		},
		{
			name: "with archived filter",
			filters: &RepositoryFilters{
				IsArchived: &falseVal,
			},
			want: &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
				IsArchived: &falseVal,
			},
		},
		{
			name:    "nil filters",
			filters: nil,
			want: &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     1,
					PageSize: 30,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildSearchRepoOptions(tt.filters)
			if !cmp.Equal(tt.want, opts) {
				t.Error(cmp.Diff(tt.want, opts))
			}
		})
	}
}
