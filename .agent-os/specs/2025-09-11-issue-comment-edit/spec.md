# Spec Requirements Document

> Spec: Issue Comment Edit
> Created: 2025-09-11

## Overview

Add a new MCP tool `issue_comment_edit` that enables AI agents to edit existing comments on Forgejo/Gitea repository issues, providing complete comment lifecycle management capabilities.

## User Stories

### Edit Issue Comment

As an AI agent, I want to edit existing comments on Forgejo/Gitea repository issues, so that I can update information, correct errors, or provide additional context without creating duplicate comments.

The workflow involves identifying a specific comment by its ID, providing updated content, and receiving confirmation of the successful edit with the updated comment metadata. This enables agents to maintain accurate and up-to-date discussions in issue threads.

## Spec Scope

1. **Issue Comment Editing** - Modify the content of existing comments on repository issues
2. **Comment Validation** - Ensure edited comments meet content and format requirements
3. **Permission Handling** - Verify user has authorization to edit the specified comment
4. **Response Formatting** - Return structured confirmation with updated comment metadata

## Out of Scope

- Comment deletion functionality
- Bulk comment editing operations
- Comment reaction management
- Issue editing capabilities
- Pull request comment editing

## Expected Deliverable

1. A functional MCP tool `issue_comment_edit` that successfully updates comment content and returns confirmation
2. Proper error handling for invalid inputs, permission issues, and API failures
3. Integration with existing MCP server following established patterns from other comment tools