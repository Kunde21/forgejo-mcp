package servertest

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIssueCommentEditWithInvalidToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "invalid-token", // Use invalid token
	})
	if err := ts.Initialize(); err != nil {
		t.Fatal(err)
	}
	client := ts.Client()

	// Test with invalid token (should fail)
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_comment_edit",
		Arguments: map[string]any{
			"repository":   "testuser/testrepo",
			"issue_number": 1,
			"comment_id":   123,
			"new_content":  "Updated content",
		},
	})

	// Should fail with invalid token
	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if result != nil && !result.IsError {
		t.Error("Expected error result for invalid token")
	}
	
	t.Logf("Result with invalid token: err=%v, result.IsError=%v", err, result != nil && result.IsError)
}
