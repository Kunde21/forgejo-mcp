package gitea

import (
	"context"
	"testing"

	"github.com/kunde21/forgejo-mcp/remote"
)

// This is a compile-time check - if the methods don't exist, this won't compile
var _ remote.ClientInterface = (*GiteaClient)(nil)

func TestGiteaClient_EditPullRequest_Validation(t *testing.T) {
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
			client := &GiteaClient{}
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

func TestGiteaClient_EditPullRequest_NilClient(t *testing.T) {
	t.Parallel()

	// Test that EditPullRequest handles nil client gracefully
	client := &GiteaClient{}
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
