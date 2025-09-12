package gitea

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type mockGiteaClient struct {
	listPullRequestCommentsFunc  func(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error)
	createPullRequestCommentFunc func(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error)
	editPullRequestCommentFunc   func(ctx context.Context, args EditPullRequestCommentArgs) (*PullRequestComment, error)
}

func (m *mockGiteaClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
	return nil, nil
}

func (m *mockGiteaClient) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*IssueComment, error) {
	return nil, nil
}

func (m *mockGiteaClient) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
	return nil, nil
}

func (m *mockGiteaClient) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*IssueComment, error) {
	return nil, nil
}

func (m *mockGiteaClient) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
	return nil, nil
}

func (m *mockGiteaClient) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
	if m.listPullRequestCommentsFunc != nil {
		return m.listPullRequestCommentsFunc(ctx, repo, pullRequestNumber, limit, offset)
	}
	return &PullRequestCommentList{}, nil
}

func (m *mockGiteaClient) CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error) {
	if m.createPullRequestCommentFunc != nil {
		return m.createPullRequestCommentFunc(ctx, repo, pullRequestNumber, comment)
	}
	return &PullRequestComment{
		ID:        1,
		Content:   comment,
		Author:    "testuser",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}, nil
}

func (m *mockGiteaClient) EditPullRequestComment(ctx context.Context, args EditPullRequestCommentArgs) (*PullRequestComment, error) {
	if m.editPullRequestCommentFunc != nil {
		return m.editPullRequestCommentFunc(ctx, args)
	}
	return &PullRequestComment{
		ID:        args.CommentID,
		Content:   args.NewContent,
		Author:    "testuser",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}, nil
}

func TestService_ListPullRequestComments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		repo              string
		pullRequestNumber int
		limit             int
		offset            int
		mockResponse      *PullRequestCommentList
		mockError         error
		expectedError     string
		expectedResult    *PullRequestCommentList
	}{
		{
			name:              "valid request",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			mockResponse: &PullRequestCommentList{
				Comments: []PullRequestComment{
					{
						ID:        1,
						Content:   "Test comment",
						Author:    "testuser",
						CreatedAt: "2024-01-01T00:00:00Z",
						UpdatedAt: "2024-01-01T00:00:00Z",
					},
				},
				Total:  1,
				Limit:  10,
				Offset: 0,
			},
			expectedResult: &PullRequestCommentList{
				Comments: []PullRequestComment{
					{
						ID:        1,
						Content:   "Test comment",
						Author:    "testuser",
						CreatedAt: "2024-01-01T00:00:00Z",
						UpdatedAt: "2024-01-01T00:00:00Z",
					},
				},
				Total:  1,
				Limit:  10,
				Offset: 0,
			},
		},
		{
			name:              "invalid repository format",
			repo:              "invalid-repo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			expectedError:     "repository validation failed: repository must be in format 'owner/repo'",
		},
		{
			name:              "empty repository",
			repo:              "",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			expectedError:     "repository validation failed: repository cannot be empty",
		},
		{
			name:              "invalid pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: 0,
			limit:             10,
			offset:            0,
			expectedError:     "pull request number validation failed: pull request number must be positive",
		},
		{
			name:              "negative pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: -1,
			limit:             10,
			offset:            0,
			expectedError:     "pull request number validation failed: pull request number must be positive",
		},
		{
			name:              "invalid limit - too low",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             0,
			offset:            0,
			expectedError:     "pagination validation failed: limit must be between 1 and 100",
		},
		{
			name:              "invalid limit - too high",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             200,
			offset:            0,
			expectedError:     "pagination validation failed: limit must be between 1 and 100",
		},
		{
			name:              "invalid offset - negative",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            -1,
			expectedError:     "pagination validation failed: offset must be non-negative",
		},
		{
			name:              "empty result",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			mockResponse: &PullRequestCommentList{
				Comments: []PullRequestComment{},
				Total:    0,
				Limit:    10,
				Offset:   0,
			},
			expectedResult: &PullRequestCommentList{
				Comments: []PullRequestComment{},
				Total:    0,
				Limit:    10,
				Offset:   0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockGiteaClient{
				listPullRequestCommentsFunc: func(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
					return tc.mockResponse, tc.mockError
				},
			}

			service := NewService(mockClient)
			result, err := service.ListPullRequestComments(context.Background(), tc.repo, tc.pullRequestNumber, tc.limit, tc.offset)

			if tc.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tc.expectedError)
				} else if err.Error() != tc.expectedError {
					t.Errorf("Expected error %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil && tc.expectedResult != nil {
				t.Errorf("Expected result, got nil")
				return
			}

			if result != nil && tc.expectedResult == nil {
				t.Errorf("Expected nil result, got %v", result)
				return
			}

			if len(result.Comments) != len(tc.expectedResult.Comments) {
				t.Errorf("Expected %d comments, got %d", len(tc.expectedResult.Comments), len(result.Comments))
			}

			for i, expectedComment := range tc.expectedResult.Comments {
				if i >= len(result.Comments) {
					t.Errorf("Missing comment at index %d", i)
					continue
				}
				actualComment := result.Comments[i]
				if actualComment != expectedComment {
					t.Errorf("Comment %d mismatch: expected %+v, got %+v", i, expectedComment, actualComment)
				}
			}

			if result.Total != tc.expectedResult.Total ||
				result.Limit != tc.expectedResult.Limit ||
				result.Offset != tc.expectedResult.Offset {
				t.Errorf("Pagination metadata mismatch: expected %+v, got %+v",
					map[string]int{"total": tc.expectedResult.Total, "limit": tc.expectedResult.Limit, "offset": tc.expectedResult.Offset},
					map[string]int{"total": result.Total, "limit": result.Limit, "offset": result.Offset})
			}
		})
	}
}

