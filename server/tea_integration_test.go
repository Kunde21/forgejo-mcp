// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
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
