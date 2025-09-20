# Agent Guidelines for forejo-mcp

## Build/Test Commands
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests (includes Forgejo/Gitea integration tests)
- `go test ./path/to/package -run TestName` - Run single test
- `go test -run Integration ./...` - Run integration tests only
- `go test -v ./...` - Run tests with verbose output
- `go test -cover ./...` - Run tests with coverage report
- `go mod tidy` - Clean up dependencies
- `go vet ./...` - Static analysis
- `goimports -w .` - Format code

## Code Style Guidelines
- **Imports**: Use standard library first, then third-party, then local packages with blank lines between groups
- **Formatting**: Use `goimports` for consistent formatting
- **Types**: Prefer explicit types; use interfaces for behavior contracts
- **Naming**: Use camelCase for local variables, PascalCase for exported identifiers
- **Error Handling**: Always check errors; use `fmt.Errorf` for wrapping with context
- **Documentation**: Add godoc comments for all exported functions, types, and constants
- **Testing**: Use table-driven tests; file names end with `_test.go`

## Project Structure
- Follow standard Go project layout
- Keep main package in `cmd/` directory for executables
