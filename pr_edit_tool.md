# Plan: Add Pull Request Edit Tool

## Overview
This plan outlines the addition of a `pr_edit` tool to the forgejo-mcp server, enabling users to edit pull request metadata (title, body, state, base branch) in Forgejo/Gitea repositories.

## Current State
The forgejo-mcp server currently provides these PR-related tools:
- `pr_list` - List pull requests
- `pr_comment_list` - List PR comments  
- `pr_comment_create` - Create PR comments
- `pr_comment_edit` - Edit PR comments

**Missing**: Tool to edit the pull request itself (title, body, state, base branch).

## Implementation Plan

### 1. Remote Interface Extension (`remote/interface.go`)

**Add new types:**
```go
// EditPullRequestArgs represents the arguments for editing a pull request
type EditPullRequestArgs struct {
    Repository        string `json:"repository"`
    PullRequestNumber int    `json:"pull_request_number"`
    Title             string `json:"title,omitempty"`
    Body              string `json:"body,omitempty"`
    State             string `json:"state,omitempty"` // "open", "closed"
    BaseBranch        string `json:"base_branch,omitempty"`
}

// PullRequestEditor defines the interface for editing pull requests
type PullRequestEditor interface {
    EditPullRequest(ctx context.Context, args EditPullRequestArgs) (*PullRequest, error)
}
```

**Update ClientInterface:**
Add `PullRequestEditor` to the combined interface.

### 2. Forgejo Client Implementation (`remote/forgejo/pull_requests.go`)

**Implement `EditPullRequest` method:**
- Parse repository string ("owner/repo")
- Validate pull request number
- Use Forgejo SDK's `EditPullRequest` method
- Convert between Forgejo SDK types and our `PullRequest` struct
- Handle validation and error cases
- Return updated pull request metadata

**Validation rules:**
- Repository format validation
- Pull request number must be positive
- At least one editable field must be provided
- State validation (open/closed only)

### 3. Gitea Client Implementation (`remote/gitea/gitea_client.go`)

**Implement `EditPullRequest` method:**
- Mirror Forgejo implementation patterns
- Use Gitea SDK's `EditPullRequest` functionality
- Ensure consistent behavior across both client types

### 4. Server Handler (`server/pr_edit.go` - new file)

**Create handler function:**
```go
func (s *Server) handlePullRequestEdit(ctx context.Context, request *mcp.CallToolRequest, args PullRequestEditArgs) (*mcp.CallToolResult, *PullRequestEditResult, error)
```

**Handler components:**
- `PullRequestEditArgs` struct with validation tags
- `PullRequestEditResult` struct for response data
- Input validation using ozzo-validation
- Repository resolution (directory parameter support)
- Service layer call
- Success/error response formatting

**Tool parameters:**
- `repository` (string, optional): "owner/repo" format
- `directory` (string, optional): Local directory for auto-resolution  
- `pull_request_number` (int, required): PR number to edit
- `title` (string, optional): New PR title
- `body` (string, optional): New PR description
- `state` (string, optional): New state ("open", "closed")
- `base_branch` (string, optional): New base branch

**Validation rules:**
- At least one of repository or directory must be provided
- Pull request number must be positive
- At least one of title, body, state, or base_branch must be provided
- State must be "open" or "closed" if provided
- Directory must be absolute path and exist

### 5. Server Registration (`server/server.go`)

**Add tool registration:**
```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_edit",
    Description: "Edit a pull request in a Forgejo/Gitea repository",
}, s.handlePullRequestEdit)
```

### 6. Testing Implementation

#### Server Tests (`server_test/pr_edit_test.go` - new file)
- Table-driven tests for all scenarios
- Success cases: edit title, body, state, base branch
- Error cases: validation errors, permission errors, not found
- Concurrent request handling tests
- Directory parameter resolution tests
- Integration tests with mock servers

#### Forgejo Client Tests (`remote/forgejo/forgejo_client_test.go`)
- Unit tests for `EditPullRequest` method
- Validation tests for all input parameters
- Error handling tests (nil client, invalid inputs)
- Mock server integration tests

#### Gitea Client Tests (`remote/gitea/client_test.go`)
- Mirror Forgejo client test structure
- Ensure consistent behavior across implementations

### 7. Documentation

**Godoc comments for all exported functions:**
- Comprehensive parameter documentation
- Return value documentation
- Migration notes following existing patterns
- Usage examples in comments

**Tool description:**
- Clear description of functionality
- Parameter explanations
- Usage notes and limitations

### 8. Implementation Details

**Code Style Compliance:**
- Follow Go style guide from `.agent-os/standards/code-style/go-style.md`
- Use existing patterns from PR comment editing
- Maintain consistent error handling
- Use ozzo-validation for input validation

**Error Handling:**
- Structured error responses with `TextErrorf`
- Proper error wrapping with `%w` verb
- Context validation first
- Immediate error returns

**Response Format:**
- Success: Updated PR metadata with confirmation message
- Error: Structured error with descriptive message
- Consistent with existing tool response patterns

**Concurrency:**
- Thread-safe implementation
- Proper context handling
- Resource cleanup

### 9. Dependencies

**No new dependencies required:**
- Uses existing Forgejo/Gitea SDKs
- Leverages existing validation library
- Follows current MCP SDK patterns

### 10. Migration Considerations

**Backward compatibility:**
- No breaking changes to existing tools
- Follows established patterns
- Maintains consistent API design

**Future extensibility:**
- Interface design allows for additional PR editing features
- Consistent with existing tool architecture
- Supports future Forgejo/Gitea API changes

## Success Criteria

1. ✅ Tool successfully edits pull requests in both Forgejo and Gitea
2. ✅ All validation scenarios handled correctly
3. ✅ Comprehensive test coverage (unit + integration)
4. ✅ Consistent with existing codebase patterns
5. ✅ Proper error handling and response formatting
6. ✅ Documentation follows project standards
7. ✅ No breaking changes to existing functionality

## Implementation Order

1. Remote interface extension
2. Forgejo client implementation
3. Gitea client implementation  
4. Server handler implementation
5. Tool registration
6. Testing implementation
7. Documentation updates

This plan ensures a systematic approach to adding the PR edit functionality while maintaining consistency with the existing codebase architecture and following established patterns.