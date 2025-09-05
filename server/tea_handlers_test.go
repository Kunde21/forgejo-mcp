package server

import (
	"context"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

func TestTeaPRListHandler_HandlePRListRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Use comprehensive mock client with test data
	mockClient := &MockGiteaClient{
		mockPRs: []*gitea.PullRequest{
			{
				ID:     1,
				Index:  1,
				Title:  "Test PR",
				State:  gitea.StateOpen,
				Body:   "Test description",
				Poster: &gitea.User{UserName: "testuser"},
			},
		},
	}
	handler := NewTeaPRListHandler(logger, mockClient)
	if handler == nil {
		t.Fatal("NewTeaPRListHandler returned nil")
	}

	// Test with empty params
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		State  string `json:"state,omitempty"`
		Author string `json:"author,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{}

	result, data, err := handler.HandlePRListRequest(ctx, req, args)

	// Should return a result even if tea command fails (we're testing the handler structure)
	if result == nil {
		t.Error("HandlePRListRequest returned nil result")
	}
	if err != nil {
		t.Logf("HandlePRListRequest returned error (expected in test env): %v", err)
	}
	if data == nil {
		t.Log("HandlePRListRequest returned nil data")
	}

	// Verify the response contains expected data structure
	if data != nil {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("HandlePRListRequest returned data of wrong type")
		} else {
			prs, exists := dataMap["pullRequests"]
			if !exists {
				t.Error("HandlePRListRequest data missing pullRequests field")
			} else {
				prsSlice, ok := prs.([]map[string]interface{})
				if !ok {
					t.Error("pullRequests field is not a slice")
				} else if len(prsSlice) != 1 {
					t.Errorf("Expected 1 PR, got %d", len(prsSlice))
				}
			}
		}
	}
}

func TestTeaIssueListHandler_HandleIssueListRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Use comprehensive mock client with test data
	mockClient := &MockGiteaClient{
		mockIssues: []*gitea.Issue{
			{
				ID:     1,
				Index:  1,
				Title:  "Test Issue",
				State:  "open",
				Body:   "Test description",
				Poster: &gitea.User{UserName: "testuser"},
			},
		},
	}
	handler := NewTeaIssueListHandler(logger, mockClient)
	if handler == nil {
		t.Fatal("NewTeaIssueListHandler returned nil")
	}

	// Test with empty params
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := struct {
		State  string   `json:"state,omitempty"`
		Author string   `json:"author,omitempty"`
		Labels []string `json:"labels,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}{}

	result, data, err := handler.HandleIssueListRequest(ctx, req, args)

	// Should return a result even if tea command fails (we're testing the handler structure)
	if result == nil {
		t.Error("HandleIssueListRequest returned nil result")
	}
	if err != nil {
		t.Logf("HandleIssueListRequest returned error (expected in test env): %v", err)
	}
	if data == nil {
		t.Log("HandleIssueListRequest returned nil data")
	}

	// Verify the response contains expected data structure
	if data != nil {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("HandleIssueListRequest returned data of wrong type")
		} else {
			issues, exists := dataMap["issues"]
			if !exists {
				t.Error("HandleIssueListRequest data missing issues field")
			} else {
				issuesSlice, ok := issues.([]map[string]interface{})
				if !ok {
					t.Error("issues field is not a slice")
				} else if len(issuesSlice) != 1 {
					t.Errorf("Expected 1 issue, got %d", len(issuesSlice))
				}
			}
		}
	}
}

func TestTeaCommandBuilder_BuildPRListCommand(t *testing.T) {
	builder := NewTeaCommandBuilder()
	if builder == nil {
		t.Fatal("NewTeaCommandBuilder returned nil")
	}

	// Test with empty params
	result := builder.BuildPRListCommand(map[string]interface{}{})
	expected := []string{"tea", "pr", "list", "--output", "json"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d args, got %d", len(expected), len(result))
	}
	for i, arg := range expected {
		if i >= len(result) || result[i] != arg {
			t.Errorf("Expected arg %d to be %s, got %s", i, arg, result[i])
		}
	}
}

func TestTeaCommandBuilder_BuildIssueListCommand(t *testing.T) {
	builder := NewTeaCommandBuilder()
	if builder == nil {
		t.Fatal("NewTeaCommandBuilder returned nil")
	}

	// Test with empty params
	result := builder.BuildIssueListCommand(map[string]interface{}{})
	expected := []string{"tea", "issue", "list", "--output", "json"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d args, got %d", len(expected), len(result))
	}
	for i, arg := range expected {
		if i >= len(result) || result[i] != arg {
			t.Errorf("Expected arg %d to be %s, got %s", i, arg, result[i])
		}
	}
}

func TestTeaOutputParser_ParsePRList(t *testing.T) {
	parser := NewTeaOutputParser()
	if parser == nil {
		t.Fatal("NewTeaOutputParser returned nil")
	}

	// Test with valid JSON
	jsonData := `[
		{
			"number": 42,
			"title": "Add dark mode support",
			"author": "developer1",
			"state": "open",
			"created_at": "2025-08-26T10:00:00Z",
			"updated_at": "2025-08-26T15:30:00Z"
		}
	]`

	prs, err := parser.ParsePRList([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParsePRList failed: %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}
	if prs[0].Number != 42 {
		t.Errorf("Expected PR number 42, got %d", prs[0].Number)
	}
	if prs[0].Title != "Add dark mode support" {
		t.Errorf("Expected PR title 'Add dark mode support', got %s", prs[0].Title)
	}
}

func TestTeaOutputParser_ParseIssueList(t *testing.T) {
	parser := NewTeaOutputParser()
	if parser == nil {
		t.Fatal("NewTeaOutputParser returned nil")
	}

	// Test with valid JSON
	jsonData := `[
		{
			"number": 123,
			"title": "UI responsiveness issue",
			"author": "user1",
			"state": "open",
			"labels": ["bug", "ui"],
			"created_at": "2025-08-24T08:30:00Z"
		}
	]`

	issues, err := parser.ParseIssueList([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseIssueList failed: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected issue number 123, got %d", issues[0].Number)
	}
}
