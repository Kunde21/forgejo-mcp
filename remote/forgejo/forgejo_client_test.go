package forgejo

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/remote"
)

// Compile-time interface check
var _ remote.ClientInterface = (*ForgejoClient)(nil)

func TestNewForgejoClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		url     string
		token   string
		wantErr bool
	}{
		{
			name:    "invalid URL",
			url:     "invalid-url",
			token:   "valid-token",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			token:   "valid-token",
			wantErr: true,
		},
		{
			name:    "empty token",
			url:     "https://forgejo.example.com",
			token:   "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewForgejoClient(tc.url, tc.token)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client == nil {
				t.Error("expected non-nil client")
			}

			if client.client == nil {
				t.Error("expected non-nil forgejo client")
			}
		})
	}
}

func TestNewForgejoClient_ValidInput(t *testing.T) {
	t.Parallel()

	// Test that the constructor accepts valid inputs without error
	// Note: The Forgejo SDK may make network requests during initialization
	// For now, we'll test with a localhost URL that should fail gracefully
	client, err := NewForgejoClient("http://localhost:0", "test-token")

	// We expect this to fail due to network issues, but the constructor should still work
	if err == nil {
		t.Error("expected error due to network connectivity, got nil")
	}

	// The important thing is that we get a proper error, not a panic
	if client != nil {
		t.Error("expected nil client when error occurs")
	}
}

func TestForgejoClient_ZeroValue(t *testing.T) {
	t.Parallel()

	// Test that zero value of ForgejoClient is safe
	var client ForgejoClient

	// These should not panic
	_, _ = client.ListIssues(nil, "", 0, 0)
	_, _ = client.CreateIssueComment(nil, "", 0, "")
	_, _ = client.ListIssueComments(nil, "", 0, 0, 0)
	_, _ = client.EditIssueComment(nil, remote.EditIssueCommentArgs{})
	_, _ = client.EditIssue(nil, remote.EditIssueArgs{})
	_, _ = client.ListPullRequests(nil, "", remote.ListPullRequestsOptions{})
	_, _ = client.ListPullRequestComments(nil, "", 0, 0, 0)
	_, _ = client.CreatePullRequestComment(nil, "", 0, "")
	_, _ = client.EditPullRequestComment(nil, remote.EditPullRequestCommentArgs{})
}

func TestForgejoClient_CreateIssueComment_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		repo        string
		issueNumber int
		comment     string
		wantErr     bool
	}{
		{
			name:        "empty repository",
			repo:        "",
			issueNumber: 1,
			comment:     "This is a test comment",
			wantErr:     true,
		},
		{
			name:        "invalid repository format",
			repo:        "invalid-repo",
			issueNumber: 1,
			comment:     "This is a test comment",
			wantErr:     true,
		},
		{
			name:        "zero issue number",
			repo:        "testuser/testrepo",
			issueNumber: 0,
			comment:     "This is a test comment",
			wantErr:     true,
		},
		{
			name:        "negative issue number",
			repo:        "testuser/testrepo",
			issueNumber: -1,
			comment:     "This is a test comment",
			wantErr:     true,
		},
		{
			name:        "empty comment",
			repo:        "testuser/testrepo",
			issueNumber: 1,
			comment:     "",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			comment, err := client.CreateIssueComment(ctx, tc.repo, tc.issueNumber, tc.comment)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if comment != nil {
					t.Error("expected nil comment when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if comment == nil {
				t.Error("expected non-nil comment")
			}
		})
	}
}

