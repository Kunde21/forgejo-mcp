package servertest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCategory defines the type of test for organization
type TestCategory string

const (
	Unit        TestCategory = "Unit"
	Integration TestCategory = "Integration"
	Acceptance  TestCategory = "Acceptance"
	Validation  TestCategory = "validation"
	Success     TestCategory = "success"
	Pagination  TestCategory = "pagination"
	Performance TestCategory = "performance"
	Error       TestCategory = "error"
)

// prListTestCase represents a comprehensive test case for pull request listing
type prListTestCase struct {
	name           string
	category       TestCategory
	setupMock      func(*MockGiteaServer)
	setupDir       func(t *testing.T) string // Optional function to set up a temporary directory
	arguments      map[string]any
	expect         *mcp.CallToolResult
	expectError    bool
	errorSubstring string
	timeout        time.Duration
	validateFunc   func(*testing.T, *mcp.CallToolResult)
}

func TestPullRequestsList(t *testing.T) {
	t.Parallel()
	testCases := []prListTestCase{
		{
			name: "acceptance",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Feature: Add dark mode", State: "open"},
					{ID: 2, Number: 2, Title: "Fix: Memory leak", State: "open"},
					{ID: 3, Number: 3, Title: "Bug: Login fails", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Feature: Add dark mode",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Fix: Memory leak",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "pagination",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("testuser", "testrepo", prs)
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 1; i <= 10; i++ {
							prs = append(prs, map[string]any{
								"id":      float64(i),
								"number":  float64(i),
								"title":   fmt.Sprintf("Pull Request %d", i),
								"body":    "",
								"state":   "open",
								"user":    "testuser",
								"created": "2025-09-11T10:30:00Z",
								"updated": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
					}(),
				},
			},
		},
		{
			name: "invalid repository format",
			arguments: map[string]any{
				"repository": "invalid-repo-format",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: repository: repository must be in format 'owner/repo'."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "missing repository",
			arguments: map[string]any{
				"limit":  10,
				"offset": 0,
				"state":  "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: at least one of directory or repository must be provided; repository: at least one of directory or repository must be provided."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "invalid limit",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      200, // Invalid: > 100
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: limit: must be no greater than 100."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "invalid state",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "invalid-state", // Invalid state
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: state: state must be one of: open, closed, all."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "default values",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Default Test PR", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Default Test PR",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "invalid offset",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     -1, // Invalid: negative
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: offset: must be no less than 0."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		// Integration test cases
		{
			name: "complete workflow",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Feature: Add dark mode", State: "open"},
					{ID: 2, Number: 2, Title: "Fix: Memory leak", State: "open"},
					{ID: 3, Number: 3, Title: "Bug: Login fails", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Feature: Add dark mode",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Fix: Memory leak",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "successful parameters - minimal",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Test PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Test PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "successful parameters - all parameters",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      5,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Test PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Test PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "successful parameters - closed state",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo-closed", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo-closed",
				"state":      "closed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 pull requests"},
				},
				StructuredContent: map[string]any{},
			},
		},
		{
			name: "successful parameters - all states",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"state":      "all",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Test PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Test PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "state filtering - open",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
					{ID: 3, Number: 3, Title: "Closed PR 1", State: "closed"},
					{ID: 4, Number: 4, Title: "Closed PR 2", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Open PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Open PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "state filtering - closed",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
					{ID: 3, Number: 3, Title: "Closed PR 1", State: "closed"},
					{ID: 4, Number: 4, Title: "Closed PR 2", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
				"state":      "closed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(3),
							"number":  float64(3),
							"title":   "Closed PR 1",
							"body":    "",
							"state":   "closed",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(4),
							"number":  float64(4),
							"title":   "Closed PR 2",
							"body":    "",
							"state":   "closed",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "state filtering - all",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
					{ID: 3, Number: 3, Title: "Closed PR 1", State: "closed"},
					{ID: 4, Number: 4, Title: "Closed PR 2", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
				"state":      "all",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 4 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Open PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Open PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(3),
							"number":  float64(3),
							"title":   "Closed PR 1",
							"body":    "",
							"state":   "closed",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(4),
							"number":  float64(4),
							"title":   "Closed PR 2",
							"body":    "",
							"state":   "closed",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "state filtering - default state",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("owner", "repo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
					{ID: 3, Number: 3, Title: "Closed PR 1", State: "closed"},
					{ID: 4, Number: 4, Title: "Closed PR 2", State: "closed"},
				})
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 2 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Open PR 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
						map[string]any{
							"id":      float64(2),
							"number":  float64(2),
							"title":   "Open PR 2",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "pagination - first page",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 1; i <= 10; i++ {
							prs = append(prs, map[string]any{
								"id":      float64(i),
								"number":  float64(i),
								"title":   fmt.Sprintf("Pull Request %d", i),
								"body":    "",
								"state":   "open",
								"user":    "testuser",
								"created": "2025-09-11T10:30:00Z",
								"updated": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
					}(),
				},
			},
		},
		{
			name: "pagination - second page",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo-page2", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo-page2",
				"limit":      10,
				"offset":     10,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 10 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 11; i <= 20; i++ {
							prs = append(prs, map[string]any{
								"id":      float64(i),
								"number":  float64(i),
								"title":   fmt.Sprintf("Pull Request %d", i),
								"body":    "",
								"state":   "open",
								"user":    "testuser",
								"created": "2025-09-11T10:30:00Z",
								"updated": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
					}(),
				},
			},
		},
		{
			name: "pagination - third page",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo-page3", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo-page3",
				"limit":      10,
				"offset":     20,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 5 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 21; i <= 25; i++ {
							prs = append(prs, map[string]any{
								"id":      float64(i),
								"number":  float64(i),
								"title":   fmt.Sprintf("Pull Request %d", i),
								"body":    "",
								"state":   "open",
								"user":    "testuser",
								"created": "2025-09-11T10:30:00Z",
								"updated": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
					}(),
				},
			},
		},
		{
			name: "pagination - beyond available data",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo-beyond", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo-beyond",
				"limit":      10,
				"offset":     30,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 pull requests"},
				},
				StructuredContent: map[string]any{},
			},
		},
		{
			name: "pagination - single item pages",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo-single", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo-single",
				"limit":      1,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Pull Request 1",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "pagination - large limit",
			setupMock: func(mock *MockGiteaServer) {
				var prs []MockPullRequest
				for i := 1; i <= 25; i++ {
					prs = append(prs, MockPullRequest{
						ID:     i,
						Number: i,
						Title:  fmt.Sprintf("Pull Request %d", i),
						State:  "open",
					})
				}
				mock.AddPullRequests("owner", "repo", prs)
			},
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      100,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 25 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": func() []any {
						var prs []any
						for i := 1; i <= 25; i++ {
							prs = append(prs, map[string]any{
								"id":      float64(i),
								"number":  float64(i),
								"title":   fmt.Sprintf("Pull Request %d", i),
								"body":    "",
								"state":   "open",
								"user":    "testuser",
								"created": "2025-09-11T10:30:00Z",
								"updated": "2025-09-11T10:30:00Z",
								"head": map[string]any{
									"ref": "feature-branch",
									"sha": "abc123",
								},
								"base": map[string]any{
									"ref": "main",
									"sha": "def456",
								},
							})
						}
						return prs
					}(),
				},
			},
		},
		{
			name: "permission errors",
			arguments: map[string]any{
				"repository": "owner/repo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 pull requests"},
				},
				StructuredContent: map[string]any{},
			},
		},
		{
			name: "API failures",
			arguments: map[string]any{
				"repository": "nonexistent/repo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 0 pull requests"},
				},
				StructuredContent: map[string]any{},
			},
		},
		{
			name: "directory parameter - valid git repository",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Directory-based PR", State: "open"},
				})
			},
			setupDir: func(t *testing.T) string {
				return createTempGitRepo(t, "testuser", "testrepo")
			},
			arguments: map[string]any{
				"directory": "", // Will be set dynamically
				"limit":     10,
				"offset":    0,
				"state":     "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Directory-based PR",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
		{
			name: "directory parameter - non-existent directory",
			arguments: map[string]any{
				"directory": "/non/existent/directory",
				"limit":     10,
				"offset":    0,
				"state":     "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "directory parameter - directory without git",
			setupDir: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "forgejo-test-no-git-*")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}
				// Don't create .git directory - this should fail validation
				t.Cleanup(func() {
					os.RemoveAll(tempDir)
				})
				return tempDir
			},
			arguments: map[string]any{
				"directory": "", // Will be set dynamically
				"limit":     10,
				"offset":    0,
				"state":     "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid request: directory: invalid directory."},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "both directory and repository provided (directory takes precedence)",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR", State: "open"},
				})
			},
			arguments: map[string]any{
				"directory":  "/tmp/test-repo",
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Failed to resolve directory: repository validate failed for /tmp/test-repo: directory does not exist"},
				},
				StructuredContent: map[string]any{},
				IsError:           true,
			},
		},
		{
			name: "repository parameter test",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Legacy repository PR", State: "open"},
				})
			},
			arguments: map[string]any{
				"repository": "testuser/testrepo",
				"limit":      10,
				"offset":     0,
				"state":      "open",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Found 1 pull requests"},
				},
				StructuredContent: map[string]any{
					"pull_requests": []any{
						map[string]any{
							"id":      float64(1),
							"number":  float64(1),
							"title":   "Legacy repository PR",
							"body":    "",
							"state":   "open",
							"user":    "testuser",
							"created": "2025-09-11T10:30:00Z",
							"updated": "2025-09-11T10:30:00Z",
							"head": map[string]any{
								"ref": "feature-branch",
								"sha": "abc123",
							},
							"base": map[string]any{
								"ref": "main",
								"sha": "def456",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Set up temporary directory if needed
			var tempDir string
			if tc.setupDir != nil {
				tempDir = tc.setupDir(t)
				// Update arguments with the actual temp directory path
				args := make(map[string]any)
				for k, v := range tc.arguments {
					args[k] = v
				}
				if dir, ok := args["directory"].(string); ok && dir == "" {
					args["directory"] = tempDir
				}
				tc.arguments = args
			}

			ts := NewTestServer(t, t.Context(), map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "pr_list",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call pr_list tool: %v", err)
			}

			// Special handling for directory without git test (dynamic path)
			if tc.name == "directory parameter - directory without git" {
				if !result.IsError {
					t.Errorf("Expected error result for test case: %s", tc.name)
				}
				textContent := GetTextContent(result.Content)
				if !strings.Contains(textContent, "Failed to resolve directory: not a git repository:") {
					t.Errorf("Expected error message containing 'Failed to resolve directory: not a git repository:', got: %s", textContent)
				}
			} else {
				if !cmp.Equal(tc.expect, result) {
					t.Error(cmp.Diff(tc.expect, result))
				}
			}
		})
	}
}

func TestListPullRequestsConcurrent(t *testing.T) {
	mock := NewMockGiteaServer(t)
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Concurrent PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Concurrent PR 2", State: "open"},
	})
	ts := NewTestServer(t, t.Context(), map[string]string{
		"FORGEJO_REMOTE_URL": mock.URL(),
		"FORGEJO_AUTH_TOKEN": "mock-token",
	})
	if err := ts.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test server: %v", err)
	}

	const numGoroutines = 5
	results := make(chan error, numGoroutines)
	for range numGoroutines {
		go func() {
			_, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name: "pr_list",
				Arguments: map[string]any{
					"repository": "testuser/testrepo",
					"limit":      10,
					"offset":     0,
					"state":      "open",
				},
			})
			results <- err
		}()
	}
	for range numGoroutines {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
