# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-09-07-list-issues-tool/spec.md

> Created: 2025-09-07
> Status: Ready for Implementation

## Tasks

- [ ] 1. Extend Configuration Management
  - [ ] 1.1 Write tests for configuration loading with new fields
  - [ ] 1.2 Add RemoteURL and AuthToken fields to Config struct
  - [ ] 1.3 Update LoadConfig() to read FORGEJO_REMOTE_URL and FORGEJO_AUTH_TOKEN
  - [ ] 1.4 Add configuration validation for required API fields
  - [ ] 1.5 Verify configuration tests pass

- [ ] 2. Add Gitea SDK Dependency
  - [ ] 2.1 Write tests for dependency resolution
  - [ ] 2.2 Add code.gitea.io/sdk/gitea to go.mod
  - [ ] 2.3 Run go mod tidy to resolve dependencies
  - [ ] 2.4 Verify dependency tests pass

- [ ] 3. Create Remote Package Structure
  - [ ] 3.1 Write tests for IssueLister interface
  - [ ] 3.2 Create remote/gitea/interface.go with IssueLister interface
  - [ ] 3.3 Create remote/gitea/gitea_client.go with SDK client implementation
  - [ ] 3.4 Create remote/gitea/service.go with business logic
  - [ ] 3.5 Implement dependency injection in server package
  - [ ] 3.6 Verify remote package tests pass

- [ ] 4. Implement MCP Tool
  - [ ] 4.1 Write tests for list_issues tool handler
  - [ ] 4.2 Add list_issues tool registration in NewServer()
  - [ ] 4.3 Implement handleListIssues() method with input validation
  - [ ] 4.4 Add pagination parameter handling (limit, offset)
  - [ ] 4.5 Implement error handling and response formatting
  - [ ] 4.6 Verify tool implementation tests pass

- [ ] 5. Enhance Test Harness
  - [ ] 5.1 Write tests for mock Gitea server functionality
  - [ ] 5.2 Add MockGiteaServer struct to harness.go
  - [ ] 5.3 Implement mock API endpoints for issues listing
  - [ ] 5.4 Add configuration injection for test scenarios
  - [ ] 5.5 Support both real and mock API responses
  - [ ] 5.6 Verify test harness tests pass

- [ ] 6. Add Acceptance Tests
  - [ ] 6.1 Write acceptance tests for successful issue listing
  - [ ] 6.2 Add tests for pagination parameters
  - [ ] 6.3 Implement error handling test scenarios
  - [ ] 6.4 Add input validation tests
  - [ ] 6.5 Test concurrent request handling
  - [ ] 6.6 Verify acceptance tests pass

- [ ] 7. Update Documentation
  - [ ] 7.1 Write tests for documentation completeness
  - [ ] 7.2 Add list_issues tool to README.md Tools List
  - [ ] 7.3 Document required configuration variables
  - [ ] 7.4 Add usage examples and parameters
  - [ ] 7.5 Update configuration section
  - [ ] 7.6 Verify documentation tests pass
