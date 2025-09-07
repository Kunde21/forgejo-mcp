# Go Code Style Guide for forgejo-mcp

## Package Structure
- Use `main` package for executable code
- Use descriptive package names for test packages (e.g., `servertest`)
- Follow standard Go project layout conventions

## Imports
```go
import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)
```
- Standard library imports first
- Third-party imports second
- Local packages last
- Blank lines between import groups
- Use `goimports` for automatic formatting

## Naming Conventions
- **Exported types/functions**: PascalCase (e.g., `Server`, `NewServer`, `Start`)
- **Unexported types/functions**: camelCase (e.g., `handleHello`, `getEnv`)
- **Struct fields**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `mcpServer`, `config`)
- **Constants**: PascalCase (e.g., `DefaultPort`)

## Type Definitions
```go
type Server struct {
	mcpServer *server.MCPServer
	config    *Config  // unexported field
}

type Config struct {
	Host string  // exported field
	Port int     // exported field
}
```

## Function Signatures
- Use pointer receivers for structs: `func (s *Server) Start() error`
- Constructor functions: `func NewServer() (*Server, error)`
- Handler functions: `func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)`

## Error Handling
```go
func (s *Server) Start() error {
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
```
- Always check and handle errors
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Return errors immediately when appropriate
- Use `log.Fatalf` only in main function for fatal errors

## Documentation
```go
// Server represents the MCP server instance
type Server struct {
	mcpServer *server.MCPServer
	config    *Config
}

// NewServer creates a new MCP server instance
func NewServer() (*Server, error) {
	// implementation
}

// handleHello handles the hello tool request
func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Basic validation - check if request is valid
	if ctx == nil {
		return mcp.NewToolResultError("Context is required"), nil
	}
	// implementation
}
```
- Add godoc comments for all exported functions, types, and constants
- Use complete sentences starting with the name of the item
- Add inline comments for complex business logic
- Keep comments concise and focused on "why" rather than "what"

## Testing
```go
func TestServerLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	ts := NewTestServer(t, ctx)
	if err := ts.Start(); err != nil {
		t.Fatal("Failed to start server:", err)
	}
	// test implementation
}
```
- Test files end with `_test.go`
- Test functions start with `Test`
- Use `context.WithTimeout` for test timeouts
- Use `t.Context` for test context
- Use `t.Cleanup` for resource cleanup
- Use descriptive test names that explain what is being tested
- Use `t.Fatal` for setup failures, `t.Error` for assertion failures
- Create concrete expected values and errors, avoid existence checks
- Use `cmp.Equal` for test validation

```go
if !cmp.Equal(wantErr, err, cmpopts.EquateErrors()) {
	t.Error(cmp.Diff(wantErr, err, cmpopts.EquateErrors()))
}
if !cmp.Equal(want, got) {
	t.Error(cmp.Diff(want, got))
}
```

## Code Organization
- Group related functionality together
- Keep files focused on a single action or record type
- Keep functions focused on single responsibility
- Use concise, meaningful variable names
- Avoid deep nesting - return early when possible
- Use blank lines to separate logical sections

## Dependencies
- Keep dependencies minimal and well-maintained
- Use specific versions in `go.mod`
- Run `goimports -w .` regularly to clean up unused imports
- Run `go mod tidy` regularly to clean up unused dependencies

## Build and Quality
- Use `go build ./...` to build all packages
- Use `go test ./...` to run all tests
- Use `go vet ./...` for static analysis
- Use `goimports -w .` for consistent import formatting
- Ensure all code passes `go vet` checks

This style guide reflects the patterns observed in the current codebase and should be followed for all new code additions to maintain consistency.
