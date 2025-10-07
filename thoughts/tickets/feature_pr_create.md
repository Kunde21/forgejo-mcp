---
type: feature
priority: medium
created: 2025-10-06T00:00:00Z
created_by: GLM 4.6
status: researched
tags: [pr, create, forgejo, gitea, mcp, pull request]
keywords: [pr, pull request, create, forgejo, gitea, branch, fork, reviewer, template, validation, directory, repository]
patterns: [tool implementation, parameter validation, branch detection, remote detection, error handling, directory resolution, template processing]
---

# FEATURE-004: Add tool to create pull requests on repository

## Description
Create a new MCP tool that allows users to create pull requests on Forgejo/Gitea repositories through an Agent interface. The tool should support creating PRs from the current branch (or specified branch) to a target branch, handling both same-repo and fork-to-repo scenarios, following the existing patterns in the repository.

## Context
Developers working on the repository need the ability to quickly create pull requests when they have changes ready for review. This tool integrates with the existing Forgejo/Gitea MCP server to provide seamless PR creation capabilities, connecting the entire development process within the agent.

## Requirements
- Create PRs from current or specified branch to target branch (default or specified)
- Support both same-repo and fork-to-repo PR creation
- Auto-detect current git branch and remote configuration when not provided
- Support PR templates from .gitea/PULL_REQUEST_TEMPLATE.md on target branch
- Allow draft PRs with ability to change between draft and ready-to-review
- Support single reviewer assignment
- Validate branch status and conflicts before creation
- Use markdown formatting for PR description
- Follow existing tool patterns in the repository

### Functional Requirements
- [ ] Create PR with title and description
- [ ] Auto-detect current branch and remote when not specified
- [ ] Support custom target branch (not just default)
- [ ] Handle fork-to-repo PR creation
- [ ] Load and use PR templates when available
- [ ] Support draft and ready-to-review PRs
- [ ] Assign single reviewer to PR
- [ ] Validate source branch exists and has new commits
- [ ] Detect and report branch conflicts with details
- [ ] Ensure source branch is not behind target branch
- [ ] Return warnings for self-assigned reviewers
- [ ] Support special characters in PR titles
- [ ] Handle markdown formatting in descriptions
- [ ] Return errors for read-only repositories
- [ ] Handle protection rule failures gracefully

### Non-Functional Requirements
- [ ] Follow existing code patterns and conventions
- [ ] Integrate with existing error handling
- [ ] Support both Forgejo and Gitea backends
- [ ] Validate all inputs before remote operations
- [ ] Return detailed error messages for user action

## Current State
Repository has existing tools for listing PRs, creating/editing PR comments, and editing PRs. No tool exists for creating new pull requests.

## Desired State
New tool `forgejo_pr_create` that can create PRs from branches following the established patterns.

## Research Context

### Keywords to Search
- pr - Existing PR-related functionality
- pull request - PR creation and management
- create - Creation patterns in existing tools
- forgejo - Forgejo client implementation
- gitea - Gitea client implementation
- branch - Branch detection and validation
- fork - Fork handling logic
- reviewer - Reviewer assignment patterns
- template - Template loading and processing
- validation - Input validation patterns
- directory - Directory parameter handling
- repository - Repository resolution logic

### Patterns to Investigate
- tool implementation - How existing MCP tools are structured
- parameter validation - Input validation patterns
- branch detection - How git branches are detected and validated
- remote detection - How git remotes are resolved
- error handling - Error reporting and user feedback
- directory resolution - How directory parameters are processed
- template processing - How templates are loaded and processed

### Key Decisions Made
- Use existing authentication system
- Follow established tool naming conventions
- Support both same-repo and fork-to-repo scenarios
- Auto-detect branch and remote when not provided
- Support single reviewer only (multiple reviewers future scope)
- No attachments in initial implementation
- No preview/dry-run mode initially
- No local git operations (branch must be pushed)
- No PR title deduplication
- No issue references in titles (only descriptions)

## Success Criteria

### Automated Verification
- [ ] Tool creates PRs successfully from same repo
- [ ] Tool creates PRs successfully from forks
- [ ] Branch detection works correctly
- [ ] Remote detection works correctly
- [ ] Template loading works when template exists
- [ ] Conflict detection reports detailed errors
- [ ] Behind-branch detection prevents creation
- [ ] Draft PR creation works
- [ ] Reviewer assignment works
- [ ] Error conditions are properly reported

### Manual Verification
- [ ] PR appears on repository with correct content
- [ ] PR is created from correct source to target branch
- [ ] Template content is used when available
- [ ] Draft status can be toggled via edit tool
- [ ] Reviewer is assigned correctly
- [ ] Warnings appear for self-assigned reviewers
- [ ] Tool integrates seamlessly with existing tools

## Related Information
- Existing PR listing tools
- PR comment creation/editing tools
- PR edit tool
- Issue creation tool (for patterns)
- Repository resolver implementation

## Notes
- Need to research Forgejo/Gitea PR creation API endpoints
- Should investigate how existing tools handle branch detection
- Need to understand current authentication flow for private repos
- Must ensure source branch is pushed to remote before PR creation
- Need to handle both master and main default branches
- Should research PR template location and format
- Need to understand fork-to-repo PR creation flow
