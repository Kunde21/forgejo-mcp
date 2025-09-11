# Pull Request Comment Create Tool - Spec Lite

> Feature: Add `pr_comment_create` MCP tool
> Created: 2025-09-12

## Overview
Add MCP tool to create comments on Forgejo/Gitea pull requests, following existing patterns with validation only in server handler.

## Tool Details
- **Name**: `pr_comment_create`
- **Purpose**: Create comments on pull requests
- **Parameters**: repository (string), pull_request_number (int), comment (string)
- **Validation**: Server handler only using ozzo-validation
- **Architecture**: Interface → Client → Service → Server (no validation duplication)

## Implementation Plan
1. **Interface**: Add `PullRequestCommenter` interface and args struct
2. **Client**: Implement `CreatePullRequestComment` using Gitea SDK (no validation)
3. **Service**: Add method that calls client directly (no validation)
4. **Server**: Add handler with ozzo-validation and tool registration
5. **Tests**: Comprehensive unit, integration, and acceptance tests
6. **Docs**: Update README with usage examples

## Success Criteria
- ✅ Tool successfully creates PR comments
- ✅ Validation only in server handler
- ✅ No validation duplication in service layer
- ✅ All tests pass, no regressions
- ✅ Documentation complete

## Files to Modify
- `remote/gitea/interface.go`
- `remote/gitea/gitea_client.go`
- `remote/gitea/service.go`
- `server/pr_comments.go`
- `server/server.go`
- `server_test/` (multiple test files)
- `README.md`

## Status: Planning Complete ✅