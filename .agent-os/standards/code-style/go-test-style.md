# Go Test Style Guide for forgejo-mcp

## Overview

This guide documents the testing approach used in the forgejo-mcp project, focusing on acceptance testing patterns and best practices. The testing framework emphasizes end-to-end validation of MCP tool functionality using a comprehensive test harness.

## Test Categories

### 1. Unit Tests
- Focus on individual function behavior
- Use table-driven tests with descriptive test cases
- Mock external dependencies where appropriate
- Located in the same package as the code being tested

### 2. Integration Tests
- Validate MCP protocol interactions
- Test client-server communication
- Verify tool discovery and execution
- Located in `server_test` package

### 3. Acceptance Tests
- End-to-end validation of complete workflows
- Real-world scenario testing
- Concurrent request handling
- Performance and edge case validation

## Test Harness Structure

The test harness consists of two main components:

### TestServer
A wrapper around the MCP server that provides:
- In-memory transport for fast execution
- Automatic cleanup with `t.Cleanup()`
- Context management with timeouts
- Client session management

```go
mock := NewMockGiteaServer(t)
ts := NewTestServer(t, ctx, map[string]string{
    "FORGEJO_REMOTE_URL": mock.URL(),
    "FORGEJO_AUTH_TOKEN": "mock-token",
})
if err := ts.Initialize(); err != nil {
    t.Fatal(err)
}
client := ts.Client()
```

### MockGiteaServer
A mock HTTP server that simulates Gitea API responses:
- Supports multiple endpoint types (issues, comments, pull requests)
- Handles authentication and error scenarios
- Provides data management methods
- Uses modern Go 1.22+ routing patterns

```go
mock := NewMockGiteaServer(t)
mock.AddIssues("testuser", "testrepo", []MockIssue{
    {Index: 1, Title: "Bug: Login fails", State: "open"},
})
```

## Test Structure Patterns

### Table-Driven Tests
Use struct-based test cases for comprehensive validation:

```go
type TestCase struct {
    name      string
    setupMock func(*MockGiteaServer)
    arguments map[string]any
    expect    *mcp.CallToolResult
}

testCases := []TestCase{
    {
        name: "successful comment edit",
        setupMock: func(mock *MockGiteaServer) {
            mock.AddComments("testuser", "testrepo", []MockComment{
                {ID: 123, Content: "Original comment", Author: "testuser"},
            })
        },
        arguments: map[string]any{
            "repository":   "testuser/testrepo",
            "issue_number": 1,
            "comment_id":   123,
            "new_content":  "Updated comment content",
        },
        expect: &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{Text: "Comment edited successfully..."},
            },
            IsError: false,
        },
    },
}
```

### Test Execution Pattern
Follow this standard execution pattern:

```go
for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
        t.Cleanup(cancel)

        mock := NewMockGiteaServer(t)
        if tc.setupMock != nil {
            tc.setupMock(mock)
        }

        ts := NewTestServer(t, ctx, map[string]string{
            "FORGEJO_REMOTE_URL": mock.URL(),
            "FORGEJO_AUTH_TOKEN": "mock-token",
        })
        if err := ts.Initialize(); err != nil {
            t.Fatal(err)
        }
        client := ts.Client()

        result, err := client.CallTool(ctx, &mcp.CallToolParams{
            Name:      "tool_name",
            Arguments: tc.arguments,
        })
        if err != nil {
            t.Fatal(err)
        }
        
        if !cmp.Equal(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
            t.Error(cmp.Diff(tc.expect, result, cmpopts.IgnoreUnexported(mcp.TextContent{})))
        }
    })
}
```

## Acceptance Testing Patterns

### Real-World Scenarios
Test actual use cases that users will encounter:

```go
{
    name: "real world scenario - status update",
    setupMock: func(mock *MockGiteaServer) {
        mock.AddComments("testuser", "testrepo", []MockComment{
            {ID: 1, Content: "Working on this issue", Author: "testuser"},
        })
    },
    arguments: map[string]any{
        "repository":   "testuser/testrepo",
        "issue_number": 1,
        "comment_id":   1,
        "new_content":  "I've completed the implementation and added comprehensive tests. Ready for review.",
    },
    // ... expected result
},
```

### Performance Testing
Validate handling of large content and edge cases:

