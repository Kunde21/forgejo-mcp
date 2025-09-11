# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-11-pr-comment-list/spec.md

## Technical Requirements

### Data Structures (in remote/gitea/interface.go)
- **PullRequestComment struct**: Similar to IssueComment but for PR comments with fields for ID, Body, User, CreatedAt, UpdatedAt
- **PullRequestCommentList struct**: Collection with pagination metadata including Comments array, TotalCount, Limit, Offset
- **ListPullRequestCommentsArgs struct**: Arguments with validation tags for Repository, PullRequestNumber, Limit, Offset
- **PullRequestCommentLister interface**: Method signature for listing PR comments with proper error handling

### Service Layer (in remote/gitea/service.go)
- **ListPullRequestComments() method**: Business logic and client delegation with comprehensive error handling
- **No validation**: Service layer assumes pre-validated inputs from server handler
- **Error handling**: Consistent error messages and formatting following existing patterns

### Client Implementation (in remote/gitea/gitea_client.go)
- **ListPullRequestComments() method**: Gitea SDK integration using gitea-sdk's ListPullRequestComments()
- **Repository parsing**: Parse owner/repo format and convert between Gitea SDK and internal structs
- **Pagination handling**: Convert between our pagination format and Gitea's ListOptions with proper bounds checking

### Server Handler (in new file server/pr_comments.go)
- **PullRequestCommentListArgs struct**: Handler arguments with ozzo-validation tags for all parameters
- **PullRequestCommentListResult struct**: Response data structure with success message and structured data
- **handlePullRequestCommentList() function**: MCP tool handler following established patterns with proper request/response handling
- **Complete validation**: All input validation performed here using ozzo-validation before calling service layer

### Server Registration (in server/server.go)
- **Tool registration**: Add tool registration in NewFromService() using mcp.AddTool()
- **Tool configuration**: Name "pr_comment_list", description, and schema definition
- **Integration**: Ensure proper integration with existing MCP server infrastructure

### Input Validation (Server Handler Only)
- **Repository format**: Regex validation ^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$ in server handler
- **Pull request number**: Positive integer validation with minimum value of 1 in server handler
- **Limit**: Range validation 1-100 with default value of 15 in server handler
- **Offset**: Non-negative integer validation with default value of 0 in server handler
- **No service layer validation**: All validation responsibility moved to server handler to eliminate duplication

### Response Format
- **Success response**: Structured result with comments array, pagination metadata, and success message
- **Error response**: Consistent error formatting using TextErrorf() with descriptive messages
- **Data structure**: JSON-compatible format with proper field naming and types

### Testing Requirements
- **Unit tests**: Test validation logic, happy path, and error cases following table-driven pattern
- **Integration tests**: Test with mock Gitea client for end-to-end functionality
- **Handler tests**: Test MCP server integration and response formatting
- **Edge cases**: Test boundary conditions, invalid inputs, and error scenarios

## External Dependencies

No new external dependencies are required. The implementation will use existing dependencies:
- **github.com/mark3labs/mcp-go**: For MCP server functionality
- **gitea-sdk**: For Gitea API integration (already in use)
- **go-ozzo/ozzo-validation**: For input validation (already in use)