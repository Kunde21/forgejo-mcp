
# MCP SDK Migration Plan

## Overview
This document outlines the comprehensive plan to remove all custom MCP server implementation code and replace it with the official "github.com/modelcontextprotocol/go-sdk/mcp" library for full MCP protocol compliance.

## Current Custom MCP Implementation Analysis

The codebase currently has a hybrid approach with both MCP SDK usage and extensive custom implementation:

### Custom Components to Remove:
1. **Custom Server Architecture** (`server/server.go`):
   - `Server` struct with custom transport/dispatcher/processor
   - `Transport` interface and `StdioTransport`/`SSETransport` implementations
   - `RequestDispatcher`, `MessageProcessor`, `ToolRegistry`
   - Custom authentication system (`AuthState`, `AuthenticatedToolHandler`)

2. **Custom MCP Server Wrapper** (`server/mcp_server.go`):
   - `MCPServer` struct that wraps MCP SDK server
   - Custom tool registration using `mcp.AddTool`
   - Custom handler functions for tools
   - Mixed MCP SDK and custom authentication

3. **Custom Transport Layer** (`server/transport.go`):
   - Custom transport interfaces and SSE implementation
   - JSON-RPC message processing
   - Connection management adapters

4. **Custom Tool System** (`server/tools.go`):
   - Tool definitions, registry, validation, and manifest generation

5. **Multiple Handler Implementations**:
   - `handlers.go`: Mock data handlers
   - `tea_handlers.go`: Tea CLI integration handlers
   - Gitea SDK integration handlers

## Migration Plan

### Phase 1: Core MCP SDK Migration
1. **Replace Custom Server with Pure MCP SDK**:
   - Remove `server/server.go` entirely
   - Remove `server/mcp_server.go`
   - Remove `server/transport.go`
   - Remove `server/tools.go`
   - Keep only essential handlers from `server/handlers.go` and `server/tea_handlers.go`

2. **Update Command Layer** (`cmd/serve.go`):
   - Simplify to use only MCP SDK server creation
   - Remove custom transport selection logic
   - Use MCP SDK's built-in transport handling

3. **Update MCP Configuration** (`mcp.json`):
   - Ensure it points to the simplified MCP SDK server

### Phase 2: Handler Consolidation
1. **Consolidate Tool Handlers**:
   - Keep only the most functional handler implementations
   - Remove mock data handlers
   - Standardize on Gitea SDK integration handlers
   - Ensure all handlers use MCP SDK's `CallToolRequest`/`CallToolResult` types

2. **Authentication Integration**:
   - Move authentication logic to a middleware approach
   - Use MCP SDK's request handling pipeline
   - Remove custom `AuthState` and authentication wrappers

### Phase 3: Dependency Cleanup
1. **Update go.mod**:
   - Ensure `github.com/modelcontextprotocol/go-sdk/mcp` is the primary MCP dependency
   - Remove any unnecessary dependencies used only by custom implementation

2. **File Structure Simplification**:
   - Remove obsolete server files
   - Consolidate remaining handlers into fewer files
   - Update imports throughout the codebase

### Phase 4: Testing and Validation
1. **Update Tests**:
   - Remove tests for custom components
   - Update integration tests to work with MCP SDK
   - Ensure MCP protocol compliance

2. **Configuration Updates**:
   - Update any configuration files that reference removed components
   - Ensure MCP server configuration works with SDK defaults

## Key Benefits of This Migration

1. **Reduced Complexity**: Eliminates ~2000+ lines of custom MCP implementation
2. **Better Maintainability**: Uses official MCP SDK with community support
3. **Protocol Compliance**: Ensures proper MCP protocol implementation
4. **Simplified Architecture**: Single server implementation instead of hybrid approach
5. **Future-Proof**: Automatic updates with MCP SDK improvements

## Files to Remove:
- `server/server.go` (entire file)
- `server/mcp_server.go` (entire file)
- `server/transport.go` (entire file)
- `server/tools.go` (entire file)
- `server/handlers.go` (mock handlers only)
- `server/tea_handlers.go` (custom tea integration, keep Gitea SDK parts)

## Files to Modify:
- `cmd/serve.go` (simplify to pure MCP SDK usage)
- `server/tea_handlers.go` (keep only Gitea SDK handlers, adapt to MCP SDK)
- `go.mod` (dependency cleanup)
- Test files (update for new architecture)

## Implementation Notes

- All interactions and tool calls must use the "github.com/modelcontextprotocol/go-sdk/mcp" library
- Authentication should be integrated as MCP SDK middleware
- Tool handlers should use MCP SDK's standard request/response types
- Transport handling should rely on MCP SDK's built-in capabilities
- Configuration should leverage MCP SDK defaults where possible

This migration will result in a much cleaner, more maintainable codebase that fully leverages the official MCP SDK while preserving the core Forgejo integration functionality.
