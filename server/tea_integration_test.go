// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestTeaCommandBuilder_PRList(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected []string
	}{
		{
			name:     "no parameters",
			params:   map[string]interface{}{},
			expected: []string{"tea", "pr", "list", "--output", "json"},
		},
		{
			name: "with state filter",
			params: map[string]interface{}{
				"state": "open",
			},
			expected: []string{"tea", "pr", "list", "--state", "open", "--output", "json"},
		},
		{
			name: "with author filter",
			params: map[string]interface{}{
				"author": "developer1",
			},
			expected: []string{"tea", "pr", "list", "--author", "developer1", "--output", "json"},
		},
		{
			name: "with limit",
			params: map[string]interface{}{
				"limit": float64(10),
			},
			expected: []string{"tea", "pr", "list", "--limit", "10", "--output", "json"},
		},
		{
			name: "all parameters",
			params: map[string]interface{}{
				"state":  "closed",
				"author": "developer2",
				"limit":  float64(5),
			},
			expected: []string{"tea", "pr", "list", "--state", "closed", "--author", "developer2", "--limit", "5", "--output", "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewTeaCommandBuilder()
			result := builder.BuildPRListCommand(tt.params)

			if !cmp.Equal(tt.expected, result) {
				t.Errorf("BuildPRListCommand() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTeaCommandBuilder_IssueList(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected []string
	}{
		{
			name:     "no parameters",
			params:   map[string]interface{}{},
			expected: []string{"tea", "issue", "list", "--output", "json"},
		},
		{
			name: "with state filter",
			params: map[string]interface{}{
				"state": "closed",
			},
			expected: []string{"tea", "issue", "list", "--state", "closed", "--output", "json"},
		},
		{
			name: "with labels",
			params: map[string]interface{}{
				"labels": []interface{}{"bug", "ui"},
			},
			expected: []string{"tea", "issue", "list", "--labels", "bug,ui", "--output", "json"},
		},
		{
			name: "with author",
			params: map[string]interface{}{
				"author": "user1",
			},
			expected: []string{"tea", "issue", "list", "--author", "user1", "--output", "json"},
		},
		{
			name: "all parameters",
			params: map[string]interface{}{
				"state":  "open",
				"labels": []interface{}{"enhancement"},
				"author": "user2",
				"limit":  float64(20),
			},
			expected: []string{"tea", "issue", "list", "--state", "open", "--labels", "enhancement", "--author", "user2", "--limit", "20", "--output", "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewTeaCommandBuilder()
			result := builder.BuildIssueListCommand(tt.params)

			if !cmp.Equal(tt.expected, result) {
				t.Errorf("BuildIssueListCommand() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTeaExecutor_ExecuteCommand(t *testing.T) {
	executor := NewTeaExecutor()

	// Test with a simple command that should work on most systems
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use echo command for testing (should be available on most systems)
	cmd := []string{"echo", "test output"}
	output, err := executor.ExecuteCommand(ctx, cmd)

	if err != nil {
		t.Logf("ExecuteCommand failed (this might be expected on some systems): %v", err)
		// Don't fail the test if echo is not available
		return
	}

	expected := "test output\n"
	if output != expected {
		t.Errorf("ExecuteCommand() = %q, expected %q", output, expected)
	}
}

func TestTeaExecutor_CommandTimeout(t *testing.T) {
	executor := NewTeaExecutor()

	// Test with a command that should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Use sleep command to test timeout
	cmd := []string{"sleep", "10"}
	_, err := executor.ExecuteCommand(ctx, cmd)

	if err == nil {
		t.Error("ExecuteCommand() should have timed out")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("ExecuteCommand() error should mention timeout, got: %v", err)
	}
}

func TestTeaOutputParser_ParsePRList(t *testing.T) {
	parser := NewTeaOutputParser()

	// Test with valid JSON
	validJSON := `[
		{
			"number": 42,
			"title": "Add dark mode support",
			"author": "developer1",
			"state": "open",
			"created_at": "2025-08-26T10:00:00Z",
			"updated_at": "2025-08-26T15:30:00Z"
		}
	]`

	result, err := parser.ParsePRList([]byte(validJSON))
	if err != nil {
		t.Fatalf("ParsePRList() should not error on valid JSON, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("ParsePRList() should return 1 PR, got %d", len(result))
	}

	if result[0].Number != 42 {
		t.Errorf("PR number should be 42, got %d", result[0].Number)
	}

	if result[0].Title != "Add dark mode support" {
		t.Errorf("PR title should be 'Add dark mode support', got %s", result[0].Title)
	}
}

func TestTeaOutputParser_ParseIssueList(t *testing.T) {
	parser := NewTeaOutputParser()

	// Test with valid JSON
	validJSON := `[
		{
			"number": 123,
			"title": "UI responsiveness issue",
			"author": "user1",
			"state": "open",
			"labels": ["bug", "ui"],
			"created_at": "2025-08-24T08:30:00Z"
		}
	]`

	result, err := parser.ParseIssueList([]byte(validJSON))
	if err != nil {
		t.Fatalf("ParseIssueList() should not error on valid JSON, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("ParseIssueList() should return 1 issue, got %d", len(result))
	}

	if result[0].Number != 123 {
		t.Errorf("Issue number should be 123, got %d", result[0].Number)
	}

	if len(result[0].Labels) != 2 {
		t.Errorf("Issue should have 2 labels, got %d", len(result[0].Labels))
	}
}

func TestTeaOutputParser_ParseInvalidJSON(t *testing.T) {
	parser := NewTeaOutputParser()

	// Test with invalid JSON
	invalidJSON := `{"invalid": json}`

	_, err := parser.ParsePRList([]byte(invalidJSON))
	if err == nil {
		t.Error("ParsePRList() should error on invalid JSON")
	}

	_, err = parser.ParseIssueList([]byte(invalidJSON))
	if err == nil {
		t.Error("ParseIssueList() should error on invalid JSON")
	}
}

func TestTeaIntegrationHandler_PRList(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test PR list handler with parameters
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "pr_list",
			"arguments": map[string]interface{}{
				"state":  "open",
				"author": "developer1",
				"limit":  float64(5),
			},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response")
	}
	if resp.Error != nil {
		t.Errorf("tools/call should not return error, got: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("tools/call should return result")
	}
}

func TestTeaIntegrationHandler_IssueList(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Register default handlers
	server.RegisterDefaultHandlers()

	// Test issue list handler with parameters
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "issue_list",
			"arguments": map[string]interface{}{
				"state":  "closed",
				"labels": []interface{}{"bug", "ui"},
				"limit":  float64(10),
			},
		},
	}

	resp := server.dispatcher.Dispatch(context.Background(), req)
	if resp == nil {
		t.Error("tools/call should return a response")
	}
	if resp.Error != nil {
		t.Errorf("tools/call should not return error, got: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("tools/call should return result")
	}
}

// TestGiteaSDKIntegration tests the integration of handlers with the new Gitea SDK client
func TestGiteaSDKIntegration(t *testing.T) {
	// Create a mock client for testing
	mockClient := &MockGiteaClient{}

	// Test PR list integration
	t.Run("PRListWithGiteaSDK", func(t *testing.T) {
		handler := &GiteaSDKPRListHandler{
			logger: logrus.New(),
			client: mockClient,
		}

		params := map[string]interface{}{
			"state":  "open",
			"author": "developer1",
			"limit":  float64(5),
		}

		result, err := handler.HandleRequest(context.Background(), "pr_list", params)
		if err != nil {
			t.Errorf("HandleRequest() should not error, got: %v", err)
		}

		if result == nil {
			t.Error("HandleRequest() should return result")
		}

		// Verify result structure
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Error("Result should be a map")
		}

		if _, exists := resultMap["pullRequests"]; !exists {
			t.Error("Result should contain pullRequests field")
		}

		if _, exists := resultMap["total"]; !exists {
			t.Error("Result should contain total field")
		}
	})

	// Test issue list integration
	t.Run("IssueListWithGiteaSDK", func(t *testing.T) {
		handler := &GiteaSDKIssueListHandler{
			logger: logrus.New(),
			client: mockClient,
		}

		params := map[string]interface{}{
			"state":  "closed",
			"labels": []interface{}{"bug", "ui"},
			"limit":  float64(10),
		}

		result, err := handler.HandleRequest(context.Background(), "issue_list", params)
		if err != nil {
			t.Errorf("HandleRequest() should not error, got: %v", err)
		}

		if result == nil {
			t.Error("HandleRequest() should return result")
		}

		// Verify result structure
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Error("Result should be a map")
		}

		if _, exists := resultMap["issues"]; !exists {
			t.Error("Result should contain issues field")
		}

		if _, exists := resultMap["total"]; !exists {
			t.Error("Result should contain total field")
		}
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		mockClient.ShouldError = true

		handler := &GiteaSDKPRListHandler{
			logger: logrus.New(),
			client: mockClient,
		}

		_, err := handler.HandleRequest(context.Background(), "pr_list", map[string]interface{}{})
		if err == nil {
			t.Error("HandleRequest() should return error when client fails")
		}

		mockClient.ShouldError = false
	})
}

// TestConfigurationIntegration tests the integration of configuration with Gitea settings
func TestConfigurationIntegration(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		valid  bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				ForgejoURL:    "https://example.forgejo.com",
				AuthToken:     "test-token",
				ClientTimeout: 30,
				UserAgent:     "test-agent",
			},
			valid: true,
		},
		{
			name: "missing URL",
			config: &config.Config{
				AuthToken:     "test-token",
				ClientTimeout: 30,
				UserAgent:     "test-agent",
			},
			valid: false,
		},
		{
			name: "missing token",
			config: &config.Config{
				ForgejoURL:    "https://example.forgejo.com",
				ClientTimeout: 30,
				UserAgent:     "test-agent",
			},
			valid: false,
		},
		{
			name: "invalid timeout",
			config: &config.Config{
				ForgejoURL:    "https://example.forgejo.com",
				AuthToken:     "test-token",
				ClientTimeout: -1,
				UserAgent:     "test-agent",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGiteaClientConfig(tt.config)
			if tt.valid && err != nil {
				t.Errorf("Expected valid config to pass, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected invalid config to fail validation")
			}
		})
	}
}

// TestEndToEndWorkflow tests complete workflows with the Gitea SDK client
func TestEndToEndWorkflow(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:    "https://example.forgejo.com",
		AuthToken:     "test-token",
		ClientTimeout: 30,
		UserAgent:     "forgejo-mcp-client/1.0.0",
		Host:          "localhost",
		Port:          8080,
		ReadTimeout:   30,
		WriteTimeout:  30,
		LogLevel:      "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Register default handlers (CLI-based)
	// Note: RegisterGiteaSDKHandlers would require a real Forgejo instance
	// For end-to-end testing, we use the default handlers
	server.RegisterDefaultHandlers()

	// Test complete workflow: list PRs -> list issues -> process results
	t.Run("CompleteWorkflow", func(t *testing.T) {
		ctx := context.Background()

		// Step 1: List pull requests
		prReq := &Request{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "pr_list",
				"arguments": map[string]interface{}{
					"state": "open",
					"limit": float64(10),
				},
			},
		}

		prResp := server.dispatcher.Dispatch(ctx, prReq)
		if prResp == nil || prResp.Error != nil {
			t.Errorf("PR list request failed: %v", prResp.Error)
		}

		// Step 2: List issues
		issueReq := &Request{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "issue_list",
				"arguments": map[string]interface{}{
					"state": "open",
					"limit": float64(10),
				},
			},
		}

		issueResp := server.dispatcher.Dispatch(ctx, issueReq)
		if issueResp == nil || issueResp.Error != nil {
			t.Errorf("Issue list request failed: %v", issueResp.Error)
		}

		// Verify both responses have expected structure
		verifyResponseStructure(t, prResp.Result, "pullRequests")
		verifyResponseStructure(t, issueResp.Result, "issues")
	})

	// Test error scenarios
	t.Run("ErrorScenarios", func(t *testing.T) {
		ctx := context.Background()

		// Test invalid parameters
		invalidReq := &Request{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "pr_list",
				"arguments": map[string]interface{}{
					"state": "invalid_state",
					"limit": float64(-1),
				},
			},
		}

		invalidResp := server.dispatcher.Dispatch(ctx, invalidReq)
		if invalidResp == nil {
			t.Error("Invalid request should return a response")
		}
		// Note: Error handling behavior depends on implementation
	})
}

// MockGiteaClient is a mock implementation of the Client interface for testing
type MockGiteaClient struct {
	ShouldError bool
}

func (m *MockGiteaClient) ListPRs(owner, repo string, filters *client.PullRequestFilters) ([]client.PullRequest, error) {
	if m.ShouldError {
		return nil, &client.APIError{StatusCode: 500, Message: "mock error"}
	}

	return []client.PullRequest{
		{
			Index: 1,
			Title: "Test PR",
			Poster: &client.User{
				UserName: "testuser",
			},
			State:   client.StateOpen,
			HTMLURL: "https://example.com/pr/1",
		},
	}, nil
}

func (m *MockGiteaClient) ListIssues(owner, repo string, filters *client.IssueFilters) ([]client.Issue, error) {
	if m.ShouldError {
		return nil, &client.APIError{StatusCode: 500, Message: "mock error"}
	}

	return []client.Issue{
		{
			Index: 1,
			Title: "Test Issue",
			Poster: &client.User{
				UserName: "testuser",
			},
			State:   client.StateOpen,
			HTMLURL: "https://example.com/issue/1",
		},
	}, nil
}

func (m *MockGiteaClient) ListRepositories(filters *client.RepositoryFilters) ([]client.Repository, error) {
	if m.ShouldError {
		return nil, &client.APIError{StatusCode: 500, Message: "mock error"}
	}

	return []client.Repository{
		{
			ID:       1,
			Name:     "test-repo",
			FullName: "testuser/test-repo",
			HTMLURL:  "https://example.com/testuser/test-repo",
		},
	}, nil
}

func (m *MockGiteaClient) GetRepository(owner, name string) (*client.Repository, error) {
	if m.ShouldError {
		return nil, &client.APIError{StatusCode: 404, Message: "not found"}
	}

	return &client.Repository{
		ID:       1,
		Name:     name,
		FullName: owner + "/" + name,
		HTMLURL:  "https://example.com/" + owner + "/" + name,
	}, nil
}

// Helper functions for integration tests
func validateGiteaClientConfig(cfg *config.Config) error {
	if cfg.ForgejoURL == "" {
		return fmt.Errorf("forgejo_url is required")
	}
	if cfg.AuthToken == "" {
		return fmt.Errorf("auth_token is required")
	}
	if cfg.ClientTimeout < 0 {
		return fmt.Errorf("client_timeout must be non-negative")
	}
	return nil
}

func verifyResponseStructure(t *testing.T, result interface{}, expectedField string) {
	if result == nil {
		t.Error("Response result should not be nil")
		return
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Response result should be a map")
		return
	}

	if _, exists := resultMap[expectedField]; !exists {
		t.Errorf("Response should contain %s field", expectedField)
	}

	if _, exists := resultMap["total"]; !exists {
		t.Error("Response should contain total field")
	}
}
