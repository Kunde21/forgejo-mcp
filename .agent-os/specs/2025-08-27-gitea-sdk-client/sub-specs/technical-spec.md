# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-08-27-gitea-sdk-client/spec.md

> Created: 2025-08-27
> Version: 1.0.0

## Technical Requirements

### Client Interface Design
- Define `Client` interface in `client/client.go` with methods for API operations
- Implement `ForgejoClient` struct using Gitea SDK with configuration and authentication
- Constructor `New(baseURL string, token string) (*ForgejoClient, error)` creates authenticated client
- Support configurable HTTP timeout with default of 30 seconds for API calls

### Gitea SDK Integration
- Use `code.gitea.io/sdk/gitea` package as the core HTTP client
- Configure client with proper base URL, authentication token, and HTTP settings
- Handle different authentication methods (token, basic auth) through SDK options
- Set appropriate user agent and request headers for Forgejo compatibility

### High-Level Methods
- `ListPRs(owner, repo string, filters map[string]interface{}) ([]types.PullRequest, error)` for pull requests
- `ListIssues(owner, repo string, filters map[string]interface{}) ([]types.Issue, error)` for issues
- Methods handle complete flow: build request → execute API call → transform response → return structs
- Support filter parameters: state (open/closed/all), labels, assignee, author, milestone

### API Request Building
- `buildPRListOptions(filters map[string]interface{}) ListPullRequestsOptions` constructs PR API options
- `buildIssueListOptions(filters map[string]interface{}) ListIssueOption` constructs issue API options
- Map Go filter parameters to Gitea SDK option structs
- Handle pagination parameters (page, limit) for large result sets
- Validate filter parameters before API calls

### Response Transformation
- `transformPullRequest(sdkPR *gitea.PullRequest) types.PullRequest` converts SDK types to internal types
- `transformIssue(sdkIssue *gitea.Issue) types.Issue` converts SDK types to internal types
- Map all relevant fields from Gitea SDK response to internal type structures
- Handle nil/optional fields appropriately in transformation functions
- Preserve all metadata (timestamps, URLs, etc.) during transformation

### Error Handling Strategy
- Custom error types: `APIError`, `AuthError`, `NetworkError`, `ValidationError`
- Wrap Gitea SDK errors with context using `fmt.Errorf` with %w verb
- Include HTTP status codes and response bodies in error messages for debugging
- Distinguish between client errors (4xx) and server errors (5xx) for appropriate handling
- Handle rate limiting and temporary failures with exponential backoff

### Testing Approach
- Mock Gitea SDK client using interface for unit testing
- Test request building with various filter combinations
- Test response transformation with sample Gitea API responses
- Integration tests with actual Forgejo instance when available
- Test error scenarios including network failures and authentication issues

### Performance Considerations
- Reuse ForgejoClient instance across multiple calls to maintain connection pooling
- Implement efficient pagination handling for large result sets
- Consider concurrent API calls for independent operations
- Cache authentication validation to avoid repeated token checks
- Benchmark API call performance and response transformation

## Approach

The Gitea SDK client will be implemented as a standalone package that provides a clean Go interface for interacting with Forgejo repositories through the official API. The client will abstract away the complexity of HTTP communication, authentication, and response handling, making it easy for the MCP server to integrate with Forgejo instances.

The implementation will follow a layered architecture:
1. **HTTP client layer**: Gitea SDK handles HTTP communication and authentication
2. **Request builder layer**: Constructs properly formatted API requests with filters
3. **Response transformer layer**: Converts Gitea SDK types to internal data structures
4. **Public API layer**: Provides high-level methods for common operations

## External Dependencies

- **Gitea SDK**: `code.gitea.io/sdk/gitea@v0.21.0` - Official Go client for Gitea/Forgejo API
- **Standard library packages**: `context`, `fmt`, `time`, `net/http`
- **Internal packages**: `types` package for data structures, `auth` for token management, `config` for client configuration