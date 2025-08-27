# Go Style Guide

## Overview

This document outlines the coding standards and style guidelines for Go code in this project. These guidelines ensure consistency, readability, and maintainability across the codebase.

## General Principles

- **Clarity over Cleverness**: Code should be clear and understandable
- **Consistency**: Follow established patterns throughout the codebase
- **Documentation**: All exported functions, types, and constants must be documented
- **Testing**: Use table-driven tests with explicit expected values

## Imports

### Standard Library First
```go
import (
    "net/url"
    "testing"
    "time"

    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
)
```

### Grouping
- Standard library imports first
- Third-party packages second
- Local/project packages last
- Use blank lines to separate groups

## Testing

### Table-Driven Tests
Use table-driven tests for comprehensive coverage with explicit expected values:

```go
func TestNewClientValidation(t *testing.T) {
    tests := []struct {
        name    string
        baseURL string
        token   string
        want    *ForgejoClient
        wantErr error
    }{
        {
            name:    "valid inputs",
            baseURL: "https://example.com",
            token:   "test-token",
            want: &ForgejoClient{
                baseURL:   exampleCom,
                token:     "test-token",
                timeout:   30 * time.Second,
                userAgent: "forgejo-mcp-client/1.0.0",
            },
            wantErr: nil,
        },
        // ... more test cases
    }

    for _, tst := range tests {
        t.Run(tst.name, func(t *testing.T) {
            client, err := New(tst.baseURL, tst.token)
            if !cmp.Equal(tst.wantErr, err) {
                t.Error(cmp.Diff(tst.wantErr, err))
            }
            if !cmp.Equal(tst.want, client, cmp.AllowUnexported(ForgejoClient{})) {
                t.Error(cmp.Diff(tst.want, client, cmp.AllowUnexported(ForgejoClient{})))
            }
        })
    }
}
```

### Test Structure
- Use `t.Run()` for subtests
- Use `cmp.Equal()` and `cmp.Diff()` for assertions
- Use `cmp.AllowUnexported()` for comparing structs with unexported fields
- Use `any` instead of `interface{}` for type declarations
- Use global variables in `init()` functions for test fixtures

### Test Naming
- Test functions: `TestFunctionName`
- Benchmark functions: `BenchmarkFunctionName`
- Example functions: `ExampleFunctionName`

## Structs and Types

### Field Alignment
```go
type ForgejoClient struct {
    baseURL   string        // grouped by type
    token     string
    timeout   time.Duration
    userAgent string
}
```

### Constructor Functions
```go
// New creates a new ForgejoClient with default configuration
func New(baseURL, token string) (*ForgejoClient, error) {
    return NewWithConfig(baseURL, token, DefaultConfig())
}

// NewWithConfig creates a new ForgejoClient with custom configuration
func NewWithConfig(baseURL, token string, config *ClientConfig) (*ForgejoClient, error) {
    // implementation
}
```

## Error Handling

### Custom Error Types
```go
type ValidationError struct {
    Message string
    Field   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}
```

### Error Checking
```go
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}
```

## Constants and Variables

### Constants
```go
const (
    StateOpen   StateType = "open"
    StateClosed StateType = "closed"
    StateAll    StateType = "all"
)
```

### Global Variables in Tests
```go
var exampleCom *url.URL

func init() {
    exampleCom, _ = url.Parse("https://example.com")
}
```

## Functions

### Naming
- Exported functions: `PascalCase`
- Unexported functions: `camelCase`
- Constants: `PamelCase`
- Test functions: `TestPascalCase`
- Benchmark functions: `BenchmarkPascalCase`

### Parameters and Return Values
```go
func NewWithConfig(baseURL, token string, config *ClientConfig) (*ForgejoClient, error) {
    // implementation
}
```

## Documentation

### Package Comments
```go
// Package client provides a Gitea SDK client for Forgejo repositories
package client
```

### Function Comments
```go
// New creates a new ForgejoClient with the given base URL and token
func New(baseURL, token string) (*ForgejoClient, error) {
    // implementation
}
```

### Struct Comments
```go
// ForgejoClient implements the Client interface using the Gitea SDK
type ForgejoClient struct {
    // implementation
}
```

## Interface Design

### Interface Definition
```go
// Client defines the interface for interacting with Forgejo repositories
type Client interface {
    ListPRs(owner, repo string, filters map[string]interface{}) ([]PullRequest, error)
    ListIssues(owner, repo string, filters map[string]interface{}) ([]Issue, error)
}
```

### Interface Compliance Validation
```go
// Test that ForgejoClient implements Client interface
var _ Client = (*ForgejoClient)(nil)
```

## Comparison and Assertion

### Using go-cmp
Use Options in `github.com/google/go-cmp/cmp` and `github.com/google/go-cmp/cmp/cmpopts` for modifying comparisons

```go
import (
    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
)

// For structs with unexported fields
if !cmp.Equal(want, got, cmp.AllowUnexported(ForgejoClient{})) {
    t.Error(cmp.Diff(want, got, cmp.AllowUnexported(ForgejoClient{})))
}

// For ignoring specific fields
if !cmp.Equal(want, got, cmpopts.IgnoreFields(ForgejoClient{}, "internalField")) {
    t.Error(cmp.Diff(want, got, cmpopts.IgnoreFields(ForgejoClient{}, "internalField")))
}
```

## Best Practices

### Avoid Global State
- Prefer dependency injection over global variables
- Use interfaces for testability

### Keep Functions Small
- Functions should do one thing well
- Break large functions into smaller, focused functions

### Use Context for Cancellation
```go
func (c *ForgejoClient) ListPRs(ctx context.Context, owner, repo string) ([]PullRequest, error) {
    // implementation with context support
}
```

### Handle Errors Properly
- Always check for errors
- Use error wrapping with `fmt.Errorf` and `%w` verb
- Return meaningful error messages

This style guide reflects the current codebase patterns and should be updated as new patterns emerge.
