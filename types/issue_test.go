package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIssueValidate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		issue   *Issue
		wantErr bool
	}{
		{
			name: "valid issue",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			issue: &Issue{
				ID:        0,
				Number:    1,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
		{
			name: "invalid number",
			issue: &Issue{
				ID:        1,
				Number:    0,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "",
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "Test Issue",
				State:     "invalid",
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
		{
			name: "zero created time",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: time.Time{}},
				UpdatedAt: Timestamp{Time: now},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
		{
			name: "zero updated time",
			issue: &Issue{
				ID:        1,
				Number:    1,
				Title:     "Test Issue",
				State:     IssueStateOpen,
				CreatedAt: Timestamp{Time: now},
				UpdatedAt: Timestamp{Time: time.Time{}},
				URL:       "https://example.com/issue/1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.issue.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Issue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIssueHasLabel(t *testing.T) {
	issue := &Issue{
		Labels: []PRLabel{
			{Name: "bug"},
			{Name: "enhancement"},
			{Name: "documentation"},
		},
	}

	tests := []struct {
		name     string
		label    string
		expected bool
	}{
		{
			name:     "has bug label",
			label:    "bug",
			expected: true,
		},
		{
			name:     "has enhancement label",
			label:    "enhancement",
			expected: true,
		},
		{
			name:     "does not have feature label",
			label:    "feature",
			expected: false,
		},
		{
			name:     "empty label name",
			label:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := issue.HasLabel(tt.label); got != tt.expected {
				t.Errorf("Issue.HasLabel(%q) = %v, want %v", tt.label, got, tt.expected)
			}
		})
	}
}

func TestIssueJSONMarshal(t *testing.T) {
	now := time.Now()
	issue := &Issue{
		ID:        1,
		Number:    1,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     IssueStateOpen,
		Author:    &User{ID: 1, Username: "testuser"},
		Labels:    []PRLabel{{ID: 1, Name: "bug", Color: "#ff0000"}},
		CreatedAt: Timestamp{Time: now},
		UpdatedAt: Timestamp{Time: now},
		URL:       "https://example.com/issue/1",
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Issue
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Compare key fields
	if unmarshaled.ID != issue.ID {
		t.Errorf("ID = %v, want %v", unmarshaled.ID, issue.ID)
	}
	if unmarshaled.Title != issue.Title {
		t.Errorf("Title = %v, want %v", unmarshaled.Title, issue.Title)
	}
	if unmarshaled.State != issue.State {
		t.Errorf("State = %v, want %v", unmarshaled.State, issue.State)
	}
	if unmarshaled.Author.Username != issue.Author.Username {
		t.Errorf("Author.Username = %v, want %v", unmarshaled.Author.Username, issue.Author.Username)
	}
}

func TestMilestoneValidate(t *testing.T) {
	tests := []struct {
		name      string
		milestone *Milestone
		wantErr   bool
	}{
		{
			name: "valid milestone",
			milestone: &Milestone{
				ID:    1,
				Title: "v1.0.0",
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			milestone: &Milestone{
				ID:    0,
				Title: "v1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			milestone: &Milestone{
				ID:    1,
				Title: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.milestone.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Milestone.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIssueStateConstants(t *testing.T) {
	// Test that the constants are defined correctly
	if IssueStateOpen != "open" {
		t.Errorf("IssueStateOpen = %v, want %v", IssueStateOpen, "open")
	}
	if IssueStateClosed != "closed" {
		t.Errorf("IssueStateClosed = %v, want %v", IssueStateClosed, "closed")
	}
}
