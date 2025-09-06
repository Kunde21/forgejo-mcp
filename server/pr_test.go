package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TestSDKPRListHandler tests the SDK-based PR list handler
func TestSDKPRListHandler_HandlePRListRequest(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockRepos: []*gitea.Repository{
			{
				Name: testRepoName,
				Owner: &gitea.User{
					UserName: testUser,
				},
			},
		},
		MockPRs: []*gitea.PullRequest{
			{
				ID:    1,
				Index: 1,
				Title: "Test PR",
				State: gitea.StateOpen,
				Body:  "Test description",
				Poster: &gitea.User{
					UserName: testUser,
				},
			},
		},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := PRListArgs{Repository: testRepo, State: "open"}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	// Verify no error occurred
	if err != nil {
		t.Fatalf("HandlePRListRequest failed: %v", err)
	}

	// Verify result is not nil
	if result == nil {
		t.Fatal("HandlePRListRequest returned nil result")
	}

	// Verify data contains expected structure
	if data == nil {
		t.Fatal("HandlePRListRequest returned nil data")
	}
	want := map[string]any{
		"pullRequests": []map[string]any{
			{
				"author": testUser,
				"number": int64(1),
				"state":  "open",
				"title":  "Test PR",
				"type":   "pull_request",
				"url":    "",
			},
		},
		"repository": map[string]any{
			"archived":    false,
			"cloneUrl":    "",
			"description": "",
			"fork":        false,
			"forks":       0,
			"fullName":    "",
			"id":          int64(0),
			"name":        testRepoName,
			"owner": map[string]any{
				"email":    "",
				"fullName": "",
				"id":       int64(0),
				"username": testUser,
			},
			"private": false,
			"size":    0,
			"sshUrl":  "",
			"stars":   0,
			"url":     "",
		},
		"total": 1,
	}
	if !cmp.Equal(want, data) {
		t.Error(cmp.Diff(want, data))
	}
}

// TestSDKPRListHandler_EmptyResults tests handling of empty PR results
func TestSDKPRListHandler_EmptyResults(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockRepos: []*gitea.Repository{{Name: testRepoName, Owner: &gitea.User{UserName: testUser}}},
		MockPRs:   []*gitea.PullRequest{}, // Empty results
	}
	handler := &SDKPRListHandler{logger: logger, client: mockClient}

	ctx := t.Context()
	req := &mcp.CallToolRequest{}
	args := PRListArgs{Repository: testRepo, State: "closed"}
	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandlePRListRequest failed: %v", err)
	}
	want := map[string]any{
		"pullRequests": []map[string]any{},
		"repository": map[string]any{
			"archived":    false,
			"cloneUrl":    "",
			"description": "",
			"fork":        false,
			"forks":       0,
			"fullName":    "",
			"id":          int64(0),
			"name":        "test-repo",
			"owner": map[string]any{
				"email":    "",
				"fullName": "",
				"id":       int64(0),
				"username": "test-user",
			},
			"private": false,
			"size":    0,
			"sshUrl":  "",
			"stars":   0,
			"url":     "",
		},
		"total": 0,
	}
	if !cmp.Equal(want, data) {
		t.Error(cmp.Diff(want, data))
	}
	wantResult := TextResult("Found 0 pull requests")
	if !cmp.Equal(wantResult, result) {
		t.Error(cmp.Diff(wantResult, result))
	}

}

// TestSDKErrorHandling_SDKErrorTransformation tests SDK error type handling and transformation
func TestSDKErrorHandling_SDKErrorTransformation(t *testing.T) {
	tests := []struct {
		name           string
		MockError      error
		expectedError  string
		expectedLogged bool
	}{
		{
			name:           "network error",
			MockError:      fmt.Errorf("connection refused"),
			expectedError:  "executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): connection refused",
			expectedLogged: true,
		},
		{
			name:           "authentication error",
			MockError:      fmt.Errorf("401 Unauthorized"),
			expectedError:  "executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 401 Unauthorized",
			expectedLogged: true,
		},
		{
			name:           "API error",
			MockError:      fmt.Errorf("404 Not Found"),
			expectedError:  "executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 404 Not Found",
			expectedLogged: true,
		},
		{
			name:           "wrapped error",
			MockError:      fmt.Errorf("failed to connect: %w", fmt.Errorf("timeout")),
			expectedError:  "executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): failed to connect: timeout",
			expectedLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			mockClient := &giteasdk.MockGiteaClient{
				MockError: tt.MockError,
				MockRepos: []*gitea.Repository{{ID: 10, Owner: &gitea.User{UserName: testUser}, Name: testRepoName}},
			}
			handler := &SDKPRListHandler{logger: logger, client: mockClient}

			ctx := context.Background()
			req := &mcp.CallToolRequest{}
			args := PRListArgs{CWD: testRepo}

			result, data, err := handler.HandlePRListRequest(ctx, req, args)
			// Should not return an error in the function return value
			if err != nil {
				t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
			}
			// Data should be nil on error
			if data != nil {
				t.Error("HandlePRListRequest should return nil data on error")
			}
			want := TextError(errors.New(tt.expectedError))
			if !cmp.Equal(want, result) {
				t.Error(cmp.Diff(want, result))
			}
		})
	}
}

