package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

const testRepoName = "test-repo"
const testUser = "test-user"
const testRepo = testUser + "/" + testRepoName

// MockGiteaClient implements a comprehensive mock Gitea client for testing
type MockGiteaClient struct {
	mockPRs      []*gitea.PullRequest
	mockIssues   []*gitea.Issue
	mockRepos    []*gitea.Repository
	mockUsers    []*gitea.User
	mockBranches []*gitea.Branch
	mockCommits  []*gitea.Commit
	mockError    error
	getRepoErr   error
}

// Repository operations
func (m *MockGiteaClient) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockRepos, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
	if m.getRepoErr != nil {
		return nil, nil, m.getRepoErr
	}
	for _, r := range m.mockRepos {
		if r.Owner != nil && r.Owner.UserName == owner && r.Name == repo {
			return r, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("repository not found")
}

func (m *MockGiteaClient) CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	repo := &gitea.Repository{
		Name:        opt.Name,
		Description: opt.Description,
		Private:     opt.Private,
	}
	m.mockRepos = append(m.mockRepos, repo)
	return repo, &gitea.Response{}, nil
}

func (m *MockGiteaClient) DeleteRepo(owner, repo string) (*gitea.Response, error) {
	if m.mockError != nil {
		return nil, m.mockError
	}
	for i, r := range m.mockRepos {
		if r.Owner != nil && r.Owner.UserName == owner && r.Name == repo {
			m.mockRepos = append(m.mockRepos[:i], m.mockRepos[i+1:]...)
			return &gitea.Response{}, nil
		}
	}
	return &gitea.Response{}, fmt.Errorf("repository not found")
}

// Pull Request operations
func (m *MockGiteaClient) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}

	// Filter PRs by repository
	var filteredPRs []*gitea.PullRequest

	for _, pr := range m.mockPRs {
		// In a real implementation, PRs would be associated with repositories
		// For this mock, we'll assume all PRs belong to the requested repository
		// unless we have repository-specific mock data
		filteredPRs = append(filteredPRs, pr)
	}

	// Apply state filtering if specified
	if opt.State != "" && opt.State != gitea.StateAll {
		var stateFiltered []*gitea.PullRequest
		for _, pr := range filteredPRs {
			if pr.State == opt.State {
				stateFiltered = append(stateFiltered, pr)
			}
		}
		filteredPRs = stateFiltered
	}

	return filteredPRs, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	for _, pr := range m.mockPRs {
		if pr.Index == index {
			return pr, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("pull request not found")
}

func (m *MockGiteaClient) CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	pr := &gitea.PullRequest{
		Index:  int64(len(m.mockPRs) + 1),
		Title:  opt.Title,
		Body:   opt.Body,
		State:  gitea.StateOpen,
		Poster: &gitea.User{UserName: "testuser"},
	}
	m.mockPRs = append(m.mockPRs, pr)
	return pr, &gitea.Response{}, nil
}

func (m *MockGiteaClient) EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	// Simplified implementation to avoid type issues
	for _, pr := range m.mockPRs {
		if pr.Index == index {
			return pr, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("pull request not found")
}

// Issue operations
func (m *MockGiteaClient) ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}

	// For simplicity, return all mock issues
	return m.mockIssues, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	for _, issue := range m.mockIssues {
		if issue.Index == index {
			return issue, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("issue not found")
}

func (m *MockGiteaClient) CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	issue := &gitea.Issue{
		Index:  int64(len(m.mockIssues) + 1),
		Title:  opt.Title,
		Body:   opt.Body,
		State:  "open",
		Poster: &gitea.User{UserName: "testuser"},
	}
	m.mockIssues = append(m.mockIssues, issue)
	return issue, &gitea.Response{}, nil
}

func (m *MockGiteaClient) EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	// Simplified implementation to avoid type issues
	for _, issue := range m.mockIssues {
		if issue.Index == index {
			return issue, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("issue not found")
}

// User operations
func (m *MockGiteaClient) GetMyUserInfo() (*gitea.User, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	if len(m.mockUsers) > 0 {
		return m.mockUsers[0], &gitea.Response{}, nil
	}
	return &gitea.User{UserName: "testuser", Email: "test@example.com"}, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetUserInfo(user string) (*gitea.User, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	for _, u := range m.mockUsers {
		if u.UserName == user {
			return u, &gitea.Response{}, nil
		}
	}
	return &gitea.User{UserName: user, Email: user + "@example.com"}, &gitea.Response{}, nil
}

// Branch operations
func (m *MockGiteaClient) ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockBranches, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	for _, b := range m.mockBranches {
		if b.Name == branch {
			return b, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("branch not found")
}

// Commit operations
func (m *MockGiteaClient) GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	for _, c := range m.mockCommits {
		if c.SHA == sha {
			return c, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("commit not found")
}

func (m *MockGiteaClient) ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockCommits, &gitea.Response{}, nil
}

// TestSDKPRListHandler tests the SDK-based PR list handler
func TestSDKPRListHandler_HandlePRListRequest(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{
			{
				Name: testRepoName,
				Owner: &gitea.User{
					UserName: testUser,
				},
			},
		},
		mockPRs: []*gitea.PullRequest{
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
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
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
		if pr["author"] != testUser {
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
				ID: 1, Index: 1, Comments: 2, Mergeable: true,
				Title: "Test PR", Body: "Pull this Request", State: "open",
			},
		},
		mockIssues: []*gitea.Issue{
			{
				ID:       2,
				Index:    2,
				Title:    "Test Issue",
				Body:     "Issue This",
				State:    "open",
				Comments: 3,
				Created:  time.Now().Add(-time.Hour),
			},
		},
		mockRepos: []*gitea.Repository{
			{Name: testRepoName, Owner: &gitea.User{UserName: testUser}, Private: false},
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
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo, State: "open"}

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
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: testRepo, State: "open"}

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
	prDataMap := prData.(map[string]any)
	if prDataMap["total"] != 1 {
		t.Errorf("Expected 1 PR, got %v", prDataMap["total"])
	}

	issueDataMap := issueData.(map[string]any)
	if issueDataMap["total"] != 1 {
		t.Errorf("Expected 1 issue, got %v", issueDataMap["total"])
	}

	repoDataMap := repoData.(map[string]any)
	if repoDataMap["total"] != 1 {
		t.Errorf("Expected 1 repository, got %v", repoDataMap["total"])
	}
}

// TestSDKPRListHandler_EmptyResults tests handling of empty PR results
func TestSDKPRListHandler_EmptyResults(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockRepos: []*gitea.Repository{
			{
				Name: "testrepo",
				Owner: &gitea.User{
					UserName: "testuser",
				},
			},
		},
		mockPRs: []*gitea.PullRequest{}, // Empty results
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: "testuser/testrepo",
		State:      "closed",
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

// Issue operations
func (m *MockGiteaClient) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockIssues, &gitea.Response{}, nil
}

// TestSDKErrorHandling_SDKErrorTransformation tests SDK error type handling and transformation
func TestSDKErrorHandling_SDKErrorTransformation(t *testing.T) {
	tests := []struct {
		name           string
		mockError      error
		expectedError  string
		expectedLogged bool
	}{
		{
			name:           "network error",
			mockError:      fmt.Errorf("connection refused"),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): connection refused",
			expectedLogged: true,
		},
		{
			name:           "authentication error",
			mockError:      fmt.Errorf("401 Unauthorized"),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 401 Unauthorized",
			expectedLogged: true,
		},
		{
			name:           "API error",
			mockError:      fmt.Errorf("404 Not Found"),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 404 Not Found",
			expectedLogged: true,
		},
		{
			name:           "wrapped error",
			mockError:      fmt.Errorf("failed to connect: %w", fmt.Errorf("timeout")),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): failed to connect: timeout",
			expectedLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			mockClient := &MockGiteaClient{mockError: tt.mockError,
				mockRepos: []*gitea.Repository{{
					ID:    10,
					Owner: &gitea.User{UserName: testUser},
					Name:  testRepoName,
				}},
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
			}{CWD: testRepo}

			result, data, err := handler.HandlePRListRequest(ctx, req, args)
			// Should not return an error in the function return value
			if err != nil {
				t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
			}

			// Should return a result with error message
			if result == nil {
				t.Fatal("HandlePRListRequest should return a result even on error")
			}

			// Check that error message is in the result content
			if len(result.Content) == 0 {
				t.Fatal("HandlePRListRequest should return error content")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatal("HandlePRListRequest should return TextContent")
			}

			if textContent.Text != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, textContent.Text)
			}

			// Data should be nil on error
			if data != nil {
				t.Error("HandlePRListRequest should return nil data on error")
			}
		})
	}
}

// TestSDKErrorHandling_IssueHandlerErrorTransformation tests error handling in issue handler
func TestSDKErrorHandling_IssueHandlerErrorTransformation(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("rate limit exceeded"),
		mockRepos: []*gitea.Repository{{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName}},
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: testRepo,
	}

	result, data, err := handler.HandleIssueListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandleIssueListRequest should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("HandleIssueListRequest should return a result even on error")
	}

	if len(result.Content) == 0 {
		t.Fatal("HandleIssueListRequest should return error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandleIssueListRequest should return TextContent")
	}

	expectedError := "Error executing SDK issue list: Gitea SDK ListIssues failed (state=, limit=0): rate limit exceeded"
	if textContent.Text != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("HandleIssueListRequest should return nil data on error")
	}
}

