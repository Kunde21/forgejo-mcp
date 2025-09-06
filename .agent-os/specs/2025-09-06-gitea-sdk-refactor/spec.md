# Spec Requirements Document

> Spec: Gitea SDK Refactor
> Created: 2025-09-06
> Status: Planning

## Overview

Separate MCP handlers from Gitea SDK implementation by moving Gitea-specific code to a dedicated `remote/gitea` package while keeping MCP handlers and input validation in the `server` package. This refactor addresses current structural issues where server/sdk_handlers.go contains both MCP handler logic and Gitea SDK implementation details, creating tight coupling and reduced maintainability.

## User Stories

As a developer working on the MCP server:
- I want clear separation between MCP protocol handling and Gitea SDK implementation so that I can modify one without affecting the other
- I want to easily test Gitea SDK functionality independently of MCP handlers so that I can write focused unit tests
- I want to add support for other Git services (GitHub, GitLab) following the same pattern so that the codebase remains extensible
- I want to reuse Gitea client functionality in other parts of the application so that I don't duplicate code
- I want to maintain all existing MCP functionality after the refactor so that the changes are transparent to users

## Spec Scope

1. Create new `remote/gitea` package structure with dedicated files for client, errors, repository operations, pull requests, issues, and Git utilities
2. Move GiteaClientInterface and SDKError from server package to remote/gitea/client.go
3. Move repository resolution logic (resolveCWDToRepository, parseGitRemoteOutput, resolveCWDFromPath) to remote/gitea/git.go
4. Move repository metadata extraction to remote/gitea/repository.go
5. Restructure MCP handlers to use dependency injection and call remote/gitea package methods
6. Split server/sdk_handlers.go into server/handlers.go, server/validation.go, and server/types.go
7. Update all imports and ensure no circular dependencies
8. Move and update test files to match new package structure
9. Preserve all existing MCP functionality and API compatibility

## Out of Scope

- Changes to MCP protocol implementation or message formats
- Modifications to configuration handling or CLI interface
- Updates to external dependencies or Gitea SDK version
- Performance optimizations or new features
- Documentation updates beyond what's needed for the refactor
- Changes to build or deployment processes

## Expected Deliverable

Clean package separation with preserved functionality:
- `server/` package contains only MCP handlers, input validation, and protocol-specific logic
- `remote/gitea/` package contains all Gitea SDK implementation details and client operations
- All existing MCP functionality remains intact and fully compatible
- Clear interfaces between packages enable independent testing and future extensibility
- No circular dependencies or tight coupling between MCP and Gitea implementation
- Comprehensive test coverage maintained across the new package structure

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-06-gitea-sdk-refactor/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-06-gitea-sdk-refactor/sub-specs/technical-spec.md