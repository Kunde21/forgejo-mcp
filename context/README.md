# Repository Context Detection

This package provides automatic repository context detection for Forgejo repositories. It can identify the current Forgejo repository from the local git environment and extract owner, repository name, and remote URL information.

## Features

- **Git Repository Detection**: Automatically detects git repositories and worktrees
- **Forgejo Remote Validation**: Validates that remotes point to Forgejo instances
- **Repository Information Parsing**: Extracts owner and repository names from various URL formats
- **Context Caching**: In-memory caching with TTL for performance optimization
- **Thread-Safe Operations**: Concurrent-safe context detection and caching

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/Kunde21/forgejo-mcp/context"
)

func main() {
    // Detect context for current directory
    ctx, err := context.DetectContext(".")
    if err != nil {
        log.Fatalf("Failed to detect context: %v", err)
    }

    fmt.Printf("Repository: %s/%s\n", ctx.Owner, ctx.Repository)
    fmt.Printf("Remote URL: %s\n", ctx.RemoteURL)
}
```

## API Reference

### Context Struct

```go
type Context struct {
    Owner      string // Repository owner/organization
    Repository string // Repository name
    RemoteURL  string // Full remote URL
}
```

### Functions

#### DetectContext(path string) (*Context, error)

Detects repository context for the given path. This is a convenience function that uses a default context manager.

```go
ctx, err := context.DetectContext("/path/to/repo")
if err != nil {
    // Handle error
}
```

#### ContextManager

For advanced usage with caching and performance optimization:

```go
// Create a context manager with custom TTL
cm := context.NewContextManagerWithTTL(10 * time.Minute)

// Detect context (with caching)
ctx, err := cm.DetectContext("/path/to/repo")
if err != nil {
    // Handle error
}

// Check cache size
size := cm.CacheSize()

// Clear cache if needed
cm.ClearCache()
```

## Supported URL Formats

The package supports various git URL formats:

### HTTPS URLs
- `https://codeberg.org/user/repo.git`
- `https://codeberg.org/user/repo`
- `https://git.example.com/group/repo.git`

### SSH URLs
- `git@codeberg.org:user/repo.git`
- `git@codeberg.org:user/repo`
- `ssh://git@codeberg.org/user/repo.git`

## Supported Forgejo Instances

The package recognizes these Forgejo instances by default:

- **codeberg.org** - The main Forgejo instance
- **forgejo.org** - Official Forgejo instance
- **git.sr.ht** - SourceHut (uses Forgejo)

Custom Forgejo instances are also supported as long as they have valid hostnames.

## Git Worktree Support

The package fully supports git worktrees:

```bash
# Create a worktree
git worktree add ../feature-branch

# Detect context from worktree directory
cd ../feature-branch
```

```go
ctx, err := context.DetectContext("../feature-branch")
// Works identically to main repository
```

## Error Handling

The package provides clear error messages for common scenarios:

```go
ctx, err := context.DetectContext("/path/to/dir")
if err != nil {
    switch err.Error() {
    case "not a git repository":
        // Directory is not a git repository
    case "remote 'origin' has no URL configured":
        // Git repository has no remote configured
    case "remote is not a Forgejo instance":
        // Remote points to non-Forgejo service
    default:
        // Other parsing or validation errors
    }
}
```

## Performance Considerations

- **Caching**: Context results are cached for 5 minutes by default
- **Thread Safety**: All operations are thread-safe
- **Git Command Optimization**: Minimizes git command execution through caching

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/Kunde21/forgejo-mcp/context"
)

func main() {
    ctx, err := context.DetectContext(".")
    if err != nil {
        log.Fatalf("Context detection failed: %v", err)
    }

    fmt.Printf("Working with: %s\n", ctx.String())
    fmt.Printf("Owner: %s\n", ctx.Owner)
    fmt.Printf("Repository: %s\n", ctx.Repository)
    fmt.Printf("Remote: %s\n", ctx.RemoteURL)
}
```

### Custom Context Manager

```go
package main

import (
    "time"

    "github.com/Kunde21/forgejo-mcp/context"
)

func main() {
    // Create manager with 10-minute TTL
    cm := context.NewContextManagerWithTTL(10 * time.Minute)

    // Detect context with caching
    ctx1, _ := cm.DetectContext("./repo1")
    ctx2, _ := cm.DetectContext("./repo1") // Returns cached result

    // Clear cache when needed
    cm.ClearCache()
}
```

### Error Handling Example

```go
package main

import (
    "fmt"
    "strings"

    "github.com/Kunde21/forgejo-mcp/context"
)

func handleContextError(err error) {
    errMsg := err.Error()

    if strings.Contains(errMsg, "not a git repository") {
        fmt.Println("Please run this command from within a git repository")
    } else if strings.Contains(errMsg, "no URL configured") {
        fmt.Println("Please add a remote origin to your repository:")
        fmt.Println("  git remote add origin https://codeberg.org/user/repo.git")
    } else if strings.Contains(errMsg, "not a Forgejo instance") {
        fmt.Println("This tool only works with Forgejo repositories")
        fmt.Println("Please use a repository hosted on codeberg.org, forgejo.org, or another Forgejo instance")
    } else {
        fmt.Printf("Context detection failed: %v\n", err)
    }
}

func main() {
    _, err := context.DetectContext(".")
    if err != nil {
        handleContextError(err)
    }
}
```

## Testing

The package includes comprehensive tests covering:

- Git repository detection (regular repos and worktrees)
- Remote URL extraction and validation
- Repository information parsing
- Context manager caching behavior
- Error scenarios and edge cases
- Integration tests for complete workflows

Run tests with:

```bash
go test ./context -v
```

## Contributing

When contributing to this package:

1. Follow the existing code style and patterns
2. Add tests for new functionality
3. Update documentation for API changes
4. Ensure all tests pass with `go test ./context`

## License

This package is part of the Forgejo MCP project and follows the same license terms.