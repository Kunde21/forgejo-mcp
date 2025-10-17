---
date: 2025-10-14T17:31:16+07:00
git_commit: 812d3efde62966253e48abba52bb0122337f28e2
branch: master
repository: forgejo-mcp
topic: "Add --compat flag to serve command for response format control"
tags: [research, codebase, cli, serve, compatibility, mcp, responses]
last_updated: 2025-10-14
---

## Ticket Synopsis
The ticket requests adding a `--compat` flag to the `forgejo-mcp serve` command to control whether structured data is included in text responses for tool calls. This addresses duplicate data issues with updated MCP clients that receive both text and structured object responses.

## Summary
The research reveals a clear implementation path for the `--compat` flag. The codebase already has established patterns for CLI flags, response formatting, and server configuration. The main implementation will involve:
1. Adding the flag to `cmd/serve.go` following existing patterns
2. Extending the server configuration to include the compatibility setting
3. Modifying response helpers in `server/common.go` to conditionally include structured data in text
4. No reliable MCP client capability detection exists, making the manual flag approach necessary

## Detailed Findings

### CLI Implementation Structure
The CLI follows Cobra framework patterns with clear separation of concerns:

**Serve Command Location**: `cmd/serve.go:20-37`
- Already has three flags: `--host`, `--port`, `--debug`
- Uses `cmd.Flags().Bool()` for boolean flags
- Pattern: Define flag → Extract in `runServe()` → Pass to server constructor

**Flag Pattern Example** (`cmd/serve.go:32-34`):
```go
cmd.Flags().String("host", "localhost", "Host to bind the server to")
cmd.Flags().Int("port", 3000, "Port to bind the server to")
cmd.Flags().Bool("debug", false, "Enable debug mode (exposes hello tool)")
```

### Current Response Implementation
The codebase uses consistent response helpers that create the duplicate data issue:

**Response Helpers** (`server/common.go:13-27`):
- `TextResult()` and `TextResultf()` create text-only responses
- `TextError()` and `TextErrorf()` create error responses
- All return `*mcp.CallToolResult` with text content

**Current Pattern with Structured Data** (`server/pr_list.go:106-117`):
```go
// Builds human-readable text containing the same data
resultText := fmt.Sprintf("Found %d pull requests:\n", len(pullRequests))
for _, pr := range pullRequests {
    resultText += fmt.Sprintf("- #%d: %s (%s)\n", pr.Number, pr.Title, pr.State)
}

// Returns both text and structured data
return TextResult(resultText), &PullRequestList{PullRequests: pullRequests}, nil
```

This pattern causes the duplication - structured data appears in both the text summary and the structured object.

### Server Configuration Flow
The server has a well-defined configuration chain that can accommodate the new flag:

**Configuration Flow**:
1. CLI flags extracted in `cmd/serve.go:45-60`
2. Server created via `server.NewWithDebug(debug)` (`cmd/serve.go:65`)
3. Config loaded via `config.Load()` (`server/server.go:43`)
4. Server initialized with config (`server/server.go:61-97`)

**Missing Link**: Currently only the `debug` flag is passed to the server. Host and port flags are defined but not used.

### MCP Client Capability Detection
Research revealed no standard capability detection for structured responses:

**Key Findings**:
- MCP specification doesn't define a standard capability for structured response support
- `structuredContent` is part of the base MCP specification, not optional
- No reliable way to detect if a client can handle structured responses
- Experimental capabilities could be used but require client cooperation

**Conclusion**: Manual flag approach is appropriate given the lack of standard capability detection.

## Code References

### CLI Implementation Files
- `cmd/serve.go:32-34` - Existing flag definitions
- `cmd/serve.go:45-60` - Flag extraction pattern
- `cmd/serve.go:65-68` - Server creation with debug flag
- `cmd/root.go:66-67` - Persistent flag patterns

### Response Implementation Files
- `server/common.go:13-27` - Response helper functions
- `server/pr_list.go:106-117` - Example of duplicate data pattern
- `server/pr_fetch.go:131-144` - Complex text formatting example
- `server/server.go:128-199` - Tool registration patterns

### Configuration Files
- `config/config.go:9-22` - Config struct definition
- `config/config.go:25-34` - Default values pattern
- `config/config.go:37-41` - Environment variable binding
- `server/server.go:42-48` - Server constructor chain

## Architecture Insights

### Response Formatting Architecture
The current architecture has three layers:
1. **Helper Layer** (`server/common.go`): Standardized response creation
2. **Handler Layer** (various `server/*.go`): Business logic and response building
3. **Protocol Layer** (MCP SDK): Actual response transmission

The duplication occurs at the Handler Layer where both human-readable text and structured data are created from the same source data.

### Configuration Architecture
The configuration system supports multiple sources with clear priority:
1. Defaults in `config/config.go`
2. Optional YAML config file
3. Environment variables
4. CLI flags (highest priority)

This architecture can easily accommodate a new `compat` setting at any level.

### CLI Architecture
The CLI follows clean Cobra patterns:
- Each command in its own file
- Clear separation between flag definition and usage
- Consistent error handling with wrapped errors
- Graceful shutdown with signal handling

## Historical Context (from thoughts/)

### Previous Response Research
- `thoughts/research/2025-10-14_pr_fetch_implementation.md` - Documents the TextResult/TextError pattern
- `thoughts/research/2025-10-06_pr_creation_implementation.md` - Shows structured data handling
- `thoughts/research/2025-10-05_issue_creation_implementation.md` - MCP protocol layer documentation

### Implementation Patterns
- `thoughts/plans/pr_fetch_tool_implementation.md` - Tool response handling patterns
- `thoughts/architecture/go-style.md` - Response helpers and error handling
- `thoughts/architecture/go-test-style.md` - Testing patterns for responses

### Related Tickets
- `thoughts/tickets/feature_serve_compat_flag.md` - This ticket, already analyzed

## Related Research
- `thoughts/research/2025-10-14_pr_fetch_implementation.md` - Tool response patterns
- `thoughts/research/2025-10-06_pr_creation_implementation.md` - Response formatting
- `thoughts/research/2025-10-05_issue_creation_implementation.md` - MCP protocol basics

## Open Questions

1. **Default Behavior**: Should the default be compatibility mode (current behavior) or new behavior (structured only)? The ticket specifies new behavior as default.

2. **Migration Strategy**: How to communicate this breaking change? Version tags mentioned in ticket.

3. **Text-Only Tools**: How should tools that don't have structured data behave? (e.g., hello tool)

4. **Error Responses**: Should error responses also be affected? Ticket says no changes to error responses.

5. **Testing Coverage**: Need to update tests to verify both modes work correctly.

## Implementation Recommendation

Based on the research, here's the recommended implementation approach:

1. **Add CLI Flag**: Follow existing pattern in `cmd/serve.go`
2. **Extend Config**: Add `IncludeStructuredText bool` to `config.Config`
3. **Modify Response Helpers**: Add compatibility-aware versions or modify existing ones
4. **Update Tool Handlers**: Conditionally build text based on compatibility setting
5. **Pass Flag Through Server**: Update constructor chain to pass compatibility setting

The implementation should be straightforward given the existing patterns and clear architecture.