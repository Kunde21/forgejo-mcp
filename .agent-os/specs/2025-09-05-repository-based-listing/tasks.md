# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-09-05-repository-based-listing/spec.md

> Created: 2025-09-05
> Status: Ready for Implementation

## Tasks

### Task 1: Write Repository Parameter Validation Tests
**Status:** Completed
**Priority:** High
**Description:** Create comprehensive tests for repository parameter validation logic

- [x] Write unit tests for repository format validation (owner/repo format)
- [x] Test invalid repository formats (missing slash, extra slashes, empty parts)
- [x] Create tests for repository existence validation
- [x] Test access permission validation for private repositories
- [x] Write tests for organization-owned vs user-owned repository handling
- [x] Create mock scenarios for repository not found errors
- [x] Test edge cases like special characters in repository names
- [x] Verify error messages are descriptive and actionable

### Task 2: Write API Handler Tests for Repository Parameters
**Status:** Completed
**Priority:** High
**Description:** Create tests for updated PR and issue list handlers with repository parameters

- [x] Write tests for PR list handler with repository parameter
- [x] Write tests for issue list handler with repository parameter
- [x] Test CWD parameter resolution to repository identifier
- [x] Create tests for parameter precedence (repository vs cwd)
- [x] Test missing repository parameter error handling
- [x] Write tests for both repository and cwd parameter combinations
- [x] Create integration tests for end-to-end repository-based queries
- [x] Test pagination and filtering with repository context

### Task 3: Write Response Format Tests
**Status:** Completed
**Priority:** High
**Description:** Create tests for updated response formats with repository metadata

- [x] Test PR response includes repository metadata
- [x] Test issue response includes repository metadata
- [x] Write tests for repository information in response headers
- [x] Create tests for backward compatibility of response structure
- [x] Test total count accuracy with repository filtering
- [x] Write tests for pagination metadata preservation
- [x] Create tests for error response formats
- [x] Verify response format consistency between PR and issue endpoints

### Task 4: Implement Repository Parameter Validation
**Status:** Completed
**Priority:** High
**Description:** Implement the repository validation logic in the handlers

- [x] Add repository format validation function
- [x] Implement repository existence checking via Gitea API
- [x] Add access permission validation for repositories
- [x] Create CWD to repository resolution logic
- [x] Update error handling for repository-related errors
- [x] Implement parameter validation in both PR and issue handlers
- [x] Add repository metadata extraction and caching
- [x] Update handler method signatures to accept new parameters

### Task 5: Update API Handlers for Repository-Based Queries
**Status:** Pending
**Priority:** High
**Description:** Modify the PR and issue list handlers to use repository-based filtering

- Update PR list handler to use repository parameter
- Update issue list handler to use repository parameter
- Modify query builders to filter by repository instead of user
- Update database query logic for repository-specific data
- Implement proper JOIN operations for repository data
- Add repository context to response formatting
- Update pagination logic to work with repository scope
- Maintain existing filtering capabilities (state, author, labels)

### Task 6: Update Response Models and Formatting
**Status:** Pending
**Priority:** Medium
**Description:** Update response structures to include repository metadata

- Modify PR response model to include repository information
- Modify issue response model to include repository information
- Update JSON marshaling for new response format
- Add repository metadata to individual PR/issue objects
- Update response validation and serialization
- Ensure response format consistency across endpoints
- Add repository context to response headers
- Test response size and performance impact

### Task 7: Update SDK Client Methods
**Status:** Pending
**Priority:** Medium
**Description:** Update MCP SDK client to support repository parameters

- Modify SDK client PR list method to accept repository parameter
- Modify SDK client issue list method to accept repository parameter
- Update method signatures and parameter validation
- Add CWD parameter support to SDK methods
- Update SDK documentation and examples
- Ensure backward compatibility where possible
- Test SDK integration with updated server endpoints
- Update SDK version and release notes

### Task 8: Integration and End-to-End Testing
**Status:** Pending
**Priority:** High
**Description:** Perform comprehensive integration testing of repository-based functionality

- Create end-to-end tests for repository-based PR listing
- Create end-to-end tests for repository-based issue listing
- Test authentication and authorization with repositories
- Verify error handling for various repository scenarios
- Test performance with large repository datasets
- Create tests for concurrent repository access
- Test repository switching and context changes
- Validate API documentation accuracy

### Task 9: Update Documentation and Examples
**Status:** Pending
**Priority:** Low
**Description:** Update all documentation to reflect repository-based changes

- Update API documentation for new parameters
- Create migration guide for existing integrations
- Update code examples and usage patterns
- Add repository-based examples to documentation
- Update error code documentation
- Create troubleshooting guide for common issues
- Update SDK documentation and method references
- Add performance considerations for repository queries

### Task 10: Final Test Verification and Cleanup
**Status:** Pending
**Priority:** High
**Description:** Run final verification tests and clean up deprecated code

- Run full test suite to verify all functionality
- Remove any deprecated user-based logic
- Clean up temporary test code and debugging statements
- Verify no breaking changes to existing functionality
- Test deployment and rollback procedures
- Update version numbers and changelogs
- Perform final security and performance review
- Create release notes for the repository-based changes