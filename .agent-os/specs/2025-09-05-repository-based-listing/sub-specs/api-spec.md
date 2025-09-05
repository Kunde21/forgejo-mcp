# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-09-05-repository-based-listing/spec.md

> Created: 2025-09-05
> Version: 1.0.0

## Endpoints

### Current Implementation (User-Based)

#### pr_list Tool
**Current Parameters:**
- `state` (optional): Filter by PR state ("open", "closed", "all")
- `author` (optional): Filter by author username
- `limit` (optional): Maximum number of results to return

**Current Behavior:**
- Uses hardcoded repository context ("example-owner/example-repo")
- Lists PRs from the configured repository
- Returns PR data with basic metadata

**Current Response Format:**
```json
{
  "pullRequests": [
    {
      "number": 123,
      "title": "Fix authentication bug",
      "state": "open",
      "author": "developer1",
      "createdAt": "2025-09-05T10:30:00Z",
      "updatedAt": "2025-09-05T14:20:00Z",
      "type": "pull_request",
      "url": "https://gitea.example.com/example-owner/example-repo/pulls/123"
    }
  ],
  "total": 1
}
```

#### issue_list Tool
**Current Parameters:**
- `state` (optional): Filter by issue state ("open", "closed", "all")
- `author` (optional): Filter by author username
- `labels` (optional): Array of label names to filter by
- `limit` (optional): Maximum number of results to return

**Current Behavior:**
- Uses global user context (all repositories for authenticated user)
- Lists issues across all accessible repositories
- Returns issue data with basic metadata

**Current Response Format:**
```json
{
  "issues": [
    {
      "number": 456,
      "title": "Add new feature request",
      "state": "open",
      "author": "developer2",
      "createdAt": "2025-09-04T09:15:00Z",
      "updatedAt": "2025-09-05T11:45:00Z",
      "type": "issue",
      "url": "https://gitea.example.com/example-owner/example-repo/issues/456"
    }
  ],
  "total": 1
}
```

### Proposed Changes (Repository-Based)

#### pr_list Tool (Modified)
**New Parameters:**
- `repository` (optional): Repository identifier in format "owner/repo"
- `cwd` (optional): Current working directory
- `state` (optional): Filter by PR state ("open", "closed", "all")
- `author` (optional): Filter by author username
- `limit` (optional): Maximum number of results to return

**New Behavior:**
- Accepts repository parameter or working directory to specify target repository
- One of `repository` or `cwd` is required
- Validates repository exists and user has access
- Lists PRs only from the specified repository
- Returns repository-specific PR data

**New Request Examples:**
```json
{
  "repository": "my-org/my-project",
  "state": "open",
  "limit": 10
}
```

```json
{
  "cwd": "/path/to/project/repo",
  "state": "open",
  "limit": 10
}
```

**New Response Format:**
```json
{
  "pullRequests": [
    {
      "number": 123,
      "title": "Fix authentication bug",
      "state": "open",
      "author": "developer1",
      "createdAt": "2025-09-05T10:30:00Z",
      "updatedAt": "2025-09-05T14:20:00Z",
      "type": "pull_request",
      "url": "https://gitea.example.com/my-org/my-project/pulls/123",
      "repository": {
        "owner": "my-org",
        "name": "my-project",
        "fullName": "my-org/my-project"
      }
    }
  ],
  "total": 1,
  "repository": "my-org/my-project"
}
```

#### issue_list Tool (Modified)
**New Parameters:**
- `repository` (optional): Repository identifier in format "owner/repo"
- `cwd` (optional): Current working directory
- `state` (optional): Filter by issue state ("open", "closed", "all")
- `author` (optional): Filter by author username
- `labels` (optional): Array of label names to filter by
- `limit` (optional): Maximum number of results to return

**New Behavior:**
- Accepts repository parameter or working directory to specify target repository
- One of `repository` or `cwd` is required
- Validates repository exists and user has access
- Lists issues only from the specified repository
- Returns repository-specific issue data

**New Request Example:**
```json
{
  "repository": "my-org/my-project",
  "state": "open",
  "labels": ["bug", "enhancement"]
}
```

```json
{
  "cwd": "/path/to/project/repo",
  "state": "open",
  "labels": ["bug", "enhancement"]
}
```

