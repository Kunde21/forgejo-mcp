package gitea

import (
	"context"
	"encoding/json"
	"testing"
)

type dataStructureTestCase struct {
	name     string
	input    any
	expected any
}

func TestIssue_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "basic issue",
			input: Issue{
				Number: 42,
				Title:  "Test Issue",
				State:  "open",
			},
			expected: `{"number":42,"title":"Test Issue","state":"open"}`,
		},
		{
			name: "closed issue",
			input: Issue{
				Number: 123,
				Title:  "Bug: Login fails",
				State:  "closed",
			},
			expected: `{"number":123,"title":"Bug: Login fails","state":"closed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal Issue: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled Issue
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal Issue: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestIssueComment_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "basic comment",
			input: IssueComment{
				ID:      1,
				Content: "This is a test comment",
				Author:  "testuser",
				Created: "2024-01-01T00:00:00Z",
			},
			expected: `{"id":1,"content":"This is a test comment","author":"testuser","created":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "empty comment",
			input: IssueComment{
				ID:      0,
				Content: "",
				Author:  "",
				Created: "",
			},
			expected: `{"id":0,"content":"","author":"","created":""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal IssueComment: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled IssueComment
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal IssueComment: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestIssueCommentList_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "comment list with data",
			input: IssueCommentList{
				Comments: []IssueComment{
					{
						ID:      1,
						Content: "First comment",
						Author:  "user1",
						Created: "2024-01-01T00:00:00Z",
					},
					{
						ID:      2,
						Content: "Second comment",
						Author:  "user2",
						Created: "2024-01-02T00:00:00Z",
					},
				},
				Total:  2,
				Limit:  10,
				Offset: 0,
			},
			expected: `{"comments":[{"id":1,"content":"First comment","author":"user1","created":"2024-01-01T00:00:00Z"},{"id":2,"content":"Second comment","author":"user2","created":"2024-01-02T00:00:00Z"}],"total":2,"limit":10,"offset":0}`,
		},
		{
			name: "empty comment list",
			input: IssueCommentList{
				Comments: []IssueComment{},
				Total:    0,
				Limit:    15,
				Offset:   0,
			},
			expected: `{"comments":[],"total":0,"limit":15,"offset":0}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal IssueCommentList: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled IssueCommentList
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal IssueCommentList: %v", err)
			}
			if len(unmarshaled.Comments) != len(tc.input.(IssueCommentList).Comments) {
				t.Errorf("Expected %d comments, got %d", len(tc.input.(IssueCommentList).Comments), len(unmarshaled.Comments))
			}
			for i, comment := range unmarshaled.Comments {
				if comment != tc.input.(IssueCommentList).Comments[i] {
					t.Errorf("Comment %d: expected %+v, got %+v", i, tc.input.(IssueCommentList).Comments[i], comment)
				}
			}
			if unmarshaled.Total != tc.input.(IssueCommentList).Total ||
				unmarshaled.Limit != tc.input.(IssueCommentList).Limit ||
				unmarshaled.Offset != tc.input.(IssueCommentList).Offset {
				t.Errorf("Pagination metadata mismatch: expected %+v, got %+v",
					map[string]int{"total": tc.input.(IssueCommentList).Total, "limit": tc.input.(IssueCommentList).Limit, "offset": tc.input.(IssueCommentList).Offset},
					map[string]int{"total": unmarshaled.Total, "limit": unmarshaled.Limit, "offset": unmarshaled.Offset})
			}
		})
	}
}

