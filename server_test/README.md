# Server Test Harness

This directory contains a comprehensive test harness for the Forgejo MCP server, providing utilities for testing MCP protocol interactions, tool calls, and server behavior with both real and mock backends.

## Overview

The test harness consists of two main components:

1. **TestServer** - A wrapper around the MCP server for testing
2. **MockGiteaServer** - A mock HTTP server that simulates Gitea API responses

## Components

### TestServer

The `TestServer` struct provides a controlled environment for testing MCP server functionality:

```go
type TestServer struct {
    ctx     context.Context
    cancel  context.CancelFunc
    t       *testing.T
    client  *client.Client
    once    *sync.Once
    started bool
}
```

#### Key Methods

- `NewTestServer(t, ctx, env)` - Creates a new test server instance
- `Start()` - Starts the MCP server process
- `Initialize()` - Performs MCP protocol initialization handshake
- `Client()` - Returns the MCP client for making requests
- `IsRunning()` - Checks if the server is running

#### Usage Example

```go
func TestMyTool(t *testing.T) {
    mock := NewMockGiteaServer(t)
    ts := NewTestServer(t, t.Context(), map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
        "FORGEJO_AUTH_TOKEN": "mock-token",
    })

    // Initialize the MCP connection
    if err := ts.Initialize(); err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }

    // Call a tool
    result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name: "list_issues",
            Arguments: map[string]any{
                "repository": "owner/repo",
                "limit": 10,
            },
        },
    })

    if err != nil {
        t.Fatalf("Tool call failed: %v", err)
    }

    // Assert on result
    if result.Content == nil {
        t.Error("Expected content in result")
    }
}
```

### MockGiteaServer

The `MockGiteaServer` provides a mock HTTP server that simulates Gitea API endpoints:

```go
type MockGiteaServer struct {
    server *httptest.Server
    issues map[string][]MockIssue
}
```

#### Key Methods

- `NewMockGiteaServer(t)` - Creates a new mock server
- `URL()` - Returns the mock server URL
- `AddIssues(owner, repo, issues)` - Adds mock issues for a repository

#### Mock Issue Structure

```go
type MockIssue struct {
    Index int    `json:"index"`
    Title string `json:"title"`
    State string `json:"state"`
}
```

#### Usage Example

```go
func TestWithMockData(t *testing.T) {
    mock := NewMockGiteaServer(t)

    // Add mock issues
    mock.AddIssues("testuser", "testrepo", []MockIssue{
        {Index: 1, Title: "Bug: Login fails", State: "open"},
        {Index: 2, Title: "Feature: Add dark mode", State: "open"},
        {Index: 3, Title: "Fix: Memory leak", State: "closed"},
    })

    // Use in test server
    ts := NewTestServer(t, t.Context(), map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
        "FORGEJO_AUTH_TOKEN": "mock-token",
    })
}
```

## Test Categories

### Acceptance Tests (`acceptance_test.go`)

These tests validate end-to-end functionality with realistic scenarios:

- `TestListIssuesAcceptance` - Basic issue listing functionality
- `TestListIssuesPagination` - Pagination parameter handling
- `TestListIssuesErrorHandling` - Error scenario validation
- `TestListIssuesInputValidation` - Input parameter validation
- `TestListIssuesConcurrent` - Concurrent request handling
- `TestListIssuesInvalidLimit` - Invalid parameter handling

### Integration Tests (`integration_test.go`)

These tests validate MCP protocol interactions and server behavior:

- `TestMCPInitialization` - MCP protocol handshake
- `TestToolDiscovery` - Tool listing and schema validation
- `TestHelloTool` - Basic tool execution
- `TestToolExecution` - Tool execution with various scenarios
- `TestErrorHandling` - Error handling and edge cases
- `TestConcurrentRequests` - Concurrent request processing

## Running Tests

### Run All Tests

```bash
go test ./server_test/...
```

### Run Specific Test Categories

