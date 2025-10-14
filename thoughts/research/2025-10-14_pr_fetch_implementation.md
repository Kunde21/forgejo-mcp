---
date: 2025-10-14T07:50:15Z+0000
git_commit: 415b59ffb4ed32b596768ffdc2e010ca1b36fd3e
branch: master
repository: forgejo-mcp
topic: "PR Fetch Tool Implementation Research"
tags: [research, codebase, pr, fetch, remote, interface]
last_updated: 2025-10-14
---

## Ticket Synopsis

The ticket requests implementation of a new MCP tool to fetch comprehensive information about a single pull request by number. The tool should return detailed metadata including title, description, status, author, reviewers, approvals, build status, labels, timestamps, branch information, commit details, conflict status, and size metrics. It must support all PR states (open, closed, merged, draft) and handle both same-repository and fork-based PRs.

## Summary

The codebase has a well-established pattern for PR tools but lacks a `GetPullRequest` method in the remote client interface. Both Forgejo and Gitea SDKs support single PR retrieval, but this functionality isn't exposed. The implementation will require: (1) extending the remote interface with a `GetPullRequest` method, (2) implementing it in both clients, (3) creating a server handler following established patterns, and (4) potentially extending the `PullRequest` struct to include additional metadata fields not currently exposed.

## Detailed Findings

### Remote Client Interface

**Current State**: The remote interface (`remote/interface.go:123-135`) defines a minimal `PullRequest` struct with only basic fields:
- ID, Number, Title, Body, State, User
- CreatedAt, UpdatedAt timestamps
- Head and Base branch references

**Missing Method**: No `GetPullRequest` method exists in the interface, despite both SDKs supporting it:
- Forgejo SDK: `GetPullRequest(owner, repo string, index int64) (*PullRequest, *Response, error)`
- Gitea SDK: `GetPullRequest(owner, repo string, index int64) (*PullRequest, *Response, error)`

**Available Fields Not Exposed**: The SDKs provide rich metadata not available in the current interface:
- Reviewers and approval status
- Mergeable status and conflict information
- Labels, milestones, assignees
- Comment count, merge information
- File change statistics (additions, deletions)
- Draft status indicators

### Existing PR Tool Patterns

**Tool Registration**: Tools are registered in `server/server.go:134-197` using `mcp.AddTool()` with consistent naming (`pr_{action}`) and descriptive handlers (`handlePullRequest{Action}`).

**Handler Structure**: All handlers follow the signature pattern (`server/pr_list.go:45`):
```go
func (s *Server) handlePullRequest{Action}(ctx context.Context, request *mcp.CallToolRequest, args {Tool}Args) (*mcp.CallToolResult, *{Tool}Result, error)
```

**Validation Pattern**: Uses ozzo-validation with conditional rules (`server/pr_list.go:55-81`):
- Repository format validation with regex (`owner/repo`)
- Directory validation with absolute path requirement
- Custom validators for git operations
- Consistent error formatting with `TextErrorf()`

**Repository Resolution**: Directory parameter takes precedence over repository (`server/pr_list.go:83-91`):
- Uses `RepositoryResolver` for git repository detection
- Supports fork detection for PR creation
- Error enhancement for better user experience

### Implementation Requirements

**Interface Extension**: Need to add to `remote/interface.go`:
```go
type PullRequestGetter interface {
    GetPullRequest(ctx context.Context, repo string, number int) (*PullRequest, error)
}
```

**Client Implementation**: Both Forgejo (`remote/forgejo/pull_requests.go`) and Gitea (`remote/gitea/gitea_client.go`) clients need new methods following existing patterns:
- Parse repository string
- Validate input parameters
- Call SDK's `GetPullRequest` method
- Convert response to interface format
- Handle errors consistently

**Server Handler**: New file `server/pr_fetch.go` with:
- Argument struct with repository/directory and PR number
- Result struct wrapping the PullRequest
- Handler following established signature pattern
- Validation using ozzo-validation
- Repository resolution using existing patterns

### Data Structure Considerations

**Current Limitations**: The existing `PullRequest` struct (`remote/interface.go:123-135`) lacks many required fields:
- No reviewer/approval information
- No mergeable or conflict status
- No file change statistics
- No label or milestone data
- Limited user information (just username string)

**Potential Solutions**:
1. Extend existing `PullRequest` struct (breaking change)
2. Create new `PullRequestDetails` struct for fetch tool
3. Use optional fields with backward compatibility

### Testing Patterns

**Test Structure**: Follow existing pattern in `server_test/pr_list_test.go`:
- Structured test cases with comprehensive fields
- Mock server setup for isolated testing
- Temporary directory creation for directory parameter tests
- Result comparison using go-cmp

**Test Categories**: Include validation, integration, error handling, and edge cases similar to existing PR tool tests.

## Code References

