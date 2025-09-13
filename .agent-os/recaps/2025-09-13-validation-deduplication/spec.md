# Spec Requirements Document

> Spec: Validation Deduplication
> Created: 2025-09-13

## Overview

Consolidate input validation logic by removing duplicate validation between server and service layers, keeping all validation in server handlers using inline ozzo-validation patterns.

## User Stories

### Server Layer Simplification

As a developer, I want to eliminate duplicate validation code between server and service layers, so that I can maintain a single source of truth for input validation and reduce code complexity.

The current system has validation logic duplicated across server handlers (using ozzo-validation) and service layer functions (using custom validation functions). This creates maintenance overhead and potential inconsistencies. The solution involves removing all validation from the service layer and interface layer, keeping only the inline validation in server handlers.

### Service Layer Cleanup

As a maintainer, I want the service layer to focus purely on business logic without validation concerns, so that I can achieve cleaner separation of concerns and improve code maintainability.

The service layer currently contains validation functions that duplicate the validation already performed in server handlers. By removing these validation functions and their calls, the service layer becomes simpler and more focused on its core responsibility of coordinating between the server and client layers.

## Spec Scope

1. **Remove Service Layer Validation** - Delete all validation functions from remote/gitea/service.go and remove their calls from service methods
2. **Remove Interface Layer Validation Tags** - Strip validation tags from all struct definitions in remote/gitea/interface.go
3. **Preserve Server Layer Validation** - Keep existing inline validation patterns in all server handlers using ozzo-validation
4. **Maintain Error Handling** - Ensure error messages remain consistent and user-friendly after deduplication

## Out of Scope

- Adding new validation rules or changing existing validation logic
- Modifying the server layer validation patterns or helper functions
- Changing the ozzo-validation library or adding new validation dependencies
- Altering the overall architecture or data flow between layers

## Expected Deliverable

1. All validation logic consolidated in server handlers with no duplication in service layer
2. Service layer methods simplified to pass through calls without validation checks
3. Interface layer structs cleaned of validation tags while maintaining JSON serialization
4. All existing tests pass with no functional regression in tool behavior