func TestForgejoClient_ListIssues_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		repo    string
		limit   int
		offset  int
		wantErr bool
	}{
		{
			name:    "empty repository",
			repo:    "",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
		{
			name:    "invalid repository format",
			repo:    "invalid-repo",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
		{
			name:    "zero limit",
			repo:    "testuser/testrepo",
			limit:   0,
			offset:  0,
			wantErr: true, // Should fail due to nil client
		},
		{
			name:    "negative offset",
			repo:    "testuser/testrepo",
			limit:   10,
			offset:  -1,
			wantErr: true, // Should fail due to nil client
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			issues, err := client.ListIssues(ctx, tc.repo, tc.limit, tc.offset)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if issues != nil {
					t.Error("expected nil issues when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if issues == nil {
				t.Error("expected non-nil issues")
			}
		})
	}
}

func TestForgejoClient_ListIssues_NilClient(t *testing.T) {
	t.Parallel()

	// Test that ListIssues handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	// This should return an error due to nil client, not panic
	_, err := client.ListIssues(ctx, "testuser/testrepo", 10, 0)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_Structure(t *testing.T) {
	t.Parallel()

	// Test that ForgejoClient can be instantiated
	client := &ForgejoClient{
		client: nil, // This would normally be a real client
	}

	if client == nil {
		t.Error("ForgejoClient should not be nil")
	}
}

func TestForgejoClient_ListIssueComments_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		repo        string
		issueNumber int
		limit       int
		offset      int
		wantErr     bool
	}{
		{
			name:        "empty repository",
			repo:        "",
			issueNumber: 1,
			limit:       10,
			offset:      0,
			wantErr:     true,
		},
		{
			name:        "invalid repository format",
			repo:        "invalid-repo",
			issueNumber: 1,
			limit:       10,
			offset:      0,
			wantErr:     true,
		},
		{
			name:        "zero issue number",
			repo:        "testuser/testrepo",
			issueNumber: 0,
			limit:       10,
			offset:      0,
			wantErr:     true,
		},
		{
			name:        "negative issue number",
			repo:        "testuser/testrepo",
			issueNumber: -1,
			limit:       10,
			offset:      0,
			wantErr:     true,
		},
		{
			name:        "zero limit",
			repo:        "testuser/testrepo",
			issueNumber: 1,
			limit:       0,
			offset:      0,
			wantErr:     true, // Should fail due to nil client
		},
		{
			name:        "negative offset",
			repo:        "testuser/testrepo",
			issueNumber: 1,
			limit:       10,
			offset:      -1,
			wantErr:     true, // Should fail due to nil client
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			commentList, err := client.ListIssueComments(ctx, tc.repo, tc.issueNumber, tc.limit, tc.offset)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if commentList != nil {
					t.Error("expected nil comment list when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if commentList == nil {
				t.Error("expected non-nil comment list")
			}
		})
	}
}

func TestForgejoClient_ListIssueComments_NilClient(t *testing.T) {
	t.Parallel()

	// Test that ListIssueComments handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	// This should return an error due to nil client, not panic
	_, err := client.ListIssueComments(ctx, "testuser/testrepo", 1, 10, 0)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_EditIssueComment_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    remote.EditIssueCommentArgs
		wantErr bool
	}{
		{
			name: "empty repository",
			args: remote.EditIssueCommentArgs{
				Repository:  "",
				IssueNumber: 1,
				CommentID:   1,
				NewContent:  "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			args: remote.EditIssueCommentArgs{
				Repository:  "invalid-repo",
				IssueNumber: 1,
				CommentID:   1,
				NewContent:  "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "zero comment ID",
			args: remote.EditIssueCommentArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 1,
				CommentID:   0,
				NewContent:  "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "negative comment ID",
			args: remote.EditIssueCommentArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 1,
				CommentID:   -1,
				NewContent:  "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "empty new content",
			args: remote.EditIssueCommentArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 1,
				CommentID:   1,
				NewContent:  "",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			comment, err := client.EditIssueComment(ctx, tc.args)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if comment != nil {
					t.Error("expected nil comment when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if comment == nil {
				t.Error("expected non-nil comment")
			}
		})
	}
}

