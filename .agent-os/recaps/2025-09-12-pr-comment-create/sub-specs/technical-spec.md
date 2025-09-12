# Technical Specification

> Feature: Pull Request Comment Create Tool
> Created: 2025-09-12

## Architecture Overview

The `pr_comment_create` tool will follow the established clean architecture patterns in the forgejo-mcp codebase, with a clear separation of concerns between validation (server handler) and business logic (service layer).

## Layer Architecture

### 1. Interface Layer (`remote/gitea/interface.go`)

#### PullRequestCommenter Interface
```go
// PullRequestCommenter defines the interface for creating comments on Git repository pull requests
type PullRequestCommenter interface {
    CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error)
}
```

#### CreatePullRequestCommentArgs Struct
```go
// CreatePullRequestCommentArgs represents the arguments for creating a pull request comment
type CreatePullRequestCommentArgs struct {
    Repository        string `json:"repository"`
    PullRequestNumber int    `json:"pull_request_number"`
    Comment           string `json:"comment"`
}
```

#### Interface Integration
The `GiteaClientInterface` will be extended to include `PullRequestCommenter`:
```go
// GiteaClientInterface combines all interfaces for complete Gitea operations
type GiteaClientInterface interface {
    IssueLister
    IssueCommenter
    IssueCommentLister
    IssueCommentEditor
    PullRequestLister
    PullRequestCommentLister
    PullRequestCommenter  // New interface
}
```

### 2. Client Layer (`remote/gitea/gitea_client.go`)

#### CreatePullRequestComment Method
```go
// CreatePullRequestComment creates a comment on the specified pull request
func (c *GiteaClient) CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error) {
    // Parse repository string (format: "owner/repo")
    owner, repoName, ok := strings.Cut(repo, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
    }

    // Create comment using Gitea SDK
    opts := gitea.CreateIssueCommentOption{
        Body: comment,
    }

    giteaComment, _, err := c.client.CreateIssueComment(owner, repoName, int64(pullRequestNumber), opts)
    if err != nil {
        return nil, fmt.Errorf("failed to create pull request comment: %w", err)
    }

    // Convert to our PullRequestComment struct
    prComment := &PullRequestComment{
        ID:        int(giteaComment.ID),
        Body:      giteaComment.Body,
        User:      giteaComment.Poster.UserName,
        CreatedAt: giteaComment.Created.Format("2006-01-02T15:04:05Z"),
        UpdatedAt: giteaComment.Updated.Format("2006-01-02T15:04:05Z"),
    }

    return prComment, nil
}
```

**Key Points:**
- No input validation - trust that inputs are already validated by server handler
- Uses Gitea SDK's `CreateIssueComment` (works for both issues and PRs)
- Proper error wrapping with context
- Converts Gitea SDK response to internal struct format

### 3. Service Layer (`remote/gitea/service.go`)

#### CreatePullRequestComment Method
```go
// CreatePullRequestComment creates a comment on a pull request
func (s *Service) CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*PullRequestComment, error) {
    // Call the underlying client directly (no validation)
    return s.client.CreatePullRequestComment(ctx, repo, pullRequestNumber, comment)
}
```

**Key Points:**
- No validation logic - clean separation of concerns
- Direct pass-through to client layer
- Focus on business logic delegation
- Returns client response with proper error propagation

### 4. Server Layer (`server/pr_comments.go`)

#### PullRequestCommentCreateArgs Struct
```go
// PullRequestCommentCreateArgs represents the arguments for creating a pull request comment with validation tags
type PullRequestCommentCreateArgs struct {
    Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
    PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
    Comment           string `json:"comment" validate:"required,min=1"`
}
```

#### PullRequestCommentCreateResult Struct
```go
// PullRequestCommentCreateResult represents the result data for the pr_comment_create tool
type PullRequestCommentCreateResult struct {
    Comment gitea.PullRequestComment `json:"comment"`
}
```

#### handlePullRequestCommentCreate Handler
```go
// handlePullRequestCommentCreate handles the "pr_comment_create" tool request.
// It creates a new comment on a specified Forgejo/Gitea pull request.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - pull_request_number: The pull request number to comment on (must be positive)
//   - comment: The comment content (cannot be empty)
//
// Returns:
//   - Success: Comment creation confirmation with metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handlePullRequestCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCommentCreateArgs) (*mcp.CallToolResult, *PullRequestCommentCreateResult, error) {
    // Validate context - required for proper request handling
    if ctx == nil {
        return TextError("Context is required"), nil, nil
    }

    // Validate input arguments using ozzo-validation
    if err := v.ValidateStruct(&args,
        v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
        v.Field(&args.PullRequestNumber, v.Min(1)),
        v.Field(&args.Comment, v.Required, v.Length(1, 0)), // Non-empty string
    ); err != nil {
        return TextErrorf("Invalid request: %v", err), nil, nil
    }

    // Create the comment using the service layer
    comment, err := s.giteaService.CreatePullRequestComment(ctx, args.Repository, args.PullRequestNumber, args.Comment)
    if err != nil {
        return TextErrorf("Failed to create pull request comment: %v", err), nil, nil
    }

    // Format success response with comment metadata
    responseText := fmt.Sprintf("Pull request comment created successfully. ID: %d, Created: %s\nComment body: %s",
        comment.ID, comment.CreatedAt, comment.Body)

    return TextResult(responseText), &PullRequestCommentCreateResult{Comment: *comment}, nil
}
```

