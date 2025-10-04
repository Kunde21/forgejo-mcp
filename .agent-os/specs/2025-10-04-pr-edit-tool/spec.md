# Spec Requirements Document

> Spec: pr-edit-tool
> Created: 2025-10-04

## Overview

Add a pull request edit tool to the forgejo-mcp server that enables AI agents to modify pull request metadata including title, body, state, and base branch in Forgejo/Gitea repositories. This feature completes the PR management toolset by providing the missing edit capability for pull requests themselves.

## User Stories

### PR Metadata Editing

As an AI agent, I want to edit pull request metadata (title, body, state, base branch), so that I can manage and update pull requests as part of automated workflows and repository maintenance tasks.

The agent should be able to modify any combination of editable fields on a pull request, with proper validation and error handling. The tool should support both repository specification by owner/repo format and automatic repository resolution from local directories.

## Spec Scope

1. **Pull Request Metadata Editing** - Modify title, body, state, and base branch of existing pull requests
2. **Repository Resolution** - Support both explicit repository specification and directory-based auto-resolution
3. **Input Validation** - Comprehensive validation of all parameters with clear error messages
4. **Multi-Platform Support** - Consistent behavior across both Forgejo and Gitea repositories

## Out of Scope

- Creating new pull requests (handled by existing tools)
- Managing pull request reviewers and assignees
- Modifying pull request merge settings or strategies
- Editing pull request comments (handled by existing pr_comment_edit tool)

## Expected Deliverable

1. A new `pr_edit` MCP tool that successfully modifies pull request metadata in both Forgejo and Gitea repositories
2. Comprehensive test coverage including unit tests, integration tests, and validation scenarios
3. Complete documentation following project code style guidelines