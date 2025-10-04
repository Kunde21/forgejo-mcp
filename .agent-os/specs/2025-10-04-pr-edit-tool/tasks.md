# Spec Tasks

## Tasks

- [x] 1. Extend Remote Interface and Types
  - [x] 1.1 Write tests for EditPullRequestArgs struct and PullRequestEditor interface
  - [x] 1.2 Add EditPullRequestArgs struct to remote/interface.go
  - [x] 1.3 Add PullRequestEditor interface to remote/interface.go
  - [x] 1.4 Update ClientInterface to include PullRequestEditor
  - [x] 1.5 Verify all interface tests pass

- [x] 2. Implement Forgejo Client EditPullRequest Method
  - [x] 2.1 Write tests for EditPullRequest method in remote/forgejo/forgejo_client_test.go
  - [x] 2.2 Implement EditPullRequest method in remote/forgejo/pull_requests.go
  - [x] 2.3 Add validation logic for repository format and PR number
  - [x] 2.4 Add error handling for API failures and not found cases
  - [x] 2.5 Verify Forgejo client tests pass

- [x] 3. Implement Gitea Client EditPullRequest Method
  - [x] 3.1 Write tests for EditPullRequest method in remote/gitea/client_test.go
  - [x] 3.2 Implement EditPullRequest method in remote/gitea/gitea_client.go
  - [x] 3.3 Ensure consistent behavior with Forgejo implementation
  - [x] 3.4 Add validation and error handling matching Forgejo patterns
  - [x] 3.5 Verify Gitea client tests pass

- [x] 4. Create Server Handler for PR Edit Tool
  - [x] 4.1 Write tests for handlePullRequestEdit in server_test/pr_edit_test.go
  - [x] 4.2 Create PullRequestEditArgs struct with validation tags
  - [x] 4.3 Create PullRequestEditResult struct for response
  - [x] 4.4 Implement handlePullRequestEdit function in server/pr_edit.go
  - [x] 4.5 Add input validation using ozzo-validation
  - [x] 4.6 Add repository resolution logic for directory parameter
  - [x] 4.7 Add success and error response formatting
  - [x] 4.8 Verify server handler tests pass

- [x] 5. Register Tool in Server
  - [x] 5.1 Write tests for tool registration in server_test/tool_discovery_test.go
  - [x] 5.2 Add pr_edit tool registration in server/server.go
  - [x] 5.3 Add tool import to server.go
  - [x] 5.4 Verify tool discovery tests pass

- [x] 6. Integration Testing and Validation
  - [x] 6.1 Write integration tests for complete workflow
  - [x] 6.2 Test with mock Forgejo server
  - [x] 6.3 Test with mock Gitea server
  - [x] 6.4 Test directory parameter resolution
  - [x] 6.5 Test all validation scenarios
  - [x] 6.6 Verify all integration tests pass

- [x] 7. Documentation and Code Quality
  - [x] 7.1 Add godoc comments to all exported functions
  - [x] 7.2 Update README.md with new tool documentation
  - [x] 7.3 Run go vet and goimports on all modified files
  - [x] 7.4 Verify test coverage meets project standards
  - [x] 7.5 Run full test suite to ensure no regressions