```go
func TestPullRequestCommentCreationPerformance(t *testing.T) {
    t.Parallel()
    
    // Test large content scenario
    largeComment := strings.Repeat("Detailed code review comment. ", 200) // ~10KB
    result, err := client.CallTool(ctx, &mcp.CallToolParams{
        Name: "pr_comment_create",
        Arguments: map[string]any{
            "repository":          "testuser/testrepo",
            "pull_request_number": 1,
            "comment":             largeComment,
        },
    })
    // ... validation
}
```

### Concurrent Testing
Validate thread safety and concurrent request handling:

```go
func TestCommentEditingConcurrent(t *testing.T) {
    const numGoroutines = 3
    var wg sync.WaitGroup
    results := make(chan error, numGoroutines)

    for i := range 3 {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            _, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
                Name: "issue_comment_edit",
                Arguments: map[string]any{
                    "repository":   "testuser/testrepo",
                    "issue_number": 1,
                    "comment_id":   id,
                    "new_content":  fmt.Sprintf("Concurrent edit content for comment %d", id),
                },
            })
            results <- err
        }(i + 1)
    }

    wg.Wait()
    close(results)
    for err := range results {
        if err != nil {
            t.Errorf("Concurrent request failed: %v", err)
        }
    }
}
```

## Validation Testing

### Input Validation
Test all validation error scenarios:

```go
{
    name: "validation error - invalid repository format",
    setupMock: func(mock *MockGiteaServer) {
        // No setup needed for validation errors
    },
    arguments: map[string]any{
        "repository":   "invalid-format",
        "issue_number": 1,
        "comment_id":   1,
        "new_content":  "test content",
    },
    expect: &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{
                Text: "Invalid request: repository: repository must be in format 'owner/repo'.",
            },
        },
        IsError: true,
    },
},
```

### Error Handling
Test API errors and permission scenarios:

```go
{
    name: "permission error - invalid token",
    setupMock: func(mock *MockGiteaServer) {
        mock.AddComments("testuser", "testrepo", []MockComment{
            {ID: 123, Content: "Original comment", Author: "testuser"},
        })
    },
    arguments: map[string]any{
        "repository":   "testuser/testrepo",
        "issue_number": 1,
        "comment_id":   123,
        "new_content":  "Updated content",
    },
    expect: &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{
                Text: "Failed to edit comment: failed to edit issue comment: unknown API error: 401",
            },
        },
        IsError: true,
    },
},
```

## Best Practices

### Test Organization
1. Use descriptive test names that explain the scenario
2. Group related tests with `t.Run()` subtests
3. Enable parallel execution with `t.Parallel()` where appropriate
4. Use `t.Cleanup()` for resource management
5. Set appropriate timeouts with `context.WithTimeout()`

### Mock Data Management
1. Create realistic mock data that matches expected API responses
2. Use consistent mock data across related tests
3. Test both success and error scenarios
4. Add mock data before making requests

### Error Testing
1. Test invalid parameters and edge cases
2. Validate error responses contain appropriate error messages
3. Test network failures and timeout scenarios
4. Verify proper error propagation through MCP protocol

### Result Validation
1. Use `github.com/google/go-cmp/cmp` for deep equality comparison
2. Compare both content and structure of responses
3. Handle floating point comparisons appropriately
4. Use `cmpopts.IgnoreUnexported()` for MCP content types

### Resource Management
1. Always use `t.Cleanup()` for resource cleanup
2. Cancel contexts to prevent goroutine leaks
3. Close client sessions properly
4. Use httptest.Server for HTTP mocking

## Common Test Helpers

### Response Extraction
```go
func getTextContent(content []mcp.Content) string {
    for _, c := range content {
        if textContent, ok := c.(*mcp.TextContent); ok {
            return textContent.Text
        }
    }
    return ""
}
```

### Validation Helpers
Use the test harness's built-in validation patterns rather than creating custom ones.

## Test Execution

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific test categories
go test -run Acceptance ./server_test/
go test -run Integration ./server_test/

# Run with verbose output
go test -v ./server_test/...

# Run with coverage
go test -cover ./server_test/
```

### Test Environment
Tests use environment variables for configuration:
- `FORGEJO_REMOTE_URL` - URL of the Gitea/Forgejo instance
- `FORGEJO_AUTH_TOKEN` - Authentication token

## Contributing

When adding new tests:
1. Follow existing naming conventions
2. Add appropriate mock data for your test scenarios
3. Include both positive and negative test cases
4. Document complex test setups with comments
5. Ensure tests are isolated and don't depend on external state
6. Use the established test harness patterns
7. Test concurrent behavior where relevant
8. Validate error handling for all failure modes