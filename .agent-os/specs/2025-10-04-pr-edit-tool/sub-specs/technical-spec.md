# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-04-pr-edit-tool/spec.md

> Created: 2025-10-04
> Version: 1.0.0

## Technical Requirements

- **Remote Interface Extension**: Add EditPullRequestArgs struct and PullRequestEditor interface to remote/interface.go
- **Forgejo Client Implementation**: Implement EditPullRequest method in remote/forgejo/pull_requests.go using Forgejo SDK
- **Gitea Client Implementation**: Implement EditPullRequest method in remote/gitea/gitea_client.go using Gitea SDK
- **Server Handler**: Create new server/pr_edit.go with handlePullRequestEdit function and MCP tool registration
- **Input Validation**: Use ozzo-validation for comprehensive parameter validation
- **Repository Resolution**: Support both explicit repository specification and directory-based auto-resolution
- **Error Handling**: Structured error responses with proper error wrapping and context validation
- **Response Format**: Consistent success/error response patterns matching existing tools

## Tool Parameters

- **repository** (string, optional): "owner/repo" format specification
- **directory** (string, optional): Local directory for auto-resolution
- **pull_request_number** (int, required): PR number to edit
- **title** (string, optional): New PR title
- **body** (string, optional): New PR description
- **state** (string, optional): New state ("open", "closed")
- **base_branch** (string, optional): New base branch

## External Dependencies (Conditional)

No new external dependencies required. Uses existing Forgejo/Gitea SDKs, MCP Go SDK, and ozzo-validation library already present in the codebase.