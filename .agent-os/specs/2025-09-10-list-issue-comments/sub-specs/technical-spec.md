# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-10-list-issue-comments/spec.md

## Technical Requirements

### Interface Extension (remote/gitea/interface.go)
- Add IssueCommentLister interface with ListIssueComments method signature
- Define IssueCommentList struct for collection of comments with pagination metadata
- Update GiteaClientInterface to include the new IssueCommentLister interface
- Maintain consistency with existing interface patterns and naming conventions

### Client Implementation (remote/gitea/gitea_client.go)
- Implement ListIssueComments method using Gitea SDK's ListIssueComments API
- Handle repository parsing from owner/repo format string
- Convert between Gitea SDK comment types and internal IssueComment struct
- Add pagination support with limit and offset parameters
- Implement proper error handling for API failures and invalid inputs

### Service Layer (remote/gitea/service.go)
- Extend Service struct to implement IssueCommentLister interface
- Add ListIssueComments method that assumes validated input parameters
- Focus on business logic and API communication, not input validation
- Follow existing service patterns and error handling conventions

### MCP Handler (server/handlers.go)
- Create handleListIssueComments handler function
- Define ListIssueCommentsArgs struct for input parameters with validation tags
- Implement comprehensive input parameter validation using ozzo-validation
- Validate repository format (owner/repo pattern), issue number (positive integer), and pagination parameters (limit 1-100, offset >= 0)
- Parse and validate all required and optional parameters before calling service layer
- Return structured response with comment list and metadata
- Follow existing error handling patterns with meaningful error messages
- Implement proper response formatting for MCP protocol

### Tool Registration (server/server.go)
- Register list_issue_comments tool with proper JSON schema
- Add comprehensive tool description and parameter documentation
- Include parameter types, constraints, and default values
- Follow existing tool registration patterns

### Data Structures
- Define IssueComment struct with fields: ID, Content, Author, CreatedAt
- Define IssueCommentList struct with Comments array and pagination metadata
- Define ListIssueCommentsArgs struct for handler input parameters with validation tags
- Ensure proper JSON serialization for MCP responses

## External Dependencies

No new external dependencies required. The implementation will leverage:
- Existing Gitea SDK for API communication
- Current MCP SDK for tool registration and handling
- Existing validation libraries (ozzo-validation)
- Standard Go libraries for data structures and HTTP handling

## Performance Considerations

- Implement pagination to prevent large response payloads
- Add reasonable default limits (15 comments) to balance usability and performance
- Consider caching strategies for frequently accessed issues
- Implement proper timeout handling for API calls
- Add rate limiting awareness to prevent API abuse

## Error Handling

- Validate all input parameters before API calls
- Handle repository not found errors gracefully
- Handle issue not found errors with clear messages
- Handle authentication/authorization errors appropriately
- Provide meaningful error messages for invalid pagination parameters
- Follow existing error response patterns in the codebase

## Testing Requirements

- Unit tests for all new methods in client, service, and handler layers
- Integration tests for complete MCP tool workflow
- Test edge cases: empty comments, single comment, maximum pagination
- Test error scenarios: invalid repository, non-existent issue, invalid parameters
- Update test harness to support comment listing operations
- Mock Gitea API responses for consistent testing

## Integration Points

- Extend existing comment functionality architecture
- Follow patterns established by list_issues and create_issue_comment tools
- Maintain clean interface-based design for testability
- Preserve backward compatibility with existing functionality
- Integrate with existing authentication and configuration systems