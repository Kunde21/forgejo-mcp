# Spec Requirements Document

> Spec: Pull Request Comment Create Tool
> Created: 2025-09-12

## Overview

Create a new MCP tool `pr_comment_create` that enables users to create comments on specified Forgejo/Gitea pull requests, following the established patterns in the codebase to enable AI agents to efficiently add comments and feedback to pull request discussions.

## User Stories

### Pull Request Comment Creation

As an AI agent, I want to create comments on a specific pull request, so that I can provide feedback, ask questions, or update the PR status directly from my MCP client without switching to the web interface.

### Automated PR Workflows

As an AI agent, I want to automate pull request comment workflows, so that I can quickly respond to multiple PRs with standardized responses or update team members on PR progress.

### Structured Comment Creation

As an AI agent, I want to create pull request comments with structured input validation and proper error handling, so that I can reliably add comments with proper formatting and receive clear feedback on success or failure.

## Spec Scope

1. **Pull Request Comment Creation** - Implement MCP tool to create comments on specified pull requests in Forgejo/Gitea repositories
2. **Interface Extension** - Extend the existing Gitea interface with a `PullRequestCommenter` interface and `CreatePullRequestComment` method
3. **Service Layer Integration** - Add business logic for comment creation without validation (validation handled in server handler only)
4. **Server Handler Implementation** - Implement MCP tool handler with ozzo-validation for all parameter validation
5. **Comprehensive Testing** - Add unit tests for service methods, integration tests for the tool handler, and update the existing test harness

## Out of Scope

- Pull request creation or modification
- Comment editing, deletion, or listing
- Pull request merging or status changes
- Comment threading or nested replies
- File attachments in comments
- User authentication or authorization management
- Repository administration operations

## Expected Deliverable

1. A functional MCP tool named `pr_comment_create` that successfully creates comments on pull requests with proper validation
2. Complete test coverage including unit tests, integration tests, and validation tests following existing codebase patterns
3. Documentation and examples demonstrating tool usage and integration with AI agent workflows

## Technical Implementation

### Tool Name
`pr_comment_create`

### Parameters
- `repository`: string (owner/repo format, required)
- `pull_request_number`: integer (positive, required)
- `comment`: string (non-empty, required)

### Validation Strategy
- **Server Handler Only**: All input validation performed using ozzo-validation in the server handler
- **Service Layer Trust**: Service layer assumes inputs are valid and focuses on business logic
- **No Validation Duplication**: Eliminates redundant validation between service and server layers

### Validation Rules
```go
v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'"))
v.Field(&args.PullRequestNumber, v.Min(1))
v.Field(&args.Comment, v.Required, v.Length(1, 0)) // Non-empty string
```

### Architecture Pattern
- **Interface Layer**: Add `PullRequestCommenter` interface and type definitions (no validation tags)
- **Client Layer**: Implement `CreatePullRequestComment` method using Gitea SDK (no validation)
- **Service Layer**: Add `CreatePullRequestComment` method (no validation, direct pass-through)
- **Server Layer**: Add MCP tool handler with ozzo-validation for all parameter validation

### Files to Modify
1. `remote/gitea/interface.go` - Add interface and type definitions
2. `remote/gitea/gitea_client.go` - Implement client method
3. `remote/gitea/service.go` - Add service method
4. `server/pr_comments.go` - Add MCP tool handler
5. `server/server.go` - Register new tool
6. `server_test/` - Add comprehensive tests
7. `README.md` - Update documentation

## Success Criteria

- ✅ New `pr_comment_create` tool successfully creates comments on pull requests
- ✅ Validation performed only in server handler using ozzo-validation
- ✅ Service layer has no validation logic (clean separation of concerns)
- ✅ All existing functionality remains intact (no regressions)
- ✅ Complete test coverage with all tests passing
- ✅ Proper error handling for both validation and API errors
- ✅ Documentation updated with usage examples

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-12-pr-comment-create/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-12-pr-comment-create/sub-specs/technical-spec.md