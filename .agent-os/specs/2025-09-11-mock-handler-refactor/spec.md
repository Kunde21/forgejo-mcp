# Spec Requirements Document

> Spec: Mock HTTP Handler Refactor
> Created: 2025-09-11

## Overview

Refactor the monolithic `handleRepoRequests` function in the test harness to use modern Go 1.22+ `http.ServeMux` with method-specific path patterns, improving code maintainability, testability, and extensibility for future Forgejo MCP server testing infrastructure.

## User Stories

### Test Infrastructure Maintainability

As a developer, I want to split the 310+ line monolithic HTTP handler into focused, individual handlers, so that the test code is easier to understand, maintain, and extend.

The current `handleRepoRequests` function handles 5 different API endpoints with complex string matching logic, making it difficult to test individual endpoint logic and add new endpoints without affecting existing ones. This refactoring will create separate handler functions for each endpoint and use Go's built-in routing capabilities.

### Testing Efficiency

As a developer, I want to use modern Go HTTP routing patterns with automatic path parameter extraction, so that I can write cleaner, more efficient test harness code that follows Go best practices.

The current implementation uses manual string parsing and method checking, which is error-prone and inefficient. By leveraging `http.ServeMux` with method + path patterns and `r.PathValue()` for parameter extraction, the code will be more performant and idiomatic.

### Future Extensibility

As a developer, I want a modular test harness architecture, so that I can easily add new endpoint handlers for upcoming Forgejo MCP features without modifying existing logic.

The refactored architecture will allow new endpoints to be added by simply registering new handler functions with specific patterns, without touching existing handlers. This will support the development of upcoming features like PR comment functionality, issue creation, and advanced repository operations.

## Spec Scope

1. **Split Monolithic Handler** - Break down the 310+ line `handleRepoRequests` function into 5 focused handler functions for individual endpoints
2. **Modern HTTP Routing** - Replace manual string matching with Go 1.22+ `http.ServeMux` method + path patterns
3. **Helper Function Extraction** - Create reusable utility functions for common operations like repository validation and pagination parsing
4. **Path Parameter Modernization** - Replace manual string parsing with `r.PathValue()` for automatic parameter extraction
5. **Preserve Functionality** - Ensure all existing test functionality remains intact without breaking changes

## Out of Scope

- Changes to the public API of the test harness
- Modifications to existing test files that use the harness
- New endpoint functionality beyond what currently exists
- Changes to the Forgejo MCP server production code
- Database schema changes or migrations
- API specification changes for the MCP server

## Expected Deliverable

1. Refactored `server_test/harness.go` with individual handler functions and modern routing patterns
2. All existing tests pass without modification, confirming backward compatibility
3. Code is more maintainable with clear separation of concerns and focused handlers
4. New endpoints can be added easily by registering additional handler functions with specific patterns