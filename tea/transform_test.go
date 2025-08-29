package tea

import (
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"

	"github.com/Kunde21/forgejo-mcp/client"
)

func TestTransformRepositoryToMCP(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		repo *gitea.Repository
		want map[string]interface{}
	}{
		{
			name: "complete repository",
			repo: &gitea.Repository{
				ID:          1,
				Name:        "test-repo",
				FullName:    "owner/test-repo",
				Description: "A test repository",
				HTMLURL:     "https://example.com/owner/test-repo",
				CloneURL:    "https://example.com/owner/test-repo.git",
				SSHURL:      "git@example.com:owner/test-repo.git",
				Created:     now,
				Updated:     now,
				Private:     true,
				Archived:    false,
			},
			want: map[string]interface{}{
				"id":          int64(1),
				"name":        "test-repo",
				"full_name":   "owner/test-repo",
				"description": "A test repository",
				"html_url":    "https://example.com/owner/test-repo",
				"clone_url":   "https://example.com/owner/test-repo.git",
				"ssh_url":     "git@example.com:owner/test-repo.git",
				"created_at":  now,
				"updated_at":  now,
				"private":     true,
				"archived":    false,
				"type":        "repository",
			},
		},
		{
			name: "repository with nil fields",
			repo: &gitea.Repository{
				ID:       2,
				Name:     "empty-repo",
				FullName: "owner/empty-repo",
			},
			want: map[string]interface{}{
				"id":        int64(2),
				"name":      "empty-repo",
				"full_name": "owner/empty-repo",
				"private":   false,
				"archived":  false,
				"type":      "repository",
			},
		},
		{
			name: "nil repository",
			repo: nil,
			want: map[string]interface{}{
				"type": "repository",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformRepositoryToMCP(tt.repo)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("TransformRepositoryToMCP() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestTransformIssueToMCP(t *testing.T) {
	now := time.Now()
	pastTime := time.Now().Add(-24 * time.Hour)

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	label1 := &gitea.Label{
		ID:    1,
		Name:  "bug",
		Color: "e11d21",
	}

	label2 := &gitea.Label{
		ID:    2,
		Name:  "urgent",
		Color: "e11d22",
	}

	tests := []struct {
		name  string
		issue *gitea.Issue
		want  map[string]interface{}
	}{
		{
			name: "complete issue",
			issue: &gitea.Issue{
				ID:      1,
				Title:   "Test Issue",
				Body:    "This is a test issue",
				Index:   42,
				State:   gitea.StateOpen,
				Poster:  user,
				Labels:  []*gitea.Label{label1, label2},
				Created: now,
				Updated: now,
				Closed:  &pastTime,
			},
			want: map[string]interface{}{
				"id":          int64(1),
				"number":      int64(42),
				"title":       "Test Issue",
				"body":        "This is a test issue",
				"state":       "open",
				"author":      "testuser",
				"author_name": "Test User",
				"labels": []map[string]interface{}{
					{
						"id":    int64(1),
						"name":  "bug",
						"color": "e11d21",
					},
					{
						"id":    int64(2),
						"name":  "urgent",
						"color": "e11d22",
					},
				},
				"created_at": now,
				"updated_at": now,
				"closed_at":  &pastTime,
				"type":       "issue",
			},
		},
		{
			name: "issue with nil fields",
			issue: &gitea.Issue{
				ID:    2,
				Title: "Simple Issue",
				Index: 43,
				State: gitea.StateOpen,
			},
			want: map[string]interface{}{
				"id":     int64(2),
				"number": int64(43),
				"title":  "Simple Issue",
				"state":  "open",
				"labels": []map[string]interface{}{},
				"type":   "issue",
			},
		},
		{
			name:  "nil issue",
			issue: nil,
			want: map[string]interface{}{
				"type": "issue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformIssueToMCP(tt.issue)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("TransformIssueToMCP() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestTransformPullRequestToMCP(t *testing.T) {
	now := time.Now()
	pastTime := time.Now().Add(-24 * time.Hour)

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	label1 := &gitea.Label{
		ID:    1,
		Name:  "feature",
		Color: "0052cc",
	}

	label2 := &gitea.Label{
		ID:    2,
		Name:  "review",
		Color: "0052cd",
	}

	tests := []struct {
		name string
		pr   *gitea.PullRequest
		want map[string]interface{}
	}{
		{
			name: "complete pull request",
			pr: &gitea.PullRequest{
				ID:        1,
				Title:     "Test PR",
				Body:      "This is a test pull request",
				Index:     42,
				State:     gitea.StateOpen,
				Poster:    user,
				Labels:    []*gitea.Label{label1, label2},
				Created:   &now,
				Updated:   &now,
				Closed:    &pastTime,
				Merged:    &pastTime,
				HasMerged: true,
				Base: &gitea.PRBranchInfo{
					Ref: "main",
					Sha: "abc123",
				},
				Head: &gitea.PRBranchInfo{
					Ref: "feature-branch",
					Sha: "def456",
				},
			},
			want: map[string]interface{}{
				"id":          int64(1),
				"number":      int64(42),
				"title":       "Test PR",
				"body":        "This is a test pull request",
				"state":       "open",
				"author":      "testuser",
				"author_name": "Test User",
				"labels": []map[string]interface{}{
					{
						"id":    int64(1),
						"name":  "feature",
						"color": "0052cc",
					},
					{
						"id":    int64(2),
						"name":  "review",
						"color": "0052cd",
					},
				},
				"created_at":  &now,
				"updated_at":  &now,
				"closed_at":   &pastTime,
				"merged_at":   &pastTime,
				"has_merged":  true,
				"base_branch": "main",
				"base_sha":    "abc123",
				"head_branch": "feature-branch",
				"head_sha":    "def456",
				"type":        "pull_request",
			},
		},
		{
			name: "pull request with nil fields",
			pr: &gitea.PullRequest{
				ID:    2,
				Title: "Simple PR",
				Index: 43,
				State: gitea.StateOpen,
			},
			want: map[string]interface{}{
				"id":         int64(2),
				"number":     int64(43),
				"title":      "Simple PR",
				"state":      "open",
				"labels":     []map[string]interface{}{},
				"has_merged": false,
				"type":       "pull_request",
			},
		},
		{
			name: "nil pull request",
			pr:   nil,
			want: map[string]interface{}{
				"type": "pull_request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformPullRequestToMCP(tt.pr)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("TransformPullRequestToMCP() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestTransformRepositoriesToMCP(t *testing.T) {
	now := time.Now()

	repos := []*gitea.Repository{
		{
			ID:          1,
			Name:        "repo1",
			FullName:    "owner/repo1",
			Description: "First repository",
			HTMLURL:     "https://example.com/owner/repo1",
			Created:     now,
		},
		{
			ID:          2,
			Name:        "repo2",
			FullName:    "owner/repo2",
			Description: "Second repository",
			HTMLURL:     "https://example.com/owner/repo2",
			Created:     now,
		},
	}

	want := []map[string]interface{}{
		{
			"id":          int64(1),
			"name":        "repo1",
			"full_name":   "owner/repo1",
			"description": "First repository",
			"html_url":    "https://example.com/owner/repo1",
			"created_at":  now,
			"private":     false,
			"archived":    false,
			"type":        "repository",
		},
		{
			"id":          int64(2),
			"name":        "repo2",
			"full_name":   "owner/repo2",
			"description": "Second repository",
			"html_url":    "https://example.com/owner/repo2",
			"created_at":  now,
			"private":     false,
			"archived":    false,
			"type":        "repository",
		},
	}

	got := TransformRepositoriesToMCP(repos)
	if !cmp.Equal(want, got) {
		t.Errorf("TransformRepositoriesToMCP() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestTransformIssuesToMCP(t *testing.T) {
	now := time.Now()

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
	}

	issues := []*gitea.Issue{
		{
			ID:      1,
			Title:   "Issue 1",
			Index:   42,
			State:   gitea.StateOpen,
			Poster:  user,
			Created: now,
		},
		{
			ID:      2,
			Title:   "Issue 2",
			Index:   43,
			State:   gitea.StateOpen,
			Poster:  user,
			Created: now,
		},
	}

	want := []map[string]interface{}{
		{
			"id":          int64(1),
			"number":      int64(42),
			"title":       "Issue 1",
			"state":       "open",
			"author":      "testuser",
			"author_name": "Test User",
			"created_at":  now,
			"labels":      []map[string]interface{}{},
			"type":        "issue",
		},
		{
			"id":          int64(2),
			"number":      int64(43),
			"title":       "Issue 2",
			"state":       "open",
			"author":      "testuser",
			"author_name": "Test User",
			"created_at":  now,
			"labels":      []map[string]interface{}{},
			"type":        "issue",
		},
	}

	got := TransformIssuesToMCP(issues)
	if !cmp.Equal(want, got) {
		t.Errorf("TransformIssuesToMCP() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestTransformPullRequestsToMCP(t *testing.T) {
	now := time.Now()

	user := &gitea.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
	}

	prs := []*gitea.PullRequest{
		{
			ID:      1,
			Title:   "PR 1",
			Index:   42,
			State:   gitea.StateOpen,
			Poster:  user,
			Created: &now,
		},
		{
			ID:      2,
			Title:   "PR 2",
			Index:   43,
			State:   gitea.StateOpen,
			Poster:  user,
			Created: &now,
		},
	}

	want := []map[string]interface{}{
		{
			"id":          int64(1),
			"number":      int64(42),
			"title":       "PR 1",
			"state":       "open",
			"author":      "testuser",
			"author_name": "Test User",
			"created_at":  &now,
			"labels":      []map[string]interface{}{},
			"has_merged":  false,
			"type":        "pull_request",
		},
		{
			"id":          int64(2),
			"number":      int64(43),
			"title":       "PR 2",
			"state":       "open",
			"author":      "testuser",
			"author_name": "Test User",
			"created_at":  &now,
			"labels":      []map[string]interface{}{},
			"has_merged":  false,
			"type":        "pull_request",
		},
	}

	got := TransformPullRequestsToMCP(prs)
	if !cmp.Equal(want, got) {
		t.Errorf("TransformPullRequestsToMCP() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

// TestPRTransformer tests the PRTransformer type and methods
func TestPRTransformer(t *testing.T) {
	transformer := NewPRTransformer()
	if transformer == nil {
		t.Error("Expected non-nil transformer")
	}

	now := time.Now()
	pastTime := time.Now().Add(-24 * time.Hour)

	user := &client.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	label1 := &client.Label{
		ID:    1,
		Name:  "feature",
		Color: "0052cc",
	}

	label2 := &client.Label{
		ID:    2,
		Name:  "review",
		Color: "0052cd",
	}

	prs := []client.PullRequest{
		{
			ID:        1,
			Index:     42,
			Title:     "Test PR",
			Body:      "This is a test pull request",
			State:     client.StateOpen,
			Poster:    user,
			Labels:    []*client.Label{label1, label2},
			Created:   &now,
			Updated:   &now,
			Closed:    &pastTime,
			Merged:    &pastTime,
			HasMerged: true,
			Base: &client.PRBranchInfo{
				Ref: "main",
				SHA: "abc123",
			},
			Head: &client.PRBranchInfo{
				Ref: "feature-branch",
				SHA: "def456",
			},
			HTMLURL: "https://example.com/owner/repo/pulls/42",
		},
		{
			ID:      2,
			Index:   43,
			Title:   "Simple PR",
			State:   client.StateOpen,
			Created: &now,
		},
	}

	result, err := transformer.TransformPRsToMCP(prs)
	if err != nil {
		t.Errorf("TransformPRsToMCP failed: %v", err)
	}

	if len(result) != len(prs) {
		t.Errorf("Expected %d results, got %d", len(prs), len(result))
	}

	// Test the first PR transformation
	firstPR := result[0]
	if firstPR["id"] != int64(1) {
		t.Errorf("Expected id=1, got %v", firstPR["id"])
	}

	if firstPR["title"] != "Test PR" {
		t.Errorf("Expected title='Test PR', got %v", firstPR["title"])
	}

	if firstPR["html_url"] != "https://example.com/owner/repo/pulls/42" {
		t.Errorf("Expected html_url='https://example.com/owner/repo/pulls/42', got %v", firstPR["html_url"])
	}

	// Test the second PR transformation
	secondPR := result[1]
	if secondPR["id"] != int64(2) {
		t.Errorf("Expected id=2, got %v", secondPR["id"])
	}
}

// TestIssueTransformer tests the IssueTransformer type and methods
func TestIssueTransformer(t *testing.T) {
	transformer := NewIssueTransformer()
	if transformer == nil {
		t.Error("Expected non-nil transformer")
	}

	now := time.Now()
	pastTime := time.Now().Add(-24 * time.Hour)

	user := &client.User{
		ID:       1,
		UserName: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	label1 := &client.Label{
		ID:    1,
		Name:  "bug",
		Color: "e11d21",
	}

	label2 := &client.Label{
		ID:    2,
		Name:  "urgent",
		Color: "e11d22",
	}

	issues := []client.Issue{
		{
			ID:      1,
			Index:   42,
			Title:   "Test Issue",
			Body:    "This is a test issue",
			State:   client.StateOpen,
			Poster:  user,
			Labels:  []*client.Label{label1, label2},
			Created: now,
			Updated: now,
			Closed:  &pastTime,
			HTMLURL: "https://example.com/owner/repo/issues/42",
		},
		{
			ID:      2,
			Index:   43,
			Title:   "Simple Issue",
			State:   client.StateOpen,
			Created: now,
		},
	}

	result, err := transformer.TransformIssuesToMCP(issues)
	if err != nil {
		t.Errorf("TransformIssuesToMCP failed: %v", err)
	}

	if len(result) != len(issues) {
		t.Errorf("Expected %d results, got %d", len(issues), len(result))
	}

	// Test the first issue transformation
	firstIssue := result[0]
	if firstIssue["id"] != int64(1) {
		t.Errorf("Expected id=1, got %v", firstIssue["id"])
	}

	if firstIssue["title"] != "Test Issue" {
		t.Errorf("Expected title='Test Issue', got %v", firstIssue["title"])
	}

	if firstIssue["html_url"] != "https://example.com/owner/repo/issues/42" {
		t.Errorf("Expected html_url='https://example.com/owner/repo/issues/42', got %v", firstIssue["html_url"])
	}

	// Test the second issue transformation
	secondIssue := result[1]
	if secondIssue["id"] != int64(2) {
		t.Errorf("Expected id=2, got %v", secondIssue["id"])
	}
}
