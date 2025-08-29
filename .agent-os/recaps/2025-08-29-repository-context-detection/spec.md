# Spec Requirements Document

> Spec: Repository Context Detection
> Created: 2025-08-29
> Status: Planning

## Overview

Implement automatic repository context detection to identify the current Forgejo repository from the local git environment. This feature enables AI agents to automatically determine which repository they're working with without manual configuration.

## User Stories

### Automatic Repository Detection

As an AI agent, I want to automatically detect the current Forgejo repository from the local git environment, so that I can perform repository operations without manual repository specification.

The system will:
1. Check if the current directory is a git repository
2. Extract the remote URL from the git configuration
3. Validate that the remote is a Forgejo instance
4. Parse the owner and repository name from the URL
5. Cache the context for performance

### Context Validation

As a developer using the MCP server, I want clear error messages when repository context cannot be detected, so that I can troubleshoot configuration issues.

Error scenarios include:
- Not in a git repository
- No remote configured
- Remote is not a Forgejo instance
- Invalid repository URL format

## Spec Scope

1. **Git Repository Detection** - Validate current directory is a git repository and extract remote URLs
2. **Forgejo Remote Validation** - Verify remote URLs point to valid Forgejo instances
3. **Context Manager** - Provide unified interface for repository context with caching

## Out of Scope

- Authentication validation (handled separately)
- Multiple remote support (focus on default remote)
- Repository existence validation (handled by Gitea client)
- Custom remote naming conventions

## Expected Deliverable

1. Context detection works automatically in git repositories with Forgejo remotes
2. Clear error messages for unsupported scenarios
3. Cached context for performance optimization
4. Comprehensive test coverage for all detection scenarios

## Spec Documentation

- Tasks: @.agent-os/recaps/2025-08-29-repository-context-detection/tasks.md
- Technical Specification: @.agent-os/recaps/2025-08-29-repository-context-detection/sub-specs/technical-spec.md