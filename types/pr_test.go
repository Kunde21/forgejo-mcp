package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPullRequestValidate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		pr      *PullRequest
		wantErr bool
	}{
		{
			name: "valid pull request",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			pr: &PullRequest{
				ID:         0,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "invalid number",
			pr: &PullRequest{
				ID:         1,
				Number:     0,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "empty head branch",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "empty base branch",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "empty diff URL",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      "invalid",
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "zero created time",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: time.Time{}},
				UpdatedAt:  Timestamp{Time: now},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
		{
			name: "zero updated time",
			pr: &PullRequest{
				ID:         1,
				Number:     1,
				Title:      "Test PR",
				State:      PRStateOpen,
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				CreatedAt:  Timestamp{Time: now},
				UpdatedAt:  Timestamp{Time: time.Time{}},
				URL:        "https://example.com/pr/1",
				DiffURL:    "https://example.com/pr/1.diff",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pr.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PullRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPullRequestStateMethods(t *testing.T) {
	tests := []struct {
		name     string
		state    PRState
		isOpen   bool
		isClosed bool
		isMerged bool
	}{
		{
			name:     "open state",
			state:    PRStateOpen,
			isOpen:   true,
			isClosed: false,
			isMerged: false,
		},
		{
			name:     "closed state",
			state:    PRStateClosed,
			isOpen:   false,
			isClosed: true,
			isMerged: false,
		},
		{
			name:     "merged state",
			state:    PRStateMerged,
			isOpen:   false,
			isClosed: false,
			isMerged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{State: tt.state}

			if got := pr.IsOpen(); got != tt.isOpen {
				t.Errorf("PullRequest.IsOpen() = %v, want %v", got, tt.isOpen)
			}
			if got := pr.IsClosed(); got != tt.isClosed {
				t.Errorf("PullRequest.IsClosed() = %v, want %v", got, tt.isClosed)
			}
			if got := pr.IsMerged(); got != tt.isMerged {
				t.Errorf("PullRequest.IsMerged() = %v, want %v", got, tt.isMerged)
			}
		})
	}
}

func TestPullRequestJSONMarshal(t *testing.T) {
	now := time.Now()
	pr := &PullRequest{
		ID:         1,
		Number:     1,
		Title:      "Test PR",
		Body:       "Test body",
		State:      PRStateOpen,
		Author:     &PRAuthor{Username: "testuser", AvatarURL: "https://example.com/avatar.jpg", URL: "https://example.com/user/testuser"},
		HeadBranch: "feature-branch",
		BaseBranch: "main",
		CreatedAt:  Timestamp{Time: now},
		UpdatedAt:  Timestamp{Time: now},
		Draft:      false,
		Labels:     []PRLabel{{ID: 1, Name: "bug", Color: "#ff0000"}},
		URL:        "https://example.com/pr/1",
		DiffURL:    "https://example.com/pr/1.diff",
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled PullRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Compare key fields
	if unmarshaled.ID != pr.ID {
		t.Errorf("ID = %v, want %v", unmarshaled.ID, pr.ID)
	}
	if unmarshaled.Title != pr.Title {
		t.Errorf("Title = %v, want %v", unmarshaled.Title, pr.Title)
	}
	if unmarshaled.State != pr.State {
		t.Errorf("State = %v, want %v", unmarshaled.State, pr.State)
	}
	if unmarshaled.Author.Username != pr.Author.Username {
		t.Errorf("Author.Username = %v, want %v", unmarshaled.Author.Username, pr.Author.Username)
	}
}

func TestPRAuthorValidate(t *testing.T) {
	tests := []struct {
		name    string
		author  *PRAuthor
		wantErr bool
	}{
		{
			name: "valid author",
			author: &PRAuthor{
				Username:  "testuser",
				AvatarURL: "https://example.com/avatar.jpg",
				URL:       "https://example.com/user/testuser",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			author: &PRAuthor{
				Username:  "",
				AvatarURL: "https://example.com/avatar.jpg",
				URL:       "https://example.com/user/testuser",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.author.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PRAuthor.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPRLabelValidate(t *testing.T) {
	tests := []struct {
		name    string
		label   *PRLabel
		wantErr bool
	}{
		{
			name: "valid label",
			label: &PRLabel{
				ID:    1,
				Name:  "bug",
				Color: "#ff0000",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			label: &PRLabel{
				ID:    1,
				Name:  "",
				Color: "#ff0000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.label.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PRLabel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
