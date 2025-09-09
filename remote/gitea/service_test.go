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

func TestServiceCreateIssueCommentValidation(t *testing.T) {
	// Test validation for CreateIssueComment
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	testCases := []struct {
		name        string
		repo        string
		issueNumber int
		comment     string
		expectError bool
	}{
		{"valid input", "owner/repo", 1, "Valid comment", false},
		{"empty repo", "", 1, "Comment", true},
		{"invalid repo format", "invalid-format", 1, "Comment", true},
		{"zero issue number", "owner/repo", 0, "Comment", true},
		{"negative issue number", "owner/repo", -1, "Comment", true},
		{"empty comment", "owner/repo", 1, "", true},
		{"whitespace comment", "owner/repo", 1, "   ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CreateIssueComment(ctx, tc.repo, tc.issueNumber, tc.comment)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestServiceValidateIssueNumber(t *testing.T) {
	// Test issue number validation
	service := &Service{}

	testCases := []struct {
		issueNumber int
		expectError bool
	}{
		{1, false},
		{100, false},
		{0, true},
		{-1, true},
		{-100, true},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			err := service.validateIssueNumber(tc.issueNumber)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for issue number %d", tc.issueNumber)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for issue number %d, got %v", tc.issueNumber, err)
			}
		})
	}
}

func TestServiceValidateCommentContent(t *testing.T) {
	// Test comment content validation
	service := &Service{}

	testCases := []struct {
		comment     string
		expectError bool
	}{
		{"Valid comment", false},
		{"Another valid comment", false},
		{"", true},
		{"   ", true},
		{"\t\n", true},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			err := service.validateCommentContent(tc.comment)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for comment %q", tc.comment)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for comment %q, got %v", tc.comment, err)
			}
		})
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
