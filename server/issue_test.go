package server

import (
	"context"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TestSDKResponseTransformation_Issues tests issue response transformation from SDK to MCP format
func TestSDKResponseTransformation_Issues(t *testing.T) {
	logger := logrus.New()
	handler := &SDKIssueListHandler{logger: logger}
	issues := []*gitea.Issue{
		{
			ID:      1,
			Index:   1,
			Title:   "Test Issue with full data",
			State:   "open",
			Body:    "Test description",
			Poster:  &gitea.User{UserName: "testuser"},
			Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			URL:     "https://example.com/issue/1",
		},
		{
			ID:      2,
			Index:   2,
			Title:   "Test Issue with minimal data",
			State:   "closed",
			Poster:  nil, // Test nil handling
			HTMLURL: "",
		},
	}

	result := handler.transformIssuesToResponse(issues)
	want := []map[string]any{
		{
			"number":    int64(1),
			"title":     "Test Issue with full data",
			"state":     "open",
			"author":    "testuser",
			"createdAt": "2023-01-01T12:00:00Z",
			"updatedAt": "2023-01-02T12:00:00Z",
			"type":      "issue",
			"url":       "https://example.com/issue/1",
		},
		{
			"number":    int64(2),
			"title":     "Test Issue with minimal data",
			"state":     "closed",
			"type":      "issue",
			"author":    "",
			"url":       "",
			"createdAt": "0001-01-01T00:00:00Z",
			"updatedAt": "0001-01-01T00:00:00Z",
		},
	}
	// Test first issue with full data
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestSDKResponseFormat_IssueResponseIncludesRepositoryMetadata tests that issue responses include repository metadata
func TestSDKResponseFormat_IssueResponseIncludesRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockIssues: []*gitea.Issue{
			{
				ID:      1,
				Index:   1,
				URL:     "https://localhost/issues/1",
				Title:   "Test Issue",
				State:   "open",
				Poster:  &gitea.User{UserName: "testuser"},
				Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		},
		MockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := IssueListArgs{Repository: testRepo, State: "open"}

	result, data, err := handler.HandleIssueListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandleIssueListRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("HandleIssueListRequest returned nil result")
	}
	if data == nil {
		t.Fatal("HandleIssueListRequest returned nil data")
	}

	want := map[string]any{
		"issues": []map[string]any{
			{
				"number":    int64(1),
				"title":     "Test Issue",
				"state":     "open",
				"author":    "testuser",
				"createdAt": "2023-01-01T12:00:00Z",
				"updatedAt": "2023-01-02T12:00:00Z",
				"type":      "issue",
				"url":       "https://localhost/issues/1",
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
		"total": int(1),
	}
	if !cmp.Equal(want, data) {
		t.Error(cmp.Diff(want, data))
	}
}

// TestSDKResponseFormat_IssueIndividualObjectsIncludeRepositoryMetadata tests that individual issue objects include repository metadata
func TestSDKResponseFormat_IssueIndividualObjectsIncludeRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	handler := &SDKIssueListHandler{logger: logger}

	issues := []*gitea.Issue{
		{
			ID:      1,
			Index:   1,
			Title:   "Test Issue with repository metadata",
			State:   "open",
			Poster:  &gitea.User{UserName: "testuser"},
			Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	result := handler.transformIssuesToResponse(issues)

	// Verify result structure
	if len(result) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result))
	}

	// Test issue includes repository metadata
	issue := result[0]
	want := map[string]any{
		"number":    int64(1),
		"title":     "Test Issue with repository metadata",
		"state":     "open",
		"author":    "testuser",
		"createdAt": "2023-01-01T12:00:00Z",
		"updatedAt": "2023-01-02T12:00:00Z",
		"type":      "issue",
		"url":       "",
	}
	if !cmp.Equal(want, issue) {
		t.Error(cmp.Diff(want, issue))
	}
}

// BenchmarkSDKPerformance_IssueList benchmarks SDK issue list performance
func BenchmarkSDKPerformance_IssueList(b *testing.B) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockIssues: generateBenchmarkIssues(100), // Generate test data
	}
	handler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := IssueListArgs{Repository: "testuser/test-repo"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = handler.HandleIssueListRequest(ctx, req, args)
	}
}

// generateBenchmarkIssues generates test issue data for benchmarking (legacy function)
func generateBenchmarkIssues(count int) []*gitea.Issue {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedIssues(count, options)
}
