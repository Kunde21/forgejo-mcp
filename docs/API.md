# Forgejo MCP API Documentation

This document describes the API endpoints and tools available through the Forgejo MCP server.

## Overview

The Forgejo MCP server provides a Model Context Protocol interface that allows AI agents to interact with Forgejo repositories using the official Gitea SDK for direct API integration. This provides improved performance, reliability, and comprehensive error handling compared to CLI-based approaches.

## Authentication

Authentication to the server must happen outside of the MCP calls, before the agent connects. The server expects a valid Forgejo authentication token to be configured either through environment variables or a configuration file.

## Available Tools

### PR List Tool

**Name:** `pr_list`

**Description:** Lists pull requests for a specific repository.

**Parameters:**
- `repository` (string, required): Repository identifier in format "owner/repo"
- `cwd` (string, optional): Current working directory to resolve repository from (alternative to repository parameter)

**Response:**
```json
{
  "prs": [
    {
      "number": 123,
      "title": "Fix bug in user authentication",
      "author": "john_doe",
      "state": "open",
      "created_at": "2025-08-26T10:30:00Z",
      "updated_at": "2025-08-26T10:30:00Z",
      "repository": {
        "owner": "user",
        "name": "my-project",
        "full_name": "user/my-project"
      }
    }
  ]
}
```

### Issue List Tool

**Name:** `issue_list`

**Description:** Lists issues for a specific repository.

**Parameters:**
- `repository` (string, required): Repository identifier in format "owner/repo"
- `cwd` (string, optional): Current working directory to resolve repository from (alternative to repository parameter)

**Response:**
```json
{
  "issues": [
    {
      "number": 456,
      "title": "Add support for dark mode",
      "author": "jane_smith",
      "state": "open",
      "labels": ["enhancement", "ui"],
      "created_at": "2025-08-26T09:15:00Z",
      "repository": {
        "owner": "user",
        "name": "my-project",
        "full_name": "user/my-project"
      }
    }
  ]
}
```

### Repository List Tool

**Name:** `repo_list`

**Description:** Lists all repositories accessible to the authenticated user.

**Parameters:** None

**Response:**
```json
{
  "repositories": [
    {
      "id": 123,
      "name": "my-project",
      "full_name": "user/my-project",
      "description": "A sample project",
      "private": false,
      "owner": "user",
      "url": "https://forgejo.example.com/user/my-project"
    }
  ]
}
```

## Error Handling

All tools will return appropriate error messages when operations fail. Error responses follow the MCP standard error format with descriptive error messages.

### Error Codes

#### Repository-Related Errors

- `INVALID_REPOSITORY_FORMAT`: Repository parameter doesn't match required "owner/repo" format
- `REPOSITORY_NOT_FOUND`: Specified repository doesn't exist
- `REPOSITORY_ACCESS_DENIED`: User lacks permission to access the repository
- `REPOSITORY_NOT_IN_CWD`: Current working directory doesn't contain a valid git repository
- `PRIVATE_REPOSITORY_ACCESS_DENIED`: Cannot access private repository without proper permissions

#### Authentication & Authorization Errors

- `AUTHENTICATION_FAILED`: Invalid or missing authentication token
- `AUTHORIZATION_FAILED`: User not authorized for the requested operation
- `TOKEN_EXPIRED`: Authentication token has expired
- `TOKEN_INVALID`: Authentication token is malformed or invalid

#### Parameter Validation Errors

- `MISSING_REPOSITORY_PARAMETER`: Neither repository nor cwd parameter provided
- `INVALID_PARAMETER_FORMAT`: Parameter value doesn't match expected format
- `UNSUPPORTED_PARAMETER`: Parameter not supported for this tool

#### Network & Service Errors

- `NETWORK_ERROR`: Network connectivity issues
- `SERVICE_UNAVAILABLE`: Forgejo service is temporarily unavailable
- `RATE_LIMIT_EXCEEDED`: API rate limit has been exceeded
- `TIMEOUT_ERROR`: Request timed out

#### Data Validation Errors

- `INVALID_STATE_PARAMETER`: State parameter must be 'open', 'closed', or 'all'
- `INVALID_PAGE_PARAMETER`: Page parameter must be a positive integer
- `INVALID_PER_PAGE_PARAMETER`: Per page parameter must be between 1 and 100

### Error Response Format

```json
{
  "error": {
    "code": "REPOSITORY_NOT_FOUND",
    "message": "Repository 'nonexistent/repo' not found or access denied",
    "details": {
      "repository": "nonexistent/repo",
      "suggestion": "Verify the repository exists and you have access to it"
    }
  }
}
```

### Common Error Scenarios

