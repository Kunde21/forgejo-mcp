# Spec Requirements Document

> Spec: PR Comment Edit
> Created: 2025-09-12

## Overview

Add a new MCP tool `pr_comment_edit` that enables AI agents to edit existing comments on Forgejo/Gitea repository pull requests, providing complete PR comment lifecycle management capabilities alongside existing create and list PR comment tools.

## User Stories

### Edit PR Comment

As an AI agent, I want to edit existing comments on Forgejo/Gitea repository pull requests, so that I can update information, correct errors, or provide additional context without creating duplicate comments.

The workflow involves identifying a specific PR comment by its ID, providing updated content, and receiving confirmation of the successful edit with the updated comment metadata. This enables agents to maintain accurate and up-to-date discussions in PR threads.

## Spec Scope

1. **PR Comment Editing** - Modify the content of existing comments on repository pull requests
2. **Comment Validation** - Ensure edited comments meet content and format requirements
3. **Permission Handling** - Verify user has authorization to edit the specified comment
4. **Response Formatting** - Return structured confirmation with updated comment metadata

## Out of Scope

- Comment deletion functionality
- Bulk comment editing operations
- Comment reaction management
- PR editing capabilities
- Issue comment editing (already implemented)

## Expected Deliverable

1. A functional MCP tool `pr_comment_edit` that successfully updates PR comment content and returns confirmation
2. Proper error handling for invalid inputs, permission issues, and API failures
3. Integration with existing MCP server following established patterns from other PR comment tools

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

## Implementation Plan

### Phase 1: Interface and Client Layer
1. Add `EditPullRequestComment` method to `GiteaClientInterface`
2. Create `EditPullRequestCommentArgs` struct with validation tags
3. Implement client layer method using Gitea SDK
4. Add comprehensive tests for interface and client layers

### Phase 2: Service and Handler Layer
1. Add `EditPullRequestComment` method to service layer
2. Implement `handlePullRequestCommentEdit` function with ozzo-validation
3. Create `PullRequestCommentEditResult` struct
4. Add handler layer tests

### Phase 3: Server Integration
1. Register `pr_comment_edit` tool with MCP server
2. Add tool description and metadata
3. Wire handler function to tool registration
4. Add server integration tests

### Phase 4: Testing and Documentation
1. Implement integration tests for complete PR comment edit workflow
2. Add mock server support for testing
3. Update documentation and examples
4. Perform acceptance testing

### Phase 5: Verification and Deployment
1. Run complete test suite
2. Perform code review and quality checks
3. Verify integration with existing MCP server patterns
4. Test backward compatibility with existing PR comment tools

## Dependencies

- Existing PR comment infrastructure (`pr_comment_create`, `pr_comment_list`)
- Gitea SDK support for PR comment editing
- MCP SDK v0.4.0 handler patterns
- ozzo-validation for input validation

## Risk Assessment

### Low Risk
- Following established patterns from issue comment edit implementation
- Reusing existing PR comment data structures and validation logic
- Well-tested MCP server integration patterns

### Medium Risk
- Gitea SDK may not have PR comment editing support (needs verification)
- API differences between issue and PR comment editing endpoints

### Mitigation Strategies
- Verify Gitea SDK capabilities before implementation
- Use issue comment edit as reference implementation
- Comprehensive testing at each layer
- Gradual rollout with feature flags if needed

## Success Criteria

1. **Functional**: `pr_comment_edit` tool successfully updates PR comment content
2. **Reliable**: Proper error handling for all edge cases and failure scenarios
3. **Integrated**: Seamless integration with existing MCP server and PR comment tools
4. **Tested**: Comprehensive test coverage including unit, integration, and acceptance tests
5. **Documented**: Complete documentation with usage examples and API reference