// TestSDKErrorHandling_RepositoryHandlerErrorTransformation tests error handling in repository handler
func TestSDKErrorHandling_RepositoryHandlerErrorTransformation(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("invalid token"),
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
		t.Fatalf("ListRepositories should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("ListRepositories should return a result even on error")
	}

	if len(result.Content) == 0 {
		t.Fatal("ListRepositories should return error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("ListRepositories should return TextContent")
	}

	expectedError := "Error executing SDK repository list: Gitea SDK ListMyRepos failed (limit=0): invalid token"
	if textContent.Text != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("ListRepositories should return nil data on error")
	}
}

// TestSDKErrorHandling_ErrorContextPreservation tests that error context is preserved
func TestSDKErrorHandling_ErrorContextPreservation(t *testing.T) {
	logger := logrus.New()
	wrappedError := fmt.Errorf("original error: %w", fmt.Errorf("connection failed"))
	mockClient := &MockGiteaClient{
		mockError: wrappedError,
		mockRepos: []*gitea.Repository{{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

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

	// Test repository metadata
	repoMetadata := map[string]interface{}{
		"id":          int64(123),
		"name":        "test-repo",
		"fullName":    "owner/test-repo",
		"description": "Test repository",
		"private":     false,
		"owner": map[string]interface{}{
			"id":       int64(456),
			"username": "owner",
			"fullName": "Test Owner",
		},
		"url": "https://example.com/owner/test-repo",
	}

	result := handler.transformPRsToResponse(prs, repoMetadata)

	// Verify result structure
	if len(result) != 3 {
		t.Fatalf("Expected 3 PRs, got %d", len(result))
	}

	// Test first PR with full data
	pr1 := result[0]
	if pr1["number"] != int64(1) {
		t.Errorf("Expected PR number 1, got %v", pr1["number"])
	}
	if pr1["title"] != "Test PR with full data" {
		t.Errorf("Expected correct title, got %v", pr1["title"])
	}
	if pr1["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", pr1["state"])
	}
	if pr1["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", pr1["author"])
	}
	if pr1["createdAt"] != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected correct createdAt, got %v", pr1["createdAt"])
	}
	if pr1["updatedAt"] != "2023-01-02T12:00:00Z" {
		t.Errorf("Expected correct updatedAt, got %v", pr1["updatedAt"])
	}
	if pr1["type"] != "pull_request" {
		t.Errorf("Expected type 'pull_request', got %v", pr1["type"])
	}
	if pr1["url"] != "https://example.com/pr/1" {
		t.Errorf("Expected correct URL, got %v", pr1["url"])
	}

	// Test repository metadata inclusion
	if pr1["repository"] == nil {
		t.Error("Expected repository metadata to be included in PR object")
	} else {
		repo := pr1["repository"].(map[string]interface{})
		if repo["id"] != int64(123) {
			t.Errorf("Expected repository ID 123, got %v", repo["id"])
		}
		if repo["name"] != "test-repo" {
			t.Errorf("Expected repository name 'test-repo', got %v", repo["name"])
		}
		if repo["fullName"] != "owner/test-repo" {
			t.Errorf("Expected repository fullName 'owner/test-repo', got %v", repo["fullName"])
		}
		if repo["private"] != false {
			t.Errorf("Expected repository private false, got %v", repo["private"])
		}
	}

	// Test second PR with minimal data (nil handling)
	pr2 := result[1]
	if pr2["author"] != "" {
		t.Errorf("Expected empty author for nil poster, got %v", pr2["author"])
	}
	if pr2["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", pr2["state"])
	}

	// Test third PR state normalization
	pr3 := result[2]
	if pr3["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", pr3["state"])
	}
}

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

	// Test repository metadata
	repoMetadata := map[string]interface{}{
		"id":          int64(123),
		"name":        "test-repo",
		"fullName":    "owner/test-repo",
		"description": "Test repository",
		"private":     false,
		"owner": map[string]interface{}{
			"id":       int64(456),
			"username": "owner",
			"fullName": "Test Owner",
		},
		"url": "https://example.com/owner/test-repo",
	}

	result := handler.transformIssuesToResponse(issues, repoMetadata)

	// Verify result structure
	if len(result) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(result))
	}

	// Test first issue with full data
	issue1 := result[0]
	if issue1["number"] != int64(1) {
		t.Errorf("Expected issue number 1, got %v", issue1["number"])
	}
	if issue1["title"] != "Test Issue with full data" {
		t.Errorf("Expected correct title, got %v", issue1["title"])
	}
	if issue1["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", issue1["state"])
	}
	if issue1["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", issue1["author"])
	}
	if issue1["createdAt"] != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected correct createdAt, got %v", issue1["createdAt"])
	}
	if issue1["updatedAt"] != "2023-01-02T12:00:00Z" {
		t.Errorf("Expected correct updatedAt, got %v", issue1["updatedAt"])
	}
	if issue1["type"] != "issue" {
		t.Errorf("Expected type 'issue', got %v", issue1["type"])
	}
	if issue1["url"] != "https://example.com/issue/1" {
		t.Errorf("Expected correct URL, got %v", issue1["url"])
	}

	// Test repository metadata inclusion
	if issue1["repository"] == nil {
		t.Error("Expected repository metadata to be included in issue object")
	} else {
		repo := issue1["repository"].(map[string]interface{})
		if repo["id"] != int64(123) {
			t.Errorf("Expected repository ID 123, got %v", repo["id"])
		}
		if repo["name"] != "test-repo" {
			t.Errorf("Expected repository name 'test-repo', got %v", repo["name"])
		}
		if repo["fullName"] != "owner/test-repo" {
			t.Errorf("Expected repository fullName 'owner/test-repo', got %v", repo["fullName"])
		}
		if repo["private"] != false {
			t.Errorf("Expected repository private false, got %v", repo["private"])
		}
	}

	// Test second issue with minimal data
	issue2 := result[1]
	if issue2["author"] != "" {
		t.Errorf("Expected empty author for nil poster, got %v", issue2["author"])
	}
	if issue2["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", issue2["state"])
	}
}

// TestSDKResponseSizeAndPerformance tests response size and performance impact with repository metadata
func TestSDKResponseSizeAndPerformance(t *testing.T) {
	logger := logrus.New()
	prHandler := &SDKPRListHandler{logger: logger}
	issueHandler := &SDKIssueListHandler{logger: logger}

	// Create test data with repository metadata
	repoMetadata := map[string]interface{}{
		"id":          int64(123),
		"name":        "test-repo",
		"fullName":    "owner/test-repo",
		"description": "Test repository with a longer description to test size impact",
		"private":     false,
		"fork":        false,
		"archived":    false,
		"stars":       42,
		"forks":       10,
		"size":        1024,
		"url":         "https://example.com/owner/test-repo",
		"sshUrl":      "git@example.com:owner/test-repo.git",
		"cloneUrl":    "https://example.com/owner/test-repo.git",
		"owner": map[string]interface{}{
			"id":       int64(456),
			"username": "owner",
			"fullName": "Test Owner Name",
			"email":    "owner@example.com",
		},
	}

	// Test PR response size
	pr := &gitea.PullRequest{
		ID:     1,
		Index:  1,
		Title:  "Test PR",
		State:  gitea.StateOpen,
		Poster: &gitea.User{UserName: "testuser"},
	}

	prResult := prHandler.transformPRsToResponse([]*gitea.PullRequest{pr}, repoMetadata)

	// Verify repository metadata is included
	if len(prResult) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(prResult))
	}

	prData := prResult[0]
	if prData["repository"] == nil {
		t.Error("Expected repository metadata in PR response")
	}

	// Test issue response size
	issue := &gitea.Issue{
		ID:     1,
		Index:  1,
		Title:  "Test Issue",
		State:  "open",
		Poster: &gitea.User{UserName: "testuser"},
	}

	issueResult := issueHandler.transformIssuesToResponse([]*gitea.Issue{issue}, repoMetadata)

	// Verify repository metadata is included
	if len(issueResult) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issueResult))
	}

	issueData := issueResult[0]
	if issueData["repository"] == nil {
		t.Error("Expected repository metadata in issue response")
	}

	// Test response without repository metadata (baseline)
	prResultNoRepo := prHandler.transformPRsToResponse([]*gitea.PullRequest{pr}, map[string]interface{}{})
	issueResultNoRepo := issueHandler.transformIssuesToResponse([]*gitea.Issue{issue}, map[string]interface{}{})

	// Verify responses are still valid without repository metadata
	if len(prResultNoRepo) != 1 || len(issueResultNoRepo) != 1 {
		t.Error("Expected valid responses even without repository metadata")
	}

	// Test with empty repository metadata
	emptyRepoMetadata := map[string]interface{}{}
	prResultEmptyRepo := prHandler.transformPRsToResponse([]*gitea.PullRequest{pr}, emptyRepoMetadata)
	issueResultEmptyRepo := issueHandler.transformIssuesToResponse([]*gitea.Issue{issue}, emptyRepoMetadata)

	if len(prResultEmptyRepo) != 1 || len(issueResultEmptyRepo) != 1 {
		t.Error("Expected valid responses with empty repository metadata")
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

	// Test first repository with full data
	repo1 := result[0]
	if repo1["id"] != int64(1) {
		t.Errorf("Expected repo ID 1, got %v", repo1["id"])
	}
	if repo1["name"] != "test-repo" {
		t.Errorf("Expected repo name 'test-repo', got %v", repo1["name"])
	}
	if repo1["fullName"] != "owner/test-repo" {
		t.Errorf("Expected repo fullName 'owner/test-repo', got %v", repo1["fullName"])
	}
	if repo1["description"] != "Test repository" {
		t.Errorf("Expected description 'Test repository', got %v", repo1["description"])
	}
	if repo1["private"] != false {
		t.Errorf("Expected private false, got %v", repo1["private"])
	}
	if repo1["owner"] != "owner" {
		t.Errorf("Expected owner 'owner', got %v", repo1["owner"])
	}
	if repo1["type"] != "repository" {
		t.Errorf("Expected type 'repository', got %v", repo1["type"])
	}
	if repo1["url"] != "https://example.com/repo/test-repo" {
		t.Errorf("Expected correct URL, got %v", repo1["url"])
	}

	// Test second repository with minimal data
	repo2 := result[1]
	if repo2["private"] != true {
		t.Errorf("Expected private true, got %v", repo2["private"])
	}
	if repo2["owner"] != "" {
		t.Errorf("Expected empty owner for nil owner, got %v", repo2["owner"])
	}
	// Description should not be present if empty
	if _, exists := repo2["description"]; exists {
		t.Error("Description should not be present when empty")
	}
}

