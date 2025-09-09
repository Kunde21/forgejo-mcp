# Spec Requirements Document

> Spec: Issue Comment Tool
> Created: 2025-09-09
> Status: Planning

## Overview

The goal is to add a create_issue_comment tool to the forgejo-mcp project, enabling users to programmatically add comments to issues in Forgejo/Gitea repositories through the MCP interface.

## User Stories

1. **As a developer**, I want to comment on issues in my Forgejo repository so that I can provide feedback, ask questions, or update issue status directly from my MCP client without switching to the web interface.

2. **As a project maintainer**, I want to automate issue comment workflows so that I can quickly respond to multiple issues with standardized responses or update team members on issue progress.

3. **As a CI/CD system**, I want to post automated comments to issues so that I can notify developers about build statuses, test results, or deployment outcomes related to specific issues.

## Spec Scope

1. **Core Comment Creation**: Implement a create_issue_comment tool that accepts repository, issue number, and comment content parameters to add comments to Forgejo issues.

2. **Interface Extension**: Extend the existing Gitea interface with an IssueCommenter interface and CreateIssueComment method to maintain clean architecture patterns.

3. **Service Layer Integration**: Add business logic for comment creation including repository format validation, issue number validation, and comment content validation.

4. **MCP Tool Registration**: Register the new tool with the MCP server including proper schema definition and response formatting.

5. **Comprehensive Testing**: Add unit tests for service methods, integration tests for the tool handler, and update the existing test harness to support comment operations.

## Out of Scope

- Issue editing or deletion functionality
- Comment threading or nested replies
- File attachments in comments
- Comment editing or updating
- Issue state changes (open/close/reopen)
- Label management or assignment
- Mention notifications or @mentions
- Comment reactions or emojis
- Bulk comment operations
- Comment history or retrieval

## Expected Deliverable

1. A fully functional create_issue_comment MCP tool that successfully adds comments to specified issues in Forgejo repositories and returns structured response data including comment ID, issue number, and repository information.

2. Complete test coverage with unit tests for all new service methods, integration tests for the MCP tool handler, and updated test harness functionality that validates comment creation against a mock Forgejo server.

3. Updated project documentation including README examples showing how to use the new create_issue_comment tool with proper parameter formatting and expected response handling.

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-09-issue-comment-tool/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-09-issue-comment-tool/sub-specs/technical-spec.md