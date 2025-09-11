# Go Code Style Guide for forgejo-mcp

## Package Structure
- Use `main` package for executable code
- Use descriptive package names for test packages (e.g., `servertest`)
- Follow standard Go project layout conventions

## Imports
```go
import (
	"context"
	"fmt"
	"regexp"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)
```
- Standard library imports first
- Third-party imports second
- Local packages last
- Blank lines between import groups
- Use `goimports` for automatic formatting
- Use aliases for commonly used packages (e.g., `v` for validation)

## Naming Conventions
- **Exported types/functions**: PascalCase (e.g., `Server`, `NewServer`, `Start`)
- **Unexported types/functions**: camelCase (e.g., `handleHello`, `getEnv`)
- **Struct fields**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `mcpServer`, `config`)
- **Constants**: PascalCase (e.g., `DefaultPort`, `repoReg`)

## Type Definitions
```go
type Server struct {
	mcpServer   *mcp.MCPServer
	giteaService *gitea.Service
	config      *Config  // unexported field
}

type Config struct {
	Host string  // exported field
	Port int     // exported field
}

// Handler argument types
type IssueCommentArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Comment     string `json:"comment"`
}

// Result types
type CommentResult struct {
	Comment gitea.IssueComment `json:"comment"`
}
```

## Function Signatures
- Use pointer receivers for structs: `func (s *Server) Start() error`
- Constructor functions: `func NewServer() (*Server, error)`
- Handler functions with typed arguments: `func (s *Server) handleIssueCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args IssueCommentArgs) (*mcp.CallToolResult, *CommentResult, error)`

## Validation
```go
import v "github.com/go-ozzo/ozzo-validation/v4"

// Validate input arguments using ozzo-validation
if err := v.ValidateStruct(&args,
	v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
	v.Field(&args.IssueNumber, v.Min(1)),
	v.Field(&args.Comment, v.Required, v.Length(1, 0)), // Non-empty string
); err != nil {
	return TextErrorf("Invalid request: %v", err), nil, nil
}
```
- Use `ozzo-validation` for structured input validation
- Define validation rules with clear error messages
- Validate context first: `if ctx == nil { return TextError("Context is required"), nil, nil }`
- Use regex patterns for format validation (e.g., repository names)

## Error Handling
```go
func (s *Server) handleIssueCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args IssueCommentArgs) (*mcp.CallToolResult, *CommentResult, error) {
	// Validate context
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}

	// Validate input
	if err := v.ValidateStruct(&args, /* validation rules */); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Call service layer
	comment, err := s.giteaService.CreateIssueComment(ctx, args.Repository, args.IssueNumber, args.Comment)
	if err != nil {
		return TextErrorf("Failed to create comment: %v", err), nil, nil
	}

	// Return success
	return TextResult(responseText), &CommentResult{Comment: *comment}, nil
}
```
- Always check and handle errors
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Use helper functions for consistent error responses: `TextError()`, `TextErrorf()`
- Return errors immediately when appropriate
- Use structured error responses with `IsError: true`

## Response Helpers
```go
// Common response helpers in server/common.go
func TextResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}

func TextResultf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}
}

func TextError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}, IsError: true}
}

func TextErrorf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}, IsError: true}
}
```
- Use helper functions for consistent response formatting
- Separate success and error response builders
- Support formatted strings with `f` variants

## Documentation
```go
// handleIssueCommentCreate handles the "issue_comment_create" tool request.
// It creates a new comment on a specified Forgejo/Gitea issue.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - issue_number: The issue number to comment on (must be positive)
//   - comment: The comment content (cannot be empty)
//
// Returns:
//   - Success: Comment creation confirmation with metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handleIssueCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args IssueCommentArgs) (*mcp.CallToolResult, *CommentResult, error) {
	// implementation
}
```
- Add comprehensive godoc comments for all exported functions
- Include Parameters section documenting all arguments
- Include Returns section documenting success and error cases
- Add migration notes when updating from previous patterns
- Use complete sentences starting with the name of the item
- Keep comments concise and focused on "why" rather than "what"