// TestSDKErrorHandling_ErrorContextPreservation tests that error context is preserved
func TestSDKErrorHandling_ErrorContextPreservation(t *testing.T) {
	logger := logrus.New()
	wrappedError := fmt.Errorf("original error: %w", fmt.Errorf("connection failed"))
	mockClient := &giteasdk.MockGiteaClient{
		MockError: wrappedError,
		MockRepos: []*gitea.Repository{{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := PRListArgs{Repository: testRepo}

	result, _, err := handler.HandlePRListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}

	textContent := result.Content[0].(*mcp.TextContent)
	errorMessage := textContent.Text

	// Verify that the full error context is preserved
	if !strings.Contains(errorMessage, "original error") {
		t.Error("Error message should contain original error context")
	}
	if !strings.Contains(errorMessage, "connection failed") {
		t.Error("Error message should contain nested error details")
	}
}

// TestSDKResponseTransformation_PRs tests PR response transformation from SDK to MCP format
func TestSDKResponseTransformation_PRs(t *testing.T) {
	logger := logrus.New()
	handler := &SDKPRListHandler{logger: logger}

	// Test data with various PR states and data completeness
	prs := []*gitea.PullRequest{
		{
			ID:      1,
			Index:   1,
			Title:   "Test PR with full data",
			State:   gitea.StateOpen,
			Body:    "Test description",
			Poster:  &gitea.User{UserName: "testuser"},
			Created: &[]time.Time{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}[0],
			Updated: &[]time.Time{time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)}[0],
			URL:     "https://example.com/pr/1",
		},
		{
			ID:      2,
			Index:   2,
			Title:   "Test PR with minimal data",
			State:   gitea.StateClosed,
			Poster:  nil, // Test nil handling
			Created: nil,
			Updated: nil,
			HTMLURL: "",
		},
		{
			ID:    3,
			Index: 3,
			Title: "Test PR with merged state",
			State: gitea.StateClosed, // Note: Gitea SDK doesn't distinguish merged vs closed
		},
	}

	result := handler.transformPRsToResponse(prs)
	want := []map[string]any{
		{
			"number":    int64(1),
			"title":     "Test PR with full data",
			"state":     "open",
			"author":    "testuser",
			"createdAt": "2023-01-01T12:00:00Z",
			"updatedAt": "2023-01-02T12:00:00Z",
			"type":      "pull_request",
			"url":       "https://example.com/pr/1",
		},
		{
			"number": int64(2),
			"title":  "Test PR with minimal data",
			"state":  "closed",
			"type":   "pull_request",
			"author": "",
			"url":    "",
		},
		{
			"number": int64(3),
			"title":  "Test PR with merged state",
			"state":  "closed",
			"type":   "pull_request",
			"author": "",
			"url":    "",
		},
	}
	// Test PR transformation
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestSDKResponseFormat_PRResponseIncludesRepositoryMetadata tests that PR responses include repository metadata
func TestSDKResponseFormat_PRResponseIncludesRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockPRs: []*gitea.PullRequest{
			{ID: 1, Index: 1, Title: "Test PR", State: gitea.StateOpen, URL: "https://localhost/issues/1", Poster: &gitea.User{UserName: "testuser"}},
		},
		MockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := PRListArgs{
		Repository: testRepo,
		State:      "open",
	}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandlePRListRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("HandlePRListRequest returned nil result")
	}
	want := map[string]any{
		"pullRequests": []map[string]any{
			{
				"type":   "pull_request",
				"url":    "https://localhost/issues/1",
				"author": "testuser",
				"number": int64(1),
				"state":  "open",
				"title":  "Test PR",
			},
		},
		"repository": map[string]any{
			"archived":    false,
			"cloneUrl":    "",
			"description": "",
			"fork":        false,
			"forks":       0,
			"fullName":    testRepo,
			"id":          int64(1),
			"name":        testRepoName,
			"owner": map[string]any{
				"email":    "",
				"fullName": "",
				"id":       int64(0),
				"username": testUser,
			},
			"private": false,
			"size":    0,
			"sshUrl":  "",
			"stars":   0,
			"url":     "",
		},
		"total": 1,
	}
	if !cmp.Equal(want, data) {
		t.Error(cmp.Diff(want, data))
	}
}

// TestSDKResponseFormat_ErrorResponseFormats tests error response formats
func TestSDKResponseFormat_ErrorResponseFormats(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockError: fmt.Errorf("repository not found"),
		MockRepos: []*gitea.Repository{
			{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName},
		},
	}

	handler := &SDKPRListHandler{logger: logger, client: mockClient}
	ctx := t.Context()
	req := &mcp.CallToolRequest{}
	args := PRListArgs{Repository: testRepo}
	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	// Should not return an error in the function return value
	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}
	if data != nil {
		t.Error("HandlePRListRequest should return nil data on error")
	}
	want := TextError(errors.New("executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=test-user, repo=test-repo): repository not found"))
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// BenchmarkSDKPerformance_PRList benchmarks SDK PR list performance
func BenchmarkSDKPerformance_PRList(b *testing.B) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockPRs: generateBenchmarkPRs(100), // Generate test data
	}
	handler := &SDKPRListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = handler.HandlePRListRequest(ctx, req, args)
	}
}

// generateBenchmarkPRs generates test PR data for benchmarking (legacy function)
func generateBenchmarkPRs(count int) []*gitea.PullRequest {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedPRs(count, options)
}
