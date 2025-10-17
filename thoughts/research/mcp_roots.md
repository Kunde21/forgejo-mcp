# MCP Server Library Root Path Management Investigation

## Overview

Investigation of how the MCP Go SDK handles client root path management and how the forgejo-mcp server can leverage this functionality.

## Root Structure in MCP Go SDK

The `Root` struct in the MCP Go SDK is defined as:
```go
type Root struct {
    Meta             `json:"_meta,omitempty"`  // Optional metadata
    Name string       `json:"name,omitempty"`  // Human-readable identifier
    URI  string       `json:"uri"`             // Must start with file:// for now
}
```

## Client-Side Root Management

### Adding Roots
```go
client.AddRoots(roots ...*Root)
```
- Adds specified roots, replacing any with the same URIs
- Notifies connected servers of the change

### Removing Roots
```go
client.RemoveRoots(uris ...string)
```
- Removes roots by URI
- Notifies connected servers if the list has changed
- No error if a root doesn't exist

## Server-Side Root Change Handling

### Server Options Configuration
```go
serverOptions := &mcp.ServerOptions{
    RootsListChangedHandler: func(ctx context.Context, req *mcp.RootsListChangedRequest) {
        // Handle root changes from client
        // This is called when client sends roots/list_changed notification
    },
}
```

### Accessing Current Roots
The server can request the current roots from the client:
```go
roots, err := serverSession.ListRoots(ctx, &mcp.ListRootsParams{})
```

## Current State in forgejo-mcp

The current forgejo-mcp server implementation:
- ✅ Uses the MCP Go SDK v0.4.0
- ✅ Has the infrastructure for root management
- ❌ Does not configure a `RootsListChangedHandler`
- ❌ Does not actively use client root information in tool operations

## How Root Updates Work

1. **Client adds/removes roots:** `client.AddRoots()` or `client.RemoveRoots()`
2. **Client notifies server:** Automatic notification sent to connected servers
3. **Server receives notification:** `RootsListChangedHandler` is invoked (if configured)
4. **Server can query roots:** Use `serverSession.ListRoots()` to get current root list
5. **Server uses root info:** Apply root context to tool operations (e.g., default repository paths)

## Integration Opportunities for forgejo-mcp

The forgejo-mcp server could leverage root management to:
- Use client-provided roots as default working directories for repository resolution
- Automatically detect repositories in client-specified root paths
- Provide context-aware tool operations based on the client's workspace structure
- Eliminate the need for manual directory parameters in some tools

## Example Implementation

To add root support to forgejo-mcp:
```go
// In server.go
serverOptions := &mcp.ServerOptions{
    RootsListChangedHandler: func(ctx context.Context, req *mcp.RootsListChangedRequest) {
        // Update repository resolver with new root paths
        roots, _ := req.Session.ListRoots(ctx, &mcp.ListRootsParams{})
        // Store roots for use in tool operations
    },
}

// When creating server
mcpServer := mcp.NewServer(&mcp.Implementation{
    Name:    "forgejo-mcp", 
    Version: "1.0.0",
}, serverOptions)
```

## Key Findings

1. **Comprehensive SDK Support:** The MCP Go SDK provides complete root path management capabilities
2. **Automatic Notifications:** Root changes automatically trigger notifications to connected servers
3. **Bidirectional Communication:** Both client and server can initiate root-related operations
4. **Current Gap:** forgejo-mcp doesn't currently utilize this built-in functionality
5. **Integration Potential:** Significant opportunity to improve user experience by leveraging client workspace context

## Recommendations

1. **Implement RootsListChangedHandler** to track client root changes
2. **Enhance RepositoryResolver** to use client-provided roots as default search paths
3. **Update tool handlers** to leverage root context for automatic repository detection
4. **Consider root-based caching** for improved performance in repository operations
5. **Add root-aware error handling** to provide better context when operations fail

This investigation shows that the MCP Go SDK provides comprehensive root path management capabilities that the forgejo-mcp server could leverage to provide more integrated and context-aware repository operations.