func TestListIssueCommentsArgs_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "valid args",
			input: ListIssueCommentsArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 42,
				Limit:       10,
				Offset:      0,
			},
			expected: `{"repository":"testuser/testrepo","issue_number":42,"limit":10,"offset":0}`,
		},
		{
			name: "default values",
			input: ListIssueCommentsArgs{
				Repository:  "owner/repo",
				IssueNumber: 1,
			},
			expected: `{"repository":"owner/repo","issue_number":1,"limit":0,"offset":0}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal ListIssueCommentsArgs: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled ListIssueCommentsArgs
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal ListIssueCommentsArgs: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestPullRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "basic pull request",
			input: PullRequest{
				ID:        1,
				Number:    1,
				Title:     "Add new feature",
				Body:      "This PR adds a new feature",
				State:     "open",
				User:      "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
				Head: PullRequestBranch{
					Ref: "feature-branch",
					Sha: "abc123",
				},
				Base: PullRequestBranch{
					Ref: "main",
					Sha: "def456",
				},
			},
			expected: `{"id":1,"number":1,"title":"Add new feature","body":"This PR adds a new feature","state":"open","user":"testuser","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","head":{"ref":"feature-branch","sha":"abc123"},"base":{"ref":"main","sha":"def456"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal PullRequest: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled PullRequest
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal PullRequest: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestPullRequestBranch_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "branch reference",
			input: PullRequestBranch{
				Ref: "main",
				Sha: "abc123def456",
			},
			expected: `{"ref":"main","sha":"abc123def456"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal PullRequestBranch: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled PullRequestBranch
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal PullRequestBranch: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestListPullRequestsOptions_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "all options",
			input: ListPullRequestsOptions{
				State:  "open",
				Limit:  10,
				Offset: 0,
			},
			expected: `{"state":"open","limit":10,"offset":0}`,
		},
		{
			name:     "default values",
			input:    ListPullRequestsOptions{},
			expected: `{"state":"","limit":0,"offset":0}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal ListPullRequestsOptions: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled ListPullRequestsOptions
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal ListPullRequestsOptions: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestPullRequestComment_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "basic PR comment",
			input: PullRequestComment{
				ID:        1,
				Body:      "This is a pull request comment",
				User:      "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
			},
			expected: `{"id":1,"body":"This is a pull request comment","user":"testuser","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "empty PR comment",
			input: PullRequestComment{
				ID:        0,
				Body:      "",
				User:      "",
				CreatedAt: "",
				UpdatedAt: "",
			},
			expected: `{"id":0,"body":"","user":"","created_at":"","updated_at":""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal PullRequestComment: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled PullRequestComment
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal PullRequestComment: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestPullRequestCommentList_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "PR comment list with data",
			input: PullRequestCommentList{
				Comments: []PullRequestComment{
					{
						ID:        1,
						Body:      "First PR comment",
						User:      "user1",
						CreatedAt: "2024-01-01T00:00:00Z",
						UpdatedAt: "2024-01-01T00:00:00Z",
					},
					{
						ID:        2,
						Body:      "Second PR comment",
						User:      "user2",
						CreatedAt: "2024-01-02T00:00:00Z",
						UpdatedAt: "2024-01-02T00:00:00Z",
					},
				},
				Total:  2,
				Limit:  10,
				Offset: 0,
			},
			expected: `{"comments":[{"id":1,"body":"First PR comment","user":"user1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"},{"id":2,"body":"Second PR comment","user":"user2","created_at":"2024-01-02T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}],"total":2,"limit":10,"offset":0}`,
		},
		{
			name: "empty PR comment list",
			input: PullRequestCommentList{
				Comments: []PullRequestComment{},
				Total:    0,
				Limit:    15,
				Offset:   0,
			},
			expected: `{"comments":[],"total":0,"limit":15,"offset":0}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal PullRequestCommentList: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled PullRequestCommentList
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal PullRequestCommentList: %v", err)
			}
			if len(unmarshaled.Comments) != len(tc.input.(PullRequestCommentList).Comments) {
				t.Errorf("Expected %d comments, got %d", len(tc.input.(PullRequestCommentList).Comments), len(unmarshaled.Comments))
			}
			for i, comment := range unmarshaled.Comments {
				if comment != tc.input.(PullRequestCommentList).Comments[i] {
					t.Errorf("Comment %d: expected %+v, got %+v", i, tc.input.(PullRequestCommentList).Comments[i], comment)
				}
			}
			if unmarshaled.Total != tc.input.(PullRequestCommentList).Total ||
				unmarshaled.Limit != tc.input.(PullRequestCommentList).Limit ||
				unmarshaled.Offset != tc.input.(PullRequestCommentList).Offset {
				t.Errorf("Pagination metadata mismatch: expected %+v, got %+v",
					map[string]int{"total": tc.input.(PullRequestCommentList).Total, "limit": tc.input.(PullRequestCommentList).Limit, "offset": tc.input.(PullRequestCommentList).Offset},
					map[string]int{"total": unmarshaled.Total, "limit": unmarshaled.Limit, "offset": unmarshaled.Offset})
			}
		})
	}
}

func TestListPullRequestCommentsArgs_JSONMarshaling(t *testing.T) {
	t.Parallel()
	testCases := []dataStructureTestCase{
		{
			name: "valid PR args",
			input: ListPullRequestCommentsArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 42,
				Limit:             10,
				Offset:            0,
			},
			expected: `{"repository":"testuser/testrepo","pull_request_number":42,"limit":10,"offset":0}`,
		},
		{
			name: "default PR values",
			input: ListPullRequestCommentsArgs{
				Repository:        "owner/repo",
				PullRequestNumber: 1,
			},
			expected: `{"repository":"owner/repo","pull_request_number":1,"limit":0,"offset":0}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal ListPullRequestCommentsArgs: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(data))
			}

			var unmarshaled ListPullRequestCommentsArgs
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal ListPullRequestCommentsArgs: %v", err)
			}
			if unmarshaled != tc.input {
				t.Errorf("Expected %+v, got %+v", tc.input, unmarshaled)
			}
		})
	}
}

func TestPullRequestCommentLister_Interface(t *testing.T) {
	t.Parallel()
	// This test verifies that the interface is properly defined by checking that it can be assigned
	var _ PullRequestCommentLister = (*mockPullRequestCommentLister)(nil)
}

// mockPullRequestCommentLister is a mock implementation for testing the interface
type mockPullRequestCommentLister struct{}

func (m *mockPullRequestCommentLister) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
	return &PullRequestCommentList{}, nil
}
