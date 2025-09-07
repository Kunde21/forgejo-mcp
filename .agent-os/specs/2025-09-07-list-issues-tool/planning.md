## Implementation Plan: Add "list_issues" Tool

### Overview
This plan outlines the implementation of a new MCP tool called "list_issues" that connects to a Gitea/Forgejo instance to retrieve issues for a specified repository. The tool will support pagination and return structured issue data.

### Current State Analysis
- **Existing Structure**: Basic MCP server with single "hello" tool
- **Configuration**: Simple config with Host/Port, needs extension for remote API access
- **Testing**: Integration test harness using stdio client, needs enhancement for API mocking
- **Dependencies**: Uses mark3labs/mcp-go, needs Gitea SDK addition

### Detailed Implementation Steps

#### 1. Configuration Extension
**File**: `config.go`
- Add `RemoteURL` and `AuthToken` fields to Config struct
- Update `LoadConfig()` to read `FORGEJO_REMOTE_URL` and `FORGEJO_AUTH_TOKEN` environment variables
- Add validation to ensure required fields are present when API tools are used

#### 2. Dependency Management
**File**: `go.mod`
- Add Gitea SDK dependency: `code.gitea.io/sdk/gitea`
- Run `go mod tidy` to resolve dependencies

#### 3. Remote Package Creation
**Directory**: `remote/gitea/`
**Files**:
- `interface.go`: Define `IssueLister` interface for dependency injection
- `gitea_client.go`: Implement Gitea SDK client with authentication
- `service.go`: Business logic for listing issues with pagination

**Interface Design**:
```go
type IssueLister interface {
    ListIssues(ctx context.Context, owner, repo string, limit, offset int) ([]Issue, error)
}

type Issue struct {
    Number int    `json:"number"`
    Title  string `json:"title"`
    State  string `json:"state"` // open, closed
    Status string `json:"status"` // WIP, open, closed, merged (derived from PR state)
}
```

#### 4. Tool Implementation
**File**: `main.go`
- Add `list_issues` tool registration in `NewServer()`
- Implement `handleListIssues()` method with:
  - Input validation for repository format ("owner/repo")
  - Pagination parameter handling (limit: 1-100, offset: >=0)
  - Error handling for API failures
  - Response formatting as structured MCP content

**Tool Schema**:
```json
{
  "name": "list_issues",
  "description": "List issues for a Gitea/Forgejo repository",
  "inputSchema": {
    "type": "object",
    "properties": {
      "repository": {
        "type": "string",
        "description": "Repository in format 'owner/repository'"
      },
      "limit": {
        "type": "integer",
        "description": "Maximum number of issues to return (default: 15)",
        "minimum": 1,
        "maximum": 100
      },
      "offset": {
        "type": "integer",
        "description": "Pagination offset (default: 0)",
        "minimum": 0
      }
    },
    "required": ["repository"]
  }
}
```

#### 5. Test Harness Enhancement
**File**: `server_test/harness.go`
- Add `MockGiteaServer` struct to manage httptest.Server
- Implement mock API endpoints for issues listing
- Add configuration injection for test scenarios
- Support both real API calls and mock responses

**Mock Server Features**:
- Configurable issue data for testing
- Support for pagination parameters
- Authentication validation
- Error response simulation

#### 6. Acceptance Tests
**File**: `server_test/list_issues_test.go`
- Test successful issue listing with mock data
- Test pagination parameters (limit, offset)
- Test error handling (invalid repo, auth failure, network errors)
- Test input validation (malformed repository format)
- Test concurrent requests

**Test Scenarios**:
- Valid repository with issues
- Empty repository
- Invalid repository format
- Authentication failure
- Network timeout
- Pagination edge cases

#### 7. Documentation Update
**File**: `README.md`
- Add "list_issues" to the Tools List under Issue interactions
- Document required configuration (REMOTE_URL, AUTH_TOKEN)
- Add usage examples
- Update configuration section with new environment variables

### Technical Considerations

#### Error Handling Strategy
- Network errors: Return MCP error with descriptive message
- Authentication errors: Generic "authentication failed" message
- Invalid repository: "repository not found" message
- API rate limits: "rate limit exceeded" with retry suggestion

#### Security Considerations
- Never log authentication tokens
- Validate repository format to prevent injection
- Use HTTPS for all API calls
- Implement request timeouts

#### Performance Optimization
- Implement connection pooling in Gitea client
- Add caching for frequently accessed repositories (future enhancement)
- Support concurrent API calls where appropriate

#### Testing Strategy
- Unit tests for individual components
- Integration tests for MCP protocol
- Acceptance tests for end-to-end functionality
- Mock server for reliable CI/CD testing

### Migration Path
1. **Phase 1**: Implement core functionality with mock testing
2. **Phase 2**: Add real API integration and authentication
3. **Phase 3**: Add comprehensive error handling and edge cases
4. **Phase 4**: Performance optimization and monitoring

### Success Criteria
- Tool successfully lists issues from Gitea/Forgejo API
- Proper pagination support with configurable limits
- Comprehensive error handling for all failure scenarios
- Full test coverage including acceptance tests
- Documentation updated and accurate
- No regressions in existing functionality

This plan provides a complete roadmap for implementing the list_issues tool while maintaining code quality, testability, and following Go best practices.
