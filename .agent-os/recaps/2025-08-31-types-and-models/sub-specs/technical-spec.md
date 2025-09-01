# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-08-31-types-and-models/spec.md

## Technical Requirements

### Pull Request Types (types/pr.go)
- Define `PullRequest` struct with fields: ID, Number, Title, Body, State, Author, HeadBranch, BaseBranch, CreatedAt, UpdatedAt, ClosedAt, MergedAt, Draft, Labels, Assignees, Reviewers, URL, DiffURL
- Define `PRAuthor` struct with fields: Username, AvatarURL, URL
- Define `PRLabel` struct with fields: ID, Name, Color, Description
- Define `PRState` type as string with constants: Open, Closed, Merged
- Implement `Validate() error` method on PullRequest to check required fields
- Add JSON tags with omitempty for optional fields
- Implement `IsOpen()`, `IsClosed()`, `IsMerged()` helper methods

### Issue Types (types/issue.go)
- Define `Issue` struct with fields: ID, Number, Title, Body, State, Author, Labels, Assignees, Milestone, CreatedAt, UpdatedAt, ClosedAt, CommentCount, URL
- Define `IssueAuthor` struct (can reuse PRAuthor if identical)
- Define `IssueState` type as string with constants: Open, Closed
- Define `Milestone` struct with fields: ID, Title, Description, DueDate, State
- Implement `Validate() error` method on Issue
- Add JSON tags following camelCase convention
- Implement `HasLabel(name string) bool` helper method

### Response Types (types/responses.go)
- Define `SuccessResponse` struct with fields: Success (bool), Data (interface{}), Metadata (ResponseMetadata)
- Define `ErrorResponse` struct with fields: Success (bool), Error (ErrorDetails)
- Define `ErrorDetails` struct with fields: Code (string), Message (string), Details (map[string]interface{})
- Define `ResponseMetadata` struct with fields: RequestID, Timestamp, Version
- Define `PaginatedResponse` struct extending SuccessResponse with pagination info
- Define `Pagination` struct with fields: Page, PerPage, Total, HasNext, HasPrev
- Implement standard error codes as constants (ValidationError, NotFound, Unauthorized, etc.)

### Common Types (types/common.go)
- Define `Repository` struct with fields: Owner, Name, FullName, Private, Fork, URL
- Define `User` struct with fields: ID, Username, Email, AvatarURL
- Define `Timestamp` type wrapping time.Time with custom JSON marshaling
- Implement `FilterOptions` struct for query parameters
- Define `SortOrder` type with constants: Ascending, Descending

### Validation Requirements
- All Validate() methods should check:
  - Required fields are not empty
  - URLs are valid format
  - Timestamps are reasonable (not in future, not zero)
  - States are valid enum values
- Return descriptive error messages indicating which field failed validation
- Use custom validation error type for consistent error handling

### JSON Serialization Requirements
- Use camelCase for all JSON field names
- Omit empty/nil fields with omitempty tag
- Format timestamps as RFC3339 strings
- Ensure enums serialize as strings not integers
- Handle null vs empty arrays correctly

### Integration Points
- Types must be compatible with Gitea SDK types for transformation
- Response types must match MCP protocol expectations
- Validation should align with Forgejo's business rules
- Consider versioning strategy for future changes

## Performance Criteria

- Type creation and validation should be negligible (<1ms)
- JSON marshaling/unmarshaling should handle large responses (1000+ items)
- No memory leaks from circular references
- Efficient string operations in validation methods