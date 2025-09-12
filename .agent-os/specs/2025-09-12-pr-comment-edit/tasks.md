# Implementation Tasks

> Feature: Pull Request Comment Edit Tool
> Created: 2025-09-12

## Task Breakdown

### Phase 1: Interface Layer Implementation

#### Task 1.1: Add PullRequestCommentEditor Interface
- **File**: `remote/gitea/interface.go`
- **Description**: Add `PullRequestCommentEditor` interface with `EditPullRequestComment` method
- **Implementation**:
  - Define interface with method signature
  - Add `EditPullRequestCommentArgs` struct without validation tags
  - Update `GiteaClientInterface` to include `PullRequestCommentEditor`
- **Status**: ✅ Completed

#### Task 1.2: Add Type Definitions
- **File**: `remote/gitea/interface.go`
- **Description**: Add necessary type definitions for PR comment editing
- **Implementation**:
  - Ensure `PullRequestComment` struct exists (should already be there from listing/creation)
  - Add `EditPullRequestCommentArgs` struct with proper JSON tags
- **Status**: ✅ Completed

### Phase 2: Client Layer Implementation

#### Task 2.1: Implement EditPullRequestComment Method
- **File**: `remote/gitea/gitea_client.go`
- **Description**: Implement the `EditPullRequestComment` method using Gitea SDK
- **Implementation**:
  - Parse repository string (owner/repo format)
  - Use Gitea SDK to edit comment on pull request
  - Convert Gitea SDK response to internal `PullRequestComment` struct
  - Handle errors with proper wrapping and context
  - **No input validation** - trust that inputs are already validated
- **Status**: ⏳ Pending

#### Task 2.2: Add Client Tests
- **File**: `remote/gitea/client_test.go`
- **Description**: Add unit tests for the new `EditPullRequestComment` method
- **Implementation**:
  - Test successful comment editing
  - Test error scenarios (invalid repository, PR not found, comment not found, etc.)
  - Test response conversion and formatting
- **Status**: ⏳ Pending

### Phase 3: Service Layer Implementation

#### Task 3.1: Add Service Method
- **File**: `remote/gitea/service.go`
- **Description**: Add `EditPullRequestComment` method to service layer
- **Implementation**:
  - Add method that calls client directly
  - **No validation methods** - trust that server handler already validated inputs
  - Focus on business logic and error handling from API calls
  - Return converted `PullRequestComment` struct
- **Status**: ⏳ Pending

#### Task 3.2: Add Service Tests
- **File**: `remote/gitea/service_test.go`
- **Description**: Add unit tests for the service method
- **Implementation**:
  - Test successful comment editing through service
  - Test error propagation from client
  - **No validation tests** - validation is handled in server layer
- **Status**: ⏳ Pending

### Phase 4: Server Layer Implementation

#### Task 4.1: Add Server Handler Args Struct
- **File**: `server/pr_comments.go`
- **Description**: Add `PullRequestCommentEditArgs` struct with ozzo-validation tags
- **Implementation**:
  - Define struct with proper JSON tags
  - Add ozzo-validation tags for all parameters (repository, PR number, comment ID, new content)
  - Follow existing patterns from other comment editing handlers
- **Status**: ⏳ Pending

#### Task 4.2: Implement MCP Tool Handler
- **File**: `server/pr_comments.go`
- **Description**: Implement `handlePullRequestCommentEdit` handler function
- **Implementation**:
  - Use ozzo-validation for all parameter validation
  - Call service layer to edit comment
  - Format success response with updated comment metadata
  - Handle validation errors and API errors appropriately
  - Follow existing patterns from `handleIssueCommentEdit`
- **Status**: ⏳ Pending

#### Task 4.3: Register New Tool
- **File**: `server/server.go`
- **Description**: Register the new `pr_comment_edit` tool with the MCP server
- **Implementation**:
  - Add tool registration using `mcp.AddTool`
  - Provide descriptive tool name and description
  - Follow existing registration patterns
- **Status**: ⏳ Pending

### Phase 5: Testing Implementation

#### Task 5.1: Add Handler Unit Tests
- **File**: `server_test/pr_comment_edit_test.go`
- **Description**: Add unit tests for the new MCP tool handler
- **Implementation**:
  - Test successful comment editing
  - Test validation scenarios (invalid repository, PR number, comment ID, new content)
  - Test error handling and response formatting
  - Use existing test patterns and helpers
- **Status**: ⏳ Pending

#### Task 5.2: Add Integration Tests
- **File**: `server_test/integration_test.go`
- **Description**: Add integration tests for the new tool
- **Implementation**:
  - Test end-to-end comment editing workflow
  - Test with mock Gitea server
  - Verify proper MCP protocol compliance
- **Status**: ⏳ Pending

#### Task 5.3: Update Test Harness
- **File**: `server_test/harness.go`
- **Description**: Update test harness to support PR comment editing
- **Implementation**:
  - Add mock server endpoints for PR comment editing
  - Add test helpers for PR comment edit operations
  - Follow existing harness patterns
- **Status**: ⏳ Pending

#### Task 5.4: Add Acceptance Tests
- **File**: `server_test/pr_comment_edit_acceptance_test.go`
- **Description**: Add acceptance tests following existing patterns
- **Implementation**:
  - Create comprehensive acceptance test file
  - Test real-world usage scenarios
  - Verify integration with existing tools
- **Status**: ⏳ Pending

### Phase 6: Documentation and Finalization

#### Task 6.1: Update README
- **File**: `README.md`
- **Description**: Update documentation with new tool usage
- **Implementation**:
  - Add `pr_comment_edit` tool documentation
  - Include usage examples and parameter descriptions
  - Update tool list and examples
- **Status**: ⏳ Pending

#### Task 6.2: Run Full Test Suite
- **Description**: Ensure all tests pass with new functionality
- **Implementation**:
  - Run `go test ./...` to verify all tests pass
  - Run `go vet ./...` for static analysis
  - Run `goimports -w .` for code formatting
- **Status**: ⏳ Pending

#### Task 6.3: Final Review
- **Description**: Review implementation against requirements
- **Implementation**:
  - Verify all spec requirements are met
  - Check code quality and adherence to patterns
  - Ensure no regressions in existing functionality
- **Status**: ⏳ Pending

## Task Dependencies

- **Phase 1** must be completed before Phase 2
- **Phase 2** must be completed before Phase 3
- **Phase 3** must be completed before Phase 4
- **Phase 4** must be completed before Phase 5
- **Phase 5** must be completed before Phase 6

## Quality Gates

- All tests must pass (100% test coverage for new code)
- Code must follow existing patterns and conventions
- No validation duplication between service and server layers
- Proper error handling and response formatting
- Documentation must be complete and accurate

## Success Metrics

- ✅ New `pr_comment_edit` tool successfully edits comments on pull requests
- ✅ Validation performed only in server handler using ozzo-validation
- ✅ Service layer has no validation logic (clean separation of concerns)
- ✅ All existing functionality remains intact (no regressions)
- ✅ Complete test coverage with all tests passing
- ✅ Proper error handling for both validation and API errors
- ✅ Documentation updated with usage examples