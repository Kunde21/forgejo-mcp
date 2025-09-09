# Feature Specification: Validation Layer Migration

## 1. Summary
This plan outlines the migration of all input validation logic from the Gitea service layer to the MCP handlers in the forgejo-mcp project. The goal is to eliminate validation duplication, improve consistency, and establish a clear separation of concerns where handlers handle input validation and services handle business logic.

## 2. Problem Statement
Currently, validation logic is duplicated and inconsistent between the service layer and handlers. The service layer contains validation functions that should be moved to the handlers to establish a cleaner architecture and prevent validation bypass scenarios.

*   **As a** Developer
*   **I want to** Move all input validation from the service layer to the handlers
*   **So that** I can eliminate duplication, improve consistency, and establish clear separation of concerns

## 3. Current State Analysis

### Service Layer Validation (service.go)
The service layer currently contains these validation functions:
- `validateRepository(repo string) error` - Validates owner/repo format with regex `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
- `validatePagination(limit, offset int) error` - Validates limit (1-100) and offset (≥0)
- `validateIssueNumber(issueNumber int) error` - Validates issue number (>0)
- `validateCommentContent(comment string) error` - Validates non-empty and non-whitespace content

### Handler Layer Validation (handlers.go)
The handlers currently have partial validation using ozzo-validation:
- `handleListIssues` - Validates repository format, limit (1-100), offset (≥0)
- `handleCreateIssueComment` - Validates repository (required), issue number (>0), comment (non-empty)

### Issues Identified
1. **Duplication**: Both layers validate similar parameters
2. **Inconsistency**: Different validation approaches and error messages
3. **Incomplete Coverage**: Handler validation is missing some checks (e.g., repository format validation in create_issue_comment)
4. **Architecture Violation**: Service layer should not be responsible for input validation

## 4. User Stories & Acceptance Criteria

*   **User Story 1:** Remove validation functions from service layer
    *   **AC 1:** Given the service.go file, when I remove validation functions, then the file should only contain business logic
    *   **AC 2:** Given service methods, when I remove validation calls, then they should directly call the client without validation
    *   **AC 3:** Given existing tests, when I run them, then they should still pass after validation removal

*   **User Story 2:** Consolidate validation in handlers
    *   **AC 1:** Given the handlers.go file, when I enhance validation, then all validation from service layer should be covered
    *   **AC 2:** Given handleListIssues handler, when I validate repository format, then it should use the same regex as service layer
    *   **AC 3:** Given handleCreateIssueComment handler, when I validate repository format, then it should validate owner/repo format
    *   **AC 4:** Given handleCreateIssueComment handler, when I validate comment content, then it should check for whitespace-only content

*   **User Story 3:** Ensure validation consistency
    *   **AC 1:** Given both handlers, when they validate the same parameter, then they should use identical validation rules
    *   **AC 2:** Given validation errors, when they occur, then they should have consistent error message formats
    *   **AC 3:** Given repository validation, when it fails, then it should provide clear feedback about the expected format

*   **User Story 4:** Maintain test coverage
    *   **AC 1:** Given existing service tests, when I remove validation, then I should update or remove validation-specific tests
    *   **AC 2:** Given handler tests, when I enhance validation, then I should add tests for new validation cases
    *   **AC 3:** Given all tests, when I run them, then they should pass with the new validation structure

## 5. Technical Requirements & Considerations

### Service Layer Changes
- **File**: `remote/gitea/service.go`
- **Remove Functions**:
  - `validateRepository(repo string) error`
  - `validatePagination(limit, offset int) error`
  - `validateIssueNumber(issueNumber int) error`
  - `validateCommentContent(comment string) error`
- **Update Methods**:
  - `ListIssues()` - Remove validation calls, directly call client
  - `CreateIssueComment()` - Remove validation calls, directly call client
- **Error Handling**: Remove validation-specific error wrapping

### Handler Layer Changes
- **File**: `server/handlers.go`
- **Enhance handleListIssues validation**:
  - Keep existing repository format validation using ozzo-validation v4 patterns
  - Keep existing pagination validation using `v.Min()` and `v.Max()`
  - Ensure error messages are consistent with ozzo-validation v4 conventions
- **Enhance handleCreateIssueComment validation**:
  - Add repository format validation using ozzo-validation v4 `v.By()` with custom validator
  - Enhance comment content validation using `v.By()` to check for whitespace-only content
  - Keep existing issue number validation using `v.Min(1)`
- **Validation Utilities**:
  - Create shared validation constants and helper functions using ozzo-validation v4 patterns
  - Standardize error message formats using ozzo-validation v4 error handling
  - Create custom validators using `v.By()` for complex validation logic

### Validation Rules to Implement (ozzo-validation v4 patterns)
1. **Repository Format**:
   - Pattern: `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
   - Implementation: Use `v.By()` with custom validator function
   - Error: Use `v.NewError("repository_format", "repository must be in format 'owner/repo'")`
   - Both parts validation: Use `v.ValidateStruct()` with `v.Match(repoRegex)` for owner and repo parts

