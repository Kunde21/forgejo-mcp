## Migration Plan: Replace Tea CLI with Gitea SDK

### Executive Summary

This plan outlines the migration from the current tea CLI wrapper approach to direct Gitea SDK integration. The migration will improve performance, reliability, and maintainability while providing better type safety and error handling.

### Current State Analysis

Current Implementation:

• Uses tea CLI tool via exec.Command() calls
• Complex output parsing (JSON/text fallback)
• Command building with shell escaping
• Timeout handling for CLI execution
• Supports PR and issue listing with basic filters

Key Issues with Current Approach:

• Process spawning overhead
• Complex parsing logic prone to breakage
• Limited error handling from CLI output
• Dependency on external tea binary
• No type safety for API responses

### Target State

Gitea SDK Integration:

• Direct HTTP API calls using code.gitea.io/sdk/gitea
• Native Go structs and error handling
• Rich filtering and pagination support
• Better authentication handling
• Type-safe API interactions

### Migration Strategy

#### Phase 1: Foundation (High Priority)

1. Update Dependencies
 • Add code.gitea.io/sdk/gitea v0.22.0 to go.mod
 • Remove tea-related dependencies
 • Update go.sum
2. Create Gitea Client Infrastructure
 • Implement Gitea client factory
 • Add client configuration and initialization
 • Set up proper authentication handling
3. Update Configuration
 • Remove TeaPath configuration
 • Add Gitea SDK specific settings
 • Update environment variable names
 • Maintain backward compatibility where possible


#### Phase 2: Core Implementation (High Priority)

4. Refactor Server Handlers
 • Replace TeaPRListHandler with GiteaPRListHandler
 • Replace TeaIssueListHandler with GiteaIssueListHandler
 • Remove tea command builders and parsers
 • Implement direct SDK method calls
5. Update Tool Registration
 • Modify registerTools() to use new handlers
 • Update MCP tool definitions if needed
 • Ensure backward compatibility with MCP protocol
6. Authentication Updates
 • Implement token-based authentication for SDK
 • Remove tea-specific authentication logic
 • Add proper error handling for auth failures


#### Phase 3: Testing and Validation (Medium Priority)

7. Update Unit Tests
 • Replace mocked tea command execution with SDK mocks
 • Update test data structures to match SDK types
 • Add integration tests for SDK calls
8. Update Integration Tests
 • Modify end-to-end tests to work with SDK
 • Test against real Forgejo/Gitea instances
 • Validate error handling and edge cases


#### Phase 4: Documentation and Cleanup (Medium Priority)

9. Update Documentation
 • Modify README.md to reflect SDK usage
 • Update setup instructions
 • Remove tea CLI installation requirements
 • Add SDK-specific configuration examples
10. Update Project Documentation
 • Modify PHASE1_TASKS.md to reflect completed migration
 • Update any inline code comments
 • Document SDK-specific features and benefits


### Technical Implementation Details

#### SDK Client Setup

// New Gitea client initialization
client, err := gitea.NewClient(cfg.ForgejoURL, gitea.SetToken(cfg.AuthToken))
if err != nil {
    return nil, fmt.Errorf("failed to create Gitea client: %w", err)
}

#### Handler Migration Example

// Before: Tea CLI approach
cmd := h.commandBuilder.BuildPRListCommand(params)
output, err := h.executor.ExecuteCommand(ctx, cmd)
prs, err := h.parser.ParsePRList(output)

// After: SDK approach
opts := gitea.ListPullRequestsOptions{
    State: gitea.StateOpen,
    // ... other options
}
prs, _, err := client.ListRepoPullRequests(owner, repo, opts)

#### Configuration Changes

# Before
forgejo_url: "https://example.forgejo.com"
auth_token: "token"
tea_path: "tea"

# After
forgejo_url: "https://example.forgejo.com"
auth_token: "token"
# tea_path removed

### Benefits of Migration

1. Performance Improvements
 • No process spawning overhead
 • Direct HTTP calls with connection reuse
 • Better memory usage
2. Reliability Enhancements
 • Type-safe API interactions
 • Structured error handling
 • No parsing failures from CLI output
3. Maintainability
 • Less complex codebase
 • Better testability
 • Official SDK support
4. Feature Richness
 • Full API coverage
 • Advanced filtering options
 • Better pagination support


### Risk Mitigation

1. Backward Compatibility
 • Maintain existing MCP tool interfaces
 • Keep environment variable names consistent
 • Provide migration guide for users
2. Testing Strategy
 • Comprehensive unit test coverage
 • Integration tests with real instances
 • Gradual rollout with feature flags if needed
3. Fallback Plan
 • Keep tea CLI as optional fallback
 • Allow configuration to switch between implementations
 • Easy rollback if issues arise


### Success Criteria

[ ] All existing MCP tools work with SDK
[ ] Performance improvement of at least 50%
[ ] No breaking changes to MCP protocol
[ ] All tests pass with >80% coverage
[ ] Documentation updated and accurate
[ ] End-to-end testing successful

### Timeline Estimate

• Phase 1: 1-2 days (Foundation)
• Phase 2: 2-3 days (Core Implementation)
• Phase 3: 1-2 days (Testing)
• Phase 4: 1 day (Documentation)

Total: 5-8 days for complete migration

This migration will significantly improve the project's architecture while maintaining full functionality and backward compatibility.
