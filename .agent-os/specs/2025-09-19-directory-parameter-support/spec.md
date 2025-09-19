# Spec Requirements Document

> Spec: directory-parameter-support
> Created: 2025-09-19

## Overview

Add a consistent directory parameter to all server tools that allows users to specify which directory/repository they want to work with, similar to how other MCP tools work, while maintaining backward compatibility with existing repository parameter usage.

## User Stories

### Directory-Based Repository Operations

As a forgejo-mcp user, I want to specify a local directory path when working with repository tools, so that I can work with git repositories directly from my file system without needing to manually extract owner/repo information.

**Detailed Workflow:** Users provide an absolute directory path containing a git repository. The server automatically detects the git repository, extracts remote information, and resolves it to the appropriate owner/repo format for Forgejo API calls. This eliminates the need for users to manually parse git remote URLs or remember repository naming conventions.

### Consistent Parameter Interface

As an MCP tool user, I want all forgejo-mcp tools to use the same directory parameter convention, so that I have a consistent and predictable interface across all repository operations that aligns with standard MCP tool patterns.

**Detailed Workflow:** All tools (issue_list, pr_list, issue_comment_create, etc.) accept the same directory parameter with identical validation rules and behavior. Users can learn the parameter pattern once and apply it consistently across all tools, reducing cognitive load and improving usability.

### Smooth Migration Path

As an existing forgejo-mcp user, I want to continue using the repository parameter while gradually transitioning to the new directory parameter, so that my existing workflows continue to work without interruption during the migration period.

**Detailed Workflow:** Existing users can continue using the repository parameter with deprecation warnings, while new users adopt the directory parameter. The server supports both parameters simultaneously, with clear documentation guiding users through the migration process and timeline for eventual repository parameter deprecation.

## Spec Scope

1. **Directory Parameter Implementation** - Add a consistent `directory` parameter to all 9 server tools that accepts absolute file system paths to git repositories.
2. **Repository Resolution Logic** - Implement automatic detection of git repositories and extraction of remote owner/repo information from directory paths.
3. **Backward Compatibility** - Maintain support for existing `repository` parameter with deprecation warnings and mutual exclusivity validation.
4. **Parameter Validation** - Add comprehensive validation for directory paths including existence checks, git repository detection, and remote information extraction.
5. **Testing Infrastructure** - Update existing test suite to cover directory parameter functionality, backward compatibility, and error scenarios.

## Out of Scope

- Complete removal of repository parameter (this will be done in a future major version)
- Support for non-git directories or other version control systems
- Automatic repository cloning or git operations beyond remote information extraction
- Changes to the MCP protocol or server architecture beyond parameter additions
- User interface changes or visual tooling improvements

## Expected Deliverable

1. All 9 forgejo-mcp server tools support a consistent `directory` parameter that accepts absolute file system paths to git repositories and automatically resolves them to owner/repo format for API calls.
2. Comprehensive test coverage for directory parameter functionality including validation, resolution logic, error handling, and backward compatibility scenarios.
3. Updated documentation and examples showing how to use the directory parameter, migration guidance from repository parameter, and best practices for directory-based workflows.