# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-08-31-types-and-models/spec.md

> Created: 2025-08-31
> Status: Ready for Implementation

## Tasks

- [x] 1. Implement Common Types Foundation
   - [x] 1.1 Write tests for common types (Repository, User, Timestamp, FilterOptions)
   - [x] 1.2 Create types/common.go with Repository and User structs
   - [x] 1.3 Implement custom Timestamp type with RFC3339 JSON marshaling
   - [x] 1.4 Add FilterOptions and SortOrder types with validation
   - [x] 1.5 Create validation helper functions for common checks
   - [x] 1.6 Verify all common types tests pass

- [x] 2. Implement Pull Request Types
   - [x] 2.1 Write tests for PullRequest types and validation logic
   - [x] 2.2 Create types/pr.go with PullRequest struct and all fields
   - [x] 2.3 Define PRAuthor, PRLabel, and PRState supporting types
   - [x] 2.4 Implement Validate() method with comprehensive field checking
   - [x] 2.5 Add helper methods (IsOpen, IsClosed, IsMerged)
   - [x] 2.6 Configure JSON tags with camelCase and omitempty
   - [x] 2.7 Test JSON marshaling/unmarshaling behavior
   - [x] 2.8 Verify all PR types tests pass with >90% coverage

- [x] 3. Implement Issue Types
   - [x] 3.1 Write tests for Issue types and validation
   - [x] 3.2 Create types/issue.go with Issue struct definition
   - [x] 3.3 Define IssueState enum and Milestone struct
   - [x] 3.4 Implement Validate() method for Issue
   - [x] 3.5 Add HasLabel(name string) helper method
   - [x] 3.6 Configure JSON serialization with proper tags
   - [x] 3.7 Test label checking and milestone handling
   - [x] 3.8 Verify all issue types tests pass with >90% coverage

- [x] 4. Implement Response Types
   - [x] 4.1 Write tests for MCP response types
   - [x] 4.2 Create types/responses.go with SuccessResponse structure
   - [x] 4.3 Define ErrorResponse with ErrorDetails and standard codes
   - [x] 4.4 Implement PaginatedResponse with Pagination metadata
   - [x] 4.5 Add response builder functions for common patterns
   - [x] 4.6 Test response construction and JSON output format
   - [x] 4.7 Verify all response types tests pass

- [x] 5. Integrate Types with Existing Code
   - [x] 5.1 Update server/handlers.go to use new PullRequest type
   - [x] 5.2 Update handlePRList to return typed responses
   - [x] 5.3 Update handleIssueList to use Issue type
   - [x] 5.4 Replace all map[string]interface{} usage in handlers
   - [x] 5.5 Update client transformation functions for type compatibility
   - [x] 5.6 Run integration tests to verify end-to-end functionality
   - [x] 5.7 Verify no performance regression with benchmarks
   - [x] 5.8 Ensure all existing tests still pass