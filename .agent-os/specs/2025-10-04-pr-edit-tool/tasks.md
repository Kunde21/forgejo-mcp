# Spec Tasks

## Tasks

- [ ] 1. Extend Remote Interface and Types
  - [ ] 1.1 Write tests for EditPullRequestArgs struct and PullRequestEditor interface
  - [ ] 1.2 Add EditPullRequestArgs struct to remote/interface.go
  - [ ] 1.3 Add PullRequestEditor interface to remote/interface.go
  - [ ] 1.4 Update ClientInterface to include PullRequestEditor
  - [ ] 1.5 Verify all interface tests pass

- [ ] 2. Implement Forgejo Client EditPullRequest Method
  - [ ] 2.1 Write tests for EditPullRequest method in remote/forgejo/forgejo_client_test.go
  - [ ] 2.2 Implement EditPullRequest method in remote/forgejo/pull_requests.go
  - [ ] 2.3 Add validation logic for repository format and PR number
  - [ ] 2.4 Add error handling for API failures and not found cases
  - [ ] 2.5 Verify Forgejo client tests pass

- [ ] 3. Implement Gitea Client EditPullRequest Method
  - [ ] 3.1 Write tests for EditPullRequest method in remote/gitea/client_test.go
  - [ ] 3.2 Implement EditPullRequest method in remote/gitea/gitea_client.go
  - [ ] 3.3 Ensure consistent behavior with Forgejo implementation
  - [ ] 3.4 Add validation and error handling matching Forgejo patterns
  - [ ] 3.5 Verify Gitea client tests pass

- [ ] 4. Create Server Handler for PR Edit Tool
  - [ ] 4.1 Write tests for handlePullRequestEdit in server_test/pr_edit_test.go
  - [ ] 4.2 Create PullRequestEditArgs struct with validation tags
  - [ ] 4.3 Create PullRequestEditResult struct for response
  - [ ] 4.4 Implement handlePullRequestEdit function in server/pr_edit.go
  - [ ] 4.5 Add input validation using ozzo-validation
  - [ ] 4.6 Add repository resolution logic for directory parameter
  - [ ] 4.7 Add success and error response formatting
  - [ ] 4.8 Verify server handler tests pass

- [ ] 5. Register Tool in Server
  - [ ] 5.1 Write tests for tool registration in server_test/tool_discovery_test.go
  - [ ] 5.2 Add pr_edit tool registration in server/server.go
  - [ ] 5.3 Add tool import to server.go
  - [ ] 5.4 Verify tool discovery tests pass

- [ ] 6. Integration Testing and Validation
  - [ ] 6.1 Write integration tests for complete workflow
  - [ ] 6.2 Test with mock Forgejo server
  - [ ] 6.3 Test with mock Gitea server
  - [ ] 6.4 Test directory parameter resolution
  - [ ] 6.5 Test all validation scenarios
  - [ ] 6.6 Verify all integration tests pass

- [ ] 7. Documentation and Code Quality
  - [ ] 7.1 Add godoc comments to all exported functions
  - [ ] 7.2 Update README.md with new tool documentation
  - [ ] 7.3 Run go vet and goimports on all modified files
  - [ ] 7.4 Verify test coverage meets project standards
  - [ ] 7.5 Run full test suite to ensure no regressions