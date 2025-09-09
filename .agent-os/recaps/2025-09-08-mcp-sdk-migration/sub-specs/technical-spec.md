# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-08-mcp-sdk-migration/spec.md

## Technical Requirements

### 1. Import Path Migration
- **Systematic Replacement**: Replace all occurrences of `github.com/mark3labs/mcp-go` with `github.com/modelcontextprotocol/go-sdk/mcp` across all Go source files
- **Import Aliasing**: Maintain consistent import aliases where used to minimize code changes
- **Dependency Update**: Update go.mod to remove mark3labs dependency and add official SDK dependency

### 2. Server Implementation Updates
- **Server Initialization**: Adapt the MCP server creation and startup code to match the official SDK's initialization patterns
- **Configuration Handling**: Ensure server configuration (from config.yaml) continues to work with the new SDK
- **Transport Layer**: Update any transport-specific code (stdio, SSE, WebSocket) to match official SDK interfaces

### 3. Tool Registration and Handlers
- **Tool Interface Compliance**: Update all tool handler functions to match the official SDK's tool interface signatures
- **Tool Schema Definitions**: Ensure all tool schemas remain compatible with the new SDK's schema format
- **Error Handling**: Adapt error handling patterns to match the official SDK's error types and conventions

### 4. Type and Interface Mappings
- **Request/Response Types**: Map all MCP request and response types from mark3labs to official SDK equivalents
- **Tool Context Types**: Update any tool context or execution context types to match official SDK
- **Protocol Messages**: Ensure all protocol message types align with official SDK structures

### 5. Testing Infrastructure
- **Test Harness Updates**: Modify the test harness in server_test/ to work with the official SDK
- **Mock Updates**: Update any mocked MCP interfaces to match the official SDK
- **Integration Tests**: Ensure all integration tests continue to pass without modification to test logic

### 6. Cobra CLI Integration
- **Command Handler Updates**: Ensure the serve command and other CLI commands continue to work with the new SDK
- **Configuration Loading**: Verify config loading and validation works with the official SDK

## Migration Implementation Details

### Files Requiring Updates
1. **server/server.go** - Main server implementation and MCP server initialization
2. **server/handlers.go** - Tool handler implementations and registration
3. **cmd/serve.go** - CLI command for starting the MCP server
4. **server_test/*.go** - All test files using MCP SDK types
5. **go.mod & go.sum** - Dependency management files
6. **Any additional files importing the MCP SDK**

### API Mapping Strategy
```go
// Example mapping patterns:
// mark3labs pattern -> official SDK pattern

// Server creation
// mark3labs: mcp.NewServer(...)
// official: mcp.NewServer(...) or similar API

// Tool registration
// mark3labs: server.RegisterTool(...)
// official: server.AddTool(...) or similar API

// Context handling
// mark3labs: specific context types
// official: standard context.Context with values
```

### Validation Criteria
- **Compilation**: Project must compile without errors using the official SDK
- **Tests**: All existing tests must pass without modification to test assertions
- **Functionality**: All MCP tools must function identically to their pre-migration behavior
- **Protocol Compliance**: Server must remain compliant with MCP protocol specification
- **Performance**: No degradation in server startup time or request handling performance

## External Dependencies

**New Dependency Required:**
- **github.com/modelcontextprotocol/go-sdk/mcp** - Official Model Context Protocol Go SDK
- **Justification:** Moving to the officially maintained SDK ensures long-term support, better protocol compliance, and access to official updates and bug fixes. This reduces technical debt and aligns the project with the MCP ecosystem standards.