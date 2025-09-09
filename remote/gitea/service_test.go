package gitea

import (
	"context"
	"testing"
)

func TestServiceCreateIssueComment(t *testing.T) {
	// Test CreateIssueComment method with mock client
	ctx := context.Background()

	// Create a mock client that implements GiteaClientInterface
	mockClient := &mockGiteaClient{}

	// Create service with mock client
	service := NewService(mockClient)

	// Test successful comment creation
	comment, err := service.CreateIssueComment(ctx, "owner/repo", 1, "Test comment")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if comment == nil {
		t.Error("Expected comment to be returned")
	}
	if comment.Content != "Test comment" {
		t.Errorf("Expected comment content 'Test comment', got %s", comment.Content)
	}
}

func TestServiceCreateIssueCommentWithoutValidation(t *testing.T) {
	// Test CreateIssueComment method without validation
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	// Test successful comment creation - service layer no longer validates
	testCases := []struct {
		name        string
		repo        string
		issueNumber int
		comment     string
	}{
		{"valid input", "owner/repo", 1, "Valid comment"},
		{"empty repo", "", 1, "Comment"},
		{"invalid repo format", "invalid-format", 1, "Comment"},
		{"zero issue number", "owner/repo", 0, "Comment"},
		{"negative issue number", "owner/repo", -1, "Comment"},
		{"empty comment", "owner/repo", 1, ""},
		{"whitespace comment", "owner/repo", 1, "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comment, err := service.CreateIssueComment(ctx, tc.repo, tc.issueNumber, tc.comment)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if comment == nil {
				t.Error("Expected comment to be returned")
			}
			if comment.Content != tc.comment {
				t.Errorf("Expected comment content %q, got %s", tc.comment, comment.Content)
			}
		})
	}
}

func TestServiceListIssues(t *testing.T) {
	// Test ListIssues method without validation
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	// Test successful issue listing
	issues, err := service.ListIssues(ctx, "owner/repo", 10, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if issues == nil {
		t.Error("Expected issues to be returned")
	}
}

// mockGiteaClient implements GiteaClientInterface for testing
type mockGiteaClient struct{}

func (m *mockGiteaClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
	return []Issue{}, nil
}

func (m *mockGiteaClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error) {
	return &IssueComment{
		ID:      1,
		Content: comment,
		Author:  "test-user",
		Created: "2025-09-09T10:00:00Z",
	}, nil
}
