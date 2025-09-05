package server

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

// MockGiteaClientWithErrors implements a mock Gitea client that can return errors
type MockGiteaClientWithErrors struct {
	mockPRs    []*gitea.PullRequest
	mockIssues []*gitea.Issue
	mockRepos  []*gitea.Repository
	mockError  error
}

func (m *MockGiteaClientWithErrors) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockPRs, &gitea.Response{}, nil
}

func (m *MockGiteaClientWithErrors) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockIssues, &gitea.Response{}, nil
}

func (m *MockGiteaClientWithErrors) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError
	}
	return m.mockRepos, &gitea.Response{}, nil
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
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): connection refused",
			expectedLogged: true,
		},
		{
			name:           "authentication error",
			mockError:      fmt.Errorf("401 Unauthorized"),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): 401 Unauthorized",
			expectedLogged: true,
		},
		{
			name:           "API error",
			mockError:      fmt.Errorf("404 Not Found"),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): 404 Not Found",
			expectedLogged: true,
		},
		{
			name:           "wrapped error",
			mockError:      fmt.Errorf("failed to connect: %w", fmt.Errorf("timeout")),
			expectedError:  "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): failed to connect: timeout",
			expectedLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			mockClient := &MockGiteaClientWithErrors{
				mockError: tt.mockError,
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
			}{}

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
	mockClient := &MockGiteaClientWithErrors{
		mockError: fmt.Errorf("rate limit exceeded"),
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		State  string   `json:"state,omitempty"`
		Author string   `json:"author,omitempty"`
		Labels []string `json:"labels,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}{}

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
	mockClient := &MockGiteaClientWithErrors{
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
	mockClient := &MockGiteaClientWithErrors{
		mockError: wrappedError,
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
	}{}

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
			HTMLURL: "https://example.com/pr/1",
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
			HTMLURL: "https://example.com/issue/1",
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

	// Test second issue with minimal data
	issue2 := result[1]
	if issue2["author"] != "" {
		t.Errorf("Expected empty author for nil poster, got %v", issue2["author"])
	}
	if issue2["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", issue2["state"])
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
	prResult := prHandler.transformPRsToResponse(emptyPRs)
	if len(prResult) != 0 {
		t.Errorf("Expected empty PR result, got %d items", len(prResult))
	}

	// Test empty issues
	issueHandler := &SDKIssueListHandler{logger: logger}
	emptyIssues := []*gitea.Issue{}
	issueResult := issueHandler.transformIssuesToResponse(emptyIssues)
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

// TestSDKAuthenticationErrors_InvalidToken tests authentication error handling for invalid tokens
func TestSDKAuthenticationErrors_InvalidToken(t *testing.T) {
	logger := logrus.New()
	mockClient := &MockGiteaClientWithErrors{
		mockError: fmt.Errorf("401 Unauthorized: invalid token"),
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
	}{}

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

	expectedError := "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): 401 Unauthorized: invalid token"
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
	mockClient := &MockGiteaClientWithErrors{
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
	mockClient := &MockGiteaClientWithErrors{
		mockError: fmt.Errorf("403 Forbidden: insufficient permissions"),
	}

	handler := &SDKIssueListHandler{
		logger: logger,
		client: mockClient,
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		State  string   `json:"state,omitempty"`
		Author string   `json:"author,omitempty"`
		Labels []string `json:"labels,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}{}

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
	mockClient := &MockGiteaClientWithErrors{
		mockError: fmt.Errorf("401 Unauthorized: missing authentication token"),
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
	}{}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)

	if err != nil {
		t.Fatalf("HandlePRListRequest should not return error, got: %v", err)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("HandlePRListRequest should return TextContent")
	}

	expectedError := "Error executing SDK pr list: Gitea SDK ListRepoPullRequests failed (owner=example-owner, repo=example-repo): 401 Unauthorized: missing authentication token"
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
	mockClient := &MockGiteaClientWithErrors{
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
		t.Error("ListRepositories should return nil data on rate limit error")
	}
}