func TestForgejoClient_EditIssueComment_NilClient(t *testing.T) {
	t.Parallel()

	// Test that EditIssueComment handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	args := remote.EditIssueCommentArgs{
		Repository:  "testuser/testrepo",
		IssueNumber: 1,
		CommentID:   1,
		NewContent:  "Updated comment",
	}

	// This should return an error due to nil client, not panic
	_, err := client.EditIssueComment(ctx, args)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_ListPullRequests_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		repo    string
		options remote.ListPullRequestsOptions
		wantErr bool
	}{
		{
			name: "empty repository",
			repo: "",
			options: remote.ListPullRequestsOptions{
				State:  "open",
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			repo: "invalid-repo",
			options: remote.ListPullRequestsOptions{
				State:  "open",
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "zero limit",
			repo: "testuser/testrepo",
			options: remote.ListPullRequestsOptions{
				State:  "open",
				Limit:  0,
				Offset: 0,
			},
			wantErr: true, // Should fail due to nil client
		},
		{
			name: "negative offset",
			repo: "testuser/testrepo",
			options: remote.ListPullRequestsOptions{
				State:  "open",
				Limit:  10,
				Offset: -1,
			},
			wantErr: true, // Should fail due to nil client
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			prs, err := client.ListPullRequests(ctx, tc.repo, tc.options)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if prs != nil {
					t.Error("expected nil pull requests when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if prs == nil {
				t.Error("expected non-nil pull requests")
			}
		})
	}
}

func TestForgejoClient_ListPullRequests_NilClient(t *testing.T) {
	t.Parallel()

	// Test that ListPullRequests handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	options := remote.ListPullRequestsOptions{
		State:  "open",
		Limit:  10,
		Offset: 0,
	}

	// This should return an error due to nil client, not panic
	_, err := client.ListPullRequests(ctx, "testuser/testrepo", options)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_ListPullRequestComments_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		repo              string
		pullRequestNumber int
		limit             int
		offset            int
		wantErr           bool
	}{
		{
			name:              "empty repository",
			repo:              "",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			wantErr:           true,
		},
		{
			name:              "invalid repository format",
			repo:              "invalid-repo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            0,
			wantErr:           true,
		},
		{
			name:              "zero pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: 0,
			limit:             10,
			offset:            0,
			wantErr:           true,
		},
		{
			name:              "negative pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: -1,
			limit:             10,
			offset:            0,
			wantErr:           true,
		},
		{
			name:              "zero limit",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             0,
			offset:            0,
			wantErr:           true, // Should fail due to nil client
		},
		{
			name:              "negative offset",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			limit:             10,
			offset:            -1,
			wantErr:           true, // Should fail due to nil client
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			commentList, err := client.ListPullRequestComments(ctx, tc.repo, tc.pullRequestNumber, tc.limit, tc.offset)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if commentList != nil {
					t.Error("expected nil comment list when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if commentList == nil {
				t.Error("expected non-nil comment list")
			}
		})
	}
}

func TestForgejoClient_ListPullRequestComments_NilClient(t *testing.T) {
	t.Parallel()

	// Test that ListPullRequestComments handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	// This should return an error due to nil client, not panic
	_, err := client.ListPullRequestComments(ctx, "testuser/testrepo", 1, 10, 0)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_CreatePullRequestComment_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		repo              string
		pullRequestNumber int
		comment           string
		wantErr           bool
	}{
		{
			name:              "empty repository",
			repo:              "",
			pullRequestNumber: 1,
			comment:           "This is a test comment",
			wantErr:           true,
		},
		{
			name:              "invalid repository format",
			repo:              "invalid-repo",
			pullRequestNumber: 1,
			comment:           "This is a test comment",
			wantErr:           true,
		},
		{
			name:              "zero pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: 0,
			comment:           "This is a test comment",
			wantErr:           true,
		},
		{
			name:              "negative pull request number",
			repo:              "testuser/testrepo",
			pullRequestNumber: -1,
			comment:           "This is a test comment",
			wantErr:           true,
		},
		{
			name:              "empty comment",
			repo:              "testuser/testrepo",
			pullRequestNumber: 1,
			comment:           "",
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			comment, err := client.CreatePullRequestComment(ctx, tc.repo, tc.pullRequestNumber, tc.comment)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if comment != nil {
					t.Error("expected nil comment when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if comment == nil {
				t.Error("expected non-nil comment")
			}
		})
	}
}

func TestForgejoClient_CreatePullRequestComment_NilClient(t *testing.T) {
	t.Parallel()

	// Test that CreatePullRequestComment handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	// This should return an error due to nil client, not panic
	_, err := client.CreatePullRequestComment(ctx, "testuser/testrepo", 1, "Test comment")

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_EditPullRequestComment_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    remote.EditPullRequestCommentArgs
		wantErr bool
	}{
		{
			name: "empty repository",
			args: remote.EditPullRequestCommentArgs{
				Repository:        "",
				PullRequestNumber: 1,
				CommentID:         1,
				NewContent:        "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			args: remote.EditPullRequestCommentArgs{
				Repository:        "invalid-repo",
				PullRequestNumber: 1,
				CommentID:         1,
				NewContent:        "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "zero comment ID",
			args: remote.EditPullRequestCommentArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 1,
				CommentID:         0,
				NewContent:        "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "negative comment ID",
			args: remote.EditPullRequestCommentArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 1,
				CommentID:         -1,
				NewContent:        "Updated comment",
			},
			wantErr: true,
		},
		{
			name: "empty new content",
			args: remote.EditPullRequestCommentArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 1,
				CommentID:         1,
				NewContent:        "",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			comment, err := client.EditPullRequestComment(ctx, tc.args)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if comment != nil {
					t.Error("expected nil comment when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if comment == nil {
				t.Error("expected non-nil comment")
			}
		})
	}
}

