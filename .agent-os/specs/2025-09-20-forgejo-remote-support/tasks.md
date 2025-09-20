# Spec Tasks

- [x] 1. Add Forgejo SDK Dependency and Integration
  - [x] 1.1 Write tests for dependency integration
  - [x] 1.2 Add Forgejo SDK to go.mod dependencies
  - [x] 1.3 Run go mod tidy to resolve dependencies
  - [x] 1.4 Verify all existing tests still pass with new dependency
  - [x] 1.5 Verify all tests pass

- [ ] 2. Implement ForgejoClient Structure
  - [x] 2.1 Write tests for ForgejoClient basic structure
  - [x] 2.2 Create remote/forgejo directory structure
  - [x] 2.3 Implement ForgejoClient struct and constructor
  - [x] 2.4 Add necessary imports and basic error handling
  - [x] 2.5 Verify all tests pass

- [x] 3. Implement ForgejoClient Issue Management Methods
   - [x] 3.1 Write tests for ListIssues method
   - [x] 3.2 Implement ListIssues method with Forgejo SDK
   - [x] 3.3 Write tests for CreateIssueComment method
   - [x] 3.4 Implement CreateIssueComment method
   - [x] 3.5 Write tests for ListIssueComments method
   - [x] 3.6 Implement ListIssueComments method
   - [x] 3.7 Write tests for EditIssueComment method
   - [x] 3.8 Implement EditIssueComment method
   - [x] 3.9 Verify all tests pass

- [x] 4. Implement ForgejoClient Pull Request Methods
   - [x] 4.1 Write tests for ListPullRequests method
   - [x] 4.2 Implement ListPullRequests method
   - [x] 4.3 Write tests for ListPullRequestComments method
   - [x] 4.4 Implement ListPullRequestComments method
   - [x] 4.5 Write tests for CreatePullRequestComment method
   - [x] 4.6 Implement CreatePullRequestComment method
   - [x] 4.7 Write tests for EditPullRequestComment method
   - [x] 4.8 Implement EditPullRequestComment method
   - [x] 4.9 Verify all tests pass

- [ ] 5. Extend Configuration System
  - [ ] 5.1 Write tests for configuration extension
  - [ ] 5.2 Add ClientType field to Config struct
  - [ ] 5.3 Implement validation for "gitea", "forgejo", "auto" values
  - [ ] 5.4 Update environment variable handling
  - [ ] 5.5 Verify all tests pass

- [ ] 6. Implement Automatic Remote Type Detection
  - [ ] 6.1 Write tests for version detection functionality
  - [ ] 6.2 Implement detectRemoteType function
  - [ ] 6.3 Add HTTP client for /api/v1/version endpoint
  - [ ] 6.4 Implement version string parsing logic
  - [ ] 6.5 Add error handling and fallback strategies
  - [ ] 6.6 Verify all tests pass

- [ ] 7. Update Client Factory Logic
  - [ ] 7.1 Write tests for client factory with auto-detection
  - [ ] 7.2 Modify NewFromConfig to support client type selection
  - [ ] 7.3 Integrate automatic detection for "auto" client type
  - [ ] 7.4 Add error handling for invalid client types
  - [ ] 7.5 Ensure backward compatibility with existing behavior
  - [ ] 7.6 Verify all tests pass

- [ ] 8. Create Comprehensive Test Suite
  - [ ] 8.1 Write integration tests for ForgejoClient
  - [ ] 8.2 Create tests for version detection with real API responses
  - [ ] 8.3 Test all three client types (gitea, forgejo, auto)
  - [ ] 8.4 Test error scenarios and edge cases
  - [ ] 8.5 Verify backward compatibility with existing tests
  - [ ] 8.6 Verify all tests pass

- [ ] 9. Update Documentation and Examples
  - [ ] 9.1 Update README.md with Forgejo support documentation
  - [ ] 9.2 Update config.example.yaml with ClientType examples
  - [ ] 9.3 Create migration guide for Gitea to Forgejo
  - [ ] 9.4 Document automatic detection behavior
  - [ ] 9.5 Update AGENTS.md with new build/test commands
  - [ ] 9.6 Verify all documentation is accurate

- [ ] 10. Final Validation and Deployment
  - [ ] 10.1 Run complete test suite (unit, integration, existing)
  - [ ] 10.2 Test end-to-end workflow with all client types
  - [ ] 10.3 Verify identical functionality between Gitea and Forgejo clients
  - [ ] 10.4 Test performance and memory usage
  - [ ] 10.5 Validate code quality and style compliance
  - [ ] 10.6 Verify all tests pass