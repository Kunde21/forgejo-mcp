---
date: 2025-10-16T21:18:11+07:00
git_commit: e6aa4335059181825151178d739c13818945a19c
branch: master
repository: forgejo-mcp
topic: "Notification List Tool Implementation Research"
tags: [research, codebase, notifications, mcp-tool, implementation]
last_updated: 2025-10-16
---

## Ticket Synopsis

The ticket `thoughts/tickets/feature_notification_list_tool.md` proposes adding a `notification_list` MCP tool to enable users to read their notifications from Gitea/Forgejo remotes. This would be the first notification-related tool in the forgejo-mcp project, supporting filtering by repository, read/unread status, and pagination with offset/limit parameters.

## Summary

The notification list tool implementation is highly feasible with excellent SDK support and well-established patterns in the codebase. Both Gitea and Forgejo SDKs provide identical notification APIs with comprehensive filtering and pagination support. The existing architecture provides clear patterns for remote client interfaces, server tool registration, repository resolution, and testing infrastructure that can be directly applied to notification functionality.

## Detailed Findings

### Remote Client Architecture

**Interface Pattern** (`remote/interface.go:19-306`)
The codebase uses granular interfaces that compose into a complete client:
- Each feature has dedicated interfaces (e.g., `IssueLister`, `PullRequestLister`)
- `ClientInterface` combines all sub-interfaces for complete Git operations
- Both ForgejoClient and GiteaClient implement the full interface

**Implementation Pattern** (`remote/forgejo/forgejo_client.go:11-42`)
- Structured client with SDK client wrapper
- Consistent constructor patterns with error handling
- Repository parsing using `strings.Cut()` for "owner/repo" format
- Pagination conversion from offset/limit to 1-based page numbers

**Data Transformation Pattern** (`remote/forgejo/issues.go:49-66`)
- SDK types transformed to interface types with null safety
- Pre-allocated slices with known capacity
- Default values for missing data (e.g., "unknown" for missing usernames)
- Type conversions with proper error handling

### Server Tool Architecture

**Tool Registration** (`server/server.go:136-141`)
Tools are registered using `mcp.AddTool()` with:
- Auto-generated schemas from Go structs using generics
- Consistent naming convention (snake_case)
- Handler function references

**Argument Validation Pattern** (`server/issues.go:48-74`)
- Uses ozzo-validation with comprehensive validation rules
- Conditional validation for mutually exclusive parameters
- Regex validation for repository format
- Custom validation for directory existence checks

**Repository Resolution** (`server/repository_resolver.go:244-307`)
- Automatic directory-to-repository conversion
- Git remote extraction and parsing
- Fork detection for PR operations
- Enhanced error messages for resolution failures

**Pagination Handling** (`server/issues.go:43-46`, `server/pr_list.go:46-52`)
- Consistent defaults (limit: 15, max: 100)
- Offset-based pagination with validation
- Structured response with pagination metadata

### Notification API Support

**SDK Methods Available**
Both Gitea and Forgejo SDKs provide identical notification APIs:
- `ListNotifications(opt ListNotificationOptions)` - Core listing functionality
- `GetNotification(id int64)` - Individual notification retrieval
- `CheckNotifications()` - Total notification count

**Data Structures**
```go
type NotificationThread struct {
    ID         int64                `json:"id"`
    Repository *Repository          `json:"repository"`
    Subject    *NotificationSubject `json:"subject"`
    Unread     bool                 `json:"unread"`
    Pinned     bool                 `json:"pinned"`
    UpdatedAt  time.Time            `json:"updated_at"`
    URL        string               `json:"url"`
}
```

**Filtering Capabilities**
- Status filtering: `NotifyStatus` values (Read, Unread)
- Subject type filtering: `NotifySubjectType` values (Issue, Pull, Commit)
- Time-based filtering: `Since` and `Before` fields
- Repository filtering: Requires client-side filtering (not in SDK)

### Testing Infrastructure

**Mock Server** (`server_test/harness.go:184-225`)
- Comprehensive mock server with HTTP handlers
- Mock data structures for issues, comments, and pull requests
- Test data factory methods for realistic test scenarios
- Support for all existing API endpoints

