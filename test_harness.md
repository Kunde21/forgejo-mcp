# MCP Server Test Harness Design

## Overview

This document outlines a comprehensive test harness for the Forgejo MCP server using Go's `testing` package, `exec.CommandContext`, and subprocess management. The harness enables integration testing of the MCP server's external interface through the stdio transport protocol.

## Architecture

### Core Components

1. **TestServer struct** - Manages the MCP server subprocess lifecycle
2. **MCP Protocol Client** - Handles JSON-RPC 2.0 communication
3. **Test Utilities** - Helper functions for common test operations

### TestServer Implementation

```go
type TestServer struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
    ctx    context.Context
    cancel context.CancelFunc
    t      *testing.T
}

func NewTestServer(t *testing.T) *TestServer {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

    cmd := exec.CommandContext(ctx, "go", "run", "main.go")
    cmd.Dir = "." // Set working directory to project root

    stdin, err := cmd.StdinPipe()
    if err != nil {
        t.Fatalf("Failed to create stdin pipe: %v", err)
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        t.Fatalf("Failed to create stdout pipe: %v", err)
    }

    return &TestServer{
        cmd:    cmd,
        stdin:  stdin,
        stdout: stdout,
        ctx:    ctx,
        cancel: cancel,
        t:      t,
    }
}
```

## Test Scenarios

### 1. Basic Server Lifecycle

```go
func TestServerLifecycle(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    // Start server
    err := ts.Start()
    if err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }

    // Verify server is running
    if !ts.IsRunning() {
        t.Fatal("Server process is not running")
    }

    // Test graceful shutdown
    ts.Close()
    if ts.IsRunning() {
        t.Fatal("Server did not shut down gracefully")
    }
}
```

### 2. MCP Protocol Initialization

```go
func TestMCPInitialization(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    ts.Start()

    // Send initialize request
    initReq := &InitializeRequest{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "initialize",
        Params: InitializeParams{
            ProtocolVersion: "2024-11-05",
            Capabilities:    ClientCapabilities{},
            ClientInfo: ClientInfo{
                Name:    "test-client",
                Version: "1.0.0",
            },
        },
    }

    resp, err := ts.SendRequest(initReq)
    if err != nil {
        t.Fatalf("Failed to send initialize request: %v", err)
    }

    // Verify response
    if resp.Result == nil {
        t.Fatal("Expected result in initialize response")
    }

    serverInfo := resp.Result.ServerInfo
    if serverInfo.Name != "forgejo-mcp" {
        t.Errorf("Expected server name 'forgejo-mcp', got %s", serverInfo.Name)
    }
}
```

### 3. Tool Discovery

```go
func TestToolDiscovery(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    ts.Start()
    ts.Initialize()

    // Send tools/list request
    listReq := &ToolListRequest{
        JSONRPC: "2.0",
        ID:      2,
        Method:  "tools/list",
        Params:  ToolListParams{},
    }

    resp, err := ts.SendRequest(listReq)
    if err != nil {
        t.Fatalf("Failed to list tools: %v", err)
    }

    // Verify tools are returned
    tools := resp.Result.Tools
    if len(tools) == 0 {
        t.Fatal("Expected at least one tool to be available")
    }

    // Check for hello tool
    found := false
    for _, tool := range tools {
        if tool.Name == "hello" {
            found = true
            if tool.Description != "Returns a hello world message" {
                t.Errorf("Unexpected tool description: %s", tool.Description)
            }
            break
        }
    }

    if !found {
        t.Fatal("Hello tool not found in tool list")
    }
}
```

### 4. Tool Execution

```go
func TestToolExecution(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    ts.Start()
    ts.Initialize()

    // Call hello tool
    callReq := &ToolCallRequest{
        JSONRPC: "2.0",
        ID:      3,
        Method:  "tools/call",
        Params: ToolCallParams{
            Name: "hello",
            Arguments: map[string]interface{}{},
        },
    }

    resp, err := ts.SendRequest(callReq)
    if err != nil {
        t.Fatalf("Failed to call hello tool: %v", err)
    }

    // Verify response
    if len(resp.Result.Content) == 0 {
        t.Fatal("Expected content in tool response")
    }

    textContent, ok := resp.Result.Content[0].(TextContent)
    if !ok {
        t.Fatal("Expected TextContent in response")
    }

    expected := "Hello, World!"
    if textContent.Text != expected {
        t.Errorf("Expected %q, got %q", expected, textContent.Text)
    }
}
```

### 5. Error Handling

