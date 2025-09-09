# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-09-issue-comment-tool/spec.md

> Created: 2025-09-09
> Version: 1.0.0

## Technical Requirements

### Functionality Details

The issue comment tool will implement a complete MCP tool for creating comments on Forgejo/Gitea repository issues. The implementation will follow the existing clean architecture patterns established in the forgejo-mcp project.

**Core Components:**

1. **Interface Extension** (`remote/gitea/interface.go`):
   - Add `IssueCommenter` interface with `CreateIssueComment` method
   - Define `IssueComment` struct for comment data representation
   - Maintain consistency with existing `IssueLister` interface pattern

2. **Client Implementation** (`remote/gitea/gitea_client.go`):
   - Implement `CreateIssueComment` method using Gitea SDK
   - Handle repository parsing (owner/repo format)
   - Convert between Gitea SDK comment types and internal `IssueComment` struct
   - Proper error handling with context preservation

3. **Service Layer** (`remote/gitea/service.go`):
   - Extend `Service` struct to implement `IssueCommenter` interface
   - Add comment creation business logic
   - Implement validation for repository format, issue number, and comment content
   - Follow existing validation patterns using regex and parameter checks

4. **MCP Handler** (`server/handlers.go`):
   - Create `handleCreateIssueComment` handler function
   - Implement input parameter validation using ozzo-validation
   - Return structured response with comment metadata
   - Follow existing error handling patterns with `TextErrorf` and `TextResultf`

5. **Tool Registration** (`server/server.go`):
   - Register `create_issue_comment` tool with proper schema
   - Include comprehensive tool description
   - Maintain consistency with existing tool registration pattern

### UI/UX Specifications

**Tool Interface:**
- Tool name: `create_issue_comment`
- Input parameters follow JSON schema validation
- Response includes both human-readable text and structured data
- Error messages provide clear, actionable feedback

**Parameter Schema:**
```json
{
  "repository": {
    "type": "string",
    "description": "Repository in 'owner/repo' format",
    "required": true
  },
  "issue_number": {
    "type": "integer",
    "description": "Issue number to comment on",
    "required": true,
    "minimum": 1
  },
  "comment": {
    "type": "string",
    "description": "Comment content",
    "required": true,
    "minLength": 1
  }
}
```

**Response Format:**
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

### Integration Requirements

**MCP SDK Integration:**
- Use official MCP SDK v0.4.0 handler signature pattern
- Implement `func (s *Server) handleCreateIssueComment(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error)`
- Follow existing migration patterns from previous SDK versions

**Gitea SDK Integration:**
- Leverage existing `code.gitea.io/sdk/gitea v0.22.0` dependency
- Use `CreateIssueComment` API method from Gitea SDK
- Handle repository owner/repo parsing consistently with existing `ListIssues` implementation

**Service Architecture:**
- Maintain dependency injection pattern through interfaces
- Extend existing service composition without breaking changes
- Preserve clean separation between client, service, and handler layers

### Performance Criteria

**Response Time:**
- Target < 2 seconds for comment creation under normal conditions
- Implement proper timeout handling at the MCP handler level
- Use context cancellation for request lifecycle management

**Error Handling:**
- Validate all input parameters before API calls
- Provide specific error messages for different failure scenarios
- Follow existing error wrapping patterns with `fmt.Errorf`

**Resource Usage:**
- Minimal memory footprint for comment operations
- Efficient string handling for comment content
- Proper connection management with Gitea SDK client

## External Dependencies

### Analysis of New Dependencies Required

Based on the issue_comment.md implementation plan and existing project dependencies, **no new external dependencies are required**. The implementation will leverage the existing dependency stack:

**Existing Dependencies Utilization:**

1. **Gitea SDK** (`code.gitea.io/sdk/gitea v0.22.0`):
   - Already provides `CreateIssueComment` functionality
   - No additional version upgrades needed
   - Existing authentication and connection patterns sufficient

2. **MCP SDK** (`github.com/modelcontextprotocol/go-sdk v0.4.0`):
   - Current version supports all required tool registration patterns
   - Handler signature patterns already established
   - Response formatting capabilities sufficient

3. **Validation Library** (`github.com/go-ozzo/ozzo-validation/v4 v4.3.0`):
   - Existing validation patterns cover all parameter validation needs
   - Struct validation capabilities sufficient for new tool parameters
   - No additional validation dependencies required

4. **Standard Library Dependencies:**
   - `context` package for request lifecycle management
   - `fmt` package for error formatting and response construction
   - `regexp` package for repository format validation
   - `strings` package for repository parsing

### Dependency Justification

The existing dependency stack provides complete coverage for the issue comment functionality:

- **Gitea SDK**: Already includes comment creation APIs and authentication mechanisms
- **MCP SDK**: Supports tool registration, parameter validation, and response formatting
- **Validation Library**: Handles all input validation requirements with existing patterns
- **Standard Library**: Provides all necessary utilities for string manipulation and error handling

### Risk Assessment

**Low Risk Factors:**
- All dependencies are already tested and integrated in the project
- No breaking changes expected in existing dependency versions
- Implementation follows established patterns from existing tools

**Mitigation Strategies:**
- Follow existing code patterns exactly to minimize integration issues
- Implement comprehensive testing to validate dependency interactions
- Maintain backward compatibility with existing service interfaces

### Conclusion

The issue comment tool implementation can be completed using the existing project dependencies without introducing new external packages. This approach maintains project stability, reduces complexity, and follows the established architectural patterns while delivering the required functionality.