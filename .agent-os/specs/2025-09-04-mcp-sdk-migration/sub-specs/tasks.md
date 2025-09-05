# Implementation Tasks

## Phase 1: Core Migration (High Priority)

### 1.1 Remove Custom Server Components
- Remove `server/server.go` entirely
- Remove `server/mcp_server.go` entirely
- Remove `server/transport.go` entirely
- Remove `server/tools.go` entirely
- Remove mock handlers from `server/handlers.go`

### 1.2 Update Command Layer
- Update `cmd/serve.go` to use pure MCP SDK server creation
- Remove custom transport selection logic
- Use MCP SDK's built-in transport handling
- Update `mcp.json` configuration for SDK compatibility

## Phase 2: Handler Consolidation (High Priority)

### 2.1 Consolidate Tool Handlers
- Keep only functional Gitea SDK handlers in `server/tea_handlers.go`
- Adapt all handlers to use MCP SDK's `CallToolRequest`/`CallToolResult` types
- Remove any remaining mock data handlers
- Ensure all tool calls use SDK standard types

### 2.2 Authentication Integration
- Move authentication logic to MCP SDK middleware
- Remove custom `AuthState` and authentication wrappers
- Update authentication flow to work with SDK request pipeline
- Maintain existing validation logic but adapt to SDK patterns

## Phase 3: Dependency Cleanup (Medium Priority)

### 3.1 Update Dependencies
- Update `go.mod` to ensure MCP SDK is primary MCP dependency
- Remove dependencies used only by custom implementation
- Update all imports throughout codebase to use SDK packages
- Run `go mod tidy` to clean up unused dependencies

### 3.2 File Structure Updates
- Consolidate remaining handlers into fewer files if appropriate
- Update any configuration files that reference removed components
- Ensure MCP server configuration works with SDK defaults

## Phase 4: Testing and Validation (Medium Priority)

### 4.1 Update Tests
- Remove tests for custom components that are being eliminated
- Update integration tests to work with MCP SDK architecture
- Ensure MCP protocol compliance testing
- Maintain test coverage for remaining functionality

### 4.2 Validation
- Verify all existing tools work identically after migration
- Test MCP protocol compliance with SDK
- Performance validation to ensure no regressions
- End-to-end testing with actual Forgejo repositories

## Success Criteria
- [ ] MCP server starts and runs using only SDK components
- [ ] All existing tools (PR listing, issue operations) function identically
- [ ] Codebase reduced by ~2000+ lines of custom implementation
- [ ] All tests pass with updated architecture
- [ ] MCP protocol compliance verified