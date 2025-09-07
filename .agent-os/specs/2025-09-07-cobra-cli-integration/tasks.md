# Spec Tasks

## Tasks

- [x] 1. Setup Dependencies and Project Structure
  - [x] 1.1 Write tests for dependency management and module structure
  - [x] 1.2 Add github.com/spf13/cobra to go.mod
  - [x] 1.3 Run go mod tidy to update dependencies
  - [x] 1.4 Create cmd/ directory structure
  - [x] 1.5 Verify all tests pass

- [ ] 2. Implement Root Command Infrastructure
  - [ ] 2.1 Write tests for root command initialization and global flags
  - [ ] 2.2 Create cmd/root.go with cobra root command setup
  - [ ] 2.3 Implement global flags (--config, --verbose)
  - [ ] 2.4 Add persistent pre-run hooks for config loading
  - [ ] 2.5 Verify all tests pass

- [ ] 3. Implement Serve Command
  - [ ] 3.1 Write tests for serve command and server lifecycle
  - [ ] 3.2 Create cmd/serve.go with server startup logic
  - [ ] 3.3 Move existing server initialization from main.go
  - [ ] 3.4 Add serve-specific flags (--host, --port)
  - [ ] 3.5 Implement graceful server shutdown handling
  - [ ] 3.6 Verify all tests pass

- [ ] 4. Implement Version Command
  - [ ] 4.1 Write tests for version command output
  - [ ] 4.2 Create cmd/version.go with version display logic
  - [ ] 4.3 Add build information and Go version details
  - [ ] 4.4 Implement version flag parsing and formatting
  - [ ] 4.5 Verify all tests pass

- [ ] 5. Implement Config Command
  - [ ] 5.1 Write tests for config validation and connectivity
  - [ ] 5.2 Create cmd/config.go with configuration validation
  - [ ] 5.3 Implement Forgejo instance connectivity testing
  - [ ] 5.4 Add configuration display and error reporting
  - [ ] 5.5 Verify all tests pass

- [ ] 6. Refactor Main Entry Point
  - [ ] 6.1 Write tests for main.go cobra integration
  - [ ] 6.2 Update main.go to use cobra command execution
  - [ ] 6.3 Ensure backward compatibility with existing startup
  - [ ] 6.4 Remove old server startup code from main.go
  - [ ] 6.5 Verify all tests pass

- [ ] 7. Integration Testing and Validation
  - [ ] 7.1 Write integration tests for CLI command execution
  - [ ] 7.2 Test backward compatibility with go run main.go
  - [ ] 7.3 Validate all CLI commands work correctly
  - [ ] 7.4 Test configuration validation and error handling
  - [ ] 7.5 Run full test suite and verify all tests pass

- [ ] 8. Documentation and Final Verification
  - [ ] 8.1 Update README.md with new CLI usage examples
  - [ ] 8.2 Test help system and command documentation
  - [ ] 8.3 Verify build process works with new structure
  - [ ] 8.4 Perform final integration testing
  - [ ] 8.5 Verify all tests pass