**Test Patterns** (`server_test/issue_list_test.go:59-412`)
- Table-driven test structure with parallel execution
- Mock server setup for isolated testing
- Validation of both text content and structured data
- Comprehensive coverage of validation, pagination, and error cases

## Code References

### Core Architecture Files
- `remote/interface.go:289-306` - ClientInterface composition
- `remote/forgejo/forgejo_client.go:11-42` - ForgejoClient implementation
- `server/server.go:125-229` - Tool registration and server initialization
- `server/repository_resolver.go:244-307` - Repository resolution logic

### Implementation Pattern Examples
- `server/issues.go:42-93` - Complete list tool handler implementation
- `remote/forgejo/issues.go:19-66` - Remote client method with pagination and transformation
- `server/common.go:22-28` - Error handling helpers
- `server_test/harness.go:184-225` - Mock server setup

### Validation and Testing
- `server/issues.go:48-74` - Argument validation with ozzo-validation
- `server_test/issue_list_test.go:59-412` - Comprehensive test coverage
- `server_test/helpers.go` - Test helper functions

## Architecture Insights

### Design Patterns
1. **Interface Segregation**: Granular interfaces for each feature area
2. **Composition over Inheritance**: ClientInterface combines multiple interfaces
3. **Null Safety**: Consistent handling of optional SDK fields
4. **Error Wrapping**: Context-rich error messages with original error preservation
5. **Pagination Abstraction**: Offset-based API over page-based SDKs

### Extension Points
1. **New Interfaces**: Add to `remote/interface.go` following existing patterns
2. **Client Implementation**: Create new files in `remote/forgejo/` and `remote/gitea/`
3. **Server Tools**: Add to `server/` with consistent handler patterns
4. **Testing**: Extend mock server and create comprehensive test suites

### Consistency Mechanisms
1. **Schema Generation**: Automatic JSON schema from Go structs
2. **Validation Library**: Consistent ozzo-validation patterns
3. **Response Formatting**: Standardized text and structured responses
4. **Error Handling**: Uniform error message formatting

## Historical Context (from thoughts/)

The only notification-related document found is the feature ticket itself (`thoughts/tickets/feature_notification_list_tool.md`), created on 2025-10-16. This represents the first exploration of notification functionality in the forgejo-mcp project, with no prior implementation attempts or architectural discussions.

The ticket demonstrates thoughtful planning with:
- Clear functional and non-functional requirements
- Comprehensive success criteria with automated and manual verification
- Identification of key patterns to investigate
- Specific implementation decisions made upfront

## Related Research

No existing research documents were found specifically for notifications. However, the following related research provides valuable context:

- `thoughts/research/2025-09-11-mock-handler-refactor/` - Mock server infrastructure patterns
- `thoughts/research/2025-09-19-directory-parameter-support/` - Directory parameter handling
- Various implementation plans in `thoughts/plans/` for similar list-type tools

## Open Questions

1. **Repository Filtering Implementation**: SDK doesn't support repository filtering natively - requires client-side filtering after API call
2. **Issue/PR Number Extraction**: Need to parse from notification subject URLs or make additional API calls
3. **Count Accuracy**: `CheckNotifications()` returns total count, but filtering affects actual count
4. **Time Zone Handling**: Proper formatting of timestamp fields for consistent display

## Implementation Recommendations

### Phase 1: Core Interface and Client Implementation
1. Add notification types and interfaces to `remote/interface.go`
2. Implement `ListNotifications` in both Forgejo and Gitea clients
3. Update `ClientInterface` composition to include `NotificationLister`

### Phase 2: Server Tool Implementation
1. Create `server/notifications.go` with handler following existing patterns
2. Add tool registration to `server/server.go`
3. Implement argument validation and repository resolution

### Phase 3: Testing and Documentation
1. Extend mock server with notification endpoints
2. Create comprehensive test suite following existing patterns
3. Update documentation and examples

The implementation complexity is low due to excellent SDK support and well-established patterns in the codebase. The main challenges are client-side filtering and data transformation, both of which have clear patterns to follow.