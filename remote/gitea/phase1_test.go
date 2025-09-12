package gitea

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPhase1_InterfaceImplementation(t *testing.T) {
	t.Parallel()

	// Test that EditPullRequestCommentArgs can be marshaled/unmarshaled
	args := EditPullRequestCommentArgs{
		Repository:        "testuser/testrepo",
		PullRequestNumber: 42,
		CommentID:         123,
		NewContent:        "Updated comment content",
	}

	data, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("Failed to marshal EditPullRequestCommentArgs: %v", err)
	}

	expected := `{"repository":"testuser/testrepo","pull_request_number":42,"comment_id":123,"new_content":"Updated comment content"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	var unmarshaled EditPullRequestCommentArgs
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal EditPullRequestCommentArgs: %v", err)
	}
	if unmarshaled != args {
		t.Errorf("Expected %+v, got %+v", args, unmarshaled)
	}
}

func TestPhase1_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Test that our mock implements the interface
	var _ PullRequestCommentEditor = (*phase1MockPullRequestCommentEditor)(nil)
}

type phase1MockPullRequestCommentEditor struct{}

func (m *phase1MockPullRequestCommentEditor) EditPullRequestComment(ctx context.Context, args EditPullRequestCommentArgs) (*PullRequestComment, error) {
	return &PullRequestComment{}, nil
}
