# Spec Tasks

## Tasks

- [x] 1. Remove validation functions from service layer
  - [x] 1.1 Write tests to verify current service validation behavior
  - [x] 1.2 Remove validation functions from remote/gitea/service.go
  - [x] 1.3 Update ListIssues() method to remove validation calls
  - [x] 1.4 Update CreateIssueComment() method to remove validation calls
  - [x] 1.5 Remove validation-specific error wrapping from service methods
  - [x] 1.6 Update service tests to remove validation-specific test cases
  - [x] 1.7 Verify all service tests pass after validation removal

- [x] 2. Create shared validation utilities and constants
  - [x] 2.1 Write tests for validation utility functions
  - [x] 2.2 Create validation package structure with validators.go
  - [x] 2.3 Implement repository format validation using ozzo-validation v4 patterns
  - [x] 2.4 Implement pagination validation helpers
  - [x] 2.5 Implement issue number validation helpers
  - [x] 2.6 Implement comment content validation helpers
  - [x] 2.7 Create shared validation constants (repoRegex, error messages)
  - [x] 2.8 Verify all validation utility tests pass

- [x] 3. Enhance handleListIssues handler validation
  - [x] 3.1 Write tests for enhanced handleListIssues validation
  - [x] 3.2 Review current repository format validation in handleListIssues
  - [x] 3.3 Ensure repository validation uses ozzo-validation v4 patterns
  - [x] 3.4 Review current pagination validation in handleListIssues
  - [x] 3.5 Ensure pagination validation uses v.Min() and v.Max() correctly
  - [x] 3.6 Standardize error message formats using ozzo-validation v4 conventions
  - [x] 3.7 Add comprehensive validation test cases for edge cases
  - [x] 3.8 Verify all handleListIssues validation tests pass

- [x] 4. Enhance handleCreateIssueComment handler validation
  - [x] 4.1 Write tests for enhanced handleCreateIssueComment validation
  - [x] 4.2 Add repository format validation using v.By() with custom validator
  - [x] 4.3 Enhance comment content validation to check whitespace-only content
  - [x] 4.4 Ensure issue number validation uses v.Min(1) correctly
  - [x] 4.5 Standardize error message formats with other handlers
  - [x] 4.6 Add comprehensive validation test cases for all scenarios
  - [x] 4.7 Test custom validator functions for repository format and comment content
  - [x] 4.8 Verify all handleCreateIssueComment validation tests pass

- [x] 5. Ensure validation consistency across handlers
  - [x] 5.1 Write tests to validate consistency between handlers
  - [x] 5.2 Compare validation rules for common parameters between handlers
  - [x] 5.3 Standardize repository format validation across both handlers
  - [x] 5.4 Standardize error message formats using v.NewError() patterns
  - [x] 5.5 Ensure consistent error response structures
  - [x] 5.6 Add tests to verify identical validation rules for same parameters
  - [x] 5.7 Verify repository validation provides clear format feedback
  - [x] 5.8 Verify all consistency tests pass

- [x] 6. Update handler tests with comprehensive validation coverage
  - [x] 6.1 Review existing handler test structure and patterns
  - [x] 6.2 Add validation test cases for handleListIssues handler
  - [x] 6.3 Add validation test cases for handleCreateIssueComment handler
  - [x] 6.4 Test edge cases for all validation rules (empty, invalid, boundary values)
  - [x] 6.5 Test custom validator functions thoroughly
  - [x] 6.6 Ensure error messages are consistent and helpful
  - [x] 6.7 Add tests for validation error response formatting
  - [x] 6.8 Verify all handler tests pass with new validation structure

- [x] 7. Integration testing and validation
  - [x] 7.1 Write end-to-end integration tests for validation scenarios
  - [x] 7.2 Test validation error flows from handlers to client responses
  - [x] 7.3 Verify that validation migration maintains existing behavior
  - [x] 7.4 Test that invalid inputs are properly rejected at handler level
  - [x] 7.5 Test that valid inputs flow through to service layer correctly
  - [x] 7.6 Run complete test suite to ensure no regressions
  - [x] 7.7 Verify error responses are consistent and helpful
  - [x] 7.8 Document validation architecture and patterns used

- [x] 8. Final verification and documentation
  - [x] 8.1 Run all tests to ensure complete functionality
  - [x] 8.2 Verify service layer contains only business logic (no validation)
  - [x] 8.3 Verify handler layer contains comprehensive validation
  - [x] 8.4 Check that all validation uses ozzo-validation v4 patterns
  - [x] 8.5 Verify no validation duplication exists between layers
  - [x] 8.6 Update project documentation with new validation architecture
  - [x] 8.7 Create validation usage examples and best practices
  - [x] 8.8 Final verification that all acceptance criteria are met