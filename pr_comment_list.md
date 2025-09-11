## Plan: Pull Request Comment List Tool

### Overview

Create a new MCP tool pr_comment_list that retrieves comments from a specified Forgejo/Gitea pull request with pagination support, following the established patterns in the codebase.

### Implementation Components

#### 1. Data Structures (in remote/gitea/interface.go)

• PullRequestComment struct: Similar to IssueComment but for PR comments
• PullRequestCommentList struct: Collection with pagination metadata
• ListPullRequestCommentsArgs struct: Arguments with validation tags
• PullRequestCommentLister interface: Method signature for listing PR comments

#### 2. Service Layer (in remote/gitea/service.go)

• ListPullRequestComments() method: Validation logic and client delegation
• Validation helpers: Reuse existing validateRepository(), validatePagination(), validatePullRequestNumber()

#### 3. Client Implementation (in remote/gitea/gitea_client.go)

• ListPullRequestComments() method: Gitea SDK integration
• Parse repository format and convert between Gitea SDK and internal structs
• Handle pagination using Gitea's ListOptions

#### 4. Server Handler (in new file server/pr_comments.go)

• PullRequestCommentListArgs struct: Handler arguments with validation
• PullRequestCommentListResult struct: Response data structure
• handlePullRequestCommentList() function: MCP tool handler following established patterns

#### 5. Server Registration (in server/server.go)

• Add tool registration in NewFromService() using mcp.AddTool()
• Tool name: pr_comment_list
• Description: Clear description of functionality

### Key Design Decisions

#### Arguments

• repository (string, required): "owner/repo" format
• pull_request_number (int, required): PR number to list comments from
• limit (int, optional, default 15): Max comments (1-100)
• offset (int, optional, default 0): Comments to skip for pagination

#### Validation

• Repository format: ^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$
• PR number: Must be positive integer
• Limit: 1-100 range
• Offset: Non-negative integer

#### Response Format

• Success message with comment count and pagination info
• Structured result with comments array and metadata
• Error handling following existing patterns

### Integration Points

#### Interface Updates

• Extend GiteaClientInterface to include PullRequestCommentLister
• Maintain backward compatibility with existing interfaces

#### Service Layer

• Add validation method for pull request numbers
• Reuse existing validation utilities where possible

#### Client Layer

• Use Gitea SDK's ListPullRequestComments() method
• Handle pagination conversion between our format and Gitea's format

#### Server Layer

• Follow existing handler patterns with ozzo-validation
• Use TextResult() and TextErrorf() for responses
• Maintain consistent error handling and response formatting

### Testing Strategy

• Create unit tests following existing patterns in server_test/
• Test validation logic, happy path, and error cases
• Integration tests with mock Gitea client
• Follow table-driven test pattern used throughout codebase

This plan follows the established architecture and patterns in the codebase, ensuring consistency with existing tools like issue_comment_list and pr_list.