1. **Missing Repository Parameter:**
   ```json
   {
     "tool": "pr_list",
     "parameters": {}
   }
   // Returns: MISSING_REPOSITORY_PARAMETER
   ```

2. **Invalid Repository Format:**
   ```json
   {
     "tool": "pr_list",
     "parameters": {
       "repository": "invalid-format"
     }
   }
   // Returns: INVALID_REPOSITORY_FORMAT
   ```

3. **Private Repository Access:**
   ```json
   {
     "tool": "issue_list",
     "parameters": {
       "repository": "private-org/private-repo"
     }
   }
   // Returns: REPOSITORY_ACCESS_DENIED (if user lacks access)
   ```

## Usage Examples

### Basic Repository Queries

**List PRs for a specific repository:**
```json
{
  "tool": "pr_list",
  "parameters": {
    "repository": "octocat/Hello-World"
  }
}
```

**List issues using CWD resolution:**
```json
{
  "tool": "issue_list",
  "parameters": {
    "cwd": "/home/user/projects/my-repo"
  }
}
```

### Response Examples

**PR List Response:**
```json
{
  "prs": [
    {
      "number": 42,
      "title": "Add dark mode support",
      "author": "contributor",
      "state": "open",
      "created_at": "2025-09-06T10:00:00Z",
      "repository": {
        "owner": "my-org",
        "name": "web-app",
        "full_name": "my-org/web-app"
      }
    }
  ]
}
```

**Issue List Response:**
```json
{
  "issues": [
    {
      "number": 15,
      "title": "Fix login validation",
      "author": "developer",
      "state": "open",
      "labels": ["bug", "high-priority"],
      "created_at": "2025-09-05T14:30:00Z",
      "repository": {
        "owner": "my-org",
        "name": "api-server",
        "full_name": "my-org/api-server"
      }
    }
  ]
}
```

### Error Examples

**Invalid Repository Format:**
```json
{
  "error": {
    "code": "INVALID_REPOSITORY_FORMAT",
    "message": "Repository parameter must be in format 'owner/repo'"
  }
}
```

**Repository Not Found:**
```json
{
  "error": {
    "code": "REPOSITORY_NOT_FOUND",
    "message": "Repository 'nonexistent/repo' not found or access denied"
  }
}
```

### Advanced Usage

**With Filtering and Pagination:**
```json
{
  "tool": "pr_list",
  "parameters": {
    "repository": "my-org/project",
    "state": "closed",
    "per_page": 10,
    "page": 1
  }
}
```

**Multiple Repository Queries:**
```javascript
// Query multiple repositories
const repos = ['org/repo1', 'org/repo2', 'user/personal'];

for (const repo of repos) {
  const prs = await mcp.callTool('pr_list', { repository: repo });
  console.log(`${repo}: ${prs.prs.length} open PRs`);
}
```

## Performance Considerations

### Repository-Based Query Benefits

- **Faster Queries:** Repository-specific queries are more efficient than user-based queries
- **Reduced Data Transfer:** Only relevant repository data is retrieved
- **Better Caching:** Repository-specific results can be cached more effectively
- **Lower Latency:** Direct repository queries reduce API call overhead

### Optimization Strategies

#### Use Pagination
```json
{
  "tool": "pr_list",
  "parameters": {
    "repository": "large-org/big-repo",
    "per_page": 50,
    "page": 1
  }
}
```

#### Apply Filters Early
```json
{
  "tool": "issue_list",
  "parameters": {
    "repository": "my-org/project",
    "state": "open",
    "labels": "bug,high-priority"
  }
}
```

#### Cache Repository Metadata
Repository information is included in responses and can be cached to avoid repeated lookups.

### Performance Guidelines

#### Recommended Limits
- **per_page:** 50-100 items (default: 30)
- **Concurrent Requests:** Limit to 5-10 simultaneous repository queries
- **Cache TTL:** 5-15 minutes for repository metadata

#### Monitoring Performance
- Track response times for different repository sizes
- Monitor API rate limit usage
- Log slow queries (>5 seconds) for optimization

#### Large Repository Handling
For repositories with thousands of PRs/issues:
1. Use smaller page sizes
2. Implement client-side filtering
3. Consider time-based pagination
4. Cache frequently accessed data

### Network Optimization

#### Connection Reuse
The MCP server maintains persistent connections to Forgejo, reducing connection overhead.

#### Request Batching
Multiple repository queries can be batched to reduce round trips:
```javascript
// Instead of sequential calls
const results = await Promise.all([
  mcp.callTool('pr_list', { repository: 'org/repo1' }),
  mcp.callTool('pr_list', { repository: 'org/repo2' })
]);
```

#### Error Handling Impact
Repository validation happens early, preventing unnecessary API calls for invalid repositories.
```