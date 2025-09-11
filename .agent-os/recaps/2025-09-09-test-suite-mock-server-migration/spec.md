# Spec Requirements Document

> Spec: Test Suite Mock Server Migration
> Created: 2025-09-09

## Overview

Modernize the acceptance test suite by migrating from the in-memory MockGiteaClient to the HTTP-based MockGiteaServer, providing more realistic testing and reducing code duplication.

## User Stories

### Test Suite Modernization

As a developer, I want to use a consistent HTTP-based mock server across all tests, so that I can test actual HTTP handling code paths and reduce maintenance overhead.

The current test suite has two different mock approaches: an in-memory MockGiteaClient and an HTTP-based MockGiteaServer. This creates inconsistency, maintenance burden, and doesn't test the actual HTTP communication layer. By standardizing on the HTTP mock server, we ensure all tests exercise the real code paths while providing a more realistic testing environment.

### Code Cleanup

As a maintainer, I want to remove duplicate mock client code, so that I can reduce complexity and improve code maintainability.

The harness.go file contains both MockGiteaClient and MockGiteaServer implementations, with the MockGiteaClient being deprecated. Removing this unused code will simplify the codebase and eliminate confusion for future developers.

## Spec Scope

1. **Update issue_comment_test.go** - Migrate from MockGiteaClient to MockGiteaServer for all test cases
2. **Clean up harness.go** - Remove MockGiteaClient struct and all related methods
3. **Verify MockGiteaServer functionality** - Ensure comment creation endpoints work correctly
4. **Update test setup patterns** - Standardize on NewTestServer() instead of NewTestServerWithClient()

## Out of Scope

- Adding new test functionality or test cases
- Modifying the actual server implementation or business logic
- Changing the MCP protocol or SDK integration
- Adding new mock server capabilities beyond current functionality

## Expected Deliverable

1. All test files consistently use MockGiteaServer with HTTP-based testing
2. MockGiteaClient and related code completely removed from the codebase
3. Test suite passes with no regressions in functionality
4. Documentation and comments updated to reflect the new testing approach