- `remote/interface.go:123-135` - PullRequest struct definition
- `remote/interface.go:144-147` - PullRequestLister interface
- `remote/forgejo/pull_requests.go:13-113` - Forgejo PR operations
- `remote/gitea/gitea_client.go:190-248` - Gitea PR operations
- `server/pr_list.go:45-117` - PR list handler implementation
- `server/server.go:134-197` - Tool registration pattern
- `server_test/pr_list_test.go:43-931` - Comprehensive test structure

## Architecture Insights

**Layered Architecture**: Clear separation between server tools, remote interface, and client implementations. The interface abstraction allows both Forgejo and Gitea to be treated identically at the server level.

**Consistent Patterns**: All PR tools follow identical patterns for validation, repository resolution, error handling, and response formatting. This consistency makes adding new tools straightforward.

**Interface Segregation**: Different PR operations are split across separate interfaces (PullRequestLister, PullRequestCreator, etc.), allowing clients to implement only what they need.

**SDK Abstraction**: Both clients convert SDK-specific responses to a standardized format, hiding differences between Forgejo and Gitea implementations.

## Historical Context (from thoughts/)

From `thoughts/research/2025-10-06_pr_creation_implementation.md`:
- PR tools use a 3-phase implementation approach for iterative development
- Git utilities use `os/exec` instead of external dependencies
- Error enhancement pattern wraps API errors with user-friendly guidance
- Repository resolution prioritizes directory over repository parameter

From `thoughts/tickets/feature_pr_create.md`:
- Tool registration follows `mcp.AddTool()` pattern
- Validation uses ozzo-validation with conditional rules
- Fork detection is handled automatically for complex scenarios

## Related Research

- `thoughts/research/2025-10-06_pr_creation_implementation.md` - PR creation implementation patterns
- `thoughts/research/2025-10-05_issue_creation_implementation.md` - Issue creation patterns applicable to PRs
- `thoughts/research/2025-10-06_issue_edit_implementation.md` - Edit patterns relevant for fetch implementation

## Open Questions

1. **Data Structure Scope**: Should the fetch tool return the minimal `PullRequest` struct or an extended version with additional metadata?
2. **Interface Evolution**: Is it acceptable to extend the existing `PullRequest` struct, or should a new `PullRequestDetails` type be created?
3. **Field Priority**: Which of the available SDK fields are essential for the initial implementation vs. future enhancements?
4. **Backward Compatibility**: How will extending the interface affect existing PR tools?

## Implementation Status: ✅ COMPLETED

The PR fetch tool has been successfully implemented following the recommended phased approach:

### ✅ Phase 1: Interface Extension and Client Implementation
- **Interface Extension**: Added `PullRequestGetter` interface and `PullRequestDetails` struct to `remote/interface.go`
- **Client Implementation**: Implemented `GetPullRequest` method in both Forgejo and Gitea clients
- **Comprehensive Data**: Created detailed `PullRequestDetails` struct with all available metadata fields

### ✅ Phase 2: Server Handler Implementation  
- **Handler Creation**: Implemented `server/pr_fetch.go` following established patterns
- **Tool Registration**: Registered `pr_fetch` tool in `server/server.go`
- **Validation**: Added comprehensive parameter validation using ozzo-validation
- **Repository Resolution**: Integrated with existing repository resolution patterns

### ✅ Phase 3: Testing and Integration
- **Mock Server**: Extended mock server in `server_test/harness.go` with PR fetch endpoint
- **Comprehensive Tests**: Created `server_test/pr_fetch_test.go` with full test coverage
- **Tool Discovery**: Updated tool discovery tests to include the new tool
- **Integration**: All tests passing, build successful, static analysis clean

### Key Features Implemented
- **Comprehensive Metadata**: Returns all available PR information including labels, assignees, milestones, merge status, timestamps
- **Flexible Parameters**: Supports both repository and directory parameters with existing resolution patterns
- **Error Handling**: Robust validation and error handling following project conventions
- **Backward Compatibility**: Uses new `PullRequestDetails` struct to avoid breaking existing tools
- **Full Test Coverage**: Includes success cases, error cases, validation, and edge cases

### Files Modified/Created
1. `remote/interface.go` - Added interface and detailed struct
2. `remote/forgejo/pull_requests.go` - Added GetPullRequest implementation  
3. `remote/gitea/gitea_client.go` - Added GetPullRequest implementation
4. `server/pr_fetch.go` - Created new handler
5. `server/server.go` - Registered tool
6. `server_test/harness.go` - Extended mock server
7. `server_test/pr_fetch_test.go` - Created comprehensive tests
8. `server_test/tool_discovery_test.go` - Updated tool count and descriptions

The implementation successfully delivers a production-ready PR fetch tool that integrates seamlessly with the existing codebase architecture and follows all established patterns and conventions.