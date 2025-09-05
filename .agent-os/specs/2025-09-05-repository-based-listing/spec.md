# Spec Requirements Document

> Spec: repository-based-listing
> Created: 2025-09-05
> Status: Planning

## Overview

Modify the existing list PR and list issue functionality to operate on a repository basis instead of a user basis. This change will allow users to retrieve pull requests and issues for a specific repository rather than for a specific user, providing more targeted and useful data retrieval capabilities.

## User Stories

- As an agent, I want the option to identify the working directory so that I don't need to identify the correct owner and repository name.
- As a developer, I want to list all pull requests for a specific repository so that I can see the current work in progress for that project
- As a project maintainer, I want to list all issues for a specific repository so that I can track bugs and feature requests for that codebase
- As an automation tool, I want to query repository-specific PRs and issues so that I can perform repository-level analysis and reporting

## Spec Scope

- Modify existing list PR API endpoint to accept repository or cwd identifier instead of user identifier
- Modify existing list issue API endpoint to accept repository or cwd identifier instead of user identifier
- Update request/response models to use repository-based parameters
- Remove all deprecated logic, no backward compatibility is necessary
- Update API documentation to reflect the new repository-based approach

## Out of Scope

- User-based listing functionality (deprecated)
- Repository creation or modification capabilities
- Authentication and authorization changes
- UI/frontend changes for displaying the data

## Expected Deliverable

A modified MCP server implementation that supports repository-based listing of pull requests and issues, with updated API endpoints that accept repository identifiers and return repository-specific data. The implementation should include proper error handling, validation, and documentation updates.

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-05-repository-based-listing/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-05-repository-based-listing/sub-specs/technical-spec.md
