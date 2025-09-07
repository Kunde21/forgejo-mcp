# Spec Requirements Document

> Spec: list-issues-tool
> Created: 2025-09-07
> Status: Planning

## Overview
Implement a new MCP tool called "list_issues" that connects to a Gitea/Forgejo instance to retrieve and display issues for a specified repository with pagination support.

## User Stories

### Repository Issue Management
As a developer using AI assistance, I want to list issues from a Gitea/Forgejo repository so that I can understand current work items and priorities.

**Workflow:**
1. User provides repository in "owner/repository" format
2. AI agent calls list_issues tool with repository parameter
3. Tool connects to configured Gitea/Forgejo instance
4. Returns structured list of issues with number, title, and status
5. User can specify pagination parameters for large repositories

### Issue Status Tracking
As a project manager, I want to see issue status information so that I can track project progress and identify bottlenecks.

**Workflow:**
1. Tool returns issues with status information (open, closed, WIP)
2. Status derived from issue state and any associated pull requests
3. Clear indication of issue progression and completion status

## Spec Scope

1. **list_issues Tool** - MCP tool that accepts repository, limit, and offset parameters
2. **Gitea SDK Integration** - Direct API integration using official Gitea SDK
3. **Configuration Management** - Environment variable support for remote URL and auth token
4. **Pagination Support** - Configurable limit (1-100) and offset parameters
5. **Error Handling** - Comprehensive error handling for API failures and invalid inputs

## Out of Scope

- Issue creation or modification functionality
- Pull request management (separate tools)
- Repository creation or management
- User authentication management
- Advanced filtering beyond basic pagination

## Expected Deliverable

1. MCP server accepts list_issues tool calls and returns properly formatted issue data
2. Tool validates repository format and pagination parameters
3. Successful API integration with Gitea/Forgejo instances
4. Comprehensive test coverage including mock server scenarios

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-07-list-issues-tool/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-07-list-issues-tool/sub-specs/technical-spec.md