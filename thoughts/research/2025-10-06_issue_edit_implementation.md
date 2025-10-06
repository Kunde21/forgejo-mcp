---
date: 2025-10-06T14:30:00Z
git_commit: cd44ba0cef44d91a5a0a88df34b029d60974a7dc
branch: mcp-go
repository: forgejo-mcp
topic: "Issue Editing Tool Implementation Research"
tags: [research, codebase, issue-edit, mcp, forgejo, gitea]
last_updated: 2025-10-06T14:30:00Z
---

## Ticket Synopsis
The ticket requests implementation of a new MCP tool `forgejo_issue_edit` to enable editing existing issues on Forgejo/Gitea repositories. The tool should support updating issue titles, bodies, labels, assignees, and state through the MCP interface, with partial update support and proper validation.

## Summary
The codebase has a well-established pattern for implementing MCP tools, with pull request editing serving as the perfect template for issue editing. The implementation requires extending the Issue struct, adding new interfaces, and following the established validation and error handling patterns. The Forgejo SDK provides full support for issue editing with all required fields.

## Detailed Findings

### MCP Tool Architecture
The server/ directory contains all MCP tool implementations following a consistent pattern:
- Tool registration in `server/server.go` using `mcp.AddTool()`
- Handler functions with signature: `handleToolName(ctx, request, args) (*mcp.CallToolResult, *ResultType, error)`
- Validation using ozzo-validation with conditional rules
- Repository resolution via `RepositoryResolver` for directory parameters
- Consistent error handling with `TextErrorf()` and `TextResult()`

### Current Issue Implementation Limitations
The current Issue struct in `remote/interface.go:8-12` only contains:
```go
type Issue struct {
    Number int    `json:"number"`
    Title  string `json:"title"`
    State  string `json:"state"`
}
```

Missing fields needed for editing:
- `Body string` - Issue description/content
- `Labels []Label` - Issue labels
- `Assignees []User` - Assigned users
- `Updated string` - Last update timestamp
- `Created string` - Creation timestamp

### Pull Request Edit Pattern (Template)
The PR edit implementation in `server/pr_edit.go:14-140` provides the exact pattern to follow:
- Optional fields with `omitzero` JSON tags
- At least one field must be provided validation
- State validation with enum values ("open", "closed")
- Repository/directory mutual exclusivity with conditional validation
- Partial update support at the client level

### Remote Client Implementation Pattern
The Forgejo client in `remote/forgejo/pull_requests.go:310-416` shows how to implement partial updates:
```go
var editOptions forgejo.EditPullRequestOption
hasChanges := false

if args.Title != "" {
    editOptions.Title = args.Title
    hasChanges = true
}

if !hasChanges {
    return nil, fmt.Errorf("no changes specified")
}
```

### Forgejo SDK Support
The Forgejo SDK's `EditIssueOption` supports all required fields:
- `Title string` - Required field
- `Body *string` - Pointer for optional updates
- `Assignees []string` - Array of usernames
- `Milestone *int64` - Milestone ID
- `State *StateType` - "open" or "closed"
- `Deadline *time.Time` - Due date

### Validation Requirements
Based on existing patterns, issue editing needs:
- Title: 1-255 characters when provided
- Body: 1-65535 characters when provided
- State: Must be "open" or "closed"
- Issue number: Positive integer
- Repository: "owner/repo" format or directory resolution

## Code References

### Core Files to Modify
- `remote/interface.go:8-12` - Extend Issue struct and add EditIssueArgs/IssueEditor
- `server/server.go` - Register new tool
- `server/issues.go` - Add handleIssueEdit function
- `remote/forgejo/issues.go` - Add EditIssue method
- `remote/gitea/gitea_client.go` - Add EditIssue method

### Reference Implementations
- `server/pr_edit.go:14-140` - Complete edit handler pattern
- `remote/forgejo/pull_requests.go:310-416` - Partial update client pattern
- `server/issues.go:95-135` - Issue validation patterns
- `server_test/pr_edit_test.go` - Testing patterns

## Architecture Insights

### Design Patterns
1. **Interface-First Design**: Always extend interfaces before implementing clients
2. **Partial Update Pattern**: Use `hasChanges` flag to avoid empty API calls
3. **Conditional Validation**: Use `v.When()` for repository/directory validation
4. **Repository Resolution**: Automatic detection from directory path
5. **Error Wrapping**: Provide context for all errors

### Data Flow
1. MCP request → Server handler
2. Validation (ozzo-validation)
3. Repository resolution (if directory provided)
4. Remote client call
5. Forgejo/Gitea SDK call
6. Response conversion
7. Return structured result

### Testing Strategy
- Table-driven tests with mock servers
- Test all validation scenarios
- Test partial update combinations
- Test error conditions
- Integration tests with real repositories

## Historical Context (from thoughts/)

From `thoughts/research/2025-10-05_issue_creation_implementation.md`:
- Established MCP tool implementation patterns
- Reference to CRUD operations in `server/issue_comments.go`
- PR edit implementation as template for complex validation

From `thoughts/plans/issue-creation-tool-implementation.md`:
- Handler pattern with consistent signature
- Validation with ozzo-validation
- Repository resolution via `RepositoryResolver`
- Interface-based client architecture

From `thoughts/reviews/feature_issue_create-review.md`:
- Validation of following established patterns
- Importance of consistency across tools

## Related Research
- `thoughts/research/2025-10-05_issue_creation_implementation.md` - Issue creation patterns
- `pr_edit_tool.md` - Pull request edit implementation plan
- `thoughts/tickets/feature_issue_create.md` - Related issue creation ticket

## Open Questions
1. Should labels/assignees validation check existence against repository?
2. How to handle concurrent edits (optimistic locking)?
3. Whether to support milestone editing (requires additional API calls)?
4. Should deadline editing be included in initial implementation?

## Implementation Recommendation
Follow the PR edit pattern exactly:
1. Add `EditIssueArgs` with optional fields using `omitzero`
2. Implement `IssueEditor` interface in both clients
3. Use partial update pattern with `hasChanges` detection
4. Apply same validation rules as issue creation
5. Return updated Issue object with all fields populated

The infrastructure is complete - this is a straightforward implementation following established patterns.