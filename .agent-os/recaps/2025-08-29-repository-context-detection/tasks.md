# Spec Tasks

## Tasks

- [x] 1. Git Repository Detection
  - [x] 1.1 Write tests for git repository detection functions (including worktrees)
  - [x] 1.2 Implement IsGitRepository function with .git directory and worktree support
  - [x] 1.3 Implement GetRemoteURL function for remote URL extraction
  - [x] 1.4 Add error handling for git command failures
  - [x] 1.5 Verify all tests pass

- [x] 2. Forgejo Remote Validation
  - [x] 2.1 Write tests for Forgejo remote validation functions
  - [x] 2.2 Implement IsForgejoRemote function for URL validation
  - [x] 2.3 Implement ParseRepository function for owner/repo extraction
  - [x] 2.4 Add support for SSH and HTTPS URL formats
  - [x] 2.5 Verify all tests pass

- [x] 3. Context Manager
  - [x] 3.1 Write tests for Context struct and manager functions
  - [x] 3.2 Define Context struct with Owner, Repository, RemoteURL fields
  - [x] 3.3 Implement DetectContext function as main entry point
  - [x] 3.4 Add in-memory caching with TTL for performance
  - [x] 3.5 Implement thread-safe operations with proper locking
  - [x] 3.6 Verify all tests pass

- [x] 4. Integration and Testing
  - [x] 4.1 Write integration tests for complete context detection flow
  - [x] 4.2 Test error scenarios and edge cases
  - [x] 4.3 Add comprehensive documentation and examples
  - [x] 4.4 Update main.go or server integration if needed
  - [x] 4.5 Verify all tests pass with >80% coverage