**Key Points:**
- All validation performed using ozzo-validation
- Comprehensive validation for repository format, PR number, and comment content
- Proper error handling for validation and API errors
- Structured response with both human-readable text and JSON data
- Follows existing patterns from `handleIssueCommentCreate`

### 5. Server Registration (`server/server.go`)

#### Tool Registration
```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_comment_create",
    Description: "Create a comment on a Forgejo/Gitea repository pull request",
}, s.handlePullRequestCommentCreate)
```

## Data Flow

1. **MCP Request** → Server Handler with validation
2. **Validated Request** → Service Layer (no validation)
3. **Service Call** → Client Layer (no validation)
4. **Client Call** → Gitea SDK API
5. **API Response** → Client conversion
6. **Converted Response** → Service return
7. **Service Response** → Handler formatting
8. **Formatted Response** → MCP Response

## Validation Strategy

### Server Layer Validation (Only)
- **Repository Format**: Regex validation for `owner/repo` format
- **Pull Request Number**: Must be positive integer (> 0)
- **Comment Content**: Must be non-empty string (not just whitespace)

### Service Layer Trust
- Assumes all inputs are already validated
- No validation logic duplication
- Focus on business logic and error handling

### Client Layer Trust
- Assumes all inputs are already validated
- No validation logic
- Focus on API interaction and data conversion

## Error Handling

### Validation Errors
- Handled in server handler using ozzo-validation
- Return structured error responses with clear messages
- Examples: "repository must be in format 'owner/repo'", "pull_request_number must be positive"

### API Errors
- Handled in client layer with proper error wrapping
- Propagated through service layer to server handler
- Return structured error responses with API context
- Examples: "failed to create pull request comment: PR not found"

### Context Errors
- Handled at each layer with proper context checking
- Return appropriate error messages for context cancellation

## Testing Strategy

### Unit Tests
- **Client Layer**: Test API interaction and data conversion
- **Service Layer**: Test business logic delegation (no validation tests)
- **Server Layer**: Test validation logic and response formatting

### Integration Tests
- **End-to-End**: Test complete workflow from MCP request to response
- **Error Scenarios**: Test various error conditions and handling
- **Mock Server**: Use existing test harness with mock Gitea server

### Acceptance Tests
- **Real-world Scenarios**: Test typical usage patterns
- **Integration**: Test with existing tools and workflows
- **Performance**: Verify response time requirements

## MCP Protocol Compliance

### Request Format
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_comment_create",
    "arguments": {
      "repository": "owner/repo",
      "pull_request_number": 42,
      "comment": "This is a helpful comment on the pull request."
    }
  }
}
```

### Success Response Format
```json
{
  "content": [
    {
      "type": "text",
      "text": "Pull request comment created successfully. ID: 123, Created: 2025-09-12T14:30:45Z\nComment body: This is a helpful comment on the pull request."
    }
  ],
  "data": {
    "comment": {
      "id": 123,
      "body": "This is a helpful comment on the pull request.",
      "user": "username",
      "created_at": "2025-09-12T14:30:45Z",
      "updated_at": "2025-09-12T14:30:45Z"
    }
  }
}
```

### Error Response Format
```json
{
  "content": [
    {
      "type": "text",
      "text": "Invalid request: repository must be in format 'owner/repo'"
    }
  ],
  "isError": true
}
```

## Dependencies

### External Dependencies
- `github.com/go-ozzo/ozzo-validation/v4` - Input validation
- `code.gitea.io/sdk/gitea` - Gitea SDK for API interaction
- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP SDK v0.4.0

### Internal Dependencies
- Existing `PullRequestComment` struct (from listing functionality)
- Existing `repoReg` regex pattern (from other handlers)
- Existing `TextResult` and `TextError` helper functions
- Existing test harness and mock server infrastructure

## Performance Considerations

- **Response Time**: Target <2 seconds for typical operations
- **Memory Usage**: Minimal memory footprint with proper cleanup
- **Concurrency**: Safe for concurrent use with context handling
- **Error Recovery**: Graceful handling of API failures and timeouts

## Security Considerations

- **Input Validation**: Comprehensive validation to prevent injection attacks
- **Error Messages**: Sanitized error responses to avoid information leakage
- **Authentication**: Relies on existing Gitea token-based authentication
- **Authorization**: Respects Gitea repository permissions and access controls