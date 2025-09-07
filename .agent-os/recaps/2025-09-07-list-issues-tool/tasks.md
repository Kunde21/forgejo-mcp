# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-09-07-list-issues-tool/spec.md

> Created: 2025-09-07
> Status: Completed

## Tasks

- [x] 1. Extend Configuration Management
   - [x] 1.1 Write tests for configuration loading with new fields
   - [x] 1.2 Add RemoteURL and AuthToken fields to Config struct
   - [x] 1.3 Update LoadConfig() to read FORGEJO_REMOTE_URL and FORGEJO_AUTH_TOKEN
   - [x] 1.4 Add configuration validation for required API fields
   - [x] 1.5 Verify configuration tests pass

- [x] 2. Add Gitea SDK Dependency
   - [x] 2.1 Write tests for dependency resolution
   - [x] 2.2 Add code.gitea.io/sdk/gitea to go.mod
   - [x] 2.3 Run go mod tidy to resolve dependencies
   - [x] 2.4 Verify dependency tests pass

- [x] 3. Create Remote Package Structure
   - [x] 3.1 Write tests for IssueLister interface
   - [x] 3.2 Create remote/gitea/interface.go with IssueLister interface
   - [x] 3.3 Create remote/gitea/gitea_client.go with SDK client implementation
   - [x] 3.4 Create remote/gitea/service.go with business logic
   - [x] 3.5 Implement dependency injection in server package
   - [x] 3.6 Verify remote package tests pass

- [x] 4. Implement MCP Tool
   - [x] 4.1 Write tests for list_issues tool handler
   - [x] 4.2 Add list_issues tool registration in NewServer()
   - [x] 4.3 Implement handleListIssues() method with input validation
   - [x] 4.4 Add pagination parameter handling (limit, offset)
   - [x] 4.5 Implement error handling and response formatting
   - [x] 4.6 Verify tool implementation tests pass

- [x] 5. Enhance Test Harness
  - [x] 5.1 Write tests for mock Gitea server functionality
  - [x] 5.2 Add MockGiteaServer struct to harness.go
  - [x] 5.3 Implement mock API endpoints for issues listing
  - [x] 5.4 Add configuration injection for test scenarios
  - [x] 5.5 Support both real and mock API responses
  - [x] 5.6 Verify test harness tests pass

- [x] 6. Add Acceptance Tests
  - [x] 6.1 Write acceptance tests for successful issue listing
  - [x] 6.2 Add tests for pagination parameters
  - [x] 6.3 Implement error handling test scenarios
  - [x] 6.4 Add input validation tests
  - [x] 6.5 Test concurrent request handling
  - [x] 6.6 Verify acceptance tests pass

- [x] 7. Update Documentation
  - [x] 7.1 Write tests for documentation completeness
  - [x] 7.2 Add list_issues tool to README.md Tools List
  - [x] 7.3 Document required configuration variables
  - [x] 7.4 Add usage examples and parameters
  - [x] 7.5 Update configuration section
  - [x] 7.6 Verify documentation tests pass
