---
date: 2025-10-05T13:39:10Z
git_commit: 315a4cd74d31e28db07277fc3612be03147149dd
branch: mcp-go
repository: forgejo-mcp
topic: "Issue Creation Tool Implementation Research"
tags: [research, codebase, issue, creation, forgejo, gitea, mcp, attachment]
last_updated: 2025-10-05T13:39:10Z
---

## Ticket Synopsis

The ticket requests implementation of a new MCP tool `forgejo_issue_create` to create issues on Forgejo/Gitea repositories. The tool must support creating issues with titles, bodies, and attachments (images and PDFs only), following existing patterns in the repository.

## Summary

The forgejo-mcp codebase has a well-established pattern for MCP tool implementation with comprehensive issue listing and comment management functionality. However, it lacks issue creation capabilities and any file upload functionality. The implementation will require:

1. Extending the remote interface with issue creation methods
2. Adding file upload infrastructure from scratch (no existing patterns)
3. Following established validation and error handling patterns
4. Implementing attachment support with MIME type validation

## Detailed Findings

### Existing Issue-Related Functionality

#### Current Issue Operations
- **Issue Listing**: `server/issues.go:41-92` - Lists issues with pagination
- **Issue Comments**: Full CRUD operations in `server/issue_comments.go`
- **Remote Interface**: `remote/interface.go:14-17` defines `IssueLister` interface
- **Data Structure**: Minimal `Issue` struct with Number, Title, State fields

#### Implementation Patterns
- **Tool Registration**: `server/server.go:104-152` using `mcp.AddTool()`
- **Validation**: ozzo-validation library with conditional rules
- **Repository Resolution**: `server/repository_resolver.go` for directory-to-repo mapping
- **Error Handling**: Standardized `TextError()` and `TextErrorf()` functions

### Parameter Validation Patterns

#### Conditional Validation Pattern
```go
v.Field(&args.Repository, v.When(args.Directory == "",
    v.Required.Error("at least one of directory or repository must be provided"),
    v.Match(repoReg).Error("repository must be in format 'owner/repo'"),
)),
```

#### Directory Validation Pattern
```go
v.Field(&args.Directory, v.When(args.Repository == "",
    v.Required.Error("at least one of directory or repository must be provided"),
    v.By(func(any) error {
        if !filepath.IsAbs(args.Directory) {
            return v.NewError("abs_dir", "directory must be an absolute path")
        }
        stat, err := os.Stat(args.Directory)
        if err != nil {
            return v.NewError("abs_dir", "invalid directory")
        }
        if !stat.IsDir() {
            return v.NewError("abs_dir", "does not exist")
        }
        return nil
    }),
)),
```

#### String Length Validation
```go
v.Field(&args.Title, v.When(args.Title != "",
    v.Length(1, 255).Error("title must be between 1 and 255 characters"),
)),
```

### Forgejo/Gitea Client Patterns

#### Client Initialization
```go
func NewForgejoClient(url, token string) (*ForgejoClient, error) {
    client, err := forgejo.NewClient(url, forgejo.SetToken(token))
    if err != nil {
        return nil, fmt.Errorf("failed to create Forgejo client: %w", err)
    }
    return &ForgejoClient{client: client}, nil
}
```

#### API Call Structure
```go
func (c *ForgejoClient) ListIssues(ctx context.Context, repo string, limit, offset int) ([]remote.Issue, error) {
    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }
    
    owner, repoName, ok := strings.Cut(repo, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
    }
    
    // API call with error wrapping
    forgejoIssues, _, err := c.client.ListRepoIssues(owner, repoName, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list issues: %w", err)
    }
    
    // Transform to internal format
    issues := make([]remote.Issue, len(forgejoIssues))
    for i, gi := range forgejoIssues {
        issues[i] = remote.Issue{
            Number: int(gi.Index),
            Title:  gi.Title,
            State:  string(gi.State),
        }
    }
    
    return issues, nil
}
```

### Directory Resolution Implementation

#### Repository Resolver
- **Location**: `server/repository_resolver.go:128-307`
- **Function**: `ResolveRepository(directory string) (*RepositoryResolution, error)`
- **Features**: 
  - Validates directory exists and is a Git repository
  - Parses `.git/config` to extract remote information
  - Supports HTTPS, SSH, and Git protocol URLs
  - Returns structured resolution with directory, repository, URL, and remote name