**New Response Format:**
```json
{
  "issues": [
    {
      "number": 456,
      "title": "Add new feature request",
      "state": "open",
      "author": "developer2",
      "createdAt": "2025-09-04T09:15:00Z",
      "updatedAt": "2025-09-05T11:45:00Z",
      "type": "issue",
      "url": "https://gitea.example.com/my-org/my-project/issues/456",
      "repository": {
        "owner": "my-org",
        "name": "my-project",
        "fullName": "my-org/my-project"
      },
      "labels": ["enhancement"]
    }
  ],
  "total": 1,
  "repository": "my-org/my-project"
}
```

## Parameter Changes

### Repository Parameter
- **Type:** String (required)
- **Format:** "owner/repository" or "organization/repository"
- **Validation:**
  - Must contain exactly one forward slash
  - Owner/repository names must be valid identifiers
  - Repository must exist and be accessible to authenticated user
  - Supports both user-owned and organization-owned repositories

### Backward Compatibility
- **Breaking Change:** One of `repository` or `cwd` parameter becomes required
- **Migration Path:** Existing integrations must update to include repository parameter
- **Deprecation:** No deprecation period planned (immediate breaking change)
- **Error Handling:** Clear error messages for missing repository parameter

## Response Format Changes

### Enhanced Repository Metadata
All responses now include repository context information:

```json
{
  "repository": {
    "owner": "string",
    "name": "string", 
    "fullName": "string"
  }
}
```

### Consistent Structure
- All list operations return `total` count
- Repository information included in each item
- Standardized error response format
- Pagination metadata preserved

### Error Responses

#### Invalid Repository Format
```json
{
  "error": "INVALID_REPOSITORY_FORMAT",
  "message": "Repository must be in format 'owner/repo'",
  "details": "Received: 'invalid-format'"
}
```

#### Repository Not Found
```json
{
  "error": "REPOSITORY_NOT_FOUND", 
  "message": "Repository 'nonexistent/repo' not found or access denied",
  "details": "Verify repository exists and you have access permissions"
}
```

#### Repository Access Denied
```json
{
  "error": "REPOSITORY_ACCESS_DENIED",
  "message": "Access denied to repository 'private/repo'",
  "details": "You do not have permission to access this repository"
}
```

## Controllers

### Handler Modifications

#### SDKPRListHandler Changes
```go
// Updated method signature
func (h *SDKPRListHandler) HandlePRListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Repository string `json:"repository"`  // New parameter
    CWD        string `json:"cwd"`  // New parameter
    State      string `json:"state,omitempty"`
    Author     string `json:"author,omitempty"`
    Limit      int    `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error)

// New repository validation logic
func (h *SDKPRListHandler) validateRepository(repo string) error {
    parts := strings.Split(repo, "/")
    if len(parts) != 2 {
        return fmt.Errorf("invalid repository format: %s", repo)
    }
    owner, name := parts[0], parts[1]
    
    // Validate repository exists and is accessible
    _, _, err := h.client.GetRepo(owner, name)
    return err
}
```

#### SDKIssueListHandler Changes
```go
// Updated method signature  
func (h *SDKIssueListHandler) HandleIssueListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Repository string   `json:"repository"`  // New parameter
    CWD        string   `json:"cwd"`  // New parameter
    State      string   `json:"state,omitempty"`
    Author     string   `json:"author,omitempty"`
    Labels     []string `json:"labels,omitempty"`
    Limit      int      `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error)

// Repository validation (shared with PR handler)
func (h *SDKIssueListHandler) validateRepository(repo string) error {
    // Same validation logic as PR handler
}
```

### Tool Registration Updates
```go
// Updated tool descriptions
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_list",
    Description: "List pull requests from a specific Forgejo repository",
}, prHandler.HandlePRListRequest)

mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "issue_list", 
    Description: "List issues from a specific Forgejo repository",
}, issueHandler.HandleIssueListRequest)
```

## Migration Guide

### For Existing Integrations
1. **Update API Calls:** Add `repository` parameter to all requests
2. **Handle New Response Format:** Parse repository metadata from responses
3. **Error Handling:** Handle new repository-related error codes
4. **Testing:** Verify functionality with repository-specific data

### Example Migration
```javascript
// Before (user-based)
const prs = await client.callTool("pr_list", {
  state: "open",
  limit: 10
});

// After (repository-based)  
const prs = await client.callTool("pr_list", {
  repository: "my-org/my-project",
  state: "open", 
  limit: 10
});
```

```javascript
// Before (user-based)
const prs = await client.callTool("pr_list", {
  state: "open",
  limit: 10
});

// After (directory-based)  
const prs = await client.callTool("pr_list", {
  cwd: "/path/to/project/repo",
  state: "open", 
  limit: 10
});
```