```bash
# Acceptance tests only
go test -run Acceptance ./server_test/

# Integration tests only
go test -run Integration ./server_test/
```

### Run with Verbose Output

```bash
go test -v ./server_test/...
```

### Run with Coverage

```bash
go test -cover ./server_test/
```

## Environment Variables

The test harness supports configuration through environment variables:

- `FORGEJO_REMOTE_URL` - URL of the Gitea/Forgejo instance (defaults to mock server URL)
- `FORGEJO_AUTH_TOKEN` - Authentication token (defaults to "test-token")

## Best Practices

### Test Structure

1. **Setup**: Create mock server and test server instances
2. **Initialize**: Call `ts.Initialize()` to establish MCP connection
3. **Execute**: Make tool calls using `ts.Client().CallTool()`
4. **Assert**: Validate results and error conditions
5. **Cleanup**: Automatic cleanup via `t.Cleanup()` calls

### Mock Data Management

1. Use descriptive repository names (e.g., "testuser/testrepo")
2. Add realistic mock data that matches expected API responses
3. Test both success and error scenarios
4. Use consistent mock data across related tests

### Error Testing

1. Test invalid parameters and edge cases
2. Validate error responses contain appropriate error messages
3. Test network failures and timeout scenarios
4. Verify proper error propagation through MCP protocol

### Concurrent Testing

1. Use goroutines to simulate concurrent requests
2. Validate thread safety of server operations
3. Test resource cleanup under concurrent load
4. Verify consistent results across concurrent executions

## Example Test Patterns

### Basic Tool Testing

```go
func TestBasicTool(t *testing.T) {
    mock := NewMockGiteaServer(t)
    ts := NewTestServer(t, t.Context(), map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
    })

    if err := ts.Initialize(); err != nil {
        t.Fatal(err)
    }

    result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
        Params: mcp.CallToolParams{Name: "hello"},
    })

    if err != nil {
        t.Fatal(err)
    }

    expected := "Hello, World!"
    if text := result.Content[0].(mcp.TextContent).Text; text != expected {
        t.Errorf("Expected %q, got %q", expected, text)
    }
}
```

### Error Scenario Testing

```go
func TestErrorScenario(t *testing.T) {
    mock := NewMockGiteaServer(t)
    ts := NewTestServer(t, t.Context(), map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
    })

    if err := ts.Initialize(); err != nil {
        t.Fatal(err)
    }

    // Test with invalid repository
    result, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name: "list_issues",
            Arguments: map[string]any{
                "repository": "invalid/repo",
            },
        },
    })

    if err != nil {
        t.Fatal(err)
    }

    if result.Content == nil {
        t.Error("Expected error content")
    }
}
```

### Concurrent Load Testing

```go
func TestConcurrentLoad(t *testing.T) {
    mock := NewMockGiteaServer(t)
    ts := NewTestServer(t, t.Context(), map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
    })

    if err := ts.Initialize(); err != nil {
        t.Fatal(err)
    }

    const numRequests = 10
    results := make(chan error, numRequests)

    for range numRequests {
        go func() {
            _, err := ts.Client().CallTool(context.Background(), mcp.CallToolRequest{
                Params: mcp.CallToolParams{Name: "hello"},
            })
            results <- err
        }()
    }

    for range numRequests {
        if err := <-results; err != nil {
            t.Errorf("Concurrent request failed: %v", err)
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Server not starting**: Check that environment variables are properly set
2. **MCP initialization failures**: Verify protocol version compatibility
3. **Mock server issues**: Ensure mock data is added before making requests
4. **Timeout errors**: Increase context timeouts for complex operations

### Debug Tips

1. Use `t.Log()` to output debug information
2. Enable verbose test output with `-v` flag
3. Check mock server logs for unexpected requests
4. Validate MCP message formats in failing tests

## Contributing

When adding new tests:

1. Follow existing naming conventions
2. Add appropriate mock data for your test scenarios
3. Include both positive and negative test cases
4. Document complex test setups with comments
5. Ensure tests are isolated and don't depend on external state