#### Git Repository Detection
```go
func (r *RepositoryResolver) ValidateDirectory(directory string) error {
    if _, err := os.Stat(directory); os.IsNotExist(err) {
        return NewDirectoryNotFoundError(directory)
    }
    
    gitDir := filepath.Join(directory, ".git")
    gitInfo, err := os.Stat(gitDir)
    switch {
    case os.IsNotExist(err):
        return NewNotGitRepositoryError(directory, "no .git directory found")
    case err != nil:
        return &RepositoryError{
            Op:   "validate",
            Path: gitDir,
            Err:  fmt.Errorf("failed to access .git directory: %w", err),
        }
    case !gitInfo.IsDir():
        return NewNotGitRepositoryError(directory, ".git is not a directory")
    }
    return nil
}
```

### File Upload Requirements

#### Current State
- **No existing file upload functionality** in the codebase
- Only `TextContent` is used for responses
- No MIME type validation or file size limits implemented

#### MCP Protocol Support
- SDK supports `ImageContent`, `AudioContent`, and `BlobContent` for binary data
- Automatic base64 encoding/decoding handled by SDK
- Need to determine how clients pass binary data through MCP

#### Forgejo/Gitea API Support
- Both SDKs have `CreateReleaseAttachment()` methods
- No direct issue attachment APIs visible in current SDK
- May need custom HTTP client calls to undocumented endpoints

#### Security Requirements
- MIME type validation (images and PDFs only)
- File size limits (needs research on Forgejo/Gitea limits)
- Filename sanitization
- Temporary file secure handling

## Code References

### Core Files to Modify
- `server/server.go:104-152` - Add tool registration
- `server/issues.go` - Add `handleIssueCreate()` function
- `remote/interface.go:162-172` - Add `CreateIssue()` method to interface
- `remote/forgejo/issues.go` - Add Forgejo implementation
- `remote/gitea/gitea_client.go` - Add Gitea implementation

### Reference Implementations
- `server/issue_comments.go:45-98` - Comment creation pattern
- `server/pr_edit.go:60-96` - Complex parameter validation
- `server/common.go:10-24` - Error handling utilities
- `server_test/harness.go` - Test infrastructure

### Test Files to Update
- `server_test/issue_list_test.go` - Add issue creation tests
- `server_test/harness.go` - Add mock file upload support

## Architecture Insights

### Design Patterns
1. **Clean Architecture**: Clear separation between server handlers and remote clients
2. **Repository Pattern**: Abstraction via `remote.ClientInterface`
3. **Factory Pattern**: Client creation based on configuration
4. **Validation Layer**: Consistent ozzo-validation across all handlers
5. **Error Handling**: Structured error responses with context

### Integration Points
1. **Authentication**: Reuse existing token system
2. **Repository Resolution**: Use existing `repositoryResolver`
3. **Configuration**: Extend `config/config.go` for file upload settings
4. **Testing**: Follow existing table-driven test patterns

## Historical Context

From `thoughts/tickets/feature_issue_create.md`:
- Created today (2025-10-05) for issue creation feature
- Decisions: Support only images and PDFs, use existing auth, follow established patterns
- Constraints: Must validate inputs, support both Forgejo and Gitea, match directory parameter handling

## Related Research

None found in `thoughts/research/` directory. This appears to be the first comprehensive research on issue creation and file upload functionality.

## Open Questions

1. **Attachment API Endpoints**: Need to research actual Forgejo/Gitea issue attachment endpoints
2. **File Size Limits**: Default attachment size limits for Forgejo/Gitea instances
3. **MCP Binary Data**: How clients will pass binary data through the MCP protocol
4. **Memory Management**: Handling large file uploads without excessive memory usage
5. **Rate Limiting**: Upload rate limiting considerations

## Implementation Recommendations

### Phase 1: Core Issue Creation
1. Add `CreateIssue()` to `remote.ClientInterface`
2. Implement in both Forgejo and Gitea clients
3. Create server handler with validation
4. Add comprehensive tests

### Phase 2: File Upload Infrastructure
1. Research actual attachment API endpoints
2. Implement file validation utilities
3. Add configuration options for limits
4. Create secure temporary file handling

### Phase 3: Attachment Support
1. Extend issue creation to accept attachments
2. Implement upload workflow
3. Add attachment listing/deletion tools
4. Full integration testing

### Security Considerations
1. Implement MIME type whitelist
2. Add file size validation before upload
3. Sanitize filenames to prevent path traversal
4. Add audit logging for uploads
5. Consider virus scanning for enterprise deployments