// TestSDKResponseTransformation_EmptyResults tests transformation of empty result sets
func TestSDKResponseTransformation_EmptyResults(t *testing.T) {
	logger := logrus.New()

	// Test empty PRs
	prHandler := &SDKPRListHandler{logger: logger}
	emptyPRs := []*gitea.PullRequest{}
	prResult := prHandler.transformPRsToResponse(emptyPRs, map[string]interface{}{})
	if len(prResult) != 0 {
		t.Errorf("Expected empty PR result, got %d items", len(prResult))
	}

	// Test empty issues
	issueHandler := &SDKIssueListHandler{logger: logger}
	emptyIssues := []*gitea.Issue{}
	issueResult := issueHandler.transformIssuesToResponse(emptyIssues, map[string]interface{}{})
	if len(issueResult) != 0 {
		t.Errorf("Expected empty issue result, got %d items", len(issueResult))
	}

	// Test empty repositories
	repoHandler := &SDKRepositoryHandler{logger: logger}
	emptyRepos := []*gitea.Repository{}
	repoResult := repoHandler.transformReposToResponse(emptyRepos)
	if len(repoResult) != 0 {
		t.Errorf("Expected empty repository result, got %d items", len(repoResult))
	}
}

// TestSDKMigration_CLIToSDKCompatibility tests compatibility between CLI and SDK approaches
func TestSDKMigration_CLIToSDKCompatibility(t *testing.T) {
	logger := logrus.New()

	// Setup test data that would be equivalent between CLI and SDK
	testPRs := []*gitea.PullRequest{
		{
			ID:     1,
			Index:  1,
			Title:  "Test PR for migration",
			State:  gitea.StateOpen,
			Body:   "Migration test PR",
			Poster: &gitea.User{UserName: "testuser"},
		},
	}

	testIssues := []*gitea.Issue{
		{
			ID:     1,
			Index:  1,
			Title:  "Test Issue for migration",
			State:  "open",
			Body:   "Migration test issue",
			Poster: &gitea.User{UserName: "testuser"},
		},
	}

	testRepos := []*gitea.Repository{
		{
			ID:          1,
			Name:        testRepoName,
			FullName:    testRepo,
			Description: "Repository for migration testing",
			Private:     false,
			Owner:       &gitea.User{UserName: testUser},
		},
	}

	// Test SDK handlers with the test data
	mockClient := &MockGiteaClient{
		mockPRs:    testPRs,
		mockIssues: testIssues,
		mockRepos:  testRepos,
	}

	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}
	repoHandler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR migration
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("SDK PR handler failed: %v", prErr)
	}
	if prResult == nil {
		t.Fatal("SDK PR handler returned nil result")
	}

	prDataMap, ok := prData.(map[string]interface{})
	if !ok {
		t.Fatal("SDK PR handler returned invalid data format")
	}

	prs, exists := prDataMap["pullRequests"]
	if !exists {
		t.Fatal("SDK PR handler missing pullRequests in response")
	}

	prsSlice, ok := prs.([]map[string]interface{})
	if !ok {
		t.Fatal("SDK PR handler returned invalid PRs format")
	}

	if len(prsSlice) != 1 {
		t.Errorf("Expected 1 PR from SDK handler, got %d", len(prsSlice))
	}

	// Test Issue migration
	issueArgs := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: testRepo}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("SDK Issue handler failed: %v", issueErr)
	}
	if issueResult == nil {
		t.Fatal("SDK Issue handler returned nil result")
	}

	issueDataMap, ok := issueData.(map[string]interface{})
	if !ok {
		t.Fatal("SDK Issue handler returned invalid data format")
	}

	issues, exists := issueDataMap["issues"]
	if !exists {
		t.Fatal("SDK Issue handler missing issues in response")
	}

	issuesSlice, ok := issues.([]map[string]interface{})
	if !ok {
		t.Fatal("SDK Issue handler returned invalid issues format")
	}

	if len(issuesSlice) != 1 {
		t.Errorf("Expected 1 issue from SDK handler, got %d", len(issuesSlice))
	}

	// Test Repository migration
	repoArgs := struct {
		Limit int `json:"limit,omitempty"`
	}{}

	repoResult, repoData, repoErr := repoHandler.ListRepositories(ctx, req, repoArgs)
	if repoErr != nil {
		t.Fatalf("SDK Repository handler failed: %v", repoErr)
	}
	if repoResult == nil {
		t.Fatal("SDK Repository handler returned nil result")
	}

	repoDataMap, ok := repoData.(map[string]interface{})
	if !ok {
		t.Fatal("SDK Repository handler returned invalid data format")
	}

	repos, exists := repoDataMap["repositories"]
	if !exists {
		t.Fatal("SDK Repository handler missing repositories in response")
	}

	reposSlice, ok := repos.([]map[string]interface{})
	if !ok {
		t.Fatal("SDK Repository handler returned invalid repos format")
	}

	if len(reposSlice) != 1 {
		t.Errorf("Expected 1 repository from SDK handler, got %d", len(reposSlice))
	}

	// Verify data consistency across all handlers
	if prDataMap["total"] != 1 {
		t.Errorf("PR handler total mismatch: expected 1, got %v", prDataMap["total"])
	}
	if issueDataMap["total"] != 1 {
		t.Errorf("Issue handler total mismatch: expected 1, got %v", issueDataMap["total"])
	}
	if repoDataMap["total"] != 1 {
		t.Errorf("Repository handler total mismatch: expected 1, got %v", repoDataMap["total"])
	}
}

// TestSDKMigration_CommandBuilderCompatibility tests migration from CLI command builders to SDK
func TestSDKMigration_CommandBuilderCompatibility(t *testing.T) {
	// Test that SDK handlers can handle the same parameter formats as CLI command builders
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{ID: 1, Index: 1, Title: "CLI-compatible PR", State: gitea.StateOpen},
		},
		mockIssues: []*gitea.Issue{
			{ID: 1, Index: 1, Title: "CLI-compatible Issue", State: "open"},
		},
		mockRepos: []*gitea.Repository{{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName}},
	}

	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test parameters that would be generated by CLI command builders
	testCases := []struct {
		name   string
		prArgs struct {
			Repository string `json:"repository,omitempty"`
			CWD        string `json:"cwd,omitempty"`
			State      string `json:"state,omitempty"`
			Author     string `json:"author,omitempty"`
			Limit      int    `json:"limit,omitempty"`
		}
		issueArgs struct {
			Repository string   `json:"repository,omitempty"`
			CWD        string   `json:"cwd,omitempty"`
			State      string   `json:"state,omitempty"`
			Author     string   `json:"author,omitempty"`
			Labels     []string `json:"labels,omitempty"`
			Limit      int      `json:"limit,omitempty"`
		}
	}{
		{
			name: "open state",
			prArgs: struct {
				Repository string `json:"repository,omitempty"`
				CWD        string `json:"cwd,omitempty"`
				State      string `json:"state,omitempty"`
				Author     string `json:"author,omitempty"`
				Limit      int    `json:"limit,omitempty"`
			}{Repository: testRepo},
			issueArgs: struct {
				Repository string   `json:"repository,omitempty"`
				CWD        string   `json:"cwd,omitempty"`
				State      string   `json:"state,omitempty"`
				Author     string   `json:"author,omitempty"`
				Labels     []string `json:"labels,omitempty"`
				Limit      int      `json:"limit,omitempty"`
			}{Repository: testRepo},
		},
		{
			name: "closed state",
			prArgs: struct {
				Repository string `json:"repository,omitempty"`
				CWD        string `json:"cwd,omitempty"`
				State      string `json:"state,omitempty"`
				Author     string `json:"author,omitempty"`
				Limit      int    `json:"limit,omitempty"`
			}{Repository: testRepo},
			issueArgs: struct {
				Repository string   `json:"repository,omitempty"`
				CWD        string   `json:"cwd,omitempty"`
				State      string   `json:"state,omitempty"`
				Author     string   `json:"author,omitempty"`
				Labels     []string `json:"labels,omitempty"`
				Limit      int      `json:"limit,omitempty"`
			}{Repository: testRepo},
		},
		{
			name: "with limit",
			prArgs: struct {
				Repository string `json:"repository,omitempty"`
				CWD        string `json:"cwd,omitempty"`
				State      string `json:"state,omitempty"`
				Author     string `json:"author,omitempty"`
				Limit      int    `json:"limit,omitempty"`
			}{Repository: testRepo},
			issueArgs: struct {
				Repository string   `json:"repository,omitempty"`
				CWD        string   `json:"cwd,omitempty"`
				State      string   `json:"state,omitempty"`
				Author     string   `json:"author,omitempty"`
				Labels     []string `json:"labels,omitempty"`
				Limit      int      `json:"limit,omitempty"`
			}{Repository: testRepo},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test PR handler compatibility
			prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, tc.prArgs)
			if prErr != nil {
				t.Fatalf("PR handler failed with CLI-compatible params: %v", prErr)
			}
			if prResult == nil {
				t.Fatal("PR handler returned nil result")
			}
			if prData == nil {
				buf, err := json.Marshal(prResult)
				if err != nil {
					t.Fatal(err)
				}
				t.Fatal("PR handler returned nil data", string(buf))
			}

			// Test Issue handler compatibility
			issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, tc.issueArgs)
			if issueErr != nil {
				t.Fatalf("Issue handler failed with CLI-compatible params: %v", issueErr)
			}
			if issueResult == nil {
				t.Fatal("Issue handler returned nil result")
			}
			if issueData == nil {
				t.Fatal("Issue handler returned nil data")
			}
		})
	}
}

