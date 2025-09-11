# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-11-issue-comment-edit/spec.md

## Technical Requirements

- **Interface Layer**: Add `EditIssueComment` method to `GiteaClientInterface` with `EditIssueCommentArgs` struct containing validation tags for repository, issue_number, comment_id, and new_content parameters
- **Client Layer**: Implement `EditIssueComment` method using Gitea SDK's `EditIssueComment` function, parsing repository format (owner/repo), converting responses to our `IssueComment` struct, and handling errors with proper context
- **Handler Layer**: Add `handleIssueCommentEdit` function with input validation using ozzo-validation, service layer integration, and structured success/error responses with `CommentEditResult` struct
- **Server Registration**: Register new tool with MCP server including tool description, metadata, and handler wiring following existing patterns

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