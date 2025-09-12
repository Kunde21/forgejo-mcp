# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-09-12-pr-comment-edit/spec.md

## Endpoints

### PATCH /repos/{owner}/{repo}/issues/comments/{id}

**Purpose:** Edit an existing pull request comment
**Parameters:** 
- Path: owner (string), repo (string), id (integer)
- Body: {"body": "new comment content"}
**Response:** Updated comment object with ID, body, user, created_at, updated_at
**Errors:** 404 (not found), 403 (forbidden), 422 (validation error)

## Controllers

### handlePullRequestCommentEdit
- Action: Edit pull request comment
- Business Logic: Validate input, call Gitea SDK, format response
- Error Handling: Return structured errors for validation, API failures, permissions

## Purpose

- Provides MCP tool interface for PR comment editing
- Integrates with existing PR comment management tools
- Follows established patterns from issue comment editing