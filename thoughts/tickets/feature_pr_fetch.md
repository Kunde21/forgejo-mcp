---
type: feature
priority: high
created: 2025-01-14T10:00:00Z
created_by: Opus
status: planned
tags: [pr, fetch, validation, edit]
keywords: [pull request, fetch, get, metadata, status, author, reviewers, approvals, build status, labels, created, updated, branches, commits, conflicts, draft, forgejo, gitea]
patterns: [PR metadata fetching, remote detection, error handling, validation, tool integration]
---

# FEATURE-003: Add tool to fetch single PR information

## Description
Create a new MCP tool to fetch comprehensive information about a single pull request, providing all necessary metadata to validate the Create PR tool and enable the Edit PR tool functionality.

## Context
Agents need to retrieve detailed PR information before performing edits and to validate PR creation. This tool will complement the existing PR list tool by providing in-depth information for individual PRs.

## Requirements
- Fetch detailed information for a single PR by number
- Return comprehensive metadata for validation and editing
- Support all PR states (open, closed, merged, draft)
- Handle both same-repository and fork-based PRs
- Integrate with existing remote detection and authentication

### Functional Requirements
- Fetch PR by number with repository context
- Return title, description, status, author, reviewers, approvals, build status, labels, created/updated dates
- Include source and target branch information
- Include latest commit hash and commit count
- Include merge conflict and out-of-date status
- Include PR size metrics (files changed, additions, deletions)
- Indicate draft PR status
- Include base and fork repository references
- Return comment count (without details)
- Handle Forgejo and Gitea remotes

### Non-Functional Requirements
- No caching - fetch fresh data each call
- Clear error messages for invalid PR numbers or access denied
- Follow existing code patterns and conventions
- Use existing authentication and permission handling

## Current State
- PR list tool exists for listing multiple PRs
- Remote detection and client interfaces exist
- Create PR tool needs validation
- Edit PR tool needs data source

## Desired State
- Single PR fetch tool that returns all necessary metadata
- Seamless integration with existing PR tools
- Complete data for PR validation and editing

## Research Context
### Keywords to Search
- pull request - Core PR functionality
- fetch - Data retrieval patterns
- get - Existing get methods
- metadata - PR metadata structures
- status - PR status handling
- author - User information
- reviewers - Reviewer management
- approvals - Approval status
- build status - CI/CD integration
- labels - Label management
- created/updated - Timestamp handling
- branches - Branch information
- commits - Commit data
- conflicts - Conflict detection
- draft - Draft PR handling
- forgejo - Forgejo client
- gitea - Gitea client

### Patterns to Investigate
- PR metadata fetching - How existing tools fetch PR data
- Remote detection - How different remote types are handled
- Error handling - Consistent error response patterns
- Validation - Input validation patterns
- Tool integration - How tools are registered and exposed

### Key Decisions Made
- Single PR only (no bulk fetching)
- Metadata only (no diff content)
- No caching (out of scope)
- Authentication handled by existing systems
- Follow existing tool patterns

## Success Criteria

### Automated Verification
- [ ] Tool fetches PR data without errors
- [ ] All required fields are returned
- [ ] Error handling works for invalid inputs
- [ ] Integration tests pass with both Forgejo and Gitea

### Manual Verification
- [ ] Tool returns complete PR metadata
- [ ] Works with open, closed, merged, and draft PRs
- [ ] Handles fork-based PRs correctly
- [ ] Clear error messages for invalid PRs
- [ ] Data is sufficient for PR validation and editing

## Related Information
- Existing PR list tool (server/pr_list.go)
- Create PR tool implementation
- Remote client interfaces (remote/interface.go)
- Forgejo client (remote/forgejo/)
- Gitea client (remote/gitea/)

## Notes
- Tool should be named consistently with existing PR tools
- Must match the parameter and response patterns of other tools
- Consider the data structure needed for Edit PR tool implementation