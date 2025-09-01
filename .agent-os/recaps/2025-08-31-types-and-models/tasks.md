# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-08-31-types-and-models/spec.md

> Created: 2025-08-31
> Status: Ready for Implementation

## Tasks

- [ ] 1. Implement Common Types Foundation
  - [ ] 1.1 Write tests for common types (Repository, User, Timestamp, FilterOptions)
  - [ ] 1.2 Create types/common.go with Repository and User structs
  - [ ] 1.3 Implement custom Timestamp type with RFC3339 JSON marshaling
  - [ ] 1.4 Add FilterOptions and SortOrder types with validation
  - [ ] 1.5 Create validation helper functions for common checks
  - [ ] 1.6 Verify all common types tests pass

- [ ] 2. Implement Pull Request Types
  - [ ] 2.1 Write tests for PullRequest types and validation logic
  - [ ] 2.2 Create types/pr.go with PullRequest struct and all fields
  - [ ] 2.3 Define PRAuthor, PRLabel, and PRState supporting types
  - [ ] 2.4 Implement Validate() method with comprehensive field checking
  - [ ] 2.5 Add helper methods (IsOpen, IsClosed, IsMerged)
  - [ ] 2.6 Configure JSON tags with camelCase and omitempty
  - [ ] 2.7 Test JSON marshaling/unmarshaling behavior
  - [ ] 2.8 Verify all PR types tests pass with >90% coverage

- [ ] 3. Implement Issue Types
  - [ ] 3.1 Write tests for Issue types and validation
  - [ ] 3.2 Create types/issue.go with Issue struct definition
  - [ ] 3.3 Define IssueState enum and Milestone struct
  - [ ] 3.4 Implement Validate() method for Issue
  - [ ] 3.5 Add HasLabel(name string) helper method
  - [ ] 3.6 Configure JSON serialization with proper tags
  - [ ] 3.7 Test label checking and milestone handling
  - [ ] 3.8 Verify all issue types tests pass with >90% coverage

- [ ] 4. Implement Response Types
  - [ ] 4.1 Write tests for MCP response types
  - [ ] 4.2 Create types/responses.go with SuccessResponse structure
  - [ ] 4.3 Define ErrorResponse with ErrorDetails and standard codes
  - [ ] 4.4 Implement PaginatedResponse with Pagination metadata
  - [ ] 4.5 Add response builder functions for common patterns
  - [ ] 4.6 Test response construction and JSON output format
  - [ ] 4.7 Verify all response types tests pass

- [ ] 5. Integrate Types with Existing Code
  - [ ] 5.1 Update server/handlers.go to use new PullRequest type
  - [ ] 5.2 Update handlePRList to return typed responses
  - [ ] 5.3 Update handleIssueList to use Issue type
  - [ ] 5.4 Replace all map[string]interface{} usage in handlers
  - [ ] 5.5 Update client transformation functions for type compatibility
  - [ ] 5.6 Run integration tests to verify end-to-end functionality
  - [ ] 5.7 Verify no performance regression with benchmarks
  - [ ] 5.8 Ensure all existing tests still pass