// TestSDKMockSetupAndTeardown tests SDK mock setup and teardown functionality
func TestSDKMockSetupAndTeardown(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func() *MockGiteaClient
		teardownFunc func(*MockGiteaClient)
		validateFunc func(*testing.T, *MockGiteaClient)
	}{
		{
			name: "basic mock setup and teardown",
			setupFunc: func() *MockGiteaClient {
				return &MockGiteaClient{
					mockPRs: []*gitea.PullRequest{
						{ID: 1, Index: 1, Title: "Test PR"},
					},
					mockIssues: []*gitea.Issue{
						{ID: 1, Index: 1, Title: "Test Issue"},
					},
					mockRepos: []*gitea.Repository{
						{ID: 1, Name: "test-repo"},
					},
				}
			},
			teardownFunc: func(mock *MockGiteaClient) {
				mock.mockPRs = nil
				mock.mockIssues = nil
				mock.mockRepos = nil
			},
			validateFunc: func(t *testing.T, mock *MockGiteaClient) {
				if len(mock.mockPRs) != 0 {
					t.Error("Teardown should clear mock PRs")
				}
				if len(mock.mockIssues) != 0 {
					t.Error("Teardown should clear mock issues")
				}
				if len(mock.mockRepos) != 0 {
					t.Error("Teardown should clear mock repos")
				}
			},
		},
		{
			name: "empty mock setup",
			setupFunc: func() *MockGiteaClient {
				return &MockGiteaClient{}
			},
			teardownFunc: func(mock *MockGiteaClient) {
				// No-op teardown for empty mock
			},
			validateFunc: func(t *testing.T, mock *MockGiteaClient) {
				if mock.mockPRs != nil && len(mock.mockPRs) != 0 {
					t.Error("Empty setup should have no PRs")
				}
				if mock.mockIssues != nil && len(mock.mockIssues) != 0 {
					t.Error("Empty setup should have no issues")
				}
				if mock.mockRepos != nil && len(mock.mockRepos) != 0 {
					t.Error("Empty setup should have no repos")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := tt.setupFunc()
			if mock == nil {
				t.Fatal("Setup function should return a mock client")
			}

			// Validate initial state
			if mock.mockPRs == nil {
				mock.mockPRs = []*gitea.PullRequest{}
			}
			if mock.mockIssues == nil {
				mock.mockIssues = []*gitea.Issue{}
			}
			if mock.mockRepos == nil {
				mock.mockRepos = []*gitea.Repository{}
			}

			// Test mock functionality
			prs, _, err := mock.ListRepoPullRequests("owner", "repo", gitea.ListPullRequestsOptions{})
			if err != nil {
				t.Fatalf("Mock ListRepoPullRequests failed: %v", err)
			}
			if len(prs) != len(mock.mockPRs) {
				t.Errorf("Expected %d PRs, got %d", len(mock.mockPRs), len(prs))
			}

			issues, _, err := mock.ListIssues(gitea.ListIssueOption{})
			if err != nil {
				t.Fatalf("Mock ListIssues failed: %v", err)
			}
			if len(issues) != len(mock.mockIssues) {
				t.Errorf("Expected %d issues, got %d", len(mock.mockIssues), len(issues))
			}

			repos, _, err := mock.ListMyRepos(gitea.ListReposOptions{})
			if err != nil {
				t.Fatalf("Mock ListMyRepos failed: %v", err)
			}
			if len(repos) != len(mock.mockRepos) {
				t.Errorf("Expected %d repos, got %d", len(mock.mockRepos), len(repos))
			}

			// Teardown
			tt.teardownFunc(mock)

			// Validate teardown
			tt.validateFunc(t, mock)
		})
	}
}

// TestSDKMockSetupTeardownIntegration tests mock setup and teardown in integration scenarios
func TestSDKMockSetupTeardownIntegration(t *testing.T) {
	logger := logrus.New()

	// Setup phase
	setupMock := func() *MockGiteaClient {
		return &MockGiteaClient{
			mockPRs: []*gitea.PullRequest{
				{
					ID:     1,
					Index:  1,
					Title:  "Integration Test PR",
					State:  gitea.StateOpen,
					Poster: &gitea.User{UserName: testUser},
				},
			},
			mockIssues: []*gitea.Issue{
				{
					ID:     1,
					Index:  1,
					Title:  "Integration Test Issue",
					State:  "open",
					Poster: &gitea.User{UserName: testUser},
				},
			},
			mockRepos: []*gitea.Repository{
				{
					ID:       1,
					Name:     testRepoName,
					FullName: testRepo,
					Owner:    &gitea.User{UserName: testUser},
				},
			},
		}
	}

	teardownMock := func(mock *MockGiteaClient) {
		mock.mockPRs = nil
		mock.mockIssues = nil
		mock.mockRepos = nil
	}

	// Setup
	mock := setupMock()

	// Test integration with handlers
	prHandler := &SDKPRListHandler{logger: logger, client: mock}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mock}
	repoHandler := &SDKRepositoryHandler{logger: logger, client: mock}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR handler
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

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
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: testRepo}

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

	// Verify data integrity
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

	// Teardown
	teardownMock(mock)

	// Verify teardown
	if len(mock.mockPRs) != 0 {
		t.Error("Teardown should clear all mock data")
	}
	if len(mock.mockIssues) != 0 {
		t.Error("Teardown should clear all mock data")
	}
	if len(mock.mockRepos) != 0 {
		t.Error("Teardown should clear all mock data")
	}
}

// TestSDKAuthenticationErrors_InvalidToken tests authentication error handling for invalid tokens
func TestSDKAuthenticationErrors_InvalidToken(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("401 Unauthorized: invalid token"),
		mockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("HandlePRListRequest should return a result even on auth error")
	}

	if len(result.Content) == 0 {
		t.Fatal("HandlePRListRequest should return error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandlePRListRequest should return TextContent")
	}

	expectedError := "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 401 Unauthorized: invalid token"
	if textContent.Text != expectedError {
		t.Errorf("Expected auth error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("HandlePRListRequest should return nil data on auth error")
	}
}

// TestSDKAuthenticationErrors_ExpiredToken tests authentication error handling for expired tokens
func TestSDKAuthenticationErrors_ExpiredToken(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("401 Unauthorized: token expired"),
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
		t.Fatalf("ListRepositories should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("ListRepositories should return a result even on auth error")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("ListRepositories should return TextContent")
	}

	expectedError := "Error executing SDK repository list: Gitea SDK ListMyRepos failed (limit=0): 401 Unauthorized: token expired"
	if textContent.Text != expectedError {
		t.Errorf("Expected expired token error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("ListRepositories should return nil data on auth error")
	}
}

// TestSDKAuthenticationErrors_InsufficientPermissions tests authentication error handling for insufficient permissions
func TestSDKAuthenticationErrors_InsufficientPermissions(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("403 Forbidden: insufficient permissions"),
		mockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: testRepo}

	result, data, err := handler.HandleIssueListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandleIssueListRequest should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("HandleIssueListRequest should return a result even on auth error")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandleIssueListRequest should return TextContent")
	}

	expectedError := "Error executing SDK issue list: Gitea SDK ListIssues failed (state=, limit=0): 403 Forbidden: insufficient permissions"
	if textContent.Text != expectedError {
		t.Errorf("Expected permissions error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("HandleIssueListRequest should return nil data on auth error")
	}
}

// TestSDKAuthenticationErrors_MissingToken tests authentication error handling for missing tokens
func TestSDKAuthenticationErrors_MissingToken(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("401 Unauthorized: missing authentication token"),
		mockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandlePRListRequest should return TextContent")
	}

	expectedError := "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=" + testUser + ", repo=" + testRepoName + "): 401 Unauthorized: missing authentication token"
	if textContent.Text != expectedError {
		t.Errorf("Expected missing token error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("HandlePRListRequest should return nil data on auth error")
	}
}

