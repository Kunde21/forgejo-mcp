# Spec Tasks

## Tasks

- [ ] 1. Remove validation functions from service layer
  - [ ] 1.1 Write tests to verify current service validation behavior
  - [ ] 1.2 Remove validation functions from remote/gitea/service.go
  - [ ] 1.3 Update ListIssues() method to remove validation calls
  - [ ] 1.4 Update CreateIssueComment() method to remove validation calls
  - [ ] 1.5 Remove validation-specific error wrapping from service methods
  - [ ] 1.6 Update service tests to remove validation-specific test cases
  - [ ] 1.7 Verify all service tests pass after validation removal

- [ ] 2. Create shared validation utilities and constants
  - [ ] 2.1 Write tests for validation utility functions
  - [ ] 2.2 Create validation package structure with validators.go
  - [ ] 2.3 Implement repository format validation using ozzo-validation v4 patterns
  - [ ] 2.4 Implement pagination validation helpers
  - [ ] 2.5 Implement issue number validation helpers
  - [ ] 2.6 Implement comment content validation helpers
  - [ ] 2.7 Create shared validation constants (repoRegex, error messages)
  - [ ] 2.8 Verify all validation utility tests pass

- [ ] 3. Enhance handleListIssues handler validation
  - [ ] 3.1 Write tests for enhanced handleListIssues validation
  - [ ] 3.2 Review current repository format validation in handleListIssues
  - [ ] 3.3 Ensure repository validation uses ozzo-validation v4 patterns
  - [ ] 3.4 Review current pagination validation in handleListIssues
  - [ ] 3.5 Ensure pagination validation uses v.Min() and v.Max() correctly
  - [ ] 3.6 Standardize error message formats using ozzo-validation v4 conventions
  - [ ] 3.7 Add comprehensive validation test cases for edge cases
  - [ ] 3.8 Verify all handleListIssues validation tests pass

- [ ] 4. Enhance handleCreateIssueComment handler validation
  - [ ] 4.1 Write tests for enhanced handleCreateIssueComment validation
  - [ ] 4.2 Add repository format validation using v.By() with custom validator
  - [ ] 4.3 Enhance comment content validation to check whitespace-only content
  - [ ] 4.4 Ensure issue number validation uses v.Min(1) correctly
  - [ ] 4.5 Standardize error message formats with other handlers
  - [ ] 4.6 Add comprehensive validation test cases for all scenarios
  - [ ] 4.7 Test custom validator functions for repository format and comment content
  - [ ] 4.8 Verify all handleCreateIssueComment validation tests pass

- [ ] 5. Ensure validation consistency across handlers
  - [ ] 5.1 Write tests to validate consistency between handlers
  - [ ] 5.2 Compare validation rules for common parameters between handlers
  - [ ] 5.3 Standardize repository format validation across both handlers
  - [ ] 5.4 Standardize error message formats using v.NewError() patterns
  - [ ] 5.5 Ensure consistent error response structures
  - [ ] 5.6 Add tests to verify identical validation rules for same parameters
  - [ ] 5.7 Verify repository validation provides clear format feedback
  - [ ] 5.8 Verify all consistency tests pass

- [ ] 6. Update handler tests with comprehensive validation coverage
  - [ ] 6.1 Review existing handler test structure and patterns
  - [ ] 6.2 Add validation test cases for handleListIssues handler
  - [ ] 6.3 Add validation test cases for handleCreateIssueComment handler
  - [ ] 6.4 Test edge cases for all validation rules (empty, invalid, boundary values)
  - [ ] 6.5 Test custom validator functions thoroughly
  - [ ] 6.6 Ensure error messages are consistent and helpful
  - [ ] 6.7 Add tests for validation error response formatting
  - [ ] 6.8 Verify all handler tests pass with new validation structure

- [ ] 7. Integration testing and validation
  - [ ] 7.1 Write end-to-end integration tests for validation scenarios
  - [ ] 7.2 Test validation error flows from handlers to client responses
  - [ ] 7.3 Verify that validation migration maintains existing behavior
  - [ ] 7.4 Test that invalid inputs are properly rejected at handler level
  - [ ] 7.5 Test that valid inputs flow through to service layer correctly
  - [ ] 7.6 Run complete test suite to ensure no regressions
  - [ ] 7.7 Verify error responses are consistent and helpful
  - [ ] 7.8 Document validation architecture and patterns used

- [ ] 8. Final verification and documentation
  - [ ] 8.1 Run all tests to ensure complete functionality
  - [ ] 8.2 Verify service layer contains only business logic (no validation)
  - [ ] 8.3 Verify handler layer contains comprehensive validation
  - [ ] 8.4 Check that all validation uses ozzo-validation v4 patterns
  - [ ] 8.5 Verify no validation duplication exists between layers
  - [ ] 8.6 Update project documentation with new validation architecture
  - [ ] 8.7 Create validation usage examples and best practices
  - [ ] 8.8 Final verification that all acceptance criteria are met