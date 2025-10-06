---
type: feature
priority: medium
created: 2025-10-05T00:00:00Z
created_by: Opus
status: reviewed
tags: [issue, create, forgejo, gitea, mcp]
keywords: [issue, create, forgejo, gitea, attachment, validation, directory, repository]
patterns: [tool implementation, parameter validation, file upload, error handling, directory resolution]
---

# FEATURE-003: Add tool to create issues on repository

## Description
Create a new MCP tool that allows users to create issues on Forgejo/Gitea repositories through an Agent interface. The tool should support creating issues with titles, bodies, and attachments (images and PDFs), following the existing patterns in the repository.

## Context
Developers working on the repository need the ability to quickly create issues when they identify problems or feature requests. This tool integrates with the existing Forgejo/Gitea MCP server to provide seamless issue creation capabilities.

## Requirements
- Create issues on Forgejo/Gitea repositories with correct author, title, and body
- Support directory or repository parameter (matching existing tools)
- Handle attachments (images and PDFs only)
- Validate input before acting on remote
- Report errors back to user for resolution
- Use markdown formatting for issue body
- Follow existing tool patterns in the repository

### Functional Requirements
- [ ] Create issue with title and body
- [ ] Support directory parameter for repository resolution
- [ ] Upload attachments (images and PDFs)
- [ ] Validate attachment MIME types
- [ ] Enforce Forgejo/Gitea attachment size limits
- [ ] Return created issue details
- [ ] Handle authentication via existing token system

### Non-Functional Requirements
- [ ] Follow existing code patterns and conventions
- [ ] Integrate with existing error handling
- [ ] Support both Forgejo and Gitea backends
- [ ] Validate all inputs before remote operations

## Current State
Repository has existing tools for listing issues, creating/editing comments, and managing pull requests. No tool exists for creating new issues.

## Desired State
New tool `forgejo_issue_create` that can create issues with attachments following the established patterns.

## Research Context

### Keywords to Search
- issue - Existing issue-related functionality
- create - Creation patterns in existing tools
- forgejo - Forgejo client implementation
- gitea - Gitea client implementation
- attachment - File upload patterns
- validation - Input validation patterns
- directory - Directory parameter handling
- repository - Repository resolution logic

### Patterns to Investigate
- tool implementation - How existing MCP tools are structured
- parameter validation - Input validation patterns
- file upload - How attachments are handled in existing code
- error handling - Error reporting and user feedback
- directory resolution - How directory parameters are processed

### Key Decisions Made
- Use existing authentication system
- Follow established tool naming conventions
- Support only images and PDFs for attachments
- Match directory parameter handling from existing tools
- Use markdown for issue body formatting
- No issue templates or duplicate detection

## Success Criteria

### Automated Verification
- [ ] Tool creates issues successfully
- [ ] Attachment upload works for valid files
- [ ] Invalid MIME types are rejected
- [ ] Oversized files are rejected
- [ ] Directory resolution works correctly
- [ ] Error conditions are properly reported

### Manual Verification
- [ ] Issue appears on repository with correct content
- [ ] Attachments are properly attached
- [ ] Author is correctly set
- [ ] Tool integrates seamlessly with existing tools

## Related Information
- Existing issue listing tools
- Comment creation/editing tools
- Pull request management tools
- Repository resolver implementation

## Notes
- Need to research Forgejo/Gitea default attachment size limits
- Should investigate how existing tools handle directory parameters
- Need to understand current authentication flow for private repos