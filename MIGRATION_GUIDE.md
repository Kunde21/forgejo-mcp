# MCP SDK Migration Guide

## Overview

This guide documents the migration from the third-party `mark3labs/mcp-go` SDK to the official `github.com/modelcontextprotocol/go-sdk/mcp` SDK. This migration improves protocol compliance, provides long-term stability, and ensures official support for the MCP protocol implementation.

## Migration Timeline

- **Migration Completed**: September 8, 2025
- **SDK Version**: `github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`
- **Previous SDK**: `github.com/mark3labs/mcp-go` (deprecated)

## What Changed

### Key Improvements

1. **Official Protocol Support**: Full compliance with MCP protocol specifications
2. **Enhanced Type Safety**: Generic tool handlers with compile-time type checking
3. **Better Performance**: Optimized implementation with improved resource usage
4. **Active Maintenance**: Official SDK with guaranteed long-term support
5. **Rich Content Types**: Support for images, audio, and embedded resources
6. **Advanced Transport Options**: SSE, Command, and In-memory transports

### Breaking Changes

#### 1. Server Initialization
```go
// OLD (mark3labs)
mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")

// NEW (official)
mcpServer := mcp.NewServer(&mcp.Implementation{
    Name:    "forgejo-mcp",
    Version: "1.0.0",
}, nil)
```

#### 2. Tool Registration
```go
// OLD
mcpServer.AddTool(mcp.NewTool("list_issues",
    mcp.WithDescription("List issues from a repository"),
    mcp.WithString("repository", mcp.Required(), mcp.Description("Repository in format 'owner/repo'")),
    mcp.WithNumber("limit", mcp.DefaultNumber(15), mcp.Description("Maximum number of issues to return")),
), s.handleListIssues)

// NEW
type ListIssuesArgs struct {
    Repository string `json:"repository" jsonschema:"description=Repository in format 'owner/repo'"`
    Limit      int    `json:"limit" jsonschema:"description=Maximum number of issues to return (1-100),default=15"`
    Offset     int    `json:"offset" jsonschema:"description=Number of issues to skip (0-based),default=0"`
}

mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "list_issues",
    Description: "List issues from a Gitea/Forgejo repository",
}, func(ctx context.Context, req *mcp.CallToolRequest, args ListIssuesArgs) (*mcp.CallToolResult, any, error) {
    return s.handleListIssues(ctx, req, args)
})
```

#### 3. Handler Signatures
```go
// OLD
func (s *Server) handleListIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// NEW
func (s *Server) handleListIssues(ctx context.Context, request *mcp.CallToolRequest, args ListIssuesArgs) (*mcp.CallToolResult, any, error)
```

#### 4. Result Construction
```go
// OLD
return mcp.NewToolResultText("Hello, World!"), nil

// NEW
return &mcp.CallToolResult{
    Content: []mcp.Content{
        &mcp.TextContent{Text: "Hello, World!"},
    },
}, nil, nil
```

#### 5. Server Startup
```go
// OLD
func (s *Server) Start() error {
    return server.ServeStdio(s.mcpServer)
}

// NEW
func (s *Server) Start() error {
    return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
```

## Step-by-Step Migration Process

### Phase 1: Preparation
1. **Backup your code**: Create a git branch or backup before starting
2. **Review dependencies**: Check go.mod for any custom MCP-related dependencies
3. **Update Go version**: Ensure Go 1.24.6 or later is installed

### Phase 2: Dependency Update
1. **Update go.mod**:
   ```bash
   go mod edit -dropreplace github.com/mark3labs/mcp-go
   go mod tidy
   ```
2. **Verify new dependency**:
   ```bash
   go list -m github.com/modelcontextprotocol/go-sdk/mcp
   ```

### Phase 3: Code Migration

#### Step 1: Update Imports
Replace all occurrences of:
```go
"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
```

With:
```go
"github.com/modelcontextprotocol/go-sdk/mcp"
```

#### Step 2: Update Server Types
```go
// OLD
type Server struct {
    mcpServer    *server.MCPServer
    config       *config.Config
    giteaService *gitea.Service
}

// NEW
type Server struct {
    mcpServer    *mcp.Server
    config       *config.Config
    giteaService *gitea.Service
}
```

#### Step 3: Update Server Creation
```go
// OLD
func NewServer(config *config.Config, giteaService *gitea.Service) (*Server, error) {
    mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")
    // ...
}

// NEW
func NewServer(config *config.Config, giteaService *gitea.Service) (*Server, error) {
    mcpServer := mcp.NewServer(&mcp.Implementation{
        Name:    "forgejo-mcp",
        Version: "1.0.0",
    }, nil)
    // ...
}
```

#### Step 4: Update Tool Registration
For each tool, create argument structs and update registration:

