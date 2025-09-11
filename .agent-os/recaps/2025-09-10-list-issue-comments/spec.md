# Spec Requirements Document

> Spec: List Issue Comments Tool
> Created: 2025-09-10

## Overview

Add a list_issue_comments tool to the forgejo-mcp project, enabling users to programmatically retrieve all comments from specific issues in Forgejo/Gitea repositories through the MCP interface, complementing the existing create_issue_comment functionality.

## User Stories

### Issue Comment Retrieval

As a developer, I want to list comments on issues in my Forgejo repository so that I can review discussion history, understand context, and track issue progress without switching to the web interface.

### Automated Comment Analysis

As a project maintainer, I want to programmatically access issue comments so that I can analyze discussion patterns, identify frequently asked questions, and generate summary reports for team meetings.

### CI/CD Integration

As a CI/CD system, I want to retrieve issue comments so that I can check for specific keywords, approvals, or requirements before proceeding with automated workflows or deployments.

## Spec Scope

1. **Core Comment Listing**: Implement a list_issue_comments tool that accepts repository, issue number, and optional pagination parameters to retrieve comments from Forgejo issues.

2. **Interface Extension**: Extend the existing Gitea interface with an IssueCommentLister interface and ListIssueComments method to maintain clean architecture patterns.

3. **Service Layer Integration**: Add business logic for comment listing including repository format validation, issue number validation, and pagination parameter validation.

4. **MCP Tool Registration**: Register the new tool with the MCP server including proper schema definition and response formatting.

5. **Comprehensive Testing**: Add unit tests for service methods, integration tests for the tool handler, and update the existing test harness to support comment listing operations.

## Out of Scope

- Comment editing or deletion functionality
- Comment threading or nested replies
- File attachments in comments
- Comment creation or updating (already implemented)
- Issue state changes (open/close/reopen)
- Label management or assignment
- Comment reactions or emojis
- Bulk comment operations across multiple issues
- Comment search or filtering by content
- User profile information retrieval

## Expected Deliverable

1. A fully functional list_issue_comments MCP tool that successfully retrieves comments from specified issues in Forgejo repositories and returns structured response data including comment ID, content, author, creation date, and pagination metadata.

2. Complete test coverage with unit tests for all new service methods, integration tests for the MCP tool handler, and updated test harness functionality that validates comment listing against a mock Forgejo server.

3. Updated project documentation including README examples showing how to use the new list_issue_comments tool with proper parameter formatting and expected response handling.