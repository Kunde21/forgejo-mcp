# Gitea SDK Refactor Plan

## Goal
Separate MCP handlers from Gitea SDK implementation by moving Gitea-specific code to a dedicated `remote/gitea` package while keeping MCP handlers and input validation in the `server` package.

## Current Structure Issues
1. **server/sdk_handlers.go** contains both MCP handler logic AND Gitea SDK implementation details
2. **GiteaClientInterface** is defined in server package but should be in remote/gitea
3. **SDKError** and validation functions are mixed with MCP handlers
4. Repository resolution logic is tightly coupled to MCP handlers

## Proposed New Structure

```
/home/kunde21/development/AI/forgejo-mcp/
├── server/                    # MCP handlers and input validation only
│   ├── handlers.go           # MCP handler structs and methods
│   ├── validation.go         # Input validation functions
│   └── types.go              # MCP-specific types
├── remote/                   # Remote service implementations
│   └── gitea/               # Gitea-specific implementation
│       ├── client.go        # Gitea client wrapper and interface
│       ├── errors.go        # SDKError and error handling
│       ├── repository.go    # Repository operations
│       ├── pr.go           # Pull request operations
│       ├── issue.go        # Issue operations
│       └── git.go          # Gitea-specific utilities
└── config/                  # Configuration (unchanged)
```

## Detailed Refactor Steps

### 1. Create Remote Package Structure
- Create `/home/kunde21/development/AI/forgejo-mcp/remote/gitea/` directory
- Move Gitea-specific code from server package to remote/gitea

### 2. Move Gitea Client Interface and Errors
**From:** `server/sdk_handlers.go:15-78` (SDKError, GiteaClientInterface)
**To:** `remote/gitea/client.go`
- Move `SDKError` struct and methods
- Move `GiteaClientInterface` and rename `ClientInterface`
- Move `NewSDKError` function

### 3. Move Repository Resolution Logic
**From:** `server/sdk_handlers.go:135-226` (resolveCWD functions)
**To:** `remote/gitea/git.go`
- Move `resolveCWDToRepository`, `parseGitRemoteOutput`, `resolveCWDFromPath`

### 4. Move Repository Metadata Extraction
**From:** `server/sdk_handlers.go:228-267` (extractRepositoryMetadata)
**To:** `remote/gitea/repository.go`

### 5. Move Validation Functions
**From:** `server/sdk_handlers.go:80-133` (validation functions)
**To:** `server/validation.go`
- Keep `ValidateRepositoryFormat` in server (MCP input validation)
- Move `validateRepositoryExistence` and `validateRepositoryAccess` to `remote/gitea/repository.go`

### 6. Restructure MCP Handlers
**Current:** All handler logic in `server/sdk_handlers.go`
**New Structure:**
- `server/handlers.go` - Handler structs and core logic
- Keep MCP-specific transformation methods in server package
- Move SDK calls to remote/gitea package
- Include the request context in all calls to the remote/gitea package

### 7. Update Handler Implementation Pattern
Use dependency injection to load the SDK client interface in the handler on initialization.

Interface implementation will be responsible for interpreting the results.

```go
Instead of direct SDK calls in handlers:
```go
// Current pattern in handlers
prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)

// New pattern
prs, err := h.remote.ListPullRequests(ctx, owner, repo, opts)
```

### 8. Move Test Files
- `server/integration_test.go` → Keep in server (tests MCP handlers)
- Create `remote/gitea/issue_tets.go` for Gitea issues testing
- Create `remote/gitea/pr_tets.go` for Gitea pull request testing
- Create `remote/gitea/client_test.go` for Gitea client tests

## Code Organization Details

### server/validation.go
```go
package server

import "fmt"

// ValidateRepositoryFormat validates that a repository parameter follows the owner/repo format
func ValidateRepositoryFormat(repoParam string) (bool, error) {
    // Keep existing implementation from sdk_handlers.go:80-106
}
```

### remote/gitea/client.go
```go
package gitea

import "code.gitea.io/sdk/gitea"

// GiteaClientInterface defines the interface for Gitea client operations
type GiteaClientInterface interface {
    // Move all methods from sdk_handlers.go:46-78
}

// SDKError represents an error from the Gitea SDK with additional context
type SDKError struct {
    // Move from sdk_handlers.go:15-20
}

// NewSDKError creates a new SDK error with context
func NewSDKError(operation string, cause error, context ...string) *SDKError {
    // Move from sdk_handlers.go:33-44
}
```

### remote/gitea/git.go
```go
package gitea

import "regexp"

// resolveCWDToRepository attempts to resolve a CWD path to a repository identifier
func resolveCWDToRepository(cwd string) (string, error) {
    // Move from sdk_handlers.go:135-159
}

// parseGitRemoteOutput parses the output of 'git remote -v' to extract repository identifier
func parseGitRemoteOutput(output string) (string, error) {
    // Move from sdk_handlers.go:161-191
}

// resolveCWDFromPath is the fallback method that uses path-based resolution
func resolveCWDFromPath(cwd string) (string, error) {
    // Move from sdk_handlers.go:193-226
}
```

### remote/gitea/repository.go
```go
package gitea

import (
    "fmt"
    "strings"
)

// validateRepositoryExistence checks if a repository exists via Gitea API
func validateRepositoryExistence(client GiteaClientInterface, repoParam string) (bool, error) {
    // Move from sdk_handlers.go:108-119
}

// validateRepositoryAccess checks if the user has access to the repository
func validateRepositoryAccess(client GiteaClientInterface, repoParam string) (bool, error) {
    // Move from sdk_handlers.go:121-133
}

// extractRepositoryMetadata extracts and caches repository metadata
func extractRepositoryMetadata(client GiteaClientInterface, repoParam string) (map[string]any, error) {
    // Move from sdk_handlers.go:228-267
}
```

## Benefits of This Refactor

1. **Separation of Concerns**: MCP handlers focus on request/response, remote package handles SDK details
2. **Testability**: Can mock Gitea client interface without MCP dependencies
3. **Extensibility**: Easy to add other remote services (GitHub, GitLab) following same pattern
4. **Maintainability**: Clear boundaries between MCP protocol and remote service implementation
5. **Reusability**: Gitea client can be used independently of MCP handlers

## Implementation Order

1. Create remote/gitea package structure
2. Move GiteaClientInterface and SDKError
3. Move repository resolution utilities
4. Move validation functions appropriately
5. Restructure handlers to use new remote package
6. Update imports and fix any circular dependencies
7. Run tests to ensure functionality preserved

## Files to Create/Modify

### New Files:
- `remote/gitea/client.go`
- `remote/gitea/errors.go`
- `remote/gitea/repository.go`
- `remote/gitea/git.go`
- `remote/gitea/pr.go`
- `remote/gitea/issue.go`
- `server/validation.go`
- `server/handlers.go`
- `server/types.go`

### Modified Files:
- `server/sdk_handlers.go` → Split into multiple files
- Update imports in `cmd/serve.go` and other files
- Update test files to use new package structure

This refactor maintains all existing functionality while creating a clean separation between MCP protocol handling and Gitea SDK implementation.
