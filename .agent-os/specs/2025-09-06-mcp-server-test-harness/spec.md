# Spec Requirements Document

> Spec: MCP Server Test Harness
> Created: 2025-09-06

## Overview

Implement a comprehensive test harness for the Forgejo MCP server to enable integration testing of the MCP server's external interface through the stdio transport protocol. This will ensure reliable testing of the server's JSON-RPC 2.0 communication and subprocess management capabilities.

## User Stories

### Developer Integration Testing

As a Go developer working on the MCP server, I want to run automated integration tests that verify the complete MCP protocol flow, so that I can catch integration issues early in development.

The developer will write test cases using the TestServer struct, start the server subprocess, send MCP protocol messages, and verify responses match expected behavior for tool discovery, execution, and error handling.

### CI/CD Pipeline Validation

As a DevOps engineer maintaining the CI/CD pipeline, I want to include comprehensive integration tests that validate the MCP server's external interface, so that I can ensure production deployments are reliable.

The pipeline will execute the test harness as part of the build process, running tests for server lifecycle, protocol initialization, tool operations, and concurrent request handling to validate the complete system functionality.

### Quality Assurance Verification

As a QA engineer testing the MCP server, I want to use the test harness to simulate real client interactions, so that I can verify the server behaves correctly under various conditions.

The QA engineer will use the harness to test error scenarios, concurrent requests, and edge cases that unit tests cannot cover, ensuring the server handles all MCP protocol requirements properly.

## Spec Scope

1. **TestServer Implementation** - Create a Go struct that manages MCP server subprocess lifecycle with proper stdin/stdout pipe handling
2. **MCP Protocol Client** - Use `github.com/mark3labs/mcp-go/client` for sending requests and receiving responses
3. **Test Utilities** - Develop helper functions for common test operations like initialization and tool calling
4. **Integration Test Scenarios** - Create comprehensive test cases covering server lifecycle, protocol initialization, tool discovery, execution, and error handling
5. **Concurrent Request Testing** - Implement tests for multiple simultaneous MCP requests to validate server concurrency

## Out of Scope

- Unit tests for individual server components (already covered by existing tests)
- Performance benchmarking beyond basic concurrent request validation
- HTTP transport testing (focus on stdio protocol only)
- Authentication flow testing (server uses external authentication)
- Load testing with hundreds of concurrent clients

## Expected Deliverable

1. A complete test harness implementation in Go using the testing package and exec.CommandContext
2. Integration tests that can be run with `go test` and validate the MCP server's external interface
3. TestServer struct with methods for starting/stopping the server and sending MCP protocol messages
4. Comprehensive test coverage for all major MCP protocol operations and error scenarios
