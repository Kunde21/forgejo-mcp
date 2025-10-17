---
type: feature
priority: medium
created: 2025-10-16T00:00:00Z
status: reviewed
tags: [notifications, gitea, forgejo, mcp-tool]
keywords: [notifications, ListNotifications, NotificationThread, NotificationSubject, notification_list, repository resolver, pagination, offset/limit, unread/read status, notification count]
patterns: [Remote client interface pattern, Server tool registration pattern, Repository resolver pattern, Pagination pattern, Mock server testing pattern]
---

# FEATURE-001: Add notification list tool for Gitea/Forgejo

## Description
Add a new MCP tool called `notification_list` that allows users to read their notifications from Gitea/Forgejo remotes. This tool will enable agents to react to notifications and help developers stay aware of mentions, PR reviews, and issue updates.

## Context
The forgejo-mcp project currently supports issues, pull requests, and comments, but lacks notification functionality. Notifications are crucial for developers working on repositories to stay informed about relevant activity. This feature will enable automation workflows where agents can monitor and respond to notifications.

## Requirements

### Functional Requirements
- List user notifications from Gitea/Forgejo remotes
- Support filtering by repository (optional)
- Support filtering by read/unread status
- Include pagination with offset/limit parameters
- Return notification count along with list
- Include repository information for each notification
- Include notification type and issue/PR number
- Integrate with existing repository resolver for auto-detection
- Follow existing codebase patterns for remote clients and server tools

### Non-Functional Requirements
- Support both Gitea and Forgejo remotes (identical APIs)
- Use existing authentication/token system
- Follow existing error handling patterns
- Include comprehensive test coverage
- Maintain consistency with existing tool interfaces

## Current State
No notification functionality exists in the codebase. The project has established patterns for:
- Remote client interfaces (IssueLister, PullRequestLister, etc.)
- Server tool registration (mcp.AddTool)
- Repository resolution from directory paths
- Pagination with limit/offset
- Mock server testing infrastructure

## Desired State
Users can call the `notification_list` tool to retrieve their notifications with filtering and pagination support. The tool integrates seamlessly with existing forgejo-mcp architecture and enables agent-based notification monitoring and response workflows.

## Research Context

### Keywords to Search
- notifications - Core notification functionality
- ListNotifications - Gitea/Forgejo SDK method
- NotificationThread - Primary notification data structure
- NotificationSubject - Notification content structure
- notification_list - Tool name to implement
- repository resolver - Existing auto-detection system
- pagination - Limit/offset implementation
- offset/limit - Pagination parameters
- unread/read status - Status filtering
- notification count - Count functionality

### Patterns to Investigate
- Remote client interface pattern - How IssueLister/PullRequestLister are implemented
- Server tool registration pattern - How mcp.AddTool is used for other tools
- Repository resolver pattern - How directory paths are resolved to repositories
- Pagination pattern - How limit/offset is handled in existing tools
- Mock server testing pattern - How mock servers are set up for testing

### Key Decisions Made
- Listing only (no read/unread management)
- Support repository filtering via optional parameter
- Support status filtering (read/unread)
- Include pagination with offset/limit
- Include repository info, notification type, and issue/PR number in results
- Integrate with existing repository resolver
- Provide both count and list in response
- No time-based filtering
- No notification type filtering beyond remote defaults
- No URLs in response data
- No batch operations

## Success Criteria

### Automated Verification
- [ ] Unit tests pass for notification listing functionality
- [ ] Integration tests pass for both Gitea and Forgejo clients
- [ ] Tool registration and argument validation works correctly
- [ ] Repository resolver integration functions properly
- [ ] Pagination works with limit/offset parameters
- [ ] Mock server tests cover notification scenarios

### Manual Verification
- [ ] Tool can list notifications from a real Gitea/Forgejo instance
- [ ] Repository filtering works correctly
- [ ] Status filtering (read/unread) works correctly
- [ ] Pagination returns correct subsets of notifications
- [ ] Notification count is accurate
- [ ] Repository information is included in results
- [ ] Notification type and issue/PR numbers are included
- [ ] Error handling works for invalid repositories/tokens

## Related Information
- Existing tools: issue_list, pr_list, issue_comment_list, pr_comment_list
- Remote clients: GiteaClient, ForgejoClient
- Repository resolver: RepositoryResolver in server/
- Testing infrastructure: server_test/harness.go

## Notes
This is the first notification-related tool. Future enhancements could include read/unread management, but this initial implementation focuses on listing functionality only. The tool should enable agent workflows where notifications are monitored and appropriate actions (like reading issues/PRs) are taken in response.
