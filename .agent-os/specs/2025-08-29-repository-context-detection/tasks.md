# Spec Tasks

## Tasks

- [ ] 1. Git Repository Detection
  - [ ] 1.1 Write tests for git repository detection functions (including worktrees)
  - [ ] 1.2 Implement IsGitRepository function with .git directory and worktree support
  - [ ] 1.3 Implement GetRemoteURL function for remote URL extraction
  - [ ] 1.4 Add error handling for git command failures
  - [ ] 1.5 Verify all tests pass

- [ ] 2. Forgejo Remote Validation
  - [ ] 2.1 Write tests for Forgejo remote validation functions
  - [ ] 2.2 Implement IsForgejoRemote function for URL validation
  - [ ] 2.3 Implement ParseRepository function for owner/repo extraction
  - [ ] 2.4 Add support for SSH and HTTPS URL formats
  - [ ] 2.5 Verify all tests pass

- [ ] 3. Context Manager
  - [ ] 3.1 Write tests for Context struct and manager functions
  - [ ] 3.2 Define Context struct with Owner, Repository, RemoteURL fields
  - [ ] 3.3 Implement DetectContext function as main entry point
  - [ ] 3.4 Add in-memory caching with TTL for performance
  - [ ] 3.5 Implement thread-safe operations with proper locking
  - [ ] 3.6 Verify all tests pass

- [ ] 4. Integration and Testing
  - [ ] 4.1 Write integration tests for complete context detection flow
  - [ ] 4.2 Test error scenarios and edge cases
  - [ ] 4.3 Add comprehensive documentation and examples
  - [ ] 4.4 Update main.go or server integration if needed
  - [ ] 4.5 Verify all tests pass with >80% coverage