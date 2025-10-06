---
type: feature
priority: medium
created: 2025-10-06T00:00:00Z
created_by: Opus
status: reviewed
tags: [issue, edit, mcp, forgejo, gitea]
keywords: [issue edit, forgejo_issue_edit, gitea_issue_edit, issue_update, patch_issue]
patterns: [MCP tool implementation, issue management, remote API integration, partial updates]
---

# FEATURE-003: Add issue editing tool to MCP server

## Description
Implement a new MCP tool that allows editing existing issues on Forgejo/Gitea repositories. The tool should enable developers to update issue titles, bodies, labels, assignees, and state through the MCP interface.

## Context
Developers need to refine and add information to issues as more details are discovered during development. Currently, the MCP server supports creating and listing issues, but lacks the ability to edit existing ones. This feature completes the basic issue management lifecycle.

## Requirements
- Edit issue title, body, labels, assignee, and state
- Support partial updates (only update specified fields)
- Return the updated issue object on success
- Handle conflicts by re-reading and returning current issue state
- Enforce permissions through the remote API
- Prevent blank titles or empty bodies
- Support markdown formatting in issue bodies
- Use same directory parameter pattern as existing tools

### Functional Requirements
- Edit issue title with non-empty string validation
- Edit issue body with markdown support and non-empty validation
- Add/remove/replace labels on issues
- Add/remove/replace assignees on issues
- Change issue state (open/closed)
- Support partial field updates
- Return complete updated issue object

### Non-Functional Requirements
- Follow existing MCP tool design patterns
- Use directory parameter for repository resolution
- Handle remote API errors gracefully
- Return appropriate error messages for invalid operations
- Maintain consistency with other issue-related tools

## Current State
MCP server supports:
- Creating issues (forgejo_issue_create)
- Listing issues (forgejo_issue_list)
- Creating/editing issue comments
- No direct issue editing capability

## Desired State
New tool `forgejo_issue_edit` that can modify existing issues with full CRUD support for issue metadata.

## Research Context
Information for research agents to understand the implementation requirements.

### Keywords to Search
- forgejo_issue_edit - New tool name to implement
- issue update - Existing patterns for updating resources
- patch_issue - Partial update patterns
- issues.go - Existing issue handling code
- remote/interface.go - Remote API interface definitions

### Patterns to Investigate
- MCP tool implementation - How other tools are structured in server/
- issue management - Current issue creation/listing patterns
- partial updates - How other tools handle optional parameters
- remote API integration - Forgejo/Gitea client patterns
- error handling - Consistent error response patterns
- directory parameter - Repository resolution pattern

### Key Decisions Made
- Use same tool naming convention: forgejo_issue_edit
- Support partial updates with optional parameters
- Return updated issue object on success
- Leverage remote API for permission enforcement
- Follow existing validation patterns (non-empty title/body)
- Use directory parameter for repository identification

## Success Criteria
How to verify the ticket is complete.

### Automated Verification
- [ ] go test ./... passes with new tests
- [ ] Tool registers properly in MCP server
- [ ] Integration tests cover all edit scenarios
- [ ] Error handling tests pass

### Manual Verification
- [ ] Can edit issue title successfully
- [ ] Can edit issue body with markdown
- [ ] Can update labels and assignees
- [ ] Can change issue state
- [ ] Partial updates work correctly
- [ ] Errors returned for invalid operations
- [ ] Non-existent issues return appropriate error

## Related Information
- Existing issue tools: forgejo_issue_create, forgejo_issue_list
- Issue comment tools for reference patterns
- Remote client implementations in remote/ directory
- MCP server architecture in server/

## Notes
- Tool should be implemented in server/ directory following existing patterns
- Need to add corresponding tests in server_test/
- Consider adding mock tests for various error scenarios
- Ensure consistent parameter naming with other issue tools