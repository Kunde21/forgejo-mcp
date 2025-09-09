# Spec Requirements Document

> Spec: Validation Layer Migration
> Created: 2025-09-09
> Status: Planning

## Overview

This plan outlines the migration of all input validation logic from the Gitea service layer to the MCP handlers in the forgejo-mcp project. The goal is to eliminate validation duplication, improve consistency, and establish a clear separation of concerns where handlers handle input validation and services handle business logic.

Currently, validation logic is duplicated and inconsistent between the service layer and handlers. The service layer contains validation functions that should be moved to the handlers to establish a cleaner architecture and prevent validation bypass scenarios.

## User Stories

*   **As a** Developer
*   **I want to** Move all input validation from the service layer to the handlers
*   **So that** I can eliminate duplication, improve consistency, and establish clear separation of concerns

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

## Spec Scope

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

## Out of Scope

- Changing the validation rules themselves - only moving existing logic
- Adding new validation beyond what currently exists
- Refactoring the entire service layer architecture
- Adding validation for new tools not yet implemented
- Changing the MCP protocol or SDK usage

## Expected Deliverable

- Service layer with all validation functions removed, containing only business logic
- Handler layer with comprehensive validation using ozzo-validation v4 patterns
- Shared validation utilities and constants for consistent validation across handlers
- Updated test suite that covers all validation scenarios in handlers
- Consistent error message formats using ozzo-validation v4 conventions
- Documentation of the new validation architecture and patterns

## Spec Documentation

- Tasks: @.agent-os/specs/2025-09-09-validation-layer-migration/tasks.md
- Technical Specification: @.agent-os/specs/2025-09-09-validation-layer-migration/sub-specs/technical-spec.md