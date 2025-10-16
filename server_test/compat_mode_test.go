package servertest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCompatModeResponseFormat(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})

	// Test with compat mode enabled
	tsCompat := NewTestServerWithCompat(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, true)
	if err := tsCompat.Initialize(); err != nil {
		t.Fatal(err)
	}

	resultCompat, err := tsCompat.Client().CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list in compat mode: %v", err)
	}

	// Test with compat mode disabled
	tsRegular := NewTestServerWithCompat(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	}, false)
	if err := tsRegular.Initialize(); err != nil {
		t.Fatal(err)
	}

	resultRegular, err := tsRegular.Client().CallTool(ctx, &mcp.CallToolParams{
		Name: "issue_list",
		Arguments: map[string]any{
			"repository": "testuser/testrepo",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call issue_list in regular mode: %v", err)
	}

	// Verify text responses differ
	textCompat := GetTextContent(resultCompat.Content)
	textRegular := GetTextContent(resultRegular.Content)

	if !strings.Contains(textCompat, "Found 1 issues:") {
		t.Errorf("Compat mode should include detailed formatting, got: %s", textCompat)
	}

	if !strings.Contains(textRegular, "Found 1 issues") {
		t.Errorf("Regular mode should include simple summary, got: %s", textRegular)
	}

	if strings.Contains(textRegular, "#1: Test Issue (open)") {
		t.Errorf("Regular mode should not include detailed formatting")
	}

	// Verify structured data is identical
	structCompat := GetStructuredContent(resultCompat)
	structRegular := GetStructuredContent(resultRegular)
	if diff := cmp.Diff(structCompat, structRegular); diff != "" {
		t.Errorf("Structured data should be identical (-compat +regular):\n%s", diff)
	}
}
