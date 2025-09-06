package server

import (
	"context"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TestRepositoryBasedPRListingE2E tests end-to-end repository-based PR listing functionality
func TestRepositoryBasedPRListingE2E(t *testing.T) {
	logger := logrus.New()

	// Setup mock client with repository-specific data
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "testuser/test-repo",
				Owner:    &gitea.User{UserName: "testuser"},
				Private:  false,
			},
		},
		mockPRs: []*gitea.PullRequest{
			{
				ID:     1,
				Index:  1,
				Title:  "Repository-based PR 1",
				State:  gitea.StateOpen,
				Body:   "Test PR for repository-based listing",
				Poster: &gitea.User{UserName: "testuser"},
			},
			{
				ID:     2,
				Index:  2,
				Title:  "Repository-based PR 2",
				State:  gitea.StateClosed,
				Body:   "Another test PR",
				Poster: &gitea.User{UserName: "contributor"},
			},
		},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test 1: Repository parameter specified
	t.Run("repository_parameter_specified", func(t *testing.T) {
		args := struct {
			Repository string `json:"repository,omitempty"`
			CWD        string `json:"cwd,omitempty"`
			State      string `json:"state,omitempty"`
			Author     string `json:"author,omitempty"`
			Limit      int    `json:"limit,omitempty"`
		}{
			Repository: "testuser/test-repo",
			State:      "open",
			Limit:      10,
		}

		result, data, err := handler.HandlePRListRequest(ctx, req, args)
		if err != nil {
			t.Fatalf("HandlePRListRequest failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if data == nil {
			t.Fatal("Expected non-nil data")
		}

		// Verify response structure
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		prs, exists := dataMap["pullRequests"]
		if !exists {
			t.Fatal("Expected pullRequests field in response")
		}

		prsSlice, ok := prs.([]map[string]interface{})
		if !ok {
			t.Fatal("Expected pullRequests to be a slice")
		}

		// Should return only open PRs (1 out of 2)
		if len(prsSlice) != 1 {
			t.Errorf("Expected 1 PR, got %d", len(prsSlice))
		}

		// Verify repository metadata is included
		if len(prsSlice) > 0 {
			pr := prsSlice[0]
			if pr["repository"] == nil {
				t.Error("Expected repository metadata in PR object")
			}
		}
	})

	// Test 2: Invalid repository format
	t.Run("invalid_repository_format", func(t *testing.T) {
		args := struct {
			Repository string `json:"repository,omitempty"`
			CWD        string `json:"cwd,omitempty"`
			State      string `json:"state,omitempty"`
			Author     string `json:"author,omitempty"`
			Limit      int    `json:"limit,omitempty"`
		}{
			Repository: "invalid-format",
			State:      "open",
		}

		result, data, err := handler.HandlePRListRequest(ctx, req, args)
		// Handler should return an error for invalid repository format
		if err == nil {
			t.Fatal("HandlePRListRequest should return error for invalid repository format")
		}

		if result == nil {
			t.Fatal("Expected result even on error")
		}

		if data != nil {
			t.Error("Expected nil data on repository format error")
		}

		// Verify error message
		if len(result.Content) == 0 {
			t.Fatal("Expected error content")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !strings.Contains(textContent.Text, "invalid repository format") {
			t.Errorf("Expected repository format error, got: %s", textContent.Text)
		}
	})
}

// TestRepositoryBasedIssueListingE2E tests end-to-end repository-based issue listing functionality
func TestRepositoryBasedIssueListingE2E(t *testing.T) {
	logger := logrus.New()

	// Setup mock client with repository-specific data
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "testuser/test-repo",
				Owner:    &gitea.User{UserName: "testuser"},
				Private:  false,
			},
		},
		mockIssues: []*gitea.Issue{
			{
				ID:     1,
				Index:  1,
				Title:  "Repository-based Issue 1",
				State:  "open",
				Body:   "Test issue for repository-based listing",
				Poster: &gitea.User{UserName: "testuser"},
			},
			{
				ID:     2,
				Index:  2,
				Title:  "Repository-based Issue 2",
				State:  "closed",
				Body:   "Another test issue",
				Poster: &gitea.User{UserName: "contributor"},
			},
		},
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test 1: Repository parameter specified
	t.Run("repository_parameter_specified", func(t *testing.T) {
		args := struct {
			Repository string   `json:"repository,omitempty"`
			CWD        string   `json:"cwd,omitempty"`
			State      string   `json:"state,omitempty"`
			Author     string   `json:"author,omitempty"`
			Labels     []string `json:"labels,omitempty"`
			Limit      int      `json:"limit,omitempty"`
		}{
			Repository: "testuser/test-repo",
			State:      "open",
			Limit:      10,
		}

		result, data, err := handler.HandleIssueListRequest(ctx, req, args)
		if err != nil {
			t.Fatalf("HandleIssueListRequest failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if data == nil {
			t.Fatal("Expected non-nil data")
		}

		// Verify response structure
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		issues, exists := dataMap["issues"]
		if !exists {
			t.Fatal("Expected issues field in response")
		}

		issuesSlice, ok := issues.([]map[string]interface{})
		if !ok {
			t.Fatal("Expected issues to be a slice")
		}

		// Should return all issues (mock doesn't filter by state)
		if len(issuesSlice) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issuesSlice))
		}

		// Verify repository metadata is included
		if len(issuesSlice) > 0 {
			issue := issuesSlice[0]
			if issue["repository"] == nil {
				t.Error("Expected repository metadata in issue object")
			}
		}
	})
}
