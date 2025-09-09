# MCP SDK Migration Mapping

## Breaking Changes Between SDKs

### 1. Server Creation and Initialization
- **Old (mark3labs)**: `server.NewMCPServer(name string, version string) *MCPServer`
- **New (official)**: `mcp.NewServer(impl *Implementation, options *ServerOptions) *Server`
  - Requires Implementation struct with Name and Version fields
  - Supports ServerOptions for configuration

### 2. Tool Registration
- **Old**: `server.AddTool(tool *mcp.Tool, handler ToolHandler) error`
- **New**: `mcp.AddTool[In, Out any](s *Server, t *Tool, h ToolHandlerFor[In, Out])`
  - Generic function with type parameters for input/output
  - Handler signature includes typed arguments: `func(ctx context.Context, request *CallToolRequest, input In) (result *CallToolResult, output Out, error)`

### 3. Tool Definition and Schema
- **Old**: Tool options like `mcp.WithString()`, `mcp.WithNumber()` for schema definition
- **New**: Tool struct with Name, Description, and automatic schema generation from handler types
  - Input schema generated from handler's `In` type struct tags
  - Output schema generated from handler's `Out` type (if not `any`)

### 4. Handler Signatures
- **Old**: `func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)`
- **New**: `func(ctx context.Context, request *CallToolRequest, input In) (result *CallToolResult, output Out, error)`
  - Pointer to CallToolRequest
  - Typed input parameter
  - Separate result and output return values

### 5. Server Startup
- **Old**: `server.ServeStdio(mcpServer)`
- **New**: `server.Run(context.Background(), &mcp.StdioTransport{})`
  - Context-based execution
  - Transport passed to Run method

### 6. Result Construction
- **Old**: `mcp.NewToolResultText()`, `mcp.NewToolResultError()`, `mcp.NewToolResultStructured()`
- **New**: `*CallToolResult{Content: []Content{&TextContent{Text: text}}}`
  - Structured result with Content slice
  - Content types: TextContent, ImageContent, AudioContent, EmbeddedResource

### 7. Transport Layer
- **Old**: Implicit stdio transport via ServeStdio
- **New**: Explicit transport types: StdioTransport, SSEClientTransport, CommandTransport, etc.
  - More transport options available

## Type Mappings

### Core Types
- `server.MCPServer` → `mcp.Server`
- `mcp.CallToolRequest` → `*mcp.CallToolRequest` (now a pointer)
- `mcp.CallToolResult` → `*mcp.CallToolResult` (now a pointer)
- `mcp.Tool` → `*mcp.Tool` (now a pointer)

### Request/Response Types
- `mcp.CallToolParams` → `*mcp.CallToolParams` (pointer)
- `mcp.CallToolParamsRaw` → New type for raw parameters
- `mcp.ToolHandler` → `mcp.ToolHandlerFor[In, Out any]` (generic)

### Content Types
- Text results: `mcp.NewToolResultText(text)` → `&mcp.TextContent{Text: text}`
- Error results: `mcp.NewToolResultError(msg)` → `&mcp.TextContent{Text: msg}` with error annotation
- Structured results: `mcp.NewToolResultStructured(data, desc)` → `&mcp.EmbeddedResource{...}` or custom content

## New Features and Improvements

### 1. Generic Tool Handlers
- Type-safe input/output parameters
- Automatic JSON schema generation
- Compile-time type checking

### 2. Enhanced Transport Support
- SSE (Server-Sent Events) transport
- Command transport for subprocess communication
- Streamable HTTP transport
- In-memory transport for testing

### 3. Middleware Support
- `mcp.Middleware` type for request/response interception
- Chain multiple middlewares for logging, auth, etc.

### 4. Session Management
- `ServerSession` and `ClientSession` for connection state
- Better lifecycle management
- Concurrent connection support

### 5. Content Types
- Support for images, audio, and embedded resources
- Rich content in tool responses
- Better multimedia handling

### 6. Event Store
- `EventStore` interface for persistent event storage
- `MemoryEventStore` implementation
- Event-driven architecture support

### 7. Improved Error Handling
- Structured error types
- Better error propagation
- Resource not found helpers

### 8. Logging Integration
- `LoggingHandler` for MCP logging protocol
- Structured logging support
- Debug logging levels

## Migration Strategy

### Phase 1: Dependency Update
1. Update go.mod: Remove `github.com/mark3labs/mcp-go`, add `github.com/modelcontextprotocol/go-sdk/mcp`
2. Run `go mod tidy`
⚠️ **BLOCKED**: Official MCP Go SDK not available as Go module yet

### Phase 2: Server Initialization Migration (server/server.go)
1. **Update imports**:
   ```go
   // OLD
   "github.com/mark3labs/mcp-go/mcp"
   "github.com/mark3labs/mcp-go/server"

   // NEW
   "github.com/modelcontextprotocol/go-sdk/mcp"
   ```

2. **Update Server struct**:
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

3. **Update server creation**:
   ```go
   // OLD
   mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")

   // NEW
   mcpServer := mcp.NewServer(&mcp.Implementation{
       Name:    "forgejo-mcp",
       Version: "1.0.0",
   }, nil)
   ```

4. **Update server startup**:
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

### Phase 3: Tool Registration Migration
1. **Update tool registration**:
   ```go
   // OLD
   mcpServer.AddTool(mcp.NewTool("hello",
       mcp.WithDescription("Returns a hello world message"),
   ), s.handleHello)

   // NEW
   mcp.AddTool(mcpServer, &mcp.Tool{
       Name:        "hello",
       Description: "Returns a hello world message",
   }, s.handleHello)
   ```

2. **Update tool with parameters**:
   ```go
   // OLD
   mcpServer.AddTool(mcp.NewTool("list_issues",
       mcp.WithDescription("List issues from a Gitea/Forgejo repository"),
       mcp.WithString("repository",
           mcp.Required(),
           mcp.Description("Repository in format 'owner/repo'"),
       ),
       mcp.WithNumber("limit",
           mcp.DefaultNumber(15),
           mcp.Description("Maximum number of issues to return (1-100)"),
       ),
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

### Phase 4: Handler Implementation Migration (server/handlers.go)
1. **Update handler signatures**:
   ```go
   // OLD
   func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

   // NEW
   func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error)
   ```

2. **Update result construction**:
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

3. **Update complex results**:
   ```go
   // OLD
   return mcp.NewToolResultStructured(IssueList{Issues: issues}, fmt.Sprintf("Found %d issues", len(issues))), nil

   // NEW
   return &mcp.CallToolResult{
       Content: []mcp.Content{
           &mcp.TextContent{Text: fmt.Sprintf("Found %d issues", len(issues))},
           // For structured data, use appropriate content type
       },
   }, IssueList{Issues: issues}, nil
   ```

### Phase 5: Transport Migration
1. Replace `server.ServeStdio()` with `server.Run(ctx, &mcp.StdioTransport{})`
2. Update any transport-specific code

### Phase 6: Testing Migration
1. Update test harness to use new SDK types
2. Update mock implementations
3. Verify all tests pass

## Compatibility Notes

- **Protocol Compliance**: Official SDK maintains MCP protocol compliance
- **Backward Compatibility**: No backward compatibility with mark3labs SDK
- **Performance**: Official SDK may have better performance due to optimizations
- **Maintenance**: Official SDK has active maintenance and updates