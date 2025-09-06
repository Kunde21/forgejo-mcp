# Migration Guide: Repository-Based Queries

This guide helps you migrate from user-based to repository-based queries in the Forgejo MCP server.

## Overview

The Forgejo MCP server has been updated to support repository-specific queries instead of user-based queries. This change improves performance and provides more targeted results for AI agents working with specific repositories.

## Breaking Changes

### Before (User-Based)
```json
{
  "tool": "pr_list",
  "parameters": {}
}
```

### After (Repository-Based)
```json
{
  "tool": "pr_list",
  "parameters": {
    "repository": "owner/repo-name"
  }
}
```

## Migration Steps

### 1. Update Tool Calls

**Old Usage:**
```javascript
// Get PRs for current user context
const prs = await mcp.callTool('pr_list', {});
```

**New Usage:**
```javascript
// Get PRs for specific repository
const prs = await mcp.callTool('pr_list', {
  repository: 'my-org/my-project'
});

// Alternative: Use CWD resolution
const prs = await mcp.callTool('pr_list', {
  cwd: '/path/to/repository'
});
```

### 2. Update Response Handling

**New Response Format:**
```json
{
  "prs": [
    {
      "number": 123,
      "title": "Fix authentication bug",
      "repository": {
        "owner": "my-org",
        "name": "my-project",
        "full_name": "my-org/my-project"
      }
    }
  ]
}
```

**Migration Code:**
```javascript
// Before
const prs = response.prs;

// After - handle repository metadata
const prs = response.prs;
prs.forEach(pr => {
  console.log(`PR #${pr.number} in ${pr.repository.full_name}`);
});
```

### 3. Handle Repository Resolution

**Option 1: Explicit Repository Parameter**
```javascript
const result = await mcp.callTool('pr_list', {
  repository: 'github/octocat'
});
```

**Option 2: CWD-Based Resolution**
```javascript
const result = await mcp.callTool('pr_list', {
  cwd: process.cwd() // Server resolves to repository
});
```

### 4. Update Error Handling

**New Error Scenarios:**
- `INVALID_REPOSITORY_FORMAT`: Repository parameter doesn't match "owner/repo" format
- `REPOSITORY_NOT_FOUND`: Specified repository doesn't exist
- `REPOSITORY_ACCESS_DENIED`: User lacks permission to access repository
- `REPOSITORY_NOT_IN_CWD`: CWD doesn't contain a valid git repository

**Migration Example:**
```javascript
try {
  const result = await mcp.callTool('pr_list', {
    repository: 'invalid-format'
  });
} catch (error) {
  if (error.code === 'INVALID_REPOSITORY_FORMAT') {
    console.log('Please use format: owner/repo');
  }
}
```

## Repository Parameter Formats

### Valid Formats
- `owner/repo` - Standard format
- `organization/project` - Organization-owned repositories
- `user/personal-project` - User-owned repositories

### Invalid Formats
- `repo-only` - Missing owner
- `owner/repo/extra` - Too many parts
- `owner/` - Empty repository name
- `/repo` - Empty owner

## Testing Your Migration

### Test Cases to Verify
1. ✅ Repository parameter works with valid format
2. ✅ CWD parameter resolves correctly
3. ✅ Error handling for invalid repositories
4. ✅ Response includes repository metadata
5. ✅ Pagination and filtering still work
6. ✅ Authentication and permissions work

### Example Test Script
```javascript
// Test repository parameter
const testRepo = async () => {
  try {
    const result = await mcp.callTool('pr_list', {
      repository: 'octocat/Hello-World'
    });
    console.log('✅ Repository parameter works');
  } catch (error) {
    console.log('❌ Repository parameter failed:', error.message);
  }
};
```

## Performance Considerations

- Repository-based queries are more efficient than user-based queries
- Results are filtered at the database level
- Reduced data transfer for targeted queries
- Better caching opportunities for repository-specific data

## Rollback Plan

If you need to rollback:

1. **Temporary Workaround:** Use repository listing to find target repositories
2. **Full Rollback:** Server can be configured to support both query types during transition
3. **Gradual Migration:** Update integrations incrementally

## Support

For migration assistance:
- Check the updated API documentation
- Review the troubleshooting guide
- Test with the integration test suite
- Contact the development team for specific integration issues