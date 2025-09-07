# Spec Tasks

## Tasks

- [x] 1. Setup Dependencies and Project Structure
  - [x] 1.1 Write tests for dependency management and module structure
  - [x] 1.2 Add github.com/spf13/cobra to go.mod
  - [x] 1.3 Run go mod tidy to update dependencies
  - [x] 1.4 Create cmd/ directory structure
  - [x] 1.5 Verify all tests pass

- [x] 2. Implement Root Command Infrastructure
   - [x] 2.1 Write tests for root command initialization and global flags
   - [x] 2.2 Create cmd/root.go with cobra root command setup
   - [x] 2.3 Implement global flags (--config, --verbose)
   - [x] 2.4 Add persistent pre-run hooks for config loading
   - [x] 2.5 Verify all tests pass

- [x] 3. Implement Serve Command
   - [x] 3.1 Write tests for serve command and server lifecycle
   - [x] 3.2 Create cmd/serve.go with server startup logic
   - [x] 3.3 Move existing server initialization from main.go
   - [x] 3.4 Add serve-specific flags (--host, --port)
   - [x] 3.5 Implement graceful server shutdown handling
   - [x] 3.6 Verify all tests pass

- [x] 4. Implement Version Command
   - [x] 4.1 Write tests for version command output
   - [x] 4.2 Create cmd/version.go with version display logic
   - [x] 4.3 Add build information and Go version details
   - [x] 4.4 Implement version flag parsing and formatting
   - [x] 4.5 Verify all tests pass

- [x] 5. Implement Config Command
   - [x] 5.1 Write tests for config validation and connectivity
   - [x] 5.2 Create cmd/config.go with configuration validation
   - [x] 5.3 Implement Forgejo instance connectivity testing
   - [x] 5.4 Add configuration display and error reporting
   - [x] 5.5 Verify all tests pass

- [x] 6. Refactor Main Entry Point
   - [x] 6.1 Write tests for main.go cobra integration
   - [x] 6.2 Update main.go to use cobra command execution
   - [x] 6.3 Ensure backward compatibility with existing startup
   - [x] 6.4 Remove old server startup code from main.go
   - [x] 6.5 Verify all tests pass

- [x] 7. Integration Testing and Validation
   - [x] 7.1 Write integration tests for CLI command execution
   - [x] 7.2 Test backward compatibility with go run main.go
   - [x] 7.3 Validate all CLI commands work correctly
   - [x] 7.4 Test configuration validation and error handling
   - [x] 7.5 Run full test suite and verify all tests pass

- [x] 8. Documentation and Final Verification
  - [x] 8.1 Update README.md with new CLI usage examples
  - [x] 8.2 Test help system and command documentation
  - [x] 8.3 Verify build process works with new structure
  - [x] 8.4 Perform final integration testing
  - [x] 8.5 Verify all tests pass