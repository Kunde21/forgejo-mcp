package server

import (
	"context"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// MockGiteaClient implements a mock Gitea client for testing
type MockGiteaClient struct {
	mockPRs    []*gitea.PullRequest
	mockIssues []*gitea.Issue
	mockRepos  []*gitea.Repository
}

func (m *MockGiteaClient) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	return m.mockPRs, &gitea.Response{}, nil
}

func (m *MockGiteaClient) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	return m.mockIssues, &gitea.Response{}, nil
}

func (m *MockGiteaClient) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	return m.mockRepos, &gitea.Response{}, nil
}

// TestSDKPRListHandler tests the SDK-based PR list handler
func TestSDKPRListHandler_HandlePRListRequest(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{
				ID:    1,
				Index: 1,
				Title: "Test PR",
				State: gitea.StateOpen,
				Body:  "Test description",
				Poster: &gitea.User{
					UserName: "testuser",
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
	args := struct {
		State  string `json:"state,omitempty"`
		Author string `json:"author,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{
		State: "open",
	}

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

	// Verify the response contains pull requests
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("HandlePRListRequest returned data of wrong type")
	}

	prs, exists := dataMap["pullRequests"]
	if !exists {
		t.Fatal("HandlePRListRequest data missing pullRequests field")
	}

	prsSlice, ok := prs.([]map[string]interface{})
	if !ok {
		t.Fatal("pullRequests field is not a slice")
	}

	if len(prsSlice) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prsSlice))
	}

	// Verify PR data structure
	if len(prsSlice) > 0 {
		pr := prsSlice[0]
		if pr["number"] != int64(1) {
			t.Errorf("Expected PR number 1, got %v (type: %T)", pr["number"], pr["number"])
		}
		if pr["title"] != "Test PR" {
			t.Errorf("Expected PR title 'Test PR', got %v", pr["title"])
		}
		if pr["author"] != "testuser" {
			t.Errorf("Expected PR author 'testuser', got %v", pr["author"])
		}
		if pr["state"] != "open" {
			t.Errorf("Expected PR state 'open', got %v", pr["state"])
		}
	}
}

// TestSDKRepositoryHandler tests the SDK-based repository handler
func TestSDKRepositoryHandler_ListRepositories(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{
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
	args := struct {
		Limit int `json:"limit,omitempty"`
	}{}

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
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{}, // Empty results
	}

	handler := &SDKRepositoryHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Limit int `json:"limit,omitempty"`
	}{}

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

// TestSDKHandlersIntegration tests integration between all SDK handlers
func TestSDKHandlersIntegration(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{
				ID:     1,
				Index:  1,
				Title:  "Integration Test PR",
				State:  gitea.StateOpen,
				Poster: &gitea.User{UserName: "testuser"},
			},
		},
		mockIssues: []*gitea.Issue{
			{
				ID:     1,
				Index:  1,
				Title:  "Integration Test Issue",
				State:  gitea.StateOpen,
				Poster: &gitea.User{UserName: "testuser"},
			},
		},
		mockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "integration-repo",
				FullName: "testuser/integration-repo",
				Owner:    &gitea.User{UserName: "testuser"},
			},
		},
	}

	// Create all handlers with the same client
	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}
	repoHandler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR handler
	prArgs := struct {
		State  string `json:"state,omitempty"`
		Author string `json:"author,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{State: "open"}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}
	if prResult == nil {
		t.Fatal("PR handler returned nil result")
	}
	if prData == nil {
		t.Fatal("PR handler returned nil data")
	}

	// Test issue handler
	issueArgs := struct {
		State  string   `json:"state,omitempty"`
		Author string   `json:"author,omitempty"`
		Labels []string `json:"labels,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}{State: "open"}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}
	if issueResult == nil {
		t.Fatal("Issue handler returned nil result")
	}
	if issueData == nil {
		t.Fatal("Issue handler returned nil data")
	}

	// Test repository handler
	repoArgs := struct {
		Limit int `json:"limit,omitempty"`
	}{}

	repoResult, repoData, repoErr := repoHandler.ListRepositories(ctx, req, repoArgs)
	if repoErr != nil {
		t.Fatalf("Repository handler failed: %v", repoErr)
	}
	if repoResult == nil {
		t.Fatal("Repository handler returned nil result")
	}
	if repoData == nil {
		t.Fatal("Repository handler returned nil data")
	}

	// Verify all handlers return expected data structures
	prDataMap := prData.(map[string]interface{})
	if prDataMap["total"] != 1 {
		t.Errorf("Expected 1 PR, got %v", prDataMap["total"])
	}

	issueDataMap := issueData.(map[string]interface{})
	if issueDataMap["total"] != 1 {
		t.Errorf("Expected 1 issue, got %v", issueDataMap["total"])
	}

	repoDataMap := repoData.(map[string]interface{})
	if repoDataMap["total"] != 1 {
		t.Errorf("Expected 1 repository, got %v", repoDataMap["total"])
	}
}

// TestSDKPRListHandler_ErrorHandling tests error handling in SDK PR list handler
func TestSDKPRListHandler_ErrorHandling(t *testing.T) {
	// TODO: Implement proper error handling tests after refactoring handler
	// to check for nil client before calling methods
	t.Skip("Error handling tests need handler refactoring for nil client checks")
}

// TestSDKPRListHandler_EmptyResults tests handling of empty PR results
func TestSDKPRListHandler_EmptyResults(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{}, // Empty results
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		State  string `json:"state,omitempty"`
		Author string `json:"author,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{
		State: "closed",
	}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)

	if err != nil {
		t.Fatalf("HandlePRListRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("HandlePRListRequest returned nil result")
	}

	// Verify empty results are handled correctly
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("HandlePRListRequest returned data of wrong type")
	}

	prs, exists := dataMap["pullRequests"]
	if !exists {
		t.Fatal("HandlePRListRequest data missing pullRequests field")
	}

	prsSlice, ok := prs.([]map[string]interface{})
	if !ok {
		t.Fatal("pullRequests field is not a slice")
	}

	if len(prsSlice) != 0 {
		t.Errorf("Expected 0 PRs for empty results, got %d", len(prsSlice))
	}

	total, exists := dataMap["total"]
	if !exists {
		t.Fatal("HandlePRListRequest data missing total field")
	}

	if total != 0 {
		t.Errorf("Expected total 0 for empty results, got %v", total)
	}
}
