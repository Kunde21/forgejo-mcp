# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-09-20-forgejo-remote-support/spec.md

## Configuration API Changes

### Environment Variables
- **FORGEJO_CLIENT_TYPE**: Set to "forgejo", "gitea", or "auto" to select SDK, defaults to "gitea"
- **FORGEJO_REMOTE_URL**: Forgejo instance URL (same as existing GITEA_REMOTE_URL)
- **FORGEJO_AUTH_TOKEN**: Authentication token for Forgejo instance (same as existing GITEA_AUTH_TOKEN)

### Configuration File Structure
```yaml
host: "localhost"
port: 8080
remote_url: "https://forgejo.example.com"
auth_token: "your-token-here"
client_type: "auto"  # or "forgejo", "gitea"
```

## Client Factory API

### NewFromConfig() Enhancement
**Purpose:** Factory method to create appropriate client based on configuration with automatic detection
**Parameters:** Config struct with ClientType field
**Response:** ClientInterface implementation (GiteaClient or ForgejoClient)
**Errors:** Invalid client type, SDK initialization failures, detection failures

### Client Selection Logic
```go
func NewFromConfig(cfg *Config) (ClientInterface, error) {
    // Auto-detect if ClientType is "auto" or empty
    if cfg.ClientType == "auto" || cfg.ClientType == "" {
        detectedType, err := detectRemoteType(cfg.RemoteURL, cfg.AuthToken)
        if err != nil {
            return nil, fmt.Errorf("failed to detect remote type: %w", err)
        }
        cfg.ClientType = detectedType
    }
    
    switch strings.ToLower(cfg.ClientType) {
    case "forgejo":
        return NewForgejoClient(cfg.RemoteURL, cfg.AuthToken)
    case "gitea":
        return NewGiteaClient(cfg.RemoteURL, cfg.AuthToken)
    default:
        return nil, fmt.Errorf("invalid client type: %s", cfg.ClientType)
    }
}
```

## Version Detection API

### detectRemoteType() Function
**Purpose:** Automatically detect whether remote instance is Gitea or Forgejo
**Parameters:** remoteURL string, authToken string
**Response:** "gitea", "forgejo", or error
**Errors:** Connection failures, invalid responses, detection failures

### Version Endpoint Usage
**Endpoint:** `/api/v1/version` (both Gitea and Forgejo support this)
**Method:** GET
**Authentication:** Uses provided auth token if available
**Response Examples:**
- Forgejo: `{"version":"12.0.1-120-abfc8432+gitea-1.22.0"}`
- Gitea: `{"version":"1.25.0+dev-376-g223205cc6b"}`

### Detection Algorithm
```go
func detectRemoteType(remoteURL, authToken string) (string, error) {
    // Make HTTP request to /api/v1/version
    // Parse JSON response
    // Analyze version string for patterns:
    //   - Contains "forgejo" -> "forgejo"
    //   - Contains "gitea" -> "gitea"  
    //   - Version starts with "12." -> "forgejo"
    //   - Version starts with "1." -> "gitea"
    //   - Default to "gitea" if ambiguous
}
```

## ForgejoClient API Methods

### ListIssues
**Purpose:** Retrieve issues from Forgejo repository
**Parameters:** context, repository string (owner/repo), limit int, offset int
**Response:** []Issue slice with Number, Title, State fields
**Errors:** Invalid repository format, API errors, authentication failures

### CreateIssueComment
**Purpose:** Add comment to Forgejo issue
**Parameters:** context, repository string, issueNumber int, comment string
**Response:** nil on success
**Errors:** Invalid issue number, permission errors, API failures

### ListIssueComments
**Purpose:** Retrieve comments from Forgejo issue
**Parameters:** context, repository string, issueNumber int, limit int, offset int
**Response:** []IssueComment slice with ID, Body, User fields
**Errors:** Invalid issue number, pagination errors, API failures

### EditIssueComment
**Purpose:** Update existing comment on Forgejo issue
**Parameters:** context, repository string, commentID int, newContent string
**Response:** nil on success
**Errors:** Invalid comment ID, permission errors, API failures

### ListPullRequests
**Purpose:** Retrieve pull requests from Forgejo repository
**Parameters:** context, repository string, limit int, offset int, state string
**Response:** []PullRequest slice with Number, Title, State fields
**Errors:** Invalid repository format, API errors, authentication failures

### ListPullRequestComments
**Purpose:** Retrieve comments from Forgejo pull request
**Parameters:** context, repository string, prNumber int, limit int, offset int
**Response:** []PRComment slice with ID, Body, User fields
**Errors:** Invalid PR number, pagination errors, API failures

### CreatePullRequestComment
**Purpose:** Add comment to Forgejo pull request
**Parameters:** context, repository string, prNumber int, comment string
**Response:** nil on success
**Errors:** Invalid PR number, permission errors, API failures

### EditPullRequestComment
**Purpose:** Update existing comment on Forgejo pull request
**Parameters:** context, repository string, commentID int, newContent string
**Response:** nil on success
**Errors:** Invalid comment ID, permission errors, API failures

## Error Handling API

### Error Types
- **ClientTypeError**: Invalid client type specified in configuration
- **SDKInitError**: Failed to initialize Forgejo SDK client
- **APIError**: Forgejo API returned an error response
- **AuthError**: Authentication or authorization failure
- **ValidationError**: Invalid input parameters

### Error Response Format
All errors should include:
- Clear, actionable error message
- Underlying SDK error when applicable
- Suggested resolution when possible