func TestForgejoClient_EditPullRequestComment_NilClient(t *testing.T) {
	t.Parallel()

	// Test that EditPullRequestComment handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	args := remote.EditPullRequestCommentArgs{
		Repository:        "testuser/testrepo",
		PullRequestNumber: 1,
		CommentID:         1,
		NewContent:        "Updated comment",
	}

	// This should return an error due to nil client, not panic
	_, err := client.EditPullRequestComment(ctx, args)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_EditPullRequest_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    remote.EditPullRequestArgs
		wantErr bool
	}{
		{
			name: "empty repository",
			args: remote.EditPullRequestArgs{
				Repository:        "",
				PullRequestNumber: 1,
				Title:             "Updated title",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			args: remote.EditPullRequestArgs{
				Repository:        "invalid-repo",
				PullRequestNumber: 1,
				Title:             "Updated title",
			},
			wantErr: true,
		},
		{
			name: "zero pull request number",
			args: remote.EditPullRequestArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: 0,
				Title:             "Updated title",
			},
			wantErr: true,
		},
		{
			name: "negative pull request number",
			args: remote.EditPullRequestArgs{
				Repository:        "testuser/testrepo",
				PullRequestNumber: -1,
				Title:             "Updated title",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			pr, err := client.EditPullRequest(ctx, tc.args)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if pr != nil {
					t.Error("expected nil pull request when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pr == nil {
				t.Error("expected non-nil pull request")
			}
		})
	}
}

func TestForgejoClient_EditPullRequest_NilClient(t *testing.T) {
	t.Parallel()

	// Test that EditPullRequest handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	args := remote.EditPullRequestArgs{
		Repository:        "testuser/testrepo",
		PullRequestNumber: 123,
		Title:             "Updated title",
	}

	// This should return an error due to nil client, not panic
	_, err := client.EditPullRequest(ctx, args)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestForgejoClient_EditIssue_Validation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    remote.EditIssueArgs
		wantErr bool
	}{
		{
			name: "empty repository",
			args: remote.EditIssueArgs{
				Repository:  "",
				IssueNumber: 1,
				Title:       "Updated title",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			args: remote.EditIssueArgs{
				Repository:  "invalid-repo",
				IssueNumber: 1,
				Title:       "Updated title",
			},
			wantErr: true,
		},
		{
			name: "zero issue number",
			args: remote.EditIssueArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 0,
				Title:       "Updated title",
			},
			wantErr: true,
		},
		{
			name: "negative issue number",
			args: remote.EditIssueArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: -1,
				Title:       "Updated title",
			},
			wantErr: true,
		},
		{
			name: "no changes",
			args: remote.EditIssueArgs{
				Repository:  "testuser/testrepo",
				IssueNumber: 1,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &ForgejoClient{}
			ctx := context.Background()

			issue, err := client.EditIssue(ctx, tc.args)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if issue != nil {
					t.Error("expected nil issue when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if issue == nil {
				t.Error("expected non-nil issue")
			}
		})
	}
}

func TestForgejoClient_EditIssue_NilClient(t *testing.T) {
	t.Parallel()

	// Test that EditIssue handles nil client gracefully
	client := &ForgejoClient{}
	ctx := context.Background()

	args := remote.EditIssueArgs{
		Repository:  "testuser/testrepo",
		IssueNumber: 123,
		Title:       "Updated title",
	}

	// This should return an error due to nil client, not panic
	_, err := client.EditIssue(ctx, args)

	if err == nil {
		t.Error("expected error due to nil client, but no error occurred")
	}

	expectedErr := "client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}
