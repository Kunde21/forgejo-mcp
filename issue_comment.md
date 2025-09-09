# Plan: Add Issue Comment Tool

## Current Architecture Analysis
The forgejo-mcp project uses:
- **Official MCP SDK** (`github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`)
- **Gitea SDK** (`code.gitea.io/sdk/gitea`) for API integration
- **Clean architecture** with interfaces, services, and handlers
- **Existing tools**: `hello` (demo) and `list_issues` (functional)

## Implementation Plan

### 1. **Extend Gitea Interface** (`remote/gitea/interface.go`)
- Add `IssueCommenter` interface with `CreateIssueComment` method
- Define `IssueComment` struct for comment data

### 2. **Update Gitea Client** (`remote/gitea/gitea_client.go`)
- Implement `CreateIssueComment` method using Gitea SDK
- Handle repository parsing and API calls
- Convert between Gitea SDK and internal types

### 3. **Extend Service Layer** (`remote/gitea/service.go`)
- Add comment creation business logic
- Validate repository format and issue number
- Add comment content validation

### 4. **Add Tool Handler** (`server/handlers.go`)
- Create `handleCreateIssueComment` handler function
- Validate input parameters (repository, issue number, comment)
- Return appropriate MCP response format

### 5. **Register New Tool** (`server/server.go`)
- Add `create_issue_comment` tool registration
- Include proper description and schema

### 6. **Update Tests**
- Add unit tests for new service methods
- Add integration tests for the new tool
- Update test harness to support comment operations

## Tool Specification

**Tool Name**: `create_issue_comment`

**Parameters**:
- `repository` (string, required): Repository in "owner/repo" format
- `issue_number` (integer, required): Issue number to comment on
- `comment` (string, required): Comment content

**Response Format**:
```json
{
  "content": [{"type": "text", "text": "Comment added successfully"}],
  "structured": {
    "comment_id": 123,
    "issue_number": 42,
    "repository": "owner/repo"
  }
}
```

## Implementation Steps

1. **Phase 1**: Core functionality (interface, client, service)
2. **Phase 2**: MCP integration (handler, tool registration)
3. **Phase 3**: Testing and validation
4. **Phase 4**: Documentation updates

## Key Considerations

- **Error Handling**: Follow existing patterns with proper error wrapping
- **Validation**: Use existing ozzo-validation patterns
- **Testing**: Leverage existing test harness and mock server
- **Documentation**: Update README with tool usage examples
- **Consistency**: Follow existing code style and patterns

This plan maintains the project's clean architecture while adding the requested functionality in a maintainable and testable way.