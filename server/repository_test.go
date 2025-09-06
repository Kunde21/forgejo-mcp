package server

import (
	"context"
	"fmt"
	"testing"

	"code.gitea.io/sdk/gitea"
	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TestSDKRepositoryHandler tests the SDK-based repository handler
func TestSDKRepositoryHandler_ListRepositories(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "owner/test-repo",
				Owner: &gitea.User{
					UserName: "owner",
				},
				Description: "Test repository",
				Private:     false,
			},
		},
	}

	handler := &SDKRepositoryHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := RepoListArgs{}

	result, data, err := handler.ListRepositories(ctx, req, args)
	// Verify no error occurred
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	// Verify result is not nil
	if result == nil {
		t.Fatal("ListRepositories returned nil result")
	}

	// Verify data contains expected structure
	if data == nil {
		t.Fatal("ListRepositories returned nil data")
	}

	// Verify the response contains repositories
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("ListRepositories returned data of wrong type")
	}

	repos, exists := dataMap["repositories"]
	if !exists {
		t.Fatal("ListRepositories data missing repositories field")
	}

	reposSlice, ok := repos.([]map[string]interface{})
	if !ok {
		t.Fatal("repositories field is not a slice")
	}

	if len(reposSlice) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(reposSlice))
	}

	// Verify repository data structure
	if len(reposSlice) > 0 {
		repo := reposSlice[0]
		if repo["name"] != "test-repo" {
			t.Errorf("Expected repository name 'test-repo', got %v", repo["name"])
		}
		if repo["fullName"] != "owner/test-repo" {
			t.Errorf("Expected repository fullName 'owner/test-repo', got %v", repo["fullName"])
		}
		if repo["owner"] != "owner" {
			t.Errorf("Expected repository owner 'owner', got %v", repo["owner"])
		}
	}
}

// TestSDKRepositoryHandler_EmptyResults tests handling of empty repository results
func TestSDKRepositoryHandler_EmptyResults(t *testing.T) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockRepos: []*gitea.Repository{}, // Empty results
	}

	handler := &SDKRepositoryHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := RepoListArgs{}

	result, data, err := handler.ListRepositories(ctx, req, args)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if result == nil {
		t.Fatal("ListRepositories returned nil result")
	}

	// Verify empty results are handled correctly
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("ListRepositories returned data of wrong type")
	}

	repos, exists := dataMap["repositories"]
	if !exists {
		t.Fatal("ListRepositories data missing repositories field")
	}

	reposSlice, ok := repos.([]map[string]interface{})
	if !ok {
		t.Fatal("repositories field is not a slice")
	}

	if len(reposSlice) != 0 {
		t.Errorf("Expected 0 repositories for empty results, got %d", len(reposSlice))
	}

	total, exists := dataMap["total"]
	if !exists {
		t.Fatal("ListRepositories data missing total field")
	}

	if total != 0 {
		t.Errorf("Expected total 0 for empty results, got %v", total)
	}
}

// TestSDKResponseTransformation_Repositories tests repository response transformation from SDK to MCP format
func TestSDKResponseTransformation_Repositories(t *testing.T) {
	logger := logrus.New()
	handler := &SDKRepositoryHandler{logger: logger}

	repos := []*gitea.Repository{
		{
			ID:          1,
			Name:        "test-repo",
			FullName:    "owner/test-repo",
			Description: "Test repository",
			Private:     false,
			Owner:       &gitea.User{UserName: "owner"},
			HTMLURL:     "https://example.com/repo/test-repo",
		},
		{
			ID:          2,
			Name:        "private-repo",
			FullName:    "owner/private-repo",
			Description: "",
			Private:     true,
			Owner:       nil, // Test nil handling
			HTMLURL:     "",
		},
	}

	result := handler.transformReposToResponse(repos)

	// Verify result structure
	if len(result) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(result))
	}

	want := []map[string]any{
		{
			"id":          int64(1),
			"name":        "test-repo",
			"fullName":    "owner/test-repo",
			"description": "Test repository",
			"private":     false,
			"owner":       "owner",
			"type":        "repository",
			"url":         "https://example.com/repo/test-repo",
		},
		{
			"id":       int64(2),
			"name":     "private-repo",
			"fullName": "owner/private-repo",
			"private":  true,
			"owner":    "",
			"type":     "repository",
			"url":      "",
		},
	}
	// Test repository transformation
	if !cmp.Equal(want, result) {
		t.Error(cmp.Diff(want, result))
	}
}

// TestSDKResponseTransformation_EmptyResults tests transformation of empty result sets
func TestSDKResponseTransformation_EmptyResults(t *testing.T) {
	logger := logrus.New()

	// Test empty repositories
	repoHandler := &SDKRepositoryHandler{logger: logger}
	emptyRepos := []*gitea.Repository{}
	repoResult := repoHandler.transformReposToResponse(emptyRepos)
	if len(repoResult) != 0 {
		t.Errorf("Expected empty repository result, got %d items", len(repoResult))
	}
}

// TestSDKAuthenticationErrors tests authentication error handling across all handlers
func TestSDKAuthenticationErrors(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		args          interface{}
		expectedError string
	}{
		{
			name:          "expired token - repository handler",
			mockError:     fmt.Errorf("401 Unauthorized: token expired"),
			args:          RepoListArgs{},
			expectedError: "Error executing SDK repository list: Gitea SDK ListMyRepos failed (limit=0): 401 Unauthorized: token expired",
		},
		{
			name:          "rate limit - repository handler",
			mockError:     fmt.Errorf("429 Too Many Requests: rate limit exceeded"),
			args:          RepoListArgs{},
			expectedError: "Error executing SDK repository list: Gitea SDK ListMyRepos failed (limit=0): 429 Too Many Requests: rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			mockClient := &giteasdk.MockGiteaClient{
				MockError: tt.mockError,
			}

			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			var result *mcp.CallToolResult
			var data interface{}
			var err error

			handler := &SDKRepositoryHandler{logger: logger, client: mockClient}
			args := tt.args.(RepoListArgs)
			result, data, err = handler.ListRepositories(ctx, req, args)

			if err != nil {
				t.Fatalf("Handler should not return error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Handler should return a result even on auth error")
			}

			if len(result.Content) == 0 {
				t.Fatal("Handler should return error content")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatal("Handler should return TextContent")
			}

			if textContent.Text != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, textContent.Text)
			}

			if data != nil {
				t.Error("Handler should return nil data on auth error")
			}
		})
	}
}

// BenchmarkSDKPerformance_RepositoryList benchmarks SDK repository list performance
func BenchmarkSDKPerformance_RepositoryList(b *testing.B) {
	logger := logrus.New()
	mockClient := &giteasdk.MockGiteaClient{
		MockRepos: generateBenchmarkRepos(100), // Generate test data
	}
	handler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := RepoListArgs{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = handler.ListRepositories(ctx, req, args)
	}
}

// generateBenchmarkRepos generates test repository data for benchmarking (legacy function)
func generateBenchmarkRepos(count int) []*gitea.Repository {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedRepos(count, options)
}