```go
// Define argument struct
type ToolArgs struct {
    Param1 string `json:"param1" jsonschema:"description=Description of param1"`
    Param2 int    `json:"param2" jsonschema:"description=Description of param2,default=10"`
}

// Update registration
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "tool_name",
    Description: "Tool description",
}, func(ctx context.Context, req *mcp.CallToolRequest, args ToolArgs) (*mcp.CallToolResult, any, error) {
    return s.handleTool(ctx, req, args)
})
```

#### Step 5: Update Handler Signatures
```go
// OLD
func (s *Server) handleTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Extract parameters manually
    params := request.Params
    param1 := params["param1"].(string)
    // ...
}

// NEW
func (s *Server) handleTool(ctx context.Context, request *mcp.CallToolRequest, args ToolArgs) (*mcp.CallToolResult, any, error) {
    // Parameters are already typed
    param1 := args.Param1
    // ...
}
```

#### Step 6: Update Result Construction
```go
// OLD
return mcp.NewToolResultText("Success message"), nil

// NEW
return &mcp.CallToolResult{
    Content: []mcp.Content{
        &mcp.TextContent{Text: "Success message"},
    },
}, nil, nil
```

#### Step 7: Update Server Startup
```go
// OLD
func (s *Server) Start() error {
    return server.ServeStdio(s.mcpServer)
}

// NEW
func (s *Server) Start() error {
    return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
```

### Phase 4: Testing and Validation

#### Step 1: Build Verification
```bash
go build ./...
```

#### Step 2: Run Tests
```bash
go test ./...
```

#### Step 3: Manual Testing
1. Start the server: `./forgejo-mcp serve`
2. Test each tool with sample requests
3. Verify error handling and edge cases
4. Test with MCP clients (Claude Desktop, VS Code, etc.)

### Phase 5: Cleanup
1. **Remove deprecated code**: Clean up any workarounds or deprecated patterns
2. **Update documentation**: Update README.md and inline comments
3. **Format code**: Run `goimports -w .`
4. **Lint check**: Run `go vet ./...`

## Troubleshooting

### Common Issues

#### 1. Import Errors
**Problem**: `cannot find package "github.com/modelcontextprotocol/go-sdk/mcp"`
**Solution**:
```bash
go mod tidy
go mod download
```

#### 2. Type Mismatches
**Problem**: `cannot use mcpServer (type *mcp.Server) as type *server.MCPServer`
**Solution**: Update the Server struct field type as shown in migration examples

#### 3. Handler Signature Errors
**Problem**: `wrong number of arguments to handler function`
**Solution**: Update handler signatures to match the new generic pattern with typed arguments

#### 4. Result Construction Errors
**Problem**: `undefined: mcp.NewToolResultText`
**Solution**: Use the new `&mcp.CallToolResult{Content: []mcp.Content{...}}` pattern

#### 5. Server Startup Errors
**Problem**: `undefined: server.ServeStdio`
**Solution**: Use `mcpServer.Run(context.Background(), &mcp.StdioTransport{})`

### Rollback Instructions

If you need to rollback to the previous SDK version:

1. **Revert code changes**:
   ```bash
   git checkout <previous-commit>
   ```

2. **Restore dependencies**:
   ```bash
   go mod edit -require github.com/mark3labs/mcp-go@<previous-version>
   go mod tidy
   ```

3. **Rebuild and test**:
   ```bash
   go build ./...
   go test ./...
   ```

## Compatibility Matrix

| Component | Old SDK | New SDK | Status |
|-----------|---------|---------|--------|
| Protocol Compliance | Partial | Full | ✅ Improved |
| Tool Registration | Manual | Generic | ✅ Enhanced |
| Type Safety | Runtime | Compile-time | ✅ Improved |
| Content Types | Basic | Rich | ✅ Enhanced |
| Transport Options | Stdio only | Multiple | ✅ Enhanced |
| Error Handling | Basic | Structured | ✅ Improved |
| Performance | Good | Better | ✅ Improved |
| Maintenance | Community | Official | ✅ Guaranteed |

## Benefits of Migration

1. **Future-Proof**: Official SDK ensures compatibility with MCP protocol updates
2. **Better Performance**: Optimized implementation with lower resource usage
3. **Enhanced Security**: Official maintenance includes security patches
4. **Rich Features**: Support for advanced content types and transport options
5. **Type Safety**: Compile-time guarantees prevent runtime errors
6. **Active Development**: Regular updates and bug fixes from official maintainers

## Support

If you encounter issues during migration:

1. **Check this guide**: Review the troubleshooting section above
2. **Review examples**: Look at the updated code in the repository
3. **Test incrementally**: Migrate and test one component at a time
4. **Community support**: Check GitHub issues for similar problems

## Next Steps

After successful migration:

1. **Monitor performance**: Compare resource usage before/after migration
2. **Update client configurations**: Ensure MCP clients are compatible
3. **Review tool schemas**: Verify all tool definitions are correct
4. **Plan future updates**: Stay current with official SDK releases

---

*This migration guide was created as part of the MCP SDK migration completed on September 8, 2025.*