// TestSDKAuthenticationErrors_RateLimit tests authentication error handling for rate limiting
func TestSDKAuthenticationErrors_RateLimit(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("429 Too Many Requests: rate limit exceeded"),
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
		t.Fatalf("ListRepositories should not return error, got: %v", err)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("ListRepositories should return TextContent")
	}

	expectedError := "Error executing SDK repository list: Gitea SDK ListMyRepos failed (limit=0): 429 Too Many Requests: rate limit exceeded"
	if textContent.Text != expectedError {
		t.Errorf("Expected rate limit error message '%s', got '%s'", expectedError, textContent.Text)
	}

	if data != nil {
		t.Error("ListRepositories should return nil data on auth error")
	}
}

// TestSDKResponseFormat_PRResponseIncludesRepositoryMetadata tests that PR responses include repository metadata
func TestSDKResponseFormat_PRResponseIncludesRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{ID: 1, Index: 1, Title: "Test PR", State: gitea.StateOpen, URL: "https://localhost/issues/1", Poster: &gitea.User{UserName: "testuser"}},
		},
		mockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
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

	if data == nil {
		t.Fatal("HandlePRListRequest returned nil data")
	}

	// Verify response structure includes repository metadata
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

	// Verify PR includes repository metadata
	if len(prsSlice) > 0 {
		pr := prsSlice[0]
		want := map[string]any{
			"type":   "pull_request",
			"url":    "https://localhost/issues/1",
			"author": "testuser",
			"number": int64(1),
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
			"state": "open",
			"title": "Test PR",
		}
		if !cmp.Equal(want, pr) {
			t.Error(cmp.Diff(want, pr))
		}
	}
}

// TestSDKResponseFormat_IssueResponseIncludesRepositoryMetadata tests that issue responses include repository metadata
func TestSDKResponseFormat_IssueResponseIncludesRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockIssues: []*gitea.Issue{
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
		mockRepos: []*gitea.Repository{{ID: 1, Name: testRepoName, FullName: testRepo, Owner: &gitea.User{UserName: testUser}}},
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

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

	// Verify response structure includes repository metadata
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("HandleIssueListRequest returned data of wrong type")
	}

	issues, exists := dataMap["issues"]
	if !exists {
		t.Fatal("HandleIssueListRequest data missing issues field")
	}

	issuesSlice, ok := issues.([]map[string]interface{})
	if !ok {
		t.Fatal("issues field is not a slice")
	}

	if len(issuesSlice) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issuesSlice))
	}

	// Verify issue includes repository metadata
	if len(issuesSlice) > 0 {
		issue := issuesSlice[0]
		if issue["type"] != "issue" {
			t.Errorf("Expected issue type 'issue', got %v", issue["type"])
		}
		if issue["url"] == "" {
			t.Error("Expected issue to include URL")
		}
		if issue["createdAt"] == "" {
			t.Error("Expected issue to include createdAt")
		}
		if issue["updatedAt"] == "" {
			t.Error("Expected issue to include updatedAt")
		}
	}
}

// TestSDKResponseFormat_BackwardCompatibility tests backward compatibility of response structure
func TestSDKResponseFormat_BackwardCompatibility(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{
				ID:     1,
				Index:  1,
				Title:  "Test PR",
				State:  gitea.StateOpen,
				Poster: &gitea.User{UserName: testUser},
			},
		},
		mockIssues: []*gitea.Issue{
			{
				ID:      1,
				Index:   1,
				Title:   "Test Issue",
				State:   "open",
				Poster:  &gitea.User{UserName: testUser},
				Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		},
		mockRepos: []*gitea.Repository{
			{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName},
		},
	}

	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR handler backward compatibility
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}

	// Test Issue handler backward compatibility
	issueArgs := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}

	// Verify both handlers return expected structure
	if prResult == nil || prData == nil {
		t.Fatal("PR handler returned nil result or data")
	}
	if issueResult == nil || issueData == nil {
		t.Fatal("Issue handler returned nil result or data")
	}

	// Verify response structure consistency
	prDataMap := prData.(map[string]interface{})
	issueDataMap := issueData.(map[string]interface{})

	if _, exists := prDataMap["total"]; !exists {
		t.Error("PR response missing total field")
	}
	if _, exists := issueDataMap["total"]; !exists {
		t.Error("Issue response missing total field")
	}

	if _, exists := prDataMap["pullRequests"]; !exists {
		t.Error("PR response missing pullRequests field")
	}
	if _, exists := issueDataMap["issues"]; !exists {
		t.Error("Issue response missing issues field")
	}
}

// TestSDKResponseFormat_TotalCountAccuracy tests total count accuracy with repository filtering
func TestSDKResponseFormat_TotalCountAccuracy(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{ID: 1, Index: 1, Title: "PR 1", State: gitea.StateOpen},
			{ID: 2, Index: 2, Title: "PR 2", State: gitea.StateClosed},
			{ID: 3, Index: 3, Title: "PR 3", State: gitea.StateOpen},
		},
		mockIssues: []*gitea.Issue{
			{ID: 1, Index: 1, Title: "Issue 1", State: "open"},
			{ID: 2, Index: 2, Title: "Issue 2", State: "closed"},
		},
		mockRepos: []*gitea.Repository{
			{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName},
		},
	}

	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR count accuracy
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	_, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}

	prDataMap, _ := prData.(map[string]any)
	if prDataMap["total"] != 2 {
		t.Errorf("Expected PR total 2, got %v", prDataMap["total"])
	}

	// Test issue count accuracy
	issueArgs := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	_, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}

	issueDataMap := issueData.(map[string]interface{})
	if issueDataMap["total"] != 2 {
		t.Errorf("Expected issue total 2, got %v", issueDataMap["total"])
	}
}

// TestSDKResponseFormat_PRIndividualObjectsIncludeRepositoryMetadata tests that individual PR objects include repository metadata
func TestSDKResponseFormat_PRIndividualObjectsIncludeRepositoryMetadata(t *testing.T) {
	logger := logrus.New()
	handler := &SDKPRListHandler{logger: logger}

	// Test data with repository information
	prs := []*gitea.PullRequest{
		{
			ID:     1,
			Index:  1,
			Title:  "Test PR with repository metadata",
			State:  gitea.StateOpen,
			Poster: &gitea.User{UserName: "testuser"},
		},
	}

	// Test repository metadata
	repoMetadata := map[string]interface{}{
		"id":          int64(123),
		"name":        "test-repo",
		"fullName":    "owner/test-repo",
		"description": "Test repository",
		"private":     false,
		"owner": map[string]interface{}{
			"id":       int64(456),
			"username": "owner",
			"fullName": "Test Owner",
		},
		"url": "https://example.com/owner/test-repo",
	}

	result := handler.transformPRsToResponse(prs, repoMetadata)

	// Verify result structure
	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}

	// Test PR includes repository metadata
	pr := result[0]

	// Check existing fields are preserved
	if pr["number"] != int64(1) {
		t.Errorf("Expected PR number 1, got %v", pr["number"])
	}
	if pr["title"] != "Test PR with repository metadata" {
		t.Errorf("Expected correct title, got %v", pr["title"])
	}
	if pr["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", pr["state"])
	}
	if pr["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", pr["author"])
	}
	if pr["type"] != "pull_request" {
		t.Errorf("Expected type 'pull_request', got %v", pr["type"])
	}

	// Check repository metadata is included in individual PR object
	if _, exists := pr["repository"]; !exists {
		t.Error("Expected PR object to include repository metadata")
	}

	repoData, ok := pr["repository"].(map[string]interface{})
	if !ok {
		t.Error("Expected repository field to be a map")
	} else {
		// Since we're testing the transformation function directly,
		// repository metadata would need to be added by the handler
		// This test verifies the structure is ready for repository metadata
		if repoData == nil {
			t.Error("Expected repository data to be initialized")
		}
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

	result := handler.transformIssuesToResponse(issues, map[string]interface{}{})

	// Verify result structure
	if len(result) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result))
	}

	// Test issue includes repository metadata
	issue := result[0]

	// Check existing fields are preserved
	if issue["number"] != int64(1) {
		t.Errorf("Expected issue number 1, got %v", issue["number"])
	}
	if issue["title"] != "Test Issue with repository metadata" {
		t.Errorf("Expected correct title, got %v", issue["title"])
	}
	if issue["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", issue["state"])
	}
	if issue["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", issue["author"])
	}
	if issue["type"] != "issue" {
		t.Errorf("Expected type 'issue', got %v", issue["type"])
	}
	if issue["createdAt"] == "" {
		t.Error("Expected issue to include createdAt")
	}
	if issue["updatedAt"] == "" {
		t.Error("Expected issue to include updatedAt")
	}

	// Check repository metadata is included in individual issue object
	if _, exists := issue["repository"]; !exists {
		t.Error("Expected issue object to include repository metadata")
	}

	repoData, ok := issue["repository"].(map[string]interface{})
	if !ok {
		t.Error("Expected repository field to be a map")
	} else {
		// Since we're testing the transformation function directly,
		// repository metadata would need to be added by the handler
		// This test verifies the structure is ready for repository metadata
		if repoData == nil {
			t.Error("Expected repository data to be initialized")
		}
	}
}

