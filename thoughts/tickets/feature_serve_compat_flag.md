---
type: feature
priority: high
created: 2025-01-14T10:00:00Z
created_by: Opus
status: implemented
tags: [cli, serve, compatibility, mcp, responses]
keywords: [serve command, --compat flag, structured data, text response, tool responses, mcp client capabilities, cobra cli]
patterns: [cli flag patterns, response formatting, tool response handling, mcp protocol]
---

# FEATURE-001: Add --compat flag to serve command for response format control

## Description
Add a `--compat` flag to the `forgejo-mcp serve` command that controls whether structured data is included in text responses for tool calls. When the flag is present (compatibility mode), structured data will be included in the text response (current behavior). When the flag is not present (default), structured data will only be returned as an object, preventing duplicate data in updated MCP clients.

## Context
Updated MCP clients receive duplicate data in both text and structured object responses, polluting the context. This feature provides a compatibility mode for older clients that rely on text data only, while establishing the new default behavior of returning structured data only as objects.

## Requirements
- Add `--compat` flag to serve command
- When `--compat` is present: Include structured data in text response (current behavior)
- When `--compat` is absent: Exclude structured data from text response, return only as object
- Apply only to successful tool responses
- No changes to non-tool responses or error responses
- Default behavior: No structured data in text (new default)

### Functional Requirements
- [x] Implement CLI flag using Cobra framework
- [x] Modify tool response formatting based on flag state
- [x] Update help text with usage examples
- [x] Validate client capability to handle structured responses when possible

### Non-Functional Requirements
- No performance impact
- No logging or debug messages (stdio is used for communication)
- Server-wide setting, not per-request
- No conflicts with existing flags

## Current State
- All tool responses include structured data in both text and object format
- No mechanism to control response format
- Updated MCP clients receive duplicate data

## Desired State
- Default: Structured data only in object response
- Compatibility mode (--compat): Structured data in both text and object response
- Clear documentation for users on when to use the flag

## Research Context
Research needed on MCP client capability detection and current response implementation.

### Keywords to Search
- serve command - CLI command implementation location
- --compat flag - New flag to implement
- structured data - Current response format implementation
- text response - Response formatting logic
- tool responses - Target response type handling
- mcp client capabilities - Auto-detection research
- cobra cli - CLI framework patterns

### Patterns to Investigate
- cli flag patterns - How other flags are implemented in the codebase
- response formatting - How responses are currently structured
- tool response handling - Where tool responses are generated
- mcp protocol - Client capability detection methods

### Key Decisions Made
- Default behavior changed to exclude structured data from text
- Flag name: --compat (for compatibility mode)
- Scope: Tool responses only, not errors or non-tool responses
- Implementation: CLI flag only, no config file or environment variables
- Validation: Check if client can handle structured responses when possible
- Documentation: Update README with guidance for users

## Success Criteria

### Automated Verification
- [x] Tests pass with --compat flag (current behavior)
- [x] Tests pass without --compat flag (new behavior)
- [x] Acceptance tests updated to handle both scenarios
- [x] No regression in non-tool responses

### Manual Verification
- [x] Help text shows --compat flag with description
- [x] With --compat: Structured data appears in text response
- [x] Without --compat: Structured data only in object response
- [x] Error responses unchanged in both modes

## Related Information
- Version tags will handle the breaking change communication
- README needs update with migration guidance
- Research needed on MCP client capability auto-detection

## Notes
- Research web for MCP client capability detection methods
- Consider if server can validate client capability to handle structured responses
- Update documentation to guide users who don't see structured data in their agent