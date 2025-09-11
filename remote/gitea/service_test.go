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

func TestServiceCreateIssueCommentWithValidation(t *testing.T) {
	// Test CreateIssueComment method with validation
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	// Test cases with expected validation results
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
			comment, err := service.CreateIssueComment(ctx, tc.repo, tc.issueNumber, tc.comment)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected validation error for case '%s', but got none", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for case '%s', got %v", tc.name, err)
				}
				if comment == nil {
					t.Error("Expected comment to be returned")
				}
				if comment.Content != tc.comment {
					t.Errorf("Expected comment content %q, got %s", tc.comment, comment.Content)
				}
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

func (m *mockGiteaClient) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
	// Mock implementation for testing - returns sample comments
	return &IssueCommentList{
		Comments: []IssueComment{
			{
				ID:      1,
				Content: "First mock comment for testing",
				Author:  "test-user",
				Created: "2025-09-10T09:00:00Z",
			},
			{
				ID:      2,
				Content: "Second mock comment for testing",
				Author:  "another-user",
				Created: "2025-09-10T10:00:00Z",
			},
		},
		Total:  2,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (m *mockGiteaClient) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*IssueComment, error) {
	// Mock implementation for testing - returns updated comment
	return &IssueComment{
		ID:      args.CommentID,
		Content: args.NewContent,
		Author:  "test-user",
		Created: "2025-09-10T10:00:00Z",
	}, nil
}

func (m *mockGiteaClient) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
	// Mock implementation for testing - returns sample pull requests
	return []PullRequest{
		{
			ID:        1,
			Number:    42,
			Title:     "Test Pull Request 1",
			Body:      "This is a test pull request",
			State:     "open",
			User:      "test-user",
			CreatedAt: "2025-09-11T10:00:00Z",
			UpdatedAt: "2025-09-11T11:00:00Z",
			Head: PullRequestBranch{
				Ref: "feature-branch",
				Sha: "abc123def456",
			},
			Base: PullRequestBranch{
				Ref: "main",
				Sha: "def456abc123",
			},
		},
		{
			ID:        2,
			Number:    43,
			Title:     "Test Pull Request 2",
			Body:      "Another test pull request",
			State:     "closed",
			User:      "another-user",
			CreatedAt: "2025-09-11T09:00:00Z",
			UpdatedAt: "2025-09-11T12:00:00Z",
			Head: PullRequestBranch{
				Ref: "bugfix-branch",
				Sha: "ghi789jkl012",
			},
			Base: PullRequestBranch{
				Ref: "main",
				Sha: "def456abc123",
			},
		},
	}, nil
}

