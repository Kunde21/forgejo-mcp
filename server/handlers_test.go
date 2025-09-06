package server

import (
	"context"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	remote "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

func TestSDKPRListHandler_HandlePRListRequest_ValidRepository(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{
		MockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "owner/test-repo",
				Private:  false,
				Owner: &gitea.User{
					UserName: "owner",
				},
			},
		},
		MockPRs: []*gitea.PullRequest{
			{
				Index: 1,
				Title: "Test PR",
				State: gitea.StateOpen,
				Poster: &gitea.User{
					UserName: "testuser",
				},
			},
		},
	}

	handler := NewSDKPRListHandler(logger, mockClient)

	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: "owner/test-repo",
	}

	result, data, err := handler.HandlePRListRequest(context.Background(), req, args)

	if result == nil {
		t.Error("Expected result to not be nil")
	}
	if data == nil {
		t.Error("Expected data to not be nil")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the result contains expected data
	resultData, ok := data.(map[string]any)
	if !ok {
		t.Fatal("Expected resultData to be map[string]any")
	}
	if resultData["total"] != 1 {
		t.Errorf("Expected total to be 1, got %v", resultData["total"])
	}
}

func TestSDKPRListHandler_HandlePRListRequest_InvalidRepositoryFormat(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{}
	handler := NewSDKPRListHandler(logger, mockClient)

	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string `json:"repository,omitempty"`
		CWD        string `json:"cwd,omitempty"`
		State      string `json:"state,omitempty"`
		Author     string `json:"author,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{
		Repository: "invalid-format",
	}

	result, data, err := handler.HandlePRListRequest(context.Background(), req, args)

	if result == nil {
		t.Error("Expected result to not be nil")
	}
	if data != nil {
		t.Error("Expected data to be nil for invalid format")
	}
	if err == nil {
		t.Error("Expected error for invalid repository format")
	}
	if !strings.Contains(err.Error(), "invalid repository format") {
		t.Errorf("Expected error to contain 'invalid repository format', got %v", err)
	}
}

func TestSDKRepositoryHandler_ListRepositories_Success(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{
		MockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "owner/test-repo",
				Private:  false,
				Owner: &gitea.User{
					UserName: "owner",
				},
			},
		},
	}

	handler := NewSDKRepositoryHandler(logger, mockClient)

	req := &mcp.CallToolRequest{}
	args := struct {
		Limit int `json:"limit,omitempty"`
	}{
		Limit: 10,
	}

	result, data, err := handler.ListRepositories(context.Background(), req, args)

	if result == nil {
		t.Error("Expected result to not be nil")
	}
	if data == nil {
		t.Error("Expected data to not be nil")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the result contains expected data
	resultData, ok := data.(map[string]any)
	if !ok {
		t.Fatal("Expected resultData to be map[string]any")
	}
	if resultData["total"] != 1 {
		t.Errorf("Expected total to be 1, got %v", resultData["total"])
	}
}

func TestSDKIssueListHandler_HandleIssueListRequest_ValidRepository(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{
		MockRepos: []*gitea.Repository{
			{
				ID:       1,
				Name:     "test-repo",
				FullName: "owner/test-repo",
				Private:  false,
				Owner: &gitea.User{
					UserName: "owner",
				},
			},
		},
		MockIssues: []*gitea.Issue{
			{
				Index: 1,
				Title: "Test Issue",
				State: gitea.StateOpen,
				Poster: &gitea.User{
					UserName: "testuser",
				},
			},
		},
	}

	handler := NewSDKIssueListHandler(logger, mockClient)

	req := &mcp.CallToolRequest{}
	args := struct {
		Repository string   `json:"repository,omitempty"`
		CWD        string   `json:"cwd,omitempty"`
		State      string   `json:"state,omitempty"`
		Author     string   `json:"author,omitempty"`
		Labels     []string `json:"labels,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{
		Repository: "owner/test-repo",
		State:      "open",
	}

	result, data, err := handler.HandleIssueListRequest(context.Background(), req, args)

	if result == nil {
		t.Error("Expected result to not be nil")
	}
	if data == nil {
		t.Error("Expected data to not be nil")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the result contains expected data
	resultData, ok := data.(map[string]any)
	if !ok {
		t.Fatal("Expected resultData to be map[string]any")
	}
	if resultData["total"] != 1 {
		t.Errorf("Expected total to be 1, got %v", resultData["total"])
	}
}

func TestHandlerTransformFunctions(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{}

	// Test PR transformation
	prHandler := NewSDKPRListHandler(logger, mockClient)
	prs := []*gitea.PullRequest{
		{
			Index: 1,
			Title: "Test PR",
			State: gitea.StateOpen,
			Poster: &gitea.User{
				UserName: "testuser",
			},
		},
	}

	repoMetadata := map[string]any{"name": "test-repo"}
	transformedPRs := prHandler.transformPRsToResponse(prs, repoMetadata)

	if len(transformedPRs) != 1 {
		t.Errorf("Expected 1 transformed PR, got %d", len(transformedPRs))
	}
	if transformedPRs[0]["title"] != "Test PR" {
		t.Errorf("Expected title 'Test PR', got %v", transformedPRs[0]["title"])
	}
	if transformedPRs[0]["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", transformedPRs[0]["state"])
	}
	if transformedPRs[0]["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", transformedPRs[0]["author"])
	}

	// Test repository transformation
	repoHandler := NewSDKRepositoryHandler(logger, mockClient)
	repos := []*gitea.Repository{
		{
			ID:       1,
			Name:     "test-repo",
			FullName: "owner/test-repo",
			Private:  false,
			Owner: &gitea.User{
				UserName: "owner",
			},
		},
	}

	transformedRepos := repoHandler.transformReposToResponse(repos)

	if len(transformedRepos) != 1 {
		t.Errorf("Expected 1 transformed repo, got %d", len(transformedRepos))
	}
	if transformedRepos[0]["name"] != "test-repo" {
		t.Errorf("Expected name 'test-repo', got %v", transformedRepos[0]["name"])
	}
	if transformedRepos[0]["owner"] != "owner" {
		t.Errorf("Expected owner 'owner', got %v", transformedRepos[0]["owner"])
	}

	// Test issue transformation
	issueHandler := NewSDKIssueListHandler(logger, mockClient)
	issues := []*gitea.Issue{
		{
			Index: 1,
			Title: "Test Issue",
			State: gitea.StateOpen,
			Poster: &gitea.User{
				UserName: "testuser",
			},
		},
	}

	transformedIssues := issueHandler.transformIssuesToResponse(issues, repoMetadata)

	if len(transformedIssues) != 1 {
		t.Errorf("Expected 1 transformed issue, got %d", len(transformedIssues))
	}
	if transformedIssues[0]["title"] != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %v", transformedIssues[0]["title"])
	}
	if transformedIssues[0]["state"] != "open" {
		t.Errorf("Expected state 'open', got %v", transformedIssues[0]["state"])
	}
	if transformedIssues[0]["author"] != "testuser" {
		t.Errorf("Expected author 'testuser', got %v", transformedIssues[0]["author"])
	}
}