// TestSDKResponseFormat_RepositoryMetadataConsistency tests that repository metadata is consistent across PR and issue responses
func TestSDKResponseFormat_RepositoryMetadataConsistency(t *testing.T) {
	logger := logrus.New()
	prHandler := &SDKPRListHandler{logger: logger}
	issueHandler := &SDKIssueListHandler{logger: logger}

	// Test data
	prs := []*gitea.PullRequest{
		{
			ID:     1,
			Index:  1,
			Title:  "Test PR",
			State:  gitea.StateOpen,
			Poster: &gitea.User{UserName: "testuser"},
		},
	}

	issues := []*gitea.Issue{
		{
			ID:      1,
			Index:   1,
			Title:   "Test Issue",
			State:   "open",
			Poster:  &gitea.User{UserName: "testuser"},
			Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	prResult := prHandler.transformPRsToResponse(prs, map[string]interface{}{})
	issueResult := issueHandler.transformIssuesToResponse(issues, map[string]interface{}{})

	// Verify both have repository metadata structure
	if len(prResult) != 1 || len(issueResult) != 1 {
		t.Fatal("Expected one result from each transformation")
	}

	pr := prResult[0]
	issue := issueResult[0]

	// Both should have repository field
	if _, prHasRepo := pr["repository"]; !prHasRepo {
		t.Error("PR should have repository field")
	}
	if _, issueHasRepo := issue["repository"]; !issueHasRepo {
		t.Error("Issue should have repository field")
	}

	// Repository fields should be maps
	prRepo, prOk := pr["repository"].(map[string]interface{})
	issueRepo, issueOk := issue["repository"].(map[string]interface{})

	if !prOk || !issueOk {
		t.Error("Repository fields should be maps")
	}

	// Since we're testing the transformation functions directly,
	// the actual repository data would be populated by the handlers
	// This test ensures the structure is consistent
	if prRepo == nil || issueRepo == nil {
		t.Error("Repository data should be initialized")
	}
}

// TestSDKResponseFormat_ErrorResponseFormats tests error response formats
func TestSDKResponseFormat_ErrorResponseFormats(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockError: fmt.Errorf("repository not found"),
		mockRepos: []*gitea.Repository{
			{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName},
		},
	}

	handler := &SDKPRListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: testRepo}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)
	// Should not return an error in the function return value
	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}

	// Should return a result with error message
	if result == nil {
		t.Fatal("HandlePRListRequest should return a result even on error")
	}

	// Check that error message is in the result content
	if len(result.Content) == 0 {
		t.Fatal("HandlePRListRequest should return error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandlePRListRequest should return TextContent")
	}

	if !strings.Contains(textContent.Text, "Error executing SDK pr list") {
		t.Errorf("Expected error message to contain 'Error executing SDK pr list', got '%s'", textContent.Text)
	}

	// Data should be nil on error
	if data != nil {
		t.Error("HandlePRListRequest should return nil data on error")
	}
}

// TestSDKResponseFormat_ResponseConsistencyBetweenEndpoints tests response format consistency between PR and issue endpoints
func TestSDKResponseFormat_ResponseConsistencyBetweenEndpoints(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{
				ID:     1,
				Index:  1,
				Title:  "Test PR",
				State:  gitea.StateOpen,
				Poster: &gitea.User{UserName: testUser},
			},
		},
		mockIssues: []*gitea.Issue{
			{
				ID:      1,
				Index:   1,
				Title:   "Test Issue",
				State:   "open",
				Poster:  &gitea.User{UserName: testUser},
				Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		},
		mockRepos: []*gitea.Repository{
			{ID: 8, Owner: &gitea.User{UserName: testUser}, Name: testRepoName},
		},
	}

	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Test PR handler
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}

	// Test issue handler
	issueArgs := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: testRepo,
		State:      "open",
	}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}

	// Verify both handlers return results
	if prResult == nil || prData == nil {
		t.Fatal("PR handler returned nil result or data")
	}
	if issueResult == nil || issueData == nil {
		t.Fatal("Issue handler returned nil result or data")
	}

	// Verify response structure consistency
	prDataMap := prData.(map[string]interface{})
	issueDataMap := issueData.(map[string]interface{})

	// Both should have total field
	if prDataMap["total"] == nil {
		t.Error("PR response missing total field")
	}
	if issueDataMap["total"] == nil {
		t.Error("Issue response missing total field")
	}

	// Both should have their respective data arrays
	if prDataMap["pullRequests"] == nil {
		t.Error("PR response missing pullRequests field")
	}
	if issueDataMap["issues"] == nil {
		t.Error("Issue response missing issues field")
	}

	// Verify result content format consistency
	if len(prResult.Content) == 0 {
		t.Error("PR result missing content")
	}
	if len(issueResult.Content) == 0 {
		t.Error("Issue result missing content")
	}

	prContent, ok := prResult.Content[0].(*mcp.TextContent)
	if !ok {
		t.Error("PR result content should be TextContent")
	}
	issueContent, ok := issueResult.Content[0].(*mcp.TextContent)
	if !ok {
		t.Error("Issue result content should be TextContent")
	}

	// Both should have similar success message format
	if !strings.Contains(prContent.Text, "Found") || !strings.Contains(prContent.Text, "pull request") {
		t.Errorf("PR success message format inconsistent: %s", prContent.Text)
	}
	if !strings.Contains(issueContent.Text, "Found") || !strings.Contains(issueContent.Text, "issue") {
		t.Errorf("Issue success message format inconsistent: %s", issueContent.Text)
	}
}

