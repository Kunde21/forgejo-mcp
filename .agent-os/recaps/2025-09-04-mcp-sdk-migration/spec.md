# Spec Requirements Document

> Spec: MCP SDK Migration
> Created: 2025-09-04

## Overview

Replace the current hybrid MCP implementation with the official MCP SDK to ensure full protocol compliance, reduce maintenance overhead, and leverage community-supported components while preserving all existing Forgejo integration functionality.

## User Stories

### MCP Protocol Compliance

As a developer maintaining the Forgejo MCP server, I want to use the official MCP SDK instead of custom implementation, so that I can ensure protocol compliance and benefit from community updates and bug fixes.

**Workflow:** The server currently uses a mix of custom MCP components and SDK elements. Migration will standardize on the official SDK, removing ~2000+ lines of custom code while maintaining all existing tool functionality.

### Simplified Maintenance

As a project maintainer, I want to eliminate custom MCP server code, so that I can focus on Forgejo-specific functionality rather than MCP protocol implementation details.

**Workflow:** Current codebase has extensive custom MCP implementation across multiple files. Migration consolidates to pure SDK usage, reducing complexity and maintenance burden.

## Spec Scope

1. **Server Architecture Migration** - Replace custom server components with pure MCP SDK
2. **Transport Layer Simplification** - Remove custom transport implementations
3. **Tool Handler Consolidation** - Standardize all tool handlers to use MCP SDK types
4. **Authentication Integration** - Move auth logic to MCP SDK middleware approach
5. **Dependency Cleanup** - Remove unnecessary dependencies and update go.mod

## Out of Scope

- Adding new MCP tools or features
- Changing existing tool functionality
- Modifying the tea CLI integration logic
- Updating configuration formats beyond MCP SDK requirements

## Expected Deliverable

1. MCP server runs using only official SDK components
2. All existing tools (PR listing, issue operations) work identically
3. Reduced codebase complexity with ~2000+ lines removed
4. Full MCP protocol compliance verified through testing