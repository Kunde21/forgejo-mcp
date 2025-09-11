package server

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockGiteaClientForService implements gitea.GiteaClientInterface for testing the service layer
type mockGiteaClientForService struct{}

func (m *mockGiteaClientForService) ListIssues(ctx context.Context, repo string, limit, offset int) ([]gitea.Issue, error) {
	return []gitea.Issue{}, nil
}

func (m *mockGiteaClientForService) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*gitea.IssueComment, error) {
	return &gitea.IssueComment{
		ID:      1,
		Content: comment,
		Author:  "test-user",
		Created: "2025-09-10T10:00:00Z",
	}, nil
}

func (m *mockGiteaClientForService) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*gitea.IssueCommentList, error) {
	return &gitea.IssueCommentList{
		Comments: []gitea.IssueComment{
			{
				ID:      1,
				Content: "First test comment",
				Author:  "user1",
				Created: "2025-09-10T09:00:00Z",
			},
			{
				ID:      2,
				Content: "Second test comment",
				Author:  "user2",
				Created: "2025-09-10T10:00:00Z",
			},
		},
		Total:  2,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func TestServer_handleIssueCommentList(t *testing.T) {
	// Test handleIssueCommentList handler function
	ctx := context.Background()
	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)

	server := &Server{
		giteaService: mockService,
	}

	request := &mcp.CallToolRequest{}

	// Test successful comment listing
	result, data, err := server.handleIssueCommentList(ctx, request, struct {
		Repository  string `json:"repository"`
		IssueNumber int    `json:"issue_number"`
		Limit       int    `json:"limit"`
		Offset      int    `json:"offset"`
	}{
		Repository:  "owner/repo",
		IssueNumber: 1,
		Limit:       15,
		Offset:      0,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result to be returned")
	}
	if result.IsError {
		t.Error("Expected result to not be an error")
	}
	if data == nil {
		t.Error("Expected data to be returned")
	}

	// Test data structure
	commentListResult, ok := data.(CommentListResult)
	if !ok {
		t.Error("Expected data to be CommentListResult")
	}
	if len(commentListResult.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(commentListResult.Comments))
	}
	if commentListResult.Total != 2 {
		t.Errorf("Expected total to be 2, got %d", commentListResult.Total)
	}
}

func TestServer_handleIssueCommentListValidation(t *testing.T) {
	// Test validation for handleIssueCommentList
	ctx := context.Background()
	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)
	server := &Server{
		giteaService: mockService,
	}

	request := &mcp.CallToolRequest{}

	testCases := []struct {
		name        string
		repository  string
		issueNumber int
		limit       int
		offset      int
		expectError bool
	}{
		{"valid input", "owner/repo", 1, 15, 0, false},
		{"empty repository", "", 1, 15, 0, true},
		{"zero issue number", "owner/repo", 0, 15, 0, true},
		{"negative issue number", "owner/repo", -1, 15, 0, true},
		{"zero limit", "owner/repo", 1, 0, 0, false}, // Should pass as it defaults to 15
		{"negative limit", "owner/repo", 1, -1, 0, true},
		{"excessive limit", "owner/repo", 1, 101, 0, true},
		{"negative offset", "owner/repo", 1, 15, -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := server.handleIssueCommentList(ctx, request, struct {
				Repository  string `json:"repository"`
				IssueNumber int    `json:"issue_number"`
				Limit       int    `json:"limit"`
				Offset      int    `json:"offset"`
			}{
				Repository:  tc.repository,
				IssueNumber: tc.issueNumber,
				Limit:       tc.limit,
				Offset:      tc.offset,
			})

			if tc.expectError {
				if err == nil && (result == nil || !result.IsError) {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
				if result != nil && result.IsError {
					t.Error("Expected result to not be an error")
				}
			}
		})
	}
}

func TestServer_handleIssueCommentListDefaultLimit(t *testing.T) {
	// Test that default limit is applied when not provided
	ctx := context.Background()
	mockClient := &mockGiteaClientForService{}
	mockService := gitea.NewService(mockClient)
	server := &Server{
		giteaService: mockService,
	}

	request := &mcp.CallToolRequest{}

	result, data, err := server.handleIssueCommentList(ctx, request, struct {
		Repository  string `json:"repository"`
		IssueNumber int    `json:"issue_number"`
		Limit       int    `json:"limit"`
		Offset      int    `json:"offset"`
	}{
		Repository:  "owner/repo",
		IssueNumber: 1,
		Limit:       0, // Should default to 15
		Offset:      0,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil || result.IsError {
		t.Error("Expected successful result")
	}

	// Test that the default limit was applied
	commentListResult, ok := data.(CommentListResult)
	if !ok {
		t.Error("Expected data to be CommentListResult")
	}
	if commentListResult.Limit != 15 {
		t.Errorf("Expected default limit to be 15, got %d", commentListResult.Limit)
	}
}
