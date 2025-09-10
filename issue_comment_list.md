## Plan: List Issue Comments Tool

### Overview

Add a list_issue_comments tool to complement the existing create_issue_comment functionality, enabling users to retrieve all comments from a specific issue in Forgejo/Gitea repositories.

### Technical Architecture

#### 1. Interface Extension (remote/gitea/interface.go)

• Add IssueCommentLister interface with ListIssueComments method
• Define IssueCommentList struct for collection of comments
• Maintain consistency with existing interface patterns

#### 2. Client Implementation (remote/gitea/gitea_client.go)

• Implement ListIssueComments method using Gitea SDK
• Handle repository parsing (owner/repo format)
• Convert between Gitea SDK comment types and internal IssueComment struct
• Add pagination support (limit/offset parameters)

#### 3. Service Layer (remote/gitea/service.go)

• Extend Service struct to implement IssueCommentLister interface
• Add comment listing business logic with validation
• Implement repository format, issue number, and pagination validation
• Follow existing validation patterns

#### 4. MCP Handler (server/handlers.go)

• Create handleListIssueComments handler function
• Implement input parameter validation using ozzo-validation
• Return structured response with comment list metadata
• Follow existing error handling patterns

#### 5. Tool Registration (server/server.go)

• Register list_issue_comments tool with proper schema
• Add comprehensive tool description and parameter documentation

### Tool Specification

Tool Name: list_issue_comments

Parameters:

{
  "repository": {
    "type": "string",
    "description": "Repository in 'owner/repo' format",
    "required": true
  },
  "issue_number": {
    "type": "integer",
    "description": "Issue number to list comments from",
    "required": true,
    "minimum": 1
  },
  "limit": {
    "type": "integer",
    "description": "Maximum number of comments to return (1-100, default 15)",
    "required": false,
    "minimum": 1,
    "maximum": 100
  },
  "offset": {
    "type": "integer",
    "description": "Number of comments to skip for pagination (default 0)",
    "required": false,
    "minimum": 0
  }
}

Response Format:

{
  "content": [{"type": "text", "text": "Found 5 comments on issue #42"}],
  "structured": {
    "comments": [
      {
        "id": 123,
        "content": "This is a comment",
        "author": "username",
        "created": "2025-09-09T10:30:00Z"
      }
    ],
    "total_count": 5,
    "issue_number": 42,
    "repository": "owner/repo"
  }
}

### Implementation Tasks

1. Interface Extension
 • Add IssueCommentLister interface
 • Define IssueCommentList struct
 • Update GiteaClientInterface to include new interface
2. Client Implementation
 • Implement ListIssueComments method in GiteaClient
 • Use Gitea SDK's ListIssueComments API
 • Handle pagination and conversion logic
3. Service Layer
 • Add ListIssueComments method to Service
 • Implement validation for all parameters
 • Add business logic for comment listing
4. MCP Handler
 • Create handleListIssueComments function
 • Add parameter validation with ozzo-validation
 • Implement structured response formatting
5. Tool Registration
 • Register tool with proper schema
 • Add comprehensive documentation
6. Testing
 • Unit tests for all new methods
 • Integration tests for complete workflow
 • Update test harness for comment listing
7. Documentation
 • Update README with usage examples
 • Add tool documentation


### Dependencies

No new external dependencies required - will leverage existing Gitea SDK, MCP SDK, and validation libraries.

### Integration Points

• Extends existing comment functionality
• Follows established patterns from list_issues and create_issue_comment
• Maintains clean architecture with interface-based design
• Preserves backward compatibility

This plan provides a comprehensive approach to implementing the list issue comments tool while maintaining consistency with the existing codebase architecture and patterns.
