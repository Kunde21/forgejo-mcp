package tea

import (
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestExtractRepositoryMetadata(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		repo *gitea.Repository
		want map[string]interface{}
	}{
		{
			name: "complete repository metadata",
			repo: &gitea.Repository{
				ID:              1,
				Name:            "test-repo",
				FullName:        "owner/test-repo",
				Description:     "A test repository",
				Stars:           42,
				Forks:           15,
				OpenIssues:      3,
				OpenPulls:       2,
				Watchers:        25,
				Size:            1024,
				Created:         now,
				Updated:         now,
				Private:         true,
				Archived:        false,
				HasIssues:       true,
				HasWiki:         true,
				HasPullRequests: true,
			},
			want: map[string]interface{}{
				"stars_count":       42,
				"forks_count":       15,
				"open_issues":       3,
				"open_pulls":        2,
				"watchers_count":    25,
				"size":              1024,
				"created_at":        now,
				"updated_at":        now,
				"private":           true,
				"archived":          false,
				"has_issues":        true,
				"has_wiki":          true,
				"has_pull_requests": true,
			},
		},
		{
			name: "minimal repository metadata",
			repo: &gitea.Repository{
				ID:   2,
				Name: "simple-repo",
			},
			want: map[string]interface{}{
				"stars_count":       0,
				"forks_count":       0,
				"open_issues":       0,
				"open_pulls":        0,
				"watchers_count":    0,
				"size":              0,
				"private":           false,
				"archived":          false,
				"has_issues":        false,
				"has_wiki":          false,
				"has_pull_requests": false,
			},
		},
		{
			name: "nil repository",
			repo: nil,
			want: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractRepositoryMetadata(tt.repo)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("ExtractRepositoryMetadata() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestExtractIssueMetadata(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
	}

	milestone := &gitea.Milestone{
		ID:          1,
		Title:       "v1.0",
		Description: "Version 1.0",
	}

	tests := []struct {
		name  string
		issue *gitea.Issue
		want  map[string]interface{}
	}{
		{
			name: "complete issue metadata",
			issue: &gitea.Issue{
				ID:               1,
				Index:            42,
				Title:            "Test Issue",
				State:            gitea.StateOpen,
				Poster:           user,
				Milestone:        milestone,
				Assignees:        []*gitea.User{user},
				Comments:         5,
				Created:          now,
				Updated:          now,
				Closed:           &pastTime,
				OriginalAuthor:   "originaluser",
				OriginalAuthorID: 100,
			},
			want: map[string]interface{}{
				"id":                 int64(1),
				"number":             int64(42),
				"state":              "open",
				"author":             "testuser",
				"author_name":        "Test User",
				"milestone":          "v1.0",
				"assignees":          []string{"testuser"},
				"comments_count":     5,
				"created_at":         now,
				"updated_at":         now,
				"closed_at":          &pastTime,
				"original_author":    "originaluser",
				"original_author_id": int64(100),
			},
		},
		{
			name: "issue with multiple assignees",
			issue: &gitea.Issue{
				ID:        2,
				Index:     43,
				Title:     "Multi Assignee Issue",
				State:     gitea.StateOpen,
				Assignees: []*gitea.User{{UserName: "user1"}, {UserName: "user2"}, {UserName: "user3"}},
			},
			want: map[string]interface{}{
				"id":             int64(2),
				"number":         int64(43),
				"state":          "open",
				"assignees":      []string{"user1", "user2", "user3"},
				"comments_count": 0,
			},
		},
		{
			name: "simple issue metadata",
			issue: &gitea.Issue{
				ID:    3,
				Index: 44,
				Title: "Simple Issue",
				State: gitea.StateOpen,
			},
			want: map[string]interface{}{
				"id":             int64(3),
				"number":         int64(44),
				"state":          "open",
				"assignees":      []string{},
				"comments_count": 0,
			},
		},
		{
			name:  "nil issue",
			issue: nil,
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIssueMetadata(tt.issue)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("ExtractIssueMetadata() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestExtractPullRequestMetadata(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
	}

	milestone := &gitea.Milestone{
		ID:    1,
		Title: "v1.0",
	}

	tests := []struct {
		name string
		pr   *gitea.PullRequest
		want map[string]interface{}
	}{
		{
			name: "complete pull request metadata",
			pr: &gitea.PullRequest{
				ID:                  1,
				Index:               42,
				Title:               "Test PR",
				State:               gitea.StateOpen,
				Poster:              user,
				Milestone:           milestone,
				Assignees:           []*gitea.User{user},
				Comments:            3,
				Merged:              &pastTime,
				HasMerged:           true,
				MergedBy:            user,
				Created:             &now,
				Updated:             &now,
				Closed:              &pastTime,
				AllowMaintainerEdit: true,
			},
			want: map[string]interface{}{
				"id":                    int64(1),
				"number":                int64(42),
				"state":                 "open",
				"author":                "testuser",
				"author_name":           "Test User",
				"milestone":             "v1.0",
				"assignees":             []string{"testuser"},
				"comments_count":        3,
				"merged_at":             &pastTime,
				"has_merged":            true,
				"merged_by":             "testuser",
				"merged_by_name":        "Test User",
				"created_at":            &now,
				"updated_at":            &now,
				"closed_at":             &pastTime,
				"allow_maintainer_edit": true,
			},
		},
		{
			name: "pull request without merge info",
			pr: &gitea.PullRequest{
				ID:    2,
				Index: 43,
				Title: "Unmerged PR",
				State: gitea.StateOpen,
			},
			want: map[string]interface{}{
				"id":                    int64(2),
				"number":                int64(43),
				"state":                 "open",
				"assignees":             []string{},
				"comments_count":        0,
				"has_merged":            false,
				"allow_maintainer_edit": false,
			},
		},
		{
			name: "nil pull request",
			pr:   nil,
			want: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPullRequestMetadata(tt.pr)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("ExtractPullRequestMetadata() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestBuildContextInfo(t *testing.T) {
	labels := []*gitea.Label{
		{Name: "bug", Color: "e11d21"},
		{Name: "urgent", Color: "e11d22"},
		{Name: "feature", Color: "0052cc"},
	}

	milestone := &gitea.Milestone{
		ID:          1,
		Title:       "v1.0",
		Description: "Version 1.0 milestone",
	}

	tests := []struct {
		name          string
		labels        []*gitea.Label
		milestone     *gitea.Milestone
		wantLabels    []map[string]interface{}
		wantMilestone map[string]interface{}
	}{
		{
			name:      "complete context info",
			labels:    labels,
			milestone: milestone,
			wantLabels: []map[string]interface{}{
				{"name": "bug", "color": "e11d21"},
				{"name": "urgent", "color": "e11d22"},
				{"name": "feature", "color": "0052cc"},
			},
			wantMilestone: map[string]interface{}{
				"id":          int64(1),
				"title":       "v1.0",
				"description": "Version 1.0 milestone",
			},
		},
		{
			name:          "no labels or milestone",
			labels:        nil,
			milestone:     nil,
			wantLabels:    []map[string]interface{}{},
			wantMilestone: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labelsResult := BuildLabelsContext(tt.labels)
			if !cmp.Equal(tt.wantLabels, labelsResult) {
				t.Errorf("BuildLabelsContext() mismatch (-want +got):\n%s", cmp.Diff(tt.wantLabels, labelsResult))
			}

			milestoneResult := BuildMilestoneContext(tt.milestone)
			if !cmp.Equal(tt.wantMilestone, milestoneResult) {
				t.Errorf("BuildMilestoneContext() mismatch (-want +got):\n%s", cmp.Diff(tt.wantMilestone, milestoneResult))
			}
		})
	}
}

func TestMapResourceRelationships(t *testing.T) {
	// Create test data
	repo := &gitea.Repository{
		ID:       1,
		Name:     "test-repo",
		FullName: "owner/test-repo",
	}

	issue := &gitea.Issue{
		ID:    1,
		Index: 42,
		Title: "Test Issue",
		State: gitea.StateOpen,
		Repository: &gitea.RepositoryMeta{
			ID:   1,
			Name: "test-repo",
		},
	}

	pr := &gitea.PullRequest{
		ID:    1,
		Index: 1,
		Title: "Test PR",
		State: gitea.StateOpen,
	}

	tests := []struct {
		name         string
		repo         *gitea.Repository
		issue        *gitea.Issue
		pr           *gitea.PullRequest
		wantIssueRel map[string]interface{}
		wantPRRel    map[string]interface{}
	}{
		{
			name:  "complete relationship mapping",
			repo:  repo,
			issue: issue,
			pr:    pr,
			wantIssueRel: map[string]interface{}{
				"repository_id":   int64(1),
				"repository_name": "test-repo",
			},
			wantPRRel: map[string]interface{}{},
		},
		{
			name:         "nil resources",
			repo:         nil,
			issue:        nil,
			pr:           nil,
			wantIssueRel: map[string]interface{}{},
			wantPRRel:    map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issueRel := MapIssueRelationships(tt.issue)
			if !cmp.Equal(tt.wantIssueRel, issueRel) {
				t.Errorf("MapIssueRelationships() mismatch (-want +got):\n%s", cmp.Diff(tt.wantIssueRel, issueRel))
			}

			prRel := MapPullRequestRelationships(tt.pr)
			if !cmp.Equal(tt.wantPRRel, prRel) {
				t.Errorf("MapPullRequestRelationships() mismatch (-want +got):\n%s", cmp.Diff(tt.wantPRRel, prRel))
			}
		})
	}
}
