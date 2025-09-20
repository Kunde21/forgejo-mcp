# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-20-forgejo-remote-support/spec.md

## Technical Requirements

- **SDK Integration**: Add `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2` dependency to go.mod
- **Client Implementation**: Create `remote/forgejo/forgejo_client.go` with ForgejoClient struct implementing ClientInterface
- **Configuration Extension**: Add `ClientType` field to Config struct with validation for "gitea", "forgejo", and "auto" values
- **Factory Pattern**: Update server initialization to instantiate appropriate client based on configuration with automatic detection support
- **Interface Compliance**: Ensure ForgejoClient implements all methods from ClientInterface with identical signatures
- **Error Handling**: Implement SDK-specific error handling with clear, actionable error messages
- **Pagination Support**: Handle pagination differences between Gitea and Forgejo SDKs correctly
- **Type Mapping**: Convert Forgejo SDK types to existing interface types seamlessly
- **Version Detection**: Implement automatic remote type detection using `/api/v1/version` endpoint with intelligent version string parsing

## External Dependencies

- **forgejo-sdk/forgejo/v2** - Official Forgejo SDK for API interactions
- **Justification**: Provides native Forgejo support with better compatibility and future-proofing
- **Version Requirements**: Latest stable version (v2.x.x) from codeberg.org

## Implementation Details

### Client Structure
```go
type ForgejoClient struct {
    client *forgejo.Client
}

func NewForgejoClient(url, token string) (*ForgejoClient, error)
```

### Required Methods
- ListIssues(ctx context.Context, repo string, limit, offset int) ([]remote.Issue, error)
- CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) error
- ListIssueComments(ctx context.Context, repo string, issueNumber, limit, offset int) ([]remote.IssueComment, error)
- EditIssueComment(ctx context.Context, repo string, commentID int, newContent string) error
- ListPullRequests(ctx context.Context, repo string, limit, offset int, state string) ([]remote.PullRequest, error)
- ListPullRequestComments(ctx context.Context, repo string, prNumber, limit, offset int) ([]remote.PRComment, error)
- CreatePullRequestComment(ctx context.Context, repo string, prNumber int, comment string) error
- EditPullRequestComment(ctx context.Context, repo string, commentID int, newContent string) error

### Configuration Changes
```go
type Config struct {
    Host       string `mapstructure:"host"`
    Port       int    `mapstructure:"port"`
    RemoteURL  string `mapstructure:"remote_url"`
    AuthToken  string `mapstructure:"auth_token"`
    ClientType string `mapstructure:"client_type"` // "gitea", "forgejo", or "auto"
}
```

### Performance Considerations
- Lazy initialization of SDK clients
- Connection pooling where applicable
- Efficient pagination handling
- Minimal memory overhead for type conversions

### Version Detection Implementation

#### Detection Function
```go
func detectRemoteType(remoteURL, authToken string) (string, error) {
    // Call /api/v1/version endpoint
    // Parse response to determine if Forgejo or Gitea
    // Return "forgejo", "gitea", or error
}
```

#### Version String Analysis
- **Forgejo Detection**: Version strings containing "forgejo" or patterns like "12.x.x"
- **Gitea Detection**: Version strings containing "gitea" or patterns like "1.x.x"
- **Fallback Strategy**: Default to Gitea if detection is ambiguous

#### Auto-Detection Logic
```go
func NewFromConfig(cfg *Config) (*Server, error) {
    // Auto-detect if ClientType is "auto" or empty
    if cfg.ClientType == "auto" || cfg.ClientType == "" {
        detectedType, err := detectRemoteType(cfg.RemoteURL, cfg.AuthToken)
        if err != nil {
            return nil, fmt.Errorf("failed to detect remote type: %w", err)
        }
        cfg.ClientType = detectedType
    }
    
    // Create appropriate client based on type
    switch cfg.ClientType {
    case "forgejo":
        return createForgejoClient(cfg)
    case "gitea":
        return createGiteaClient(cfg)
    default:
        return nil, fmt.Errorf("unsupported client type: %s", cfg.ClientType)
    }
}
```