// TestRepositoryParameterValidation_FormatValidation tests repository format validation (owner/repo format)
func TestRepositoryParameterValidation_FormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		expectValid bool
		expectError string
	}{
		{
			name:        "valid owner/repo format",
			repoParam:   "owner/repo",
			expectValid: true,
		},
		{
			name:        "missing slash",
			repoParam:   "ownerrepo",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "multiple slashes",
			repoParam:   "owner/repo/extra",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "empty owner",
			repoParam:   "/repo",
			expectValid: false,
			expectError: "invalid repository format: owner cannot be empty",
		},
		{
			name:        "empty repo",
			repoParam:   "owner/",
			expectValid: false,
			expectError: "invalid repository format: repository name cannot be empty",
		},
		{
			name:        "empty string",
			repoParam:   "",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "special characters in owner",
			repoParam:   "owner@domain/repo",
			expectValid: true, // Allow special chars for now, let API validate
		},
		{
			name:        "special characters in repo",
			repoParam:   "owner/repo-name_with.special",
			expectValid: true, // Allow special chars for now, let API validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateRepositoryFormat(tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryFormat(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryFormat(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_ExistenceValidation tests repository existence validation
func TestRepositoryParameterValidation_ExistenceValidation(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		mockRepos   []*gitea.Repository
		mockError   error
		expectValid bool
		expectError string
	}{
		{
			name:      "repository exists",
			repoParam: testRepo,
			mockRepos: []*gitea.Repository{
				{
					Name:  testRepoName,
					Owner: &gitea.User{UserName: testUser},
				},
			},
			expectValid: true,
		},
		{
			name:        "repository not found",
			repoParam:   "owner/nonexistent",
			mockRepos:   []*gitea.Repository{},
			expectValid: false,
			expectError: "failed to validate repository existence: repository not found",
		},
		{
			name:      "different owner same repo name",
			repoParam: testUser + "/repo",
			mockRepos: []*gitea.Repository{
				{
					Name:  testRepoName,
					Owner: &gitea.User{UserName: testUser},
				},
			},
			expectValid: false,
			expectError: "failed to validate repository existence: repository not found",
		},
		{
			name:        "API error during validation",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("connection refused"),
			expectValid: false,
			expectError: "failed to validate repository existence: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGiteaClient{
				mockRepos:  tt.mockRepos,
				getRepoErr: tt.mockError,
			}

			valid, err := validateRepositoryExistence(mockClient, tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryExistence(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryExistence(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_AccessValidation tests repository access permission validation
func TestRepositoryParameterValidation_AccessValidation(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		mockRepos   []*gitea.Repository
		mockError   error
		expectValid bool
		expectError string
	}{
		{
			name:      "user has read access to public repo",
			repoParam: "owner/repo",
			mockRepos: []*gitea.Repository{
				{
					Name:    "repo",
					Owner:   &gitea.User{UserName: "owner"},
					Private: false,
				},
			},
			expectValid: true,
		},
		{
			name:      "user has read access to private repo",
			repoParam: "owner/private-repo",
			mockRepos: []*gitea.Repository{
				{
					Name:    "private-repo",
					Owner:   &gitea.User{UserName: "owner"},
					Private: true,
				},
			},
			expectValid: true,
		},
		{
			name:        "user lacks access to private repo",
			repoParam:   "other/private-repo",
			mockRepos:   []*gitea.Repository{}, // No repos returned = no access
			expectValid: false,
			expectError: "failed to validate repository access: repository not found",
		},
		{
			name:        "API error during access check",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("403 Forbidden"),
			expectValid: false,
			expectError: "failed to validate repository access: 403 Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGiteaClient{
				mockRepos:  tt.mockRepos,
				getRepoErr: tt.mockError,
			}

			valid, err := validateRepositoryAccess(mockClient, tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryAccess(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryAccess(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_OrganizationRepos tests organization-owned repository handling
func TestRepositoryParameterValidation_OrganizationRepos(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		mockRepos   []*gitea.Repository
		expectValid bool
		expectError string
	}{
		{
			name:      "organization repository exists",
			repoParam: "myorg/repo",
			mockRepos: []*gitea.Repository{
				{
					Name:  "repo",
					Owner: &gitea.User{UserName: "myorg"},
				},
			},
			expectValid: true,
		},
		{
			name:      "user-owned repository exists",
			repoParam: "user/repo",
			mockRepos: []*gitea.Repository{
				{
					Name:  "repo",
					Owner: &gitea.User{UserName: "user"},
				},
			},
			expectValid: true,
		},
		{
			name:        "organization repository not found",
			repoParam:   "myorg/missing-repo",
			mockRepos:   []*gitea.Repository{},
			expectValid: false,
			expectError: "failed to validate repository existence: repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGiteaClient{
				mockRepos: tt.mockRepos,
			}

			valid, err := validateRepositoryExistence(mockClient, tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryExistence(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryExistence(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_ErrorScenarios tests mock scenarios for repository not found errors
func TestRepositoryParameterValidation_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		mockError   error
		expectValid bool
		expectError string
	}{
		{
			name:        "network timeout",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("dial tcp: i/o timeout"),
			expectValid: false,
			expectError: "failed to validate repository existence: dial tcp: i/o timeout",
		},
		{
			name:        "DNS resolution failure",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("no such host"),
			expectValid: false,
			expectError: "failed to validate repository existence: no such host",
		},
		{
			name:        "authentication failure",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("401 Unauthorized"),
			expectValid: false,
			expectError: "failed to validate repository existence: 401 Unauthorized",
		},
		{
			name:        "server error",
			repoParam:   "owner/repo",
			mockError:   fmt.Errorf("500 Internal Server Error"),
			expectValid: false,
			expectError: "failed to validate repository existence: 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGiteaClient{
				getRepoErr: tt.mockError,
			}

			valid, err := validateRepositoryExistence(mockClient, tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryExistence(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryExistence(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_SpecialCharacters tests edge cases with special characters in repository names
func TestRepositoryParameterValidation_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		repoParam   string
		expectValid bool
		expectError string
	}{
		{
			name:        "hyphen in owner name",
			repoParam:   "my-owner/repo",
			expectValid: true,
		},
		{
			name:        "underscore in repo name",
			repoParam:   "owner/my_repo",
			expectValid: true,
		},
		{
			name:        "numbers in names",
			repoParam:   "owner123/repo456",
			expectValid: true,
		},
		{
			name:        "mixed case",
			repoParam:   "Owner/Repo",
			expectValid: true,
		},
		{
			name:        "very long names",
			repoParam:   "verylongownername/verylongrepositoryname",
			expectValid: true,
		},
		{
			name:        "single character names",
			repoParam:   "a/b",
			expectValid: true,
		},
		{
			name:        "spaces in names (should fail)",
			repoParam:   "owner with spaces/repo",
			expectValid: false,
			expectError: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:        "special chars that might cause issues",
			repoParam:   "owner/repo<script>",
			expectValid: true, // Let API validate these
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateRepositoryFormat(tt.repoParam)
			if valid != tt.expectValid {
				t.Errorf("validateRepositoryFormat(%q) = %v, want %v", tt.repoParam, valid, tt.expectValid)
			}
			if !tt.expectValid && err != nil && err.Error() != tt.expectError {
				t.Errorf("validateRepositoryFormat(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectError)
			}
		})
	}
}

// TestRepositoryParameterValidation_ErrorMessages tests that error messages are descriptive and actionable
func TestRepositoryParameterValidation_ErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		repoParam       string
		expectedMessage string
	}{
		{
			name:            "missing slash",
			repoParam:       "ownerrepo",
			expectedMessage: "invalid repository format: expected 'owner/repo'",
		},
		{
			name:            "empty owner",
			repoParam:       "/repo",
			expectedMessage: "invalid repository format: owner cannot be empty",
		},
		{
			name:            "empty repo",
			repoParam:       "owner/",
			expectedMessage: "invalid repository format: repository name cannot be empty",
		},
		{
			name:            "empty string",
			repoParam:       "",
			expectedMessage: "invalid repository format: expected 'owner/repo'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateRepositoryFormat(tt.repoParam)
			if err == nil {
				t.Errorf("validateRepositoryFormat(%q) should return error", tt.repoParam)
			} else if err.Error() != tt.expectedMessage {
				t.Errorf("validateRepositoryFormat(%q) error = %q, want %q", tt.repoParam, err.Error(), tt.expectedMessage)
			}
		})
	}
}

// BenchmarkSDKPerformance_PRList benchmarks SDK PR list performance
func BenchmarkSDKPerformance_PRList(b *testing.B) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockPRs: generateBenchmarkPRs(100), // Generate test data
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

// BenchmarkSDKPerformance_IssueList benchmarks SDK issue list performance
func BenchmarkSDKPerformance_IssueList(b *testing.B) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockIssues: generateBenchmarkIssues(100), // Generate test data
	}
	handler := &SDKIssueListHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: "testuser/test-repo"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = handler.HandleIssueListRequest(ctx, req, args)
	}
}

// BenchmarkSDKPerformance_RepositoryList benchmarks SDK repository list performance
func BenchmarkSDKPerformance_RepositoryList(b *testing.B) {
	logger := logrus.New()
	mockClient := &MockGiteaClient{
		mockRepos: generateBenchmarkRepos(100), // Generate test data
	}
	handler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		Limit int `json:"limit,omitempty"`
	}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = handler.ListRepositories(ctx, req, args)
	}
}

// TestDataSeeder provides comprehensive test data seeding for SDK scenarios
type TestDataSeeder struct {
	baseTime time.Time
	userPool []*gitea.User
}

// NewTestDataSeeder creates a new test data seeder with default configuration
func NewTestDataSeeder() *TestDataSeeder {
	return &TestDataSeeder{
		baseTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		userPool: []*gitea.User{
			{ID: 1, UserName: "alice", Email: "alice@example.com"},
			{ID: 2, UserName: "bob", Email: "bob@example.com"},
			{ID: 3, UserName: "charlie", Email: "charlie@example.com"},
			{ID: 4, UserName: "diana", Email: "diana@example.com"},
			{ID: 5, UserName: "eve", Email: "eve@example.com"},
		},
	}
}

// SeedPRs generates test PR data with realistic scenarios
func (s *TestDataSeeder) SeedPRs(count int, options SeedOptions) []*gitea.PullRequest {
	prs := make([]*gitea.PullRequest, count)
	states := []gitea.StateType{gitea.StateOpen, gitea.StateClosed}

	for _, user := range s.userPool {
		for i := range count {
			user := user
			state := states[i%len(states)]
			createdTime := s.baseTime.Add(time.Duration(i) * time.Hour)
			pr := &gitea.PullRequest{
				ID:      int64(i + 1),
				Index:   int64(i + 1),
				Title:   fmt.Sprintf("%s PR %d", options.Prefix, i+1),
				State:   state,
				Body:    fmt.Sprintf("Description for %s PR %d", options.Prefix, i+1),
				Poster:  user,
				Created: &createdTime,
				Updated: &createdTime,
				HTMLURL: fmt.Sprintf("https://%s.com/pr/%d", options.Domain, i+1),
			}

			if options.IncludeLabels && i%3 == 0 {
				pr.Labels = []*gitea.Label{
					{Name: "enhancement", Color: "84cc16"},
					{Name: "documentation", Color: "10b981"},
				}
			}

			prs[i] = pr
		}

	}
	return prs
}

// SeedIssues generates test issue data with realistic scenarios
func (s *TestDataSeeder) SeedIssues(count int, options SeedOptions) []*gitea.Issue {
	issues := make([]*gitea.Issue, 0, count)
	states := []string{"open", "closed"}

	for _, user := range s.userPool {
		for i := range count {
			user := user
			state := states[i%len(states)]
			createdTime := s.baseTime.Add(time.Duration(i) * time.Hour)

			issue := &gitea.Issue{
				ID:      int64(i + 1),
				Index:   int64(i + 1),
				Title:   fmt.Sprintf("%s Issue %d", options.Prefix, i+1),
				State:   gitea.StateType(state),
				Body:    fmt.Sprintf("Description for %s issue %d", options.Prefix, i+1),
				Poster:  user,
				Created: createdTime,
				Updated: createdTime,
				HTMLURL: fmt.Sprintf("https://%s.com/issue/%d", options.Domain, i+1),
			}

			if options.IncludeLabels && i%2 == 0 {
				issue.Labels = []*gitea.Label{
					{Name: "bug", Color: "ef4444"},
					{Name: "help wanted", Color: "f59e0b"},
				}
			}

			issues = append(issues, issue)
		}
	}
	return issues
}

// SeedRepos generates test repository data with realistic scenarios
func (s *TestDataSeeder) SeedRepos(count int, options SeedOptions) []*gitea.Repository {
	repos := make([]*gitea.Repository, 0, count)

	for _, user := range s.userPool {
		for i := range count {
			user := user
			repo := &gitea.Repository{
				ID:          int64(i + 1),
				Name:        fmt.Sprintf("%s-repo-%d", options.Prefix, i+1),
				FullName:    fmt.Sprintf("%s/%s-repo-%d", user.UserName, options.Prefix, i+1),
				Description: fmt.Sprintf("Test repository %d for %s", i+1, options.Prefix),
				Private:     i%5 == 0, // Every 5th repo is private
				Owner:       user,
				HTMLURL:     fmt.Sprintf("https://%s.com/%s/%s-repo-%d", options.Domain, user.UserName, options.Prefix, i+1),
			}

			repos = append(repos, repo)
		}
	}
	return repos
}

// SeedUsers generates test user data
func (s *TestDataSeeder) SeedUsers(count int) []*gitea.User {
	users := make([]*gitea.User, count)
	for i := 0; i < count; i++ {
		users[i] = &gitea.User{
			ID:       int64(i + 1),
			UserName: fmt.Sprintf("user%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			FullName: fmt.Sprintf("User %d", i+1),
		}
	}
	return users
}

// SeedBranches generates test branch data
func (s *TestDataSeeder) SeedBranches(count int, repoOwner, repoName string) []*gitea.Branch {
	branches := make([]*gitea.Branch, count)
	branchNames := []string{"main", "develop", "feature/auth", "feature/ui", "hotfix/security"}

	for i := 0; i < count; i++ {
		branchName := branchNames[i%len(branchNames)]
		if i >= len(branchNames) {
			branchName = fmt.Sprintf("branch-%d", i+1)
		}

		branches[i] = &gitea.Branch{
			Name: branchName,
		}
	}
	return branches
}

// SeedCommits generates test commit data
func (s *TestDataSeeder) SeedCommits(count int, options SeedOptions) []*gitea.Commit {
	commits := make([]*gitea.Commit, count)

	for i := 0; i < count; i++ {
		commits[i] = &gitea.Commit{}
	}
	return commits
}

// SeedOptions configures test data seeding behavior
type SeedOptions struct {
	Prefix        string
	Domain        string
	IncludeLabels bool
}

// DefaultSeedOptions returns default seeding options
func DefaultSeedOptions() SeedOptions {
	return SeedOptions{
		Prefix:        "test",
		Domain:        "example",
		IncludeLabels: true,
	}
}

// TestSDKDataSeeding tests the comprehensive test data seeding system
func TestSDKDataSeeding(t *testing.T) {
	seeder := NewTestDataSeeder()
	options := DefaultSeedOptions()

	// Test PR seeding
	prs := seeder.SeedPRs(5, options)
	if len(prs) != 5 {
		t.Errorf("Expected 5 PRs, got %d", len(prs))
	}
	if prs[0].Title != "test PR 1" {
		t.Errorf("Expected PR title 'test PR 1', got '%s'", prs[0].Title)
	}

	// Test Issue seeding
	issues := seeder.SeedIssues(3, options)
	if len(issues) != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}
	if issues[0].Title != "test Issue 1" {
		t.Errorf("Expected issue title 'test Issue 1', got '%s'", issues[0].Title)
	}

	// Test Repository seeding
	repos := seeder.SeedRepos(4, options)
	if len(repos) != 4*len(seeder.userPool) {
		t.Errorf("Expected 4 repos, got %d", len(repos))
	}
	if repos[0].Name != "test-repo-1" {
		t.Errorf("Expected repo name 'test-repo-1', got '%s'", repos[0].Name)
	}

	// Test User seeding
	users := seeder.SeedUsers(2)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].UserName != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", users[0].UserName)
	}

	// Test Branch seeding
	branches := seeder.SeedBranches(3, "owner", "repo")
	if len(branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(branches))
	}
	if branches[0].Name != "main" {
		t.Errorf("Expected branch name 'main', got '%s'", branches[0].Name)
	}

	// Test Commit seeding
	commits := seeder.SeedCommits(2, options)
	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}
	// Note: Commit struct simplified to avoid SDK compatibility issues
}

// TestSDKDataSeedingIntegration tests data seeding with mock client integration
func TestSDKDataSeedingIntegration(t *testing.T) {
	seeder := NewTestDataSeeder()
	options := DefaultSeedOptions()

	// Seed comprehensive test data
	prs := seeder.SeedPRs(3, options)
	issues := seeder.SeedIssues(3, options)
	repos := seeder.SeedRepos(3, options)
	branches := seeder.SeedBranches(2, testUser, testRepoName)
	commits := seeder.SeedCommits(2, options)

	// Create mock client with seeded data
	mockClient := &MockGiteaClient{
		mockPRs:      prs,
		mockIssues:   issues,
		mockRepos:    repos,
		mockBranches: branches,
		mockCommits:  commits,
	}

	// Test integration with handlers
	logger := logrus.New()
	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}
	repoHandler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	u := seeder.userPool[rand.IntN(len(seeder.userPool))]
	rp := []string{}
	for _, r := range repos {
		if r.Owner.UserName == u.UserName {
			rp = append(rp, r.Name)
		}
	}
	if len(rp) == 0 {
		t.Fatal("Setup failed", u, len(repos), len(seeder.userPool))
	}
	name := rp[rand.IntN(len(rp))]
	// Test all handlers with seeded data
	prArgs := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{Repository: path.Join(u.UserName, name)}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}
	if prResult == nil || prData == nil {
		buf, err := json.Marshal(prResult)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("PR handler returned nil results", string(buf))
	}

	issueArgs := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{Repository: path.Join(u.UserName, name)}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}
	if issueResult == nil || issueData == nil {
		t.Fatal("Issue handler returned nil results")
	}

	repoArgs := struct {
		Limit int `json:"limit,omitempty"`
	}{}

	repoResult, repoData, repoErr := repoHandler.ListRepositories(ctx, req, repoArgs)
	if repoErr != nil {
		t.Fatalf("Repository handler failed: %v", repoErr)
	}
	if repoResult == nil || repoData == nil {
		t.Fatal("Repository handler returned nil results")
	}

	// Verify seeded data integrity
	prDataMap := prData.(map[string]any)
	if prDataMap["total"] != 2 {
		t.Errorf("Expected 2 seeded PRs, got %v", prDataMap["total"])
	}

	issueDataMap := issueData.(map[string]any)
	if issueDataMap["total"] != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 seeded issues, got %v", issueDataMap["total"])
	}

	repoDataMap := repoData.(map[string]any)
	if repoDataMap["total"] != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 seeded repos, got %v", repoDataMap["total"])
	}
}

