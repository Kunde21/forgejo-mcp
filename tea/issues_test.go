package tea

import (
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestBuildIssueListOptions(t *testing.T) {
	since := time.Now().AddDate(0, 0, -7)
	before := time.Now()

	tests := []struct {
		name    string
		filters *IssueFilters
		want    *gitea.ListIssueOption
	}{
		{
			name:    "no filters",
			filters: &IssueFilters{},
			want:    &gitea.ListIssueOption{},
		},
		{
			name: "with pagination",
			filters: &IssueFilters{
				Page:     2,
				PageSize: 50,
			},
			want: &gitea.ListIssueOption{
				ListOptions: gitea.ListOptions{
					Page:     2,
					PageSize: 50,
				},
			},
		},
		{
			name: "with state and type",
			filters: &IssueFilters{
				State: StateOpen,
				Type:  "issue",
			},
			want: &gitea.ListIssueOption{
				State: gitea.StateOpen,
				Type:  gitea.IssueType("issue"),
			},
		},
		{
			name: "with labels and keyword",
			filters: &IssueFilters{
				Labels:  []string{"bug", "urgent"},
				KeyWord: "test",
			},
			want: &gitea.ListIssueOption{
				Labels:  []string{"bug", "urgent"},
				KeyWord: "test",
			},
		},
		{
			name: "with time filters",
			filters: &IssueFilters{
				Since:  &since,
				Before: &before,
			},
			want: &gitea.ListIssueOption{
				Since:  since,
				Before: before,
			},
		},
		{
			name: "with user filters",
			filters: &IssueFilters{
				CreatedBy:   "user1",
				AssignedBy:  "user2",
				MentionedBy: "user3",
				Owner:       "org",
				Team:        "team1",
			},
			want: &gitea.ListIssueOption{
				CreatedBy:   "user1",
				AssignedBy:  "user2",
				MentionedBy: "user3",
				Owner:       "org",
				Team:        "team1",
			},
		},
		{
			name:    "nil filters",
			filters: nil,
			want:    &gitea.ListIssueOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildIssueListOptions(tt.filters)
			if !cmp.Equal(tt.want, opts) {
				t.Error(cmp.Diff(tt.want, opts))
			}
		})
	}
}

func TestIssueFilters(t *testing.T) {
	filters := &IssueFilters{
		Page:       2,
		PageSize:   50,
		State:      StateOpen,
		Type:       "issue",
		KeyWord:    "bug",
		CreatedBy:  "user1",
		AssignedBy: "user2",
		Labels:     []string{"bug", "urgent"},
		Milestones: []string{"v1.0"},
	}

	if filters.Page != 2 {
		t.Errorf("Expected Page to be 2, got %d", filters.Page)
	}
	if filters.PageSize != 50 {
		t.Errorf("Expected PageSize to be 50, got %d", filters.PageSize)
	}
	if filters.State != StateOpen {
		t.Errorf("Expected State to be %s, got %s", StateOpen, filters.State)
	}
	if filters.Type != "issue" {
		t.Errorf("Expected Type to be 'issue', got %s", filters.Type)
	}
	if filters.KeyWord != "bug" {
		t.Errorf("Expected KeyWord to be 'bug', got %s", filters.KeyWord)
	}
	if filters.CreatedBy != "user1" {
		t.Errorf("Expected CreatedBy to be 'user1', got %s", filters.CreatedBy)
	}
	if filters.AssignedBy != "user2" {
		t.Errorf("Expected AssignedBy to be 'user2', got %s", filters.AssignedBy)
	}
	if len(filters.Labels) != 2 {
		t.Errorf("Expected Labels to have 2 items, got %d", len(filters.Labels))
	}
	if len(filters.Milestones) != 1 {
		t.Errorf("Expected Milestones to have 1 item, got %d", len(filters.Milestones))
	}
}