func TestServiceEditIssueComment(t *testing.T) {
	// Test EditIssueComment method with mock client
	ctx := context.Background()

	// Create a mock client that implements GiteaClientInterface
	mockClient := &mockGiteaClient{}

	// Create service with mock client
	service := NewService(mockClient)

	// Test successful comment editing
	args := EditIssueCommentArgs{
		Repository:  "owner/repo",
		IssueNumber: 42,
		CommentID:   123,
		NewContent:  "Updated comment content",
	}
	comment, err := service.EditIssueComment(ctx, args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if comment == nil {
		t.Error("Expected comment to be returned")
	}
	if comment.Content != "Updated comment content" {
		t.Errorf("Expected comment content 'Updated comment content', got %s", comment.Content)
	}
	if comment.ID != 123 {
		t.Errorf("Expected comment ID 123, got %d", comment.ID)
	}
}

func TestServiceEditIssueCommentValidation(t *testing.T) {
	// Test validation for EditIssueComment
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	testCases := []struct {
		name        string
		args        EditIssueCommentArgs
		expectError bool
	}{
		{
			name: "valid args",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 42,
				CommentID:   123,
				NewContent:  "Updated comment content",
			},
			expectError: false,
		},
		{
			name: "empty repository",
			args: EditIssueCommentArgs{
				Repository:  "",
				IssueNumber: 42,
				CommentID:   123,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "invalid repository format",
			args: EditIssueCommentArgs{
				Repository:  "invalid-format",
				IssueNumber: 42,
				CommentID:   123,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "zero issue number",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 0,
				CommentID:   123,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "negative issue number",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: -1,
				CommentID:   123,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "zero comment ID",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 42,
				CommentID:   0,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "negative comment ID",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 42,
				CommentID:   -1,
				NewContent:  "Updated comment content",
			},
			expectError: true,
		},
		{
			name: "empty new content",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 42,
				CommentID:   123,
				NewContent:  "",
			},
			expectError: true,
		},
		{
			name: "whitespace new content",
			args: EditIssueCommentArgs{
				Repository:  "owner/repo",
				IssueNumber: 42,
				CommentID:   123,
				NewContent:  "   ",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.EditIssueComment(ctx, tc.args)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestServiceListIssueComments(t *testing.T) {
	// Test ListIssueComments method with mock client
	ctx := context.Background()

	// Create a mock client that implements GiteaClientInterface
	mockClient := &mockGiteaClient{}

	// Create service with mock client
	service := NewService(mockClient)

	// Test successful comment listing
	commentList, err := service.ListIssueComments(ctx, "owner/repo", 1, 15, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if commentList == nil {
		t.Error("Expected comment list to be returned")
	}
	if len(commentList.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(commentList.Comments))
	}
	if commentList.Total != 2 {
		t.Errorf("Expected total to be 2, got %d", commentList.Total)
	}
	if commentList.Limit != 15 {
		t.Errorf("Expected limit to be 15, got %d", commentList.Limit)
	}
	if commentList.Offset != 0 {
		t.Errorf("Expected offset to be 0, got %d", commentList.Offset)
	}
}

func TestServiceListIssueCommentsValidation(t *testing.T) {
	// Test validation for ListIssueComments
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	testCases := []struct {
		name        string
		repo        string
		issueNumber int
		limit       int
		offset      int
		expectError bool
	}{
		{"valid input", "owner/repo", 1, 15, 0, false},
		{"empty repo", "", 1, 15, 0, true},
		{"invalid repo format", "invalid-format", 1, 15, 0, true},
		{"zero issue number", "owner/repo", 0, 15, 0, true},
		{"negative issue number", "owner/repo", -1, 15, 0, true},
		{"zero limit", "owner/repo", 1, 0, 0, true},
		{"negative limit", "owner/repo", 1, -1, 0, true},
		{"excessive limit", "owner/repo", 1, 101, 0, true},
		{"negative offset", "owner/repo", 1, 15, -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ListIssueComments(ctx, tc.repo, tc.issueNumber, tc.limit, tc.offset)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestServiceListPullRequests(t *testing.T) {
	// Test ListPullRequests method with mock client
	ctx := context.Background()

	// Create a mock client that implements GiteaClientInterface
	mockClient := &mockGiteaClient{}

	// Create service with mock client
	service := NewService(mockClient)

	// Test successful pull request listing with valid options
	options := ListPullRequestsOptions{
		State:  "open",
		Limit:  10,
		Offset: 0,
	}
	prs, err := service.ListPullRequests(ctx, "owner/repo", options)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if prs == nil {
		t.Error("Expected pull requests to be returned")
	}
	if len(prs) != 2 {
		t.Errorf("Expected 2 pull requests, got %d", len(prs))
	}
	// Verify the first pull request details
	if prs[0].ID != 1 {
		t.Errorf("Expected first PR ID to be 1, got %d", prs[0].ID)
	}
	if prs[0].Number != 42 {
		t.Errorf("Expected first PR number to be 42, got %d", prs[0].Number)
	}
	if prs[0].Title != "Test Pull Request 1" {
		t.Errorf("Expected first PR title to be 'Test Pull Request 1', got %s", prs[0].Title)
	}
	if prs[0].State != "open" {
		t.Errorf("Expected first PR state to be 'open', got %s", prs[0].State)
	}
}

func TestServiceListPullRequestsValidation(t *testing.T) {
	// Test validation for ListPullRequests
	ctx := context.Background()
	mockClient := &mockGiteaClient{}
	service := NewService(mockClient)

	testCases := []struct {
		name        string
		repo        string
		options     ListPullRequestsOptions
		expectError bool
	}{
		{
			name: "valid input with open state",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  15,
				Offset: 0,
			},
			expectError: false,
		},
		{
			name: "valid input with closed state",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "closed",
				Limit:  25,
				Offset: 5,
			},
			expectError: false,
		},
		{
			name: "valid input with all state",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "all",
				Limit:  50,
				Offset: 10,
			},
			expectError: false,
		},
		{
			name: "empty repository",
			repo: "",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  15,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "invalid repository format",
			repo: "invalid-format",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  15,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "zero limit",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  0,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "negative limit",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  -1,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "excessive limit",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  101,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "negative offset",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "open",
				Limit:  15,
				Offset: -1,
			},
			expectError: true,
		},
		{
			name: "invalid state",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "invalid-state",
				Limit:  15,
				Offset: 0,
			},
			expectError: true,
		},
		{
			name: "empty state",
			repo: "owner/repo",
			options: ListPullRequestsOptions{
				State:  "",
				Limit:  15,
				Offset: 0,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ListPullRequests(ctx, tc.repo, tc.options)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}
