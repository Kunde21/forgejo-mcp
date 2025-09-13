package servertest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestPullRequestCommentCreationPerformance tests performance and edge cases
// This acceptance test focuses on end-to-end performance scenarios
func TestPullRequestCommentCreationPerformance(t *testing.T) {
	t.Parallel()
	t.Skip()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}
	client := ts.Client()

	// Test large content scenario - should handle efficiently
	largeComment := strings.Repeat("This is a detailed code review comment with comprehensive feedback. ", 200) // ~10KB
	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: "pr_comment_create",
		Arguments: map[string]any{
			"repository":          "testuser/testrepo",
			"pull_request_number": 1,
			"comment":             largeComment,
		},
	})
	if err != nil {
		t.Fatalf("Failed to call pr_comment_create tool with large content: %v", err)
	}
	opt := cmp.FilterPath(func(p cmp.Path) bool {
		return p.Last().String() == ".Text"
	}, cmp.Comparer(func(a, b string) bool {
		prefix, test := a, b
		if len(b) < len(a) {
			prefix, test = b, a
		}
		return strings.HasPrefix(test, prefix)
	}))
	want := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Pull request comment created successfully"},
		},
		StructuredContent: map[string]any{
			"body":       largeComment,
			"created_at": "2025-09-10T10:00:00Z",
			"id":         float64(123),
			"updated_at": "2025-09-10T10:00:00Z",
			"user":       "testuser",
		},
		IsError: false,
	}
	if !cmp.Equal(want, result, opt) {
		t.Error(cmp.Diff(want, result, opt))
	}

	if result == nil || result.IsError {
		t.Fatal("Expected successful result with large content")
	}

	// Verify the result contains the large content
	if len(result.Content) == 0 {
		t.Fatal("Expected result content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Pull request comment created successfully") {
		t.Error("Expected success message for large content")
	}
}

// TestPullRequestCommentCreationConcurrentDifferentPRs tests concurrent request handling on different PRs
// This acceptance test focuses on end-to-end concurrent behavior across different pull requests
func TestPullRequestCommentCreationConcurrentDifferentPRs(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)

	mock := NewMockGiteaServer(t)
	ts := NewTestServer(t, ctx, map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		prNumber := i + 1 // Each goroutine comments on a different PR
		go func(prNum int) {
			defer wg.Done()
			_, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
				Name: "pr_comment_create",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": prNum,
					"comment":             fmt.Sprintf("Concurrent comment on PR %d", prNum),
				},
			})
			results <- err
		}(prNumber)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check results
	for err := range results {
		if err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}

func TestCreatePullRequestCommentConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddComments("testuser", "testrepo", []MockComment{}) // Start with no comments
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	results := make(chan error, numGoroutines)
	for i := range numGoroutines {
		go func(commentNum int) {
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "pr_comment_create",
				Arguments: map[string]any{
					"repository":          "testuser/testrepo",
					"pull_request_number": 1,
					"comment":             "Concurrent comment " + string(rune(commentNum+'0')),
				},
			})
			results <- err
		}(i)
	}
	for range numGoroutines {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
