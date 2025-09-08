# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-09-08-mcp-sdk-migration/spec.md

> Created: 2025-09-08
> Status: Ready for Implementation

## Tasks

### Phase 1: Analysis and Planning

- [ ] **Analyze current MCP implementation** - Review all files using mark3labs/mcp-go to understand current usage patterns
  - [ ] Document all imported types and functions from mark3labs/mcp-go
  - [ ] Identify custom extensions or workarounds built on top of the old SDK
  - [ ] Map current tool definitions and their schemas
  - [ ] List all server initialization and configuration patterns

- [ ] **Study new SDK architecture** - Deep dive into github.com/modelcontextprotocol/go-sdk/mcp
  - [ ] Review new SDK documentation and examples
  - [ ] Identify breaking changes between SDKs
  - [ ] Map old SDK concepts to new SDK equivalents
  - [ ] Document new features and improvements available

- [ ] **Create migration mapping** - Build a comprehensive migration guide
  - [ ] Create type mapping table (old types → new types)
  - [ ] Document function signature changes
  - [ ] Identify deprecated patterns and their replacements
  - [ ] Note any functionality gaps that need custom solutions

### Phase 2: Migration Implementation

- [ ] **Update go.mod dependencies**
  - [ ] Remove `github.com/mark3labs/mcp-go` dependency
  - [ ] Add `github.com/modelcontextprotocol/go-sdk/mcp` dependency
  - [ ] Run `go mod tidy` to clean up dependencies
  - [ ] Verify no conflicting dependencies exist

- [ ] **Migrate server initialization** - Update server/server.go
  - [ ] Replace old MCP server initialization with new SDK pattern
  - [ ] Update server configuration to use new SDK options
  - [ ] Migrate transport layer setup (stdio/SSE)
  - [ ] Update error handling for server lifecycle

- [ ] **Migrate tool definitions** - Update all MCP tool implementations
  - [ ] Convert ListIssues tool to new SDK schema format
  - [ ] Update tool registration with new SDK methods
  - [ ] Migrate tool parameter validation logic
  - [ ] Update tool response formatting

- [ ] **Update handler implementations** - Migrate server/handlers.go
  - [ ] Convert handler functions to new SDK signatures
  - [ ] Update context handling for new SDK
  - [ ] Migrate error responses to new format
  - [ ] Update logging and debugging output

- [ ] **Migrate types and interfaces**
  - [ ] Update all type imports throughout the codebase
  - [ ] Convert custom types that extended old SDK types
  - [ ] Update interface implementations for new SDK contracts
  - [ ] Migrate any custom middleware or interceptors

- [ ] **Update configuration handling** - Migrate config package
  - [ ] Update configuration structures for new SDK requirements
  - [ ] Migrate environment variable handling if changed
  - [ ] Update config validation for new SDK constraints
  - [ ] Ensure backward compatibility where possible

### Phase 3: Testing and Validation

- [ ] **Update unit tests**
  - [ ] Fix all compilation errors in test files
  - [ ] Update test assertions for new SDK behavior
  - [ ] Add tests for new SDK features being utilized
  - [ ] Ensure all existing tests pass with new SDK

- [ ] **Update integration tests**
  - [ ] Migrate test harness to use new SDK client
  - [ ] Update integration test scenarios for new SDK
  - [ ] Verify end-to-end functionality with new SDK
  - [ ] Test error scenarios and edge cases

- [ ] **Manual testing checklist**
  - [ ] Test server startup and shutdown
  - [ ] Verify stdio transport functionality
  - [ ] Test SSE transport if applicable
  - [ ] Validate all tool executions with sample requests
  - [ ] Test error handling and recovery scenarios

- [ ] **Performance validation**
  - [ ] Compare memory usage before and after migration
  - [ ] Benchmark tool execution times
  - [ ] Validate concurrent request handling
  - [ ] Check for any resource leaks

- [ ] **Compatibility testing**
  - [ ] Test with various MCP clients (Claude Desktop, VS Code, etc.)
  - [ ] Verify protocol version compatibility
  - [ ] Test with different Forgejo/Gitea versions
  - [ ] Validate authentication and authorization flows

### Phase 4: Documentation and Cleanup

- [ ] **Update code documentation**
  - [ ] Update all godoc comments referencing old SDK
  - [ ] Add migration notes to relevant functions
  - [ ] Document any behavior changes for users
  - [ ] Update inline comments for clarity

- [ ] **Update README.md**
  - [ ] Update installation instructions if changed
  - [ ] Document new SDK requirements
  - [ ] Add migration guide for existing users
  - [ ] Update example configurations

- [ ] **Update configuration examples**
  - [ ] Update config.example.yaml with new SDK options
  - [ ] Document any new configuration parameters
  - [ ] Add migration notes for config changes
  - [ ] Provide upgrade path documentation

- [ ] **Update MCP manifest**
  - [ ] Update mcp.json with new SDK specifications
  - [ ] Verify tool schemas match new format
  - [ ] Update version information
  - [ ] Test manifest with MCP clients

- [ ] **Create migration guide**
  - [ ] Document step-by-step migration process for users
  - [ ] List breaking changes and solutions
  - [ ] Provide troubleshooting section
  - [ ] Include rollback instructions if needed

- [ ] **Code cleanup**
  - [ ] Remove any deprecated code or workarounds
  - [ ] Clean up unused imports and variables
  - [ ] Run linters and fix any issues
  - [ ] Format code with `goimports`

### Phase 5: Release Preparation

- [ ] **Version management**
  - [ ] Update version numbers in code
  - [ ] Create changelog entry for migration
  - [ ] Tag release candidate for testing
  - [ ] Prepare release notes

- [ ] **Final validation**
  - [ ] Run full test suite
  - [ ] Perform smoke tests on all platforms
  - [ ] Verify documentation completeness
  - [ ] Get code review approval

- [ ] **Release tasks**
  - [ ] Merge migration to main branch
  - [ ] Create GitHub release with notes
  - [ ] Update any external documentation
  - [ ] Notify users of migration completion

## Success Criteria

- All tests pass with new SDK
- No regression in functionality
- Improved or equivalent performance
- Clear documentation for users
- Smooth upgrade path for existing installations