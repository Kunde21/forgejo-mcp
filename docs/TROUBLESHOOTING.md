# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Forgejo MCP server, especially after migrating to repository-based queries.

## Quick Diagnosis

### Check Server Status
```bash
# Test server connectivity
curl -X POST http://localhost:3000/tools \
  -H "Content-Type: application/json" \
  -d '{"method": "tools/list"}'
```

### Verify Configuration
```bash
# Check environment variables
echo $FORGEJO_MCP_FORGEJO_URL
echo $FORGEJO_MCP_AUTH_TOKEN

# Test token validity
curl -H "Authorization: token $FORGEJO_MCP_AUTH_TOKEN" \
  $FORGEJO_MCP_FORGEJO_URL/api/v1/user
```

## Common Issues and Solutions

### Repository Parameter Errors

#### INVALID_REPOSITORY_FORMAT
**Symptoms:** Error when using repository parameter
**Cause:** Repository identifier doesn't match "owner/repo" format
**Solution:**
```javascript
// ❌ Wrong
await mcp.callTool('pr_list', { repository: 'my-repo' });

// ✅ Correct
await mcp.callTool('pr_list', { repository: 'my-org/my-repo' });
```

#### REPOSITORY_NOT_FOUND
**Symptoms:** "Repository not found" error
**Possible Causes:**
- Repository doesn't exist
- Typo in repository name
- Repository is private and user lacks access
- Repository was deleted or renamed

**Solutions:**
1. Verify repository exists: `curl $FORGEJO_URL/api/v1/repos/owner/repo`
2. Check repository name spelling
3. Ensure user has access to private repositories
4. Update to correct repository name if renamed

#### REPOSITORY_ACCESS_DENIED
**Symptoms:** Access denied for repository
**Cause:** Insufficient permissions for private repository
**Solutions:**
1. Check user's repository permissions
2. Verify authentication token has correct scopes
3. Request access from repository owner
4. Use public repository instead

### Authentication Issues

#### AUTHENTICATION_FAILED
**Symptoms:** All requests fail with auth errors
**Cause:** Invalid or missing authentication token
**Solutions:**
1. Verify token is set: `echo $FORGEJO_MCP_AUTH_TOKEN`
2. Check token format and validity
3. Regenerate token if expired
4. Ensure token has required scopes (repo, read:user)

#### TOKEN_EXPIRED
**Symptoms:** Intermittent authentication failures
**Cause:** Forgejo token has expired
**Solutions:**
1. Generate new token in Forgejo settings
2. Update environment variable or config file
3. Restart MCP server with new token

### Network and Connectivity

#### NETWORK_ERROR
**Symptoms:** Connection timeouts or network errors
**Cause:** Network connectivity issues
**Solutions:**
1. Check Forgejo server status
2. Verify network connectivity to Forgejo instance
3. Check firewall settings
4. Try different network connection

#### SERVICE_UNAVAILABLE
**Symptoms:** 503 errors from Forgejo API
**Cause:** Forgejo server maintenance or overload
**Solutions:**
1. Check Forgejo server status page
2. Wait for maintenance to complete
3. Contact Forgejo administrator
4. Implement retry logic with exponential backoff

### Parameter Validation

#### MISSING_REPOSITORY_PARAMETER
**Symptoms:** "Missing repository parameter" error
**Cause:** Neither repository nor cwd parameter provided
**Solutions:**
```javascript
// Option 1: Explicit repository
await mcp.callTool('pr_list', { repository: 'owner/repo' });

// Option 2: Use CWD resolution
await mcp.callTool('pr_list', { cwd: process.cwd() });
```

#### REPOSITORY_NOT_IN_CWD
**Symptoms:** CWD resolution fails
**Cause:** Current directory is not a git repository or missing remote
**Solutions:**
1. Ensure you're in a git repository: `git status`
2. Verify remote is configured: `git remote -v`
3. Check remote URL format matches Forgejo instance
4. Use explicit repository parameter instead

### Performance Issues

#### Slow Response Times
**Symptoms:** Queries take longer than expected
**Possible Causes:**
- Large repository with many PRs/issues
- Network latency
- Forgejo server performance
- Inefficient query parameters

**Solutions:**
1. Use pagination: `per_page=50` instead of default 100
2. Add filters to reduce result set: `state=open`
3. Check network latency to Forgejo server
4. Monitor Forgejo server performance

#### Rate Limiting
**Symptoms:** 429 errors or rate limit exceeded
**Cause:** Too many requests in short time period
**Solutions:**
1. Implement request throttling
2. Use caching for frequently accessed data
3. Spread requests over time
4. Consider upgrading Forgejo plan for higher limits

### Data Issues

#### Empty Results
**Symptoms:** Query returns empty results unexpectedly
**Possible Causes:**
- Repository has no PRs/issues matching criteria
- Incorrect filters or parameters
- Repository is empty
- User lacks permission to see content

**Solutions:**
1. Remove filters to see all items
2. Check repository has expected content
3. Verify user permissions
4. Test with different repository

#### Incorrect Data
**Symptoms:** Data doesn't match Forgejo web interface
**Cause:** Caching or synchronization issues
**Solutions:**
1. Clear any local caches
2. Wait a few minutes for synchronization
3. Check Forgejo web interface directly
4. Report issue if persists

## Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
# Environment variable
export FORGEJO_MCP_DEBUG=true

# Or command line
forgejo-mcp serve --debug
```

Debug logs will show:
- Request/response details
- Authentication attempts
- Repository resolution steps
- API call timings

## Testing Tools

### Health Check
```bash
# Test server health
curl http://localhost:3000/health

# Test tool availability
curl -X POST http://localhost:3000/tools \
  -H "Content-Type: application/json" \
  -d '{"method": "tools/list"}'
```

### Repository Validation
```bash
# Test repository access
curl -H "Authorization: token $TOKEN" \
  $FORGEJO_URL/api/v1/repos/owner/repo
```

### Token Validation
```bash
# Test token validity
curl -H "Authorization: token $TOKEN" \
  $FORGEJO_URL/api/v1/user
```

## Getting Help

If you can't resolve an issue:

1. **Check Logs:** Enable debug mode and review server logs
2. **Test Manually:** Use curl to test API calls directly
3. **Verify Configuration:** Double-check all settings
4. **Update Dependencies:** Ensure you're using latest version
5. **Report Issue:** Create issue with debug logs and configuration details

## Prevention

- **Monitor Token Expiry:** Set reminders to rotate tokens
- **Use Environment Variables:** Avoid hardcoding sensitive data
- **Implement Error Handling:** Add try/catch blocks around MCP calls
- **Test Regularly:** Run integration tests after configuration changes
- **Keep Updated:** Use latest version for bug fixes and improvements