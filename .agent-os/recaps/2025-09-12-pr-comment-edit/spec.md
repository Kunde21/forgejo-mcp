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