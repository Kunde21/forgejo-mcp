# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-09-10-list-issue-comments/spec.md

> Created: 2025-09-10
> Status: Completed

## Tasks

- [x] 1. Interface Extension - Add IssueCommentLister interface and data structures
  - [x] 1.1 Write tests for IssueCommentLister interface and data structures
  - [x] 1.2 Define IssueComment struct with ID, Content, Author, CreatedAt fields
  - [x] 1.3 Define IssueCommentList struct with Comments array and pagination metadata
  - [x] 1.4 Define ListIssueCommentsArgs struct for handler input parameters with validation tags
  - [x] 1.5 Add IssueCommentLister interface with ListIssueComments method signature
  - [x] 1.6 Update GiteaClientInterface to include the new IssueCommentLister interface
  - [x] 1.7 Verify all tests pass

- [x] 2. Client Implementation - Implement ListIssueComments method in Gitea client
  - [x] 2.1 Write tests for ListIssueComments method in Gitea client
  - [x] 2.2 Implement ListIssueComments method using Gitea SDK's ListIssueComments API
  - [x] 2.3 Handle repository parsing from owner/repo format string
  - [x] 2.4 Convert between Gitea SDK comment types and internal IssueComment struct
  - [x] 2.5 Add pagination support with limit and offset parameters
  - [x] 2.6 Implement proper error handling for API failures and invalid inputs
  - [x] 2.7 Verify all tests pass

- [x] 3. Service Layer Integration - Extend service to implement the interface
  - [x] 3.1 Write tests for service layer ListIssueComments method
  - [x] 3.2 Extend Service struct to implement IssueCommentLister interface
  - [x] 3.3 Add ListIssueComments method that assumes validated input parameters
  - [x] 3.4 Focus on business logic and API communication, not input validation
  - [x] 3.5 Follow existing service patterns and error handling conventions
  - [x] 3.6 Verify all tests pass

- [x] 4. MCP Handler Implementation - Create handleListIssueComments function with validation
  - [x] 4.1 Write tests for handleListIssueComments handler function
  - [x] 4.2 Create handleListIssueComments handler function
  - [x] 4.3 Implement comprehensive input parameter validation using ozzo-validation
  - [x] 4.4 Validate repository format (owner/repo pattern), issue number (positive integer), and pagination parameters (limit 1-100, offset >= 0)
  - [x] 4.5 Parse and validate all required and optional parameters before calling service layer
  - [x] 4.6 Return structured response with comment list and metadata
  - [x] 4.7 Follow existing error handling patterns with meaningful error messages
  - [x] 4.8 Implement proper response formatting for MCP protocol
  - [x] 4.9 Verify all tests pass

- [x] 5. Tool Registration - Register the list_issue_comments tool with MCP server
  - [x] 5.1 Write tests for tool registration
  - [x] 5.2 Register list_issue_comments tool with proper JSON schema
  - [x] 5.3 Add comprehensive tool description and parameter documentation
  - [x] 5.4 Include parameter types, constraints, and default values
  - [x] 5.5 Follow existing tool registration patterns
  - [x] 5.6 Verify all tests pass

- [x] 6. Testing and Documentation - Add comprehensive tests and update documentation
  - [x] 6.1 Write integration tests for complete MCP tool workflow
  - [x] 6.2 Test edge cases: empty comments, single comment, maximum pagination
  - [x] 6.3 Test error scenarios: invalid repository, non-existent issue, invalid parameters
  - [x] 6.4 Update test harness to support comment listing operations
  - [x] 6.5 Mock Gitea API responses for consistent testing
  - [x] 6.6 Update project documentation including README examples
  - [x] 6.7 Add usage examples showing how to use the new list_issue_comments tool
  - [x] 6.8 Verify all tests pass

- [x] 7. Final Integration and Quality Assurance
  - [x] 7.1 Run complete test suite to ensure no regressions
  - [x] 7.2 Perform code review against project standards
  - [x] 7.3 Verify proper error handling and logging
  - [x] 7.4 Test with real Forgejo instance (if available)
  - [x] 7.5 Update any additional documentation or configuration files
  - [x] 7.6 Verify build and deployment processes work correctly