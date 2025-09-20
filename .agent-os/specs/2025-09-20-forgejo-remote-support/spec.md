# Spec Requirements Document

> Spec: forgejo-remote-support
> Created: 2025-09-20

## Overview

Add official Forgejo SDK support to the MCP server while maintaining backward compatibility with the existing Gitea SDK implementation, enabling users to interact with Forgejo repositories through the official SDK. The implementation includes automatic remote type detection using the `/api/v1/version` endpoint to intelligently select the appropriate SDK client.

## User Stories

### Forgejo Repository Integration

As a Forgejo user, I want to use the MCP server with my Forgejo repositories, so that I can leverage AI agents to manage issues, pull requests, and comments through the official Forgejo SDK.

The workflow involves configuring the MCP server to use the Forgejo client type, which will use the official Forgejo SDK (`codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2`) instead of the Gitea SDK. Users will be able to specify the client type in their configuration, and the server will automatically instantiate the appropriate client while maintaining the same interface and functionality.

### Dual SDK Support

As a system administrator, I want to support both Gitea and Forgejo repositories in the same MCP server instance, so that I can migrate gradually and maintain compatibility with existing setups.

The implementation will allow configuration-based client selection with automatic remote type detection. Users can specify "auto" as the client type, and the server will automatically detect whether the remote instance is Gitea or Forgejo by querying the `/api/v1/version` endpoint. The server will then instantiate the appropriate SDK client, ensuring all existing functionality works identically regardless of the underlying SDK.

### Automatic Remote Type Detection

As a developer, I want the MCP server to automatically detect whether my remote instance is Gitea or Forgejo, so that I don't need to manually specify the client type and can avoid configuration errors.

The workflow involves the server automatically querying the `/api/v1/version` endpoint when the client type is set to "auto" or not specified. The server will parse the version response to determine if it's a Gitea or Forgejo instance based on version string patterns (e.g., "12.x.x" for Forgejo, "1.x.x" for Gitea, or explicit "forgejo"/"gitea" identifiers in the version string). This eliminates the need for manual client type specification and provides a seamless user experience.

## Spec Scope

1. **Forgejo SDK Integration** - Add the official Forgejo SDK as a dependency and implement a ForgejoClient that mirrors the existing GiteaClient functionality.
2. **Configuration Support** - Extend the configuration system to support client type selection between "gitea", "forgejo", and "auto".
3. **Automatic Remote Detection** - Implement version endpoint detection to automatically determine remote type when client type is set to "auto".
4. **Interface Compatibility** - Ensure the ForgejoClient implements the same ClientInterface as the existing GiteaClient.
5. **Backward Compatibility** - Maintain existing Gitea SDK functionality as the default behavior.
6. **Testing Framework** - Create comprehensive tests for the new ForgejoClient implementation and version detection logic.

## Out of Scope

- Breaking changes to existing API or tool interfaces
- Removal of the existing Gitea SDK implementation
- Changes to the MCP protocol or server architecture
- User interface changes or new MCP tools
- Database schema changes or migrations

## Expected Deliverable

1. A working ForgejoClient implementation that passes all existing tests and provides identical functionality to the GiteaClient.
2. Configuration support for selecting between Gitea, Forgejo, and auto-detection with appropriate validation and error handling.
3. Automatic remote type detection using the `/api/v1/version` endpoint with intelligent version string parsing.
4. Comprehensive test coverage for the ForgejoClient including unit tests and integration tests against mock Forgejo instances.
5. Updated documentation showing how to configure and use the Forgejo SDK support with automatic detection.
6. Enhanced config command that displays detected remote type and version information.