# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-10-04-pr-edit-tool/spec.md

> Created: 2025-10-04
> Version: 1.0.0

## Endpoints

### MCP Tool: pr_edit

**Purpose:** Edit a pull request in a Forgejo/Gitea repository by modifying metadata fields
**Parameters:** 
- repository (string, optional): Repository in "owner/repo" format
- directory (string, optional): Local directory path for repository auto-resolution
- pull_request_number (int, required): Number of the pull request to edit
- title (string, optional): New title for the pull request
- body (string, optional): New description/body for the pull request
- state (string, optional): New state ("open" or "closed")
- base_branch (string, optional): New base branch for the pull request

**Response:** Updated pull request metadata with confirmation message
**Errors:** 
- Validation errors for missing required fields or invalid formats
- Repository not found or access denied
- Pull request not found or edit permissions denied
- Invalid state transition or base branch

## Controllers

### handlePullRequestEdit

**Action:** Process pull request edit requests
**Business Logic:**
- Validate input parameters using ozzo-validation
- Resolve repository from directory if provided
- Call appropriate remote client (Forgejo/Gitea) EditPullRequest method
- Format success response with updated PR metadata
- Handle and format error responses consistently

**Error Handling:**
- Structured error responses with descriptive messages
- Proper error wrapping with context
- Validation errors returned before API calls
- Remote API errors properly propagated and formatted