func TestService_CreatePullRequestComment(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		repo              string
		pullRequestNumber int
		comment           string
		mockResponse      *PullRequestComment
		mockError         error
		expectedError     string
		expectedResult    *PullRequestComment
	}{
		{
			name:              "successful comment creation",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			comment:           "This is a test comment",
			mockResponse: &PullRequestComment{
				ID:        123,
				Content:   "This is a test comment",
				Author:    "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
			},
			expectedResult: &PullRequestComment{
				ID:        123,
				Content:   "This is a test comment",
				Author:    "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
			},
		},
		{
			name:              "client error",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			comment:           "Test comment",
			mockError:         fmt.Errorf("failed to create pull request comment"),
			expectedError:     "failed to create pull request comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockGiteaClient{
				createPullRequestCommentFunc: func(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error) {
					return tc.mockResponse, tc.mockError
				},
			}

			service := NewService(mockClient)
			result, err := service.CreatePullRequestComment(context.Background(), tc.repo, tc.pullRequestNumber, tc.comment)

			if tc.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil && tc.expectedResult != nil {
				t.Errorf("Expected result, got nil")
				return
			}

			if result != nil && tc.expectedResult == nil {
				t.Errorf("Expected nil result, got %v", result)
				return
			}

			if result.ID != tc.expectedResult.ID ||
				result.Content != tc.expectedResult.Content ||
				result.Author != tc.expectedResult.Author ||
				result.CreatedAt != tc.expectedResult.CreatedAt ||
				result.UpdatedAt != tc.expectedResult.UpdatedAt {
				t.Errorf("Result mismatch: expected %+v, got %+v", tc.expectedResult, result)
			}
		})
	}
}

func TestService_EditPullRequestComment(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		args           EditPullRequestCommentArgs
		mockResponse   *PullRequestComment
		mockError      error
		expectedError  string
		expectedResult *PullRequestComment
	}{
		{
			name: "successful comment editing",
			args: EditPullRequestCommentArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 1,
				CommentID:         123,
				NewContent:        "Updated comment content",
			},
			mockResponse: &PullRequestComment{
				ID:        123,
				Content:   "Updated comment content",
				Author:    "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-02T00:00:00Z",
			},
			expectedResult: &PullRequestComment{
				ID:        123,
				Content:   "Updated comment content",
				Author:    "testuser",
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-02T00:00:00Z",
			},
		},
		{
			name: "client error",
			args: EditPullRequestCommentArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 1,
				CommentID:         123,
				NewContent:        "Updated content",
			},
			mockError:     fmt.Errorf("failed to edit pull request comment"),
			expectedError: "failed to edit pull request comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockGiteaClient{
				editPullRequestCommentFunc: func(ctx context.Context, args EditPullRequestCommentArgs) (*PullRequestComment, error) {
					return tc.mockResponse, tc.mockError
				},
			}

			service := NewService(mockClient)
			result, err := service.EditPullRequestComment(context.Background(), tc.args)

			if tc.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil && tc.expectedResult != nil {
				t.Errorf("Expected result, got nil")
				return
			}

			if result != nil && tc.expectedResult == nil {
				t.Errorf("Expected nil result, got %v", result)
				return
			}

			if result.ID != tc.expectedResult.ID ||
				result.Content != tc.expectedResult.Content ||
				result.Author != tc.expectedResult.Author ||
				result.CreatedAt != tc.expectedResult.CreatedAt ||
				result.UpdatedAt != tc.expectedResult.UpdatedAt {
				t.Errorf("Result mismatch: expected %+v, got %+v", tc.expectedResult, result)
			}
		})
	}
}
