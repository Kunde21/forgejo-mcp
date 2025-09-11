# Tool Plan: Edit Issue Comment

## Overview
Add a new MCP tool `issue_comment_edit` that allows editing existing comments on Forgejo/Gitea repository issues.

## Implementation Components

### 1. Interface Layer (`remote/gitea/interface.go`)
- Add `EditIssueComment` method to `GiteaClientInterface`
- Define `EditIssueCommentArgs` struct with validation tags
- Parameters: repository, issue_number, comment_id, new_content

### 2. Client Layer (`remote/gitea/gitea_client.go`)
- Implement `EditIssueComment` method using Gitea SDK
- Parse repository format (owner/repo)
- Use `EditIssueComment` from Gitea SDK
- Convert response to our `IssueComment` struct
- Handle errors with proper context

### 3. Handler Layer (`server/handlers.go`)
- Add `handleIssueCommentEdit` handler function
- Validate input parameters using ozzo-validation
- Call service layer to edit comment
- Return formatted success/error responses
- Add `CommentEditResult` struct for response data

### 4. Server Registration (`server/server.go`)
- Register new tool with MCP server
- Add tool description and metadata
- Wire handler to tool registration

## Tool Specification
- **Name**: `issue_comment_edit`
- **Description**: Edit an existing comment on a Forgejo/Gitea repository issue
- **Parameters**:
  - `repository` (string, required): Repository path in "owner/repo" format
  - `issue_number` (int, required): Issue number containing the comment
  - `comment_id` (int, required): ID of the comment to edit
  - `new_content` (string, required): Updated comment content

## Validation Rules
- Repository: Required, must match regex `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
- Issue number: Required, must be positive integer
- Comment ID: Required, must be positive integer
- New content: Required, non-empty string

## Error Handling
- Invalid repository format
- Non-existent issue/comment
- Permission errors
- Network/API failures
- Validation errors

## Success Response
- Confirmation message with updated comment metadata
- Return updated comment object with ID, content, author, created timestamp

## Implementation Notes
This follows the existing patterns in the codebase for issue listing, comment creation, and comment listing tools.