```go
func TestErrorHandling(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    ts.Start()
    ts.Initialize()

    // Call non-existent tool
    callReq := &ToolCallRequest{
        JSONRPC: "2.0",
        ID:      4,
        Method:  "tools/call",
        Params: ToolCallParams{
            Name: "nonexistent_tool",
            Arguments: map[string]interface{}{},
        },
    }

    resp, err := ts.SendRequest(callReq)
    if err != nil {
        t.Fatalf("Unexpected error sending request: %v", err)
    }

    // Should receive error response
    if resp.Error == nil {
        t.Fatal("Expected error response for non-existent tool")
    }

    if resp.Error.Code != -32601 { // Method not found
        t.Errorf("Expected method not found error, got code %d", resp.Error.Code)
    }
}
```

### 6. Concurrent Requests

```go
func TestConcurrentRequests(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()

    ts.Start()
    ts.Initialize()

    // Send multiple requests concurrently
    var wg sync.WaitGroup
    results := make([]*Response, 5)

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            callReq := &ToolCallRequest{
                JSONRPC: "2.0",
                ID:      id + 5,
                Method:  "tools/call",
                Params: ToolCallParams{
                    Name: "hello",
                    Arguments: map[string]interface{}{},
                },
            }

            resp, err := ts.SendRequest(callReq)
            if err != nil {
                t.Errorf("Request %d failed: %v", id, err)
                return
            }
            results[id] = resp
        }(i)
    }

    wg.Wait()

    // Verify all responses
    for i, resp := range results {
        if resp == nil {
            t.Errorf("Request %d did not receive response", i)
            continue
        }

        if len(resp.Result.Content) == 0 {
            t.Errorf("Request %d: Expected content in response", i)
        }
    }
}
```

## Protocol Message Types

### Request/Response Structures

```go
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### MCP-Specific Types

```go
type InitializeRequest struct {
    JSONRPCRequest
    Params InitializeParams `json:"params"`
}

type InitializeParams struct {
    ProtocolVersion string            `json:"protocolVersion"`
    Capabilities    ClientCapabilities `json:"capabilities"`
    ClientInfo      ClientInfo        `json:"clientInfo"`
}

type ToolCallRequest struct {
    JSONRPCRequest
    Params ToolCallParams `json:"params"`
}

type ToolCallParams struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments,omitempty"`
}
```

## Implementation Details

### Communication Protocol

- Uses JSON-RPC 2.0 over stdio
- Messages are JSON objects separated by newlines
- Server reads from stdin, writes to stdout
- Asynchronous request/response pattern with IDs

### Process Management

- Context-based timeout handling
- Proper pipe cleanup
- Signal handling for graceful shutdown
- Process exit code verification

### Test Utilities

```go
func (ts *TestServer) SendRequest(req interface{}) (*JSONRPCResponse, error) {
    // Marshal request to JSON
    data, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    // Send request with newline
    _, err = fmt.Fprintln(ts.stdin, string(data))
    if err != nil {
        return nil, err
    }

    // Read response
    scanner := bufio.NewScanner(ts.stdout)
    if !scanner.Scan() {
        return nil, scanner.Err()
    }

    // Parse response
    var resp JSONRPCResponse
    err = json.Unmarshal(scanner.Bytes(), &resp)
    return &resp, err
}

func (ts *TestServer) IsRunning() bool {
    return ts.cmd.Process != nil && ts.cmd.Process.Signal(syscall.Signal(0)) == nil
}

func (ts *TestServer) Close() {
    ts.cancel()
    ts.stdin.Close()
    ts.stdout.Close()

    if ts.cmd.Process != nil {
        ts.cmd.Wait()
    }
}
```

## Integration with Existing Tests

The test harness can be added to `server_test.go` or implemented as a separate `integration_test.go` file. It complements the existing unit tests by providing:

- End-to-end protocol testing
- Subprocess lifecycle verification
- Real communication channel testing
- Performance and reliability validation

## Benefits

1. **Comprehensive Coverage** - Tests the complete MCP server interface
2. **Isolation** - Runs server as separate process
3. **Protocol Compliance** - Validates JSON-RPC implementation
4. **Error Scenarios** - Tests failure conditions and recovery
5. **Performance** - Can measure response times and throughput
6. **CI/CD Ready** - Suitable for automated testing pipelines

## Future Extensions

- HTTP transport testing (if implemented)
- Load testing with multiple concurrent clients
- Protocol version compatibility testing
- Authentication flow testing
- Performance benchmarking