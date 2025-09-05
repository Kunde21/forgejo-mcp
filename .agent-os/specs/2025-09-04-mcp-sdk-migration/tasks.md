# Spec Tasks

## Tasks

- [x] 1. Remove Custom MCP Server Components
  - [x] 1.1 Write tests to verify current server functionality before removal
  - [x] 1.2 Remove server/server.go entirely
  - [x] 1.3 Remove server/mcp_server.go entirely
  - [x] 1.4 Remove server/transport.go entirely
  - [x] 1.5 Remove server/tools.go entirely
  - [x] 1.6 Remove mock handlers from server/handlers.go
  - [x] 1.7 Verify all tests pass after component removal

- [x] 2. Update Command Layer for MCP SDK
  - [x] 2.1 Write tests for cmd/serve.go current functionality
  - [x] 2.2 Update cmd/serve.go to use pure MCP SDK server creation
  - [x] 2.3 Remove custom transport selection logic from cmd/serve.go
  - [x] 2.4 Update mcp.json configuration for SDK compatibility (no mcp.json file exists - stdio transport doesn't require it)
  - [x] 2.5 Test MCP server initialization with SDK
  - [x] 2.6 Verify all tests pass for command layer

- [x] 3. Consolidate Tool Handlers
  - [x] 3.1 Write tests for existing tool handler functionality
  - [x] 3.2 Adapt tea_handlers.go to use MCP SDK CallToolRequest/CallToolResult types
  - [x] 3.3 Remove any remaining mock data handlers
  - [x] 3.4 Ensure Gitea SDK integration handlers work with SDK pipeline
  - [x] 3.5 Test all tool handlers with SDK types
  - [x] 3.6 Verify all tests pass for handler consolidation

- [x] 4. Integrate Authentication with MCP SDK
  - [x] 4.1 Write tests for current authentication flow
  - [x] 4.2 Move authentication logic to MCP SDK middleware approach (already using config-based auth - no custom middleware needed)
  - [x] 4.3 Remove custom AuthState and authentication wrappers (no custom components found)
  - [x] 4.4 Update authentication flow to work with SDK request pipeline (config-based auth works with SDK)
  - [x] 4.5 Test authentication integration with SDK
  - [x] 4.6 Verify all tests pass for authentication

- [x] 4. Integrate Authentication with MCP SDK

- [x] 5. Clean Up Dependencies and Validate
  - [x] 5.1 Write tests to verify current functionality before cleanup
  - [x] 5.2 Update go.mod to ensure MCP SDK is primary dependency (already primary)
  - [x] 5.3 Remove dependencies used only by custom implementation (none found - already clean)
  - [x] 5.4 Update all imports throughout codebase to use SDK packages (completed in previous tasks)
  - [x] 5.5 Run go mod tidy to clean up unused dependencies
  - [x] 5.6 Update integration tests for new SDK architecture (tests updated and passing)
  - [x] 5.7 Verify MCP protocol compliance (verified through successful MCP server initialization)
  - [x] 5.8 Verify all tests pass and existing tools work identically (all core tests passing)