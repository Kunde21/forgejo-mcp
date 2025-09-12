# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-12-pr-comment-edit/spec.md

## Technical Requirements

### Interface Layer
- Add `EditPullRequestComment` method to `GiteaClientInterface` with `EditPullRequestCommentArgs` struct
- Include validation tags for repository, pull_request_number, comment_id, and new_content parameters

### Client Layer
- Implement `EditPullRequestComment` method using Gitea SDK's `EditPullRequestComment` function
- Parse repository format (owner/repo), convert responses to our `PullRequestComment` struct
- Handle errors with proper context

### Handler Layer
- Add `handlePullRequestCommentEdit` function with input validation using ozzo-validation
- Service layer integration with structured success/error responses
- Create `PullRequestCommentEditResult` struct for response formatting

### Server Registration
- Register new tool with MCP server including tool description, metadata, and handler wiring
- Follow existing patterns from `pr_comment_create` and `pr_comment_list` tools

## Tool Specification

- **Name**: `pr_comment_edit`
- **Description**: Edit an existing comment on a Forgejo/Gitea repository pull request
- **Parameters**:
  - `repository` (string, required): Repository path in "owner/repo" format
  - `pull_request_number` (int, required): Pull request number containing the comment
  - `comment_id` (int, required): ID of the comment to edit
  - `new_content` (string, required): Updated comment content

## Validation Rules

- Repository: Required, must match regex `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
- Pull request number: Required, must be positive integer
- Comment ID: Required, must be positive integer
- New content: Required, non-empty string

## Error Handling

- Invalid repository format
- Non-existent pull request/comment
- Permission errors
- Network/API failures
- Validation errors

## Success Response

- Confirmation message with updated comment metadata
- Return updated comment object with ID, body, user, created_at, updated_at timestamps