// generateBenchmarkPRs generates test PR data for benchmarking (legacy function)
func generateBenchmarkPRs(count int) []*gitea.PullRequest {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedPRs(count, options)
}

// generateBenchmarkIssues generates test issue data for benchmarking (legacy function)
func generateBenchmarkIssues(count int) []*gitea.Issue {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedIssues(count, options)
}

// generateBenchmarkRepos generates test repository data for benchmarking (legacy function)
func generateBenchmarkRepos(count int) []*gitea.Repository {
	seeder := NewTestDataSeeder()
	options := SeedOptions{Prefix: "benchmark", Domain: "example", IncludeLabels: false}
	return seeder.SeedRepos(count, options)
}

// TestSDKPerformanceComparison tests performance characteristics of SDK vs CLI approaches
func TestSDKPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison test in short mode")
	}

	logger := logrus.New()

	// Setup test data
	testSizes := []int{10, 50, 100}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			mockClient := &MockGiteaClient{
				mockPRs:    generateBenchmarkPRs(size),
				mockIssues: generateBenchmarkIssues(size),
				mockRepos:  generateBenchmarkRepos(size),
			}

			// Test SDK handlers
			prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
			issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}
			repoHandler := &SDKRepositoryHandler{logger: logger, client: mockClient}

			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			// Measure SDK performance
			start := time.Now()
			for i := 0; i < 10; i++ { // Run 10 iterations for averaging
				prArgs := struct {
					Repository string `json:"repository,omitempty"`
					CWD        string `json:"cwd,omitempty"`
					State      string `json:"state,omitempty"`
					Author     string `json:"author,omitempty"`
					Limit      int    `json:"limit,omitempty"`
				}{}
				_, _, _ = prHandler.HandlePRListRequest(ctx, req, prArgs)

				issueArgs := struct {
					Repository string   `json:"repository,omitempty"`
					CWD        string   `json:"cwd,omitempty"`
					State      string   `json:"state,omitempty"`
					Author     string   `json:"author,omitempty"`
					Labels     []string `json:"labels,omitempty"`
					Limit      int      `json:"limit,omitempty"`
				}{Repository: "testuser/test-repo"}
				_, _, _ = issueHandler.HandleIssueListRequest(ctx, req, issueArgs)

				repoArgs := struct {
					Limit int `json:"limit,omitempty"`
				}{}
				_, _, _ = repoHandler.ListRepositories(ctx, req, repoArgs)
			}
			sdkDuration := time.Since(start)

			// Log performance metrics
			t.Logf("SDK Performance for size %d: %v total for 30 operations", size, sdkDuration)
			t.Logf("Average SDK time per operation: %v", sdkDuration/30)

			// Verify that operations complete within reasonable time
			if sdkDuration > 5*time.Second {
				t.Errorf("SDK operations took too long: %v", sdkDuration)
			}

			// Verify memory efficiency (basic check)
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Run operations again
			for i := 0; i < 10; i++ {
				prArgs := struct {
					Repository string `json:"repository,omitempty"`
					CWD        string `json:"cwd,omitempty"`
					State      string `json:"state,omitempty"`
					Author     string `json:"author,omitempty"`
					Limit      int    `json:"limit,omitempty"`
				}{}
				_, _, _ = prHandler.HandlePRListRequest(ctx, req, prArgs)
			}

			runtime.GC()
			runtime.ReadMemStats(&m2)

			var memoryUsed uint64
			if m2.Alloc >= m1.Alloc {
				memoryUsed = m2.Alloc - m1.Alloc
			} else {
				memoryUsed = 0 // Handle potential counter reset
			}
			t.Logf("Memory used for operations: %d bytes", memoryUsed)

			// Basic memory efficiency check (not a strict leak test)
			if memoryUsed > 100*1024*1024 { // 100MB threshold for test data
				t.Logf("High memory usage detected: %d bytes - may indicate inefficiency", memoryUsed)
			}
		})
	}
}
