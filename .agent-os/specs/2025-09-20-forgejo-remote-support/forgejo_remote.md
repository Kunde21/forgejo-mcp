# Forgejo Remote Implementation Plan

## Overview
This plan outlines the implementation of Forgejo remote support using the official Forgejo SDK (`codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2`) while maintaining backward compatibility with the existing Gitea SDK implementation.

## Current State Analysis

The project currently uses:
- **Gitea SDK**: `code.gitea.io/sdk/gitea v0.22.0`
- **Client Interface**: Well-defined interface in `remote/interface.go` covering issues, comments, and pull requests
- **Current Implementation**: `GiteaClient` in `remote/gitea/gitea_client.go` that implements the interface
- **Server Integration**: Uses dependency injection via `NewFromService()` for easy client swapping

## Phase 1: Add Forgejo SDK Dependency

### 1.1 Update go.mod
Add the official Forgejo SDK dependency:

```go
require (
    codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2 v2.x.x
)
```

### 1.2 Run dependency management
```bash
go mod tidy
```

## Phase 2: Create ForgejoClient Implementation

### 2.1 Create new client structure
Create `remote/forgejo/forgejo_client.go` with:
- ForgejoClient struct implementing ClientInterface
- Constructor function `NewForgejoClient(url, token string)`
- All required methods matching the interface

### 2.2 Key implementation considerations
- Use Forgejo SDK types and methods
- Map Forgejo responses to existing interface types
- Handle pagination differences
- Implement error handling for SDK-specific issues

### 2.3 Core methods to implement
- ListIssues
- CreateIssueComment
- ListIssueComments
- EditIssueComment
- ListPullRequests
- ListPullRequestComments
- CreatePullRequestComment
- EditPullRequestComment

### 2.4 Implementation approach
The ForgejoClient will mirror the existing GiteaClient structure but use the Forgejo SDK:

```go
package forgejo

import (
    "context"
    "fmt"
    "strings"

    "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
    "github.com/kunde21/forgejo-mcp/remote"
)

// ForgejoClient implements ClientInterface using the Forgejo SDK
type ForgejoClient struct {
    client *forgejo.Client
}

// NewForgejoClient creates a new Forgejo client
func NewForgejoClient(url, token string) (*ForgejoClient, error) {
    client, err := forgejo.NewClient(url, forgejo.SetToken(token))
    if err != nil {
        return nil, fmt.Errorf("failed to create Forgejo client: %w", err)
    }

    return &ForgejoClient{
        client: client,
    }, nil
}

// ListIssues retrieves issues from the specified repository
func (c *ForgejoClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]remote.Issue, error) {
    // Parse repository string (format: "owner/repo")
    owner, repoName, ok := strings.Cut(repo, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
    }

    // List issues using Forgejo SDK
    opts := forgejo.ListIssueOption{
        ListOptions: forgejo.ListOptions{
            PageSize: limit,
            Page:     offset/limit + 1, // Forgejo uses 1-based pagination
        },
        State: forgejo.StateOpen, // Only open issues for now
    }

    forgejoIssues, _, err := c.client.ListRepoIssues(owner, repoName, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list issues: %w", err)
    }

    // Convert to our Issue struct
    issues := make([]remote.Issue, len(forgejoIssues))
    for i, fi := range forgejoIssues {
        issues[i] = remote.Issue{
            Number: int(fi.Index),
            Title:  fi.Title,
            State:  string(fi.State),
        }
    }

    return issues, nil
}

// ... implement remaining methods following the same pattern
```

## Next Steps

### Phase 3: Update Server Configuration
1. **Add client type configuration**: Extend config to support client type selection ("gitea", "forgejo", "auto")
2. **Implement version detection**: Add automatic remote type detection using `/api/v1/version` endpoint
3. **Update server initialization**: Modify `NewFromConfig()` to choose between Gitea and Forgejo clients with auto-detection
4. **Maintain backward compatibility**: Default to existing Gitea client behavior

### Phase 4: Testing and Validation
1. **Unit tests**: Create comprehensive tests for the new ForgejoClient
2. **Integration tests**: Test against actual Forgejo instances
3. **Migration testing**: Ensure existing functionality still works
4. **Performance comparison**: Benchmark both implementations

### Phase 5: Documentation and Migration
1. **Update README**: Document the new Forgejo support
2. **Migration guide**: Provide guidance for users switching from Gitea to Forgejo
3. **Configuration examples**: Show how to configure for Forgejo instances

## Benefits of This Approach

1. **Gradual Migration**: Users can switch at their own pace
2. **Interface Consistency**: No changes needed to server or tool implementations
3. **Better Forgejo Support**: Official SDK provides better compatibility
4. **Future-Proof**: Easier to maintain and update
5. **Testing**: Can validate both implementations work correctly

## Configuration Changes

### Config Structure Updates
```go
type Config struct {
    Host       string `mapstructure:"host"`
    Port       int    `mapstructure:"port"`
    RemoteURL  string `mapstructure:"remote_url"`
    AuthToken  string `mapstructure:"auth_token"`
    ClientType string `mapstructure:"client_type"` // "gitea", "forgejo", or "auto"
}
```

### Environment Variables
- `FORGEJO_CLIENT_TYPE`: Set to "forgejo", "gitea", or "auto" to select SDK
- Default behavior remains Gitea SDK for backward compatibility
- "auto" enables automatic remote type detection using version endpoint

## Error Handling Strategy

- Handle SDK-specific errors appropriately
- Provide clear error messages for configuration issues
- Ensure graceful fallback to Gitea client if Forgejo client fails
- Log SDK version mismatches or compatibility issues
- Handle version detection failures with clear error messages
- Provide fallback to Gitea when auto-detection fails

## Testing Strategy

- Mock both SDKs for unit testing
- Integration tests against test Forgejo instances
- Compare API responses between implementations
- Ensure pagination and filtering work identically
- Performance benchmarking of both clients

## Version Detection Implementation

### Automatic Detection Logic
The implementation will include automatic remote type detection using the existing `/api/v1/version` endpoint:

```go
func detectRemoteType(remoteURL, authToken string) (string, error) {
    // Make HTTP request to /api/v1/version
    // Parse JSON response to extract version string
    // Analyze version string patterns:
    //   - Contains "forgejo" -> "forgejo"
    //   - Contains "gitea" -> "gitea"
    //   - Version starts with "12." -> "forgejo"
    //   - Version starts with "1." -> "gitea"
    //   - Default to "gitea" if ambiguous
}
```

### Version Response Examples
- **Forgejo**: `{"version":"12.0.1-120-abfc8432+gitea-1.22.0"}`
- **Gitea**: `{"version":"1.25.0+dev-376-g223205cc6b"}`

### Auto-Detection Benefits
- **User Convenience**: No need to manually specify client type
- **Error Prevention**: Reduces configuration errors
- **Migration Support**: Seamless transition between Gitea and Forgejo instances
- **Future-Proof**: Adapts to new version patterns automatically

This implementation provides a solid foundation for adding Forgejo support with automatic detection while maintaining stability and backward compatibility.