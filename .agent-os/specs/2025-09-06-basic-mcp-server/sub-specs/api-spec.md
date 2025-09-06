# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-09-06-basic-mcp-server/spec.md

> Created: 2025-09-06
> Version: 1.0.0

## Endpoints

### MCP Protocol Tools

The MCP server implements the Model Context Protocol (MCP) for tool execution. All tools are accessed through the standard MCP protocol endpoints.

#### Tool Execution Endpoint
- **Method**: `tools/call`
- **Protocol**: MCP JSON-RPC 2.0
- **Description**: Execute a registered tool with provided parameters

#### Tool List Endpoint
- **Method**: `tools/list`
- **Protocol**: MCP JSON-RPC 2.0
- **Description**: Retrieve list of available tools

### Hello, World! Tool Specification

#### Tool Name
`hello_world`

#### Method
`tools/call` with tool name `hello_world`

#### Parameters
```json
{
  "name": "hello_world",
  "arguments": {
    "message": "optional custom message"
  }
}
```

- `message` (optional): Custom message to include in response. If not provided, defaults to "Hello, World!"

#### Response Format
```json
{
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Hello, World! [custom message if provided]"
      }
    ]
  }
}
```

#### Example Request
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "hello_world",
    "arguments": {
      "message": "from MCP Server"
    }
  }
}
```

#### Example Response
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Hello, World! from MCP Server"
      }
    ]
  }
}
```

## Error Handling

### Standard MCP Error Codes
- `-32601`: Method not found (tool not available)
- `-32602`: Invalid params (malformed tool arguments)
- `-32000`: Server error (internal server issues)

### Tool-Specific Errors
- **Invalid Message Format**: Returns error if message parameter is not a string
- **Tool Execution Failure**: Returns server error if tool execution fails internally

### Error Response Format
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "details": "Message must be a string"
    }
  }
}
```

## Controllers

### MCP Tool Controller
- **File**: `internal/controllers/mcp_controller.go`
- **Responsibilities**:
  - Handle MCP protocol requests
  - Route tool calls to appropriate handlers
  - Validate request parameters
  - Format responses according to MCP specification
  - Handle and format errors

### Hello World Tool Handler
- **File**: `internal/tools/hello_world.go`
- **Responsibilities**:
  - Process hello_world tool requests
  - Generate appropriate response messages
  - Validate input parameters
  - Return formatted tool results