# Spec Requirements Document

> Spec: MCP SDK Migration
> Created: 2025-09-08
> Status: Planning

## Overview

Migrate the forgejo-mcp server from the third-party `github.com/mark3labs/mcp-go` SDK to the official `github.com/modelcontextprotocol/go-sdk/mcp` SDK to ensure long-term stability and protocol compliance. This migration will align the project with the official Model Context Protocol implementation, providing better support and standardization.

## User Stories

### Developer Migration Story

As a developer maintaining forgejo-mcp, I want to migrate to the official MCP Go SDK, so that the project uses the standardized and officially supported implementation.

The migration process involves updating all import statements, adapting to any API differences between the two SDKs, updating the server initialization code, and ensuring all MCP tools continue to function correctly with the new SDK. This includes thorough testing of all existing functionality to ensure no regressions are introduced during the migration.

### AI Agent Compatibility Story

As an AI agent using forgejo-mcp, I want the server to use the official MCP SDK, so that I have consistent and reliable interaction with the standardized protocol implementation.

AI agents connecting to the forgejo-mcp server should experience no disruption in service or functionality. All existing tools and commands should continue to work exactly as before, but with potentially improved stability and protocol compliance from the official SDK.

## Spec Scope

1. **SDK Dependency Migration** - Replace all imports of `github.com/mark3labs/mcp-go` with `github.com/modelcontextprotocol/go-sdk/mcp` throughout the codebase
2. **Server Initialization Updates** - Adapt the MCP server initialization and configuration code to work with the official SDK's API
3. **Tool Registration Adaptation** - Update all tool registration and handler implementations to match the official SDK's interfaces
4. **Test Suite Updates** - Modify all test files to use the new SDK and ensure comprehensive test coverage
5. **Configuration Compatibility** - Ensure all existing configuration files and patterns continue to work with the new SDK

## Out of Scope

- Adding new MCP tools or features during the migration
- Changing the existing tool functionality or behavior
- Modifying the Gitea/Forgejo integration logic
- Altering the CLI interface or commands
- Updating documentation unrelated to the SDK change

## Expected Deliverable

1. Successfully compile and run the forgejo-mcp server using the official MCP Go SDK with all existing tools functioning correctly
2. Pass all existing tests and any new tests required for SDK-specific changes
3. Maintain backward compatibility for all configuration files and CLI commands

