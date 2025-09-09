# Current MCP Implementation Analysis

## Imported Types and Functions from mark3labs/mcp-go

### Package: mcp
- **Types**:
  - CallToolRequest
  - CallToolResult
  - CallToolParams

- **Functions and Constructors**:
  - NewTool(name string, opts ...ToolOption) *Tool
  - WithDescription(desc string) ToolOption
  - WithString(name string, opts ...StringOption) ToolOption
  - Required() StringOption
  - Description(desc string) StringOption
  - WithNumber(name string, opts ...NumberOption) ToolOption
  - DefaultNumber(value int) NumberOption
  - ParseString(request CallToolRequest, param string, default string) string
  - ParseInt(request CallToolRequest, param string, default int) int
  - NewToolResultText(text string) *CallToolResult
  - NewToolResultError(message string) *CallToolResult
  - NewToolResultErrorFromErr(context string, err error) *CallToolResult
  - NewToolResultErrorf(format string, args ...interface{}) *CallToolResult
  - NewToolResultStructured(data any, description string) *CallToolResult

### Package: server
- **Types**:
  - MCPServer

- **Functions**:
  - NewMCPServer(name string, version string) *MCPServer
  - AddTool(tool *mcp.Tool, handler ToolHandler) error
  - ServeStdio(server *MCPServer) error

## Custom Extensions and Workarounds
- The Server struct wraps *server.MCPServer and adds custom fields: config *config.Config and giteaService *gitea.Service.
- Custom handler functions: handleHello and handleListIssues integrate Gitea service calls with MCP tool responses.
- Validation uses external library (go-ozzo/ozzo-validation) for tool parameters.
- No apparent workarounds; implementation follows SDK patterns but extends with Forgejo-specific logic.
- Error handling wraps errors with fmt.Errorf for context.

## Current Tool Definitions and Schemas
- **hello Tool**:
  - Description: Returns a hello world message
  - Parameters: None
  - Handler: handleHello - Returns static "Hello, World!" text response

- **list_issues Tool**:
  - Description: List issues from a Gitea/Forgejo repository
  - Parameters:
    - repository (string, required): Repository in format 'owner/repo'
    - limit (number, default 15, range 1-100): Maximum number of issues to return
    - offset (number, default 0, min 0): Number of issues to skip
  - Handler: handleListIssues - Parses params, validates, calls giteaService.ListIssues, returns structured IssueList{issues []gitea.Issue} with count description

## Server Initialization and Configuration Patterns
- **New()**: Loads config via config.Load(), calls NewFromConfig.
- **NewFromConfig(cfg *config.Config)**:
  - Validates config
  - Creates giteaClient with cfg.RemoteURL and cfg.AuthToken
  - Instantiates giteaService
  - Creates MCPServer with name "forgejo-mcp", version "1.0.0"
  - Registers tools: hello and list_issues with handlers
  - Returns wrapped Server struct
- **Start()**: Calls server.ServeStdio(mcpServer)
- **Stop()**: No-op (stdio transport runs until process ends)
- Configuration: Loaded from YAML/env, validated, used for Gitea client init. MCP server uses fixed name/version, no additional config passed to SDK.
