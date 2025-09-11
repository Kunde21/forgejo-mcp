# Spec Requirements Document

> Spec: pr-list
> Created: 2025-09-11
> Status: Planning

## Overview

This specification outlines the implementation of a new MCP tool for listing pull requests (PRs) in Forgejo/Gitea repositories. The tool will provide AI agents with the ability to retrieve PR information, following the same architectural patterns as the existing issue listing functionality. The feature will support filtering by PR state (open/closed/all) and include pagination capabilities.

## User Stories

- As an AI agent, I want to list pull requests in a repository so I can understand the current state of code review and merge activities
- As an AI agent, I want to filter pull requests by state (open/closed/all) so I can focus on specific types of PRs relevant to my task
- As an AI agent, I want to paginate through pull request results so I can handle repositories with many PRs efficiently
- As a developer, I want the PR list tool to follow the same patterns as existing tools so the codebase remains consistent and maintainable
- As a system administrator, I want proper validation and error handling so the tool fails gracefully with invalid inputs

## Spec Scope

- Implementation of a new MCP tool named "pr_list"
- Integration with the existing Forgejo client interface and service layers
- Support for repository parameter (required) to specify target repository
- Support for limit parameter (optional, default 15) for pagination
- Support for offset parameter (optional, default 0) for pagination
- Support for state parameter (optional, default "open", values: "open", "closed", "all")
- Validation using ozzo-validation library following existing patterns
- Proper error handling and response formatting
- Integration with existing MCP server registration patterns
- Unit tests following the project's testing standards
- Integration tests to verify end-to-end functionality

## Out of Scope

- Pull request creation, editing, or merging functionality
- Pull request comment listing or management
- Pull request diff viewing or code review features
- Pull request approval workflow management
- Repository management or configuration changes
- Authentication or authorization enhancements
- UI components or web interface changes
- Database schema modifications
- Changes to existing issue listing functionality

## Expected Deliverable

- Complete implementation of the pr_list tool with all specified functionality
- Updated MCP server configuration to register the new tool
- Comprehensive unit tests covering all validation scenarios and edge cases
- Integration tests verifying the tool works correctly with actual Forgejo instances
- Documentation updates following the project's documentation standards
- Code that follows the existing architectural patterns and coding standards

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-11-pr-list/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-11-pr-list/sub-specs/technical-spec.md