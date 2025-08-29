# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-08-29-repository-context-detection/spec.md

> Created: 2025-08-29
> Version: 1.0.0

## Technical Requirements

### Git Repository Detection
- Use Go's `os/exec` package to execute `git` commands
- Implement `IsGitRepository(path string) bool` function
- Check for `.git` directory existence as primary validation
- Support git worktrees by checking for `.git` file pointing to worktree structure
- Fallback to `git rev-parse --git-dir` for edge cases and worktree detection
- Handle both absolute and relative path inputs
- Validate worktree configuration and main repository relationship

### Remote URL Extraction
- Execute `git remote get-url origin` to get default remote
- Support both SSH (`git@host:user/repo.git`) and HTTPS (`https://host/user/repo`) formats
- Implement `GetRemoteURL(name string) (string, error)` function
- Handle cases where remote doesn't exist
- Return descriptive errors for git command failures

### Forgejo Remote Validation
- Implement `IsForgejoRemote(url string) bool` function
- Check for known Forgejo host patterns
- Support custom Forgejo instances beyond codeberg.org
- Validate URL format and accessibility

### Repository Information Parsing
- Implement `ParseRepository(url string) (owner, repo string, err error)` function
- Extract owner and repository name from various URL formats
- Handle SSH and HTTPS URL variations
- Strip `.git` suffix when present
- Validate extracted information format

### Context Manager Implementation
- Define `Context` struct with Owner, Repository, RemoteURL fields
- Implement `DetectContext(path string) (*Context, error)` as main entry point
- Add in-memory caching with TTL for performance
- Thread-safe implementation with proper locking
- Cache key based on working directory path

### Error Handling
- Custom error types for different failure scenarios
- Descriptive error messages for troubleshooting
- Proper error wrapping with context
- Graceful degradation when git is not available

### Performance Considerations
- Cache context results to avoid repeated git operations
- Implement cache invalidation on directory changes
- Limit cache size to prevent memory leaks
- Optimize git command execution

## Approach

[APPROACH_CONTENT]

## External Dependencies

No new external dependencies required. This feature uses:
- Go standard library (`os`, `exec`, `path/filepath`, `strings`)
- Existing project dependencies for error handling and logging