## Testing
```go
type issueCommentEditTestCase struct {
	name      string
	setupMock func(*MockGiteaServer)
	arguments map[string]any
	expect    *mcp.CallToolResult
}

func TestCommentEditingRealWorldScenarios(t *testing.T) {
	t.Parallel()
	testCases := []issueCommentEditTestCase{
		{
			name: "fix_typo",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{
						ID:      1,
						Content: "This is a commment with a typo",
						Author:  "testuser",
						Created: "2024-01-01T00:00:00Z",
					},
				})
			},
			arguments: map[string]any{
				"repository":   "testuser/testrepo",
				"issue_number": 1,
				"comment_id":   1,
				"new_content":  "This is a comment with the typo fixed",
			},
			expect: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Comment edited successfully. ID: 123, Updated: 0001-01-01T00:00:00Z\nComment body: This is a comment with the typo fixed"},
				},
				StructuredContent: map[string]any{
					"comment": map[string]any{
						"id":      float64(123),
						"content": "This is a comment with the typo fixed",
						"author":  "testuser",
						"created": "0001-01-01T00:00:00Z",
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
			ts := NewTestServer(t, t.Context(), map[string]string{
				"FORGEJO_REMOTE_URL": mock.URL(),
				"FORGEJO_AUTH_TOKEN": "mock-token",
			})
			if err := ts.Initialize(); err != nil {
				t.Fatalf("Failed to initialize test server: %v", err)
			}

			result, err := ts.Client().CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "issue_comment_edit",
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("Failed to call issue_comment_edit tool: %v", err)
			}
			if !cmp.Equal(tc.expect, result) {
				t.Error(cmp.Diff(tc.expect, result))
			}
		})
	}
}
```
- Use table-driven tests with test case structs
- Include `setupMock` function for mock server configuration
- Test both success and error scenarios
- Use `t.Parallel()` for concurrent test execution
- Use descriptive test names that explain the scenario
- Include acceptance tests for multi-step workflows
- Test concurrent request handling
- Use `cmp.Equal` for test validation with detailed diffs

## Mock Server Testing
```go
// MockGiteaServer represents a mock Gitea API server for testing
type MockGiteaServer struct {
	server        *httptest.Server
	issues        map[string][]MockIssue
	comments      map[string][]MockComment
	pullRequests  map[string][]MockPullRequest
	notFoundRepos map[string]bool // Repositories that should return 404
	nextID        int
}

// NewMockGiteaServer creates a new mock Gitea server
func NewMockGiteaServer(t *testing.T) *MockGiteaServer {
	mock := &MockGiteaServer{
		issues:        make(map[string][]MockIssue),
		comments:      make(map[string][]MockComment),
		pullRequests:  make(map[string][]MockPullRequest),
		notFoundRepos: make(map[string]bool),
		nextID:        1,
	}

	// Create HTTP handler for mock API
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", mock.handleVersion)
	handler.HandleFunc("/api/v1/repos/", mock.handleRepoRequests)

	mock.server = httptest.NewServer(handler)
	t.Cleanup(mock.server.Close)
	return mock
}
```
- Create comprehensive mock servers for API testing
- Support multiple endpoint types (issues, comments, pull requests)
- Handle authentication and error scenarios
- Use `httptest.Server` for HTTP mocking
- Register cleanup with `t.Cleanup()`

## Test Server Setup
```go
// TestServer represents a test harness for running the MCP server
type TestServer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	t       *testing.T
	client  *mcp.Client
	session *mcp.ClientSession
	once    *sync.Once
	started bool
}

// NewTestServer creates a new TestServer instance
func NewTestServer(t *testing.T, ctx context.Context, env map[string]string) *TestServer {
	if ctx == nil {
		ctx = t.Context()
	}
	ctx, cancel := context.WithCancel(ctx)
	
	// Create in-memory transports for client-server communication
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	
	// Start server and connect client
	// ... implementation
	
	ts := &TestServer{
		ctx:     ctx,
		cancel:  cancel,
		t:       t,
		client:  client,
		session: session,
		once:    &sync.Once{},
	}

	t.Cleanup(func() {
		cancel()
		if err := session.Close(); err != nil {
			t.Log(err)
		}
	})
	return ts
}
```
- Use in-memory transports for client-server communication
- Support environment variable configuration
- Handle context cancellation and cleanup
- Use `t.Cleanup()` for resource management

## Code Organization
- Group related functionality together
- Keep files focused on a single action or record type
- Keep functions focused on single responsibility
- Use concise, meaningful variable names
- Avoid deep nesting - return early when possible
- Use blank lines to separate logical sections
- Store common patterns in `server/common.go`

## Dependencies
- Keep dependencies minimal and well-maintained
- Use specific versions in `go.mod`
- Run `goimports -w .` regularly to clean up unused imports
- Run `go mod tidy` regularly to clean up unused dependencies
- Use `github.com/go-ozzo/ozzo-validation/v4` for input validation
- Use `github.com/modelcontextprotocol/go-sdk/mcp` for MCP SDK

## Build and Quality
- Use `go build ./...` to build all packages
- Use `go test ./...` to run all tests
- Use `go vet ./...` for static analysis
- Use `goimports -w .` for consistent import formatting
- Ensure all code passes `go vet` checks
- Run tests in parallel with `t.Parallel()`

This style guide reflects the patterns observed in the current codebase and should be followed for all new code additions to maintain consistency.