2. **Pagination**:
   - Limit: Use `v.Min(1).Max(100)`
   - Offset: Use `v.Min(0)`
   - Error: Leverage ozzo-validation v4's built-in error messages

3. **Issue Number**:
   - Value: Use `v.Min(1)`
   - Error: Leverage ozzo-validation v4's built-in error messages

4. **Comment Content**:
   - Non-empty: Use `v.Required`
   - Not whitespace-only: Use `v.By()` with custom validator checking `len(strings.TrimSpace(value)) > 0`
   - Error: Use `v.NewError("comment_content", "comment content cannot be only whitespace")`

### Test Updates
- **Service Tests**:
  - Remove validation-specific test cases
  - Update tests to expect validation errors from handlers, not services
  - Keep business logic tests
- **Handler Tests**:
  - Add comprehensive validation test cases
  - Test edge cases for all validation rules
  - Ensure error messages are consistent and helpful

## 6. Out of Scope
- Changing the validation rules themselves - only moving existing logic
- Adding new validation beyond what currently exists
- Refactoring the entire service layer architecture
- Adding validation for new tools not yet implemented
- Changing the MCP protocol or SDK usage

## 7. Implementation Plan

### Phase 1: Prepare Service Layer
1. Remove all validation functions from `service.go`
2. Update service methods to remove validation calls
3. Update service tests to remove validation-specific test cases
4. Ensure service methods only contain business logic

### Phase 2: Enhance Handler Validation
1. Review current handler validation in `handlers.go` and identify ozzo-validation v4 patterns
2. Add missing repository format validation to `handleCreateIssueComment` using `v.By()` with custom validator
3. Enhance comment content validation using `v.By()` to check for whitespace-only content
4. Standardize error message formats using ozzo-validation v4 error handling conventions
5. Create shared validation utilities using ozzo-validation v4 patterns:
   - Define `repoRegex` constant for repository validation
   - Create custom validator functions for complex validation logic
   - Use `v.NewError()` for consistent error message formatting

### Phase 3: Update Handler Tests
1. Add comprehensive validation test cases to handler tests using ozzo-validation v4 testing patterns
2. Test all validation rules and edge cases, including custom validator functions
3. Ensure error messages follow ozzo-validation v4 conventions and are helpful
4. Verify that validation errors are properly formatted and returned to clients using ozzo-validation v4 error structures

### Phase 4: Integration Testing
1. Run all tests to ensure nothing is broken
2. Test end-to-end functionality with validation scenarios
3. Verify that the migration maintains existing behavior
4. Check that error responses are consistent and helpful

## 8. Open Questions
- Should we create a separate validation package for shared ozzo-validation v4 utilities?
- What is the preferred error message format using ozzo-validation v4's `v.NewError()`?
- Should we add any additional validation beyond what currently exists using ozzo-validation v4 patterns?
- How should we handle internationalization of error messages with ozzo-validation v4?
- Should we add logging for validation failures using ozzo-validation v4's error handling?
- Should we create custom validator types for reusable validation logic across handlers?