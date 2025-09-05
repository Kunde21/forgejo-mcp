# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-04-mcp-sdk-migration/spec.md

## Technical Requirements

### Server Architecture Migration
- Replace `server/server.go` entirely with MCP SDK server initialization
- Replace `server/mcp_server.go` with SDK-based server wrapper
- Remove `server/transport.go` and rely on SDK's built-in transport handling
- Remove `server/tools.go` and use SDK's tool registration system
- Update `cmd/serve.go` to use pure MCP SDK server creation without custom transport selection

### Handler Consolidation
- Consolidate tool handlers in `server/tea_handlers.go` to use MCP SDK types
- Remove mock data handlers from `server/handlers.go`
- Standardize all handlers to use `CallToolRequest`/`CallToolResult` types
- Ensure Gitea SDK integration handlers work with SDK request/response pipeline

### Authentication Integration
- Move authentication logic to MCP SDK middleware approach
- Remove custom `AuthState` and authentication wrapper functions
- Integrate auth validation into SDK's request handling pipeline
- Maintain existing authentication validation logic but adapt to SDK patterns

### Code Quality and Testing
- Update all imports to use `github.com/modelcontextprotocol/go-sdk/mcp`
- Ensure MCP protocol compliance through updated integration tests
- Maintain existing test coverage while adapting to SDK architecture
- Verify performance is not impacted by migration

## External Dependencies

- **github.com/modelcontextprotocol/go-sdk/mcp** - Official MCP SDK (already in use, ensure latest version)
- Remove dependencies only used by custom implementation components (to be identified during migration)