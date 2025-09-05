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

- [ ] 2. Update Command Layer for MCP SDK
  - [ ] 2.1 Write tests for cmd/serve.go current functionality
  - [ ] 2.1 Write tests for cmd/serve.go current functionality
  - [ ] 2.2 Update cmd/serve.go to use pure MCP SDK server creation
  - [ ] 2.3 Remove custom transport selection logic from cmd/serve.go
  - [ ] 2.4 Update mcp.json configuration for SDK compatibility
  - [ ] 2.5 Test MCP server initialization with SDK
  - [ ] 2.6 Verify all tests pass for command layer

- [ ] 3. Consolidate Tool Handlers
  - [ ] 3.1 Write tests for existing tool handler functionality
  - [ ] 3.2 Adapt tea_handlers.go to use MCP SDK CallToolRequest/CallToolResult types
  - [ ] 3.3 Remove any remaining mock data handlers
  - [ ] 3.4 Ensure Gitea SDK integration handlers work with SDK pipeline
  - [ ] 3.5 Test all tool handlers with SDK types
  - [ ] 3.6 Verify all tests pass for handler consolidation

- [ ] 4. Integrate Authentication with MCP SDK
  - [ ] 4.1 Write tests for current authentication flow
  - [ ] 4.2 Move authentication logic to MCP SDK middleware approach
  - [ ] 4.3 Remove custom AuthState and authentication wrappers
  - [ ] 4.4 Update authentication flow to work with SDK request pipeline
  - [ ] 4.5 Test authentication integration with SDK
  - [ ] 4.6 Verify all tests pass for authentication

- [ ] 5. Clean Up Dependencies and Validate
  - [ ] 5.1 Write tests to verify current functionality before cleanup
  - [ ] 5.2 Update go.mod to ensure MCP SDK is primary dependency
  - [ ] 5.3 Remove dependencies used only by custom implementation
  - [ ] 5.4 Update all imports throughout codebase to use SDK packages
  - [ ] 5.5 Run go mod tidy to clean up unused dependencies
  - [ ] 5.6 Update integration tests for new SDK architecture
  - [ ] 5.7 Verify MCP protocol compliance
  - [ ] 5.8 Verify all tests pass and existing tools work identically