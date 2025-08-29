# Troubleshooting Guide

This document provides solutions for common issues you might encounter when using the Forgejo MCP server.

## Common Issues

### 1. Authentication Failures

**Symptoms:**
- "401 Unauthorized" errors
- "Invalid token" messages
- Empty or incomplete responses

**Solutions:**
1. Verify your `FORGEJO_MCP_AUTH_TOKEN` environment variable is set correctly
2. Ensure the token has appropriate permissions for the operations you're trying to perform
3. Check that your `FORGEJO_MCP_FORGEJO_URL` points to the correct Forgejo instance
4. Test your token directly with the Forgejo API using curl:
   ```bash
   curl -H "Authorization: token YOUR_TOKEN" https://your.forgejo.instance/api/v1/user
   ```

### 2. Connection Timeouts

**Symptoms:**
- "Timeout exceeded" errors
- Slow responses or hanging requests
- "Failed to connect" messages

**Solutions:**
1. Check network connectivity to your Forgejo instance
2. Increase timeout values in your configuration:
   ```yaml
   # config.yaml
   timeout: 60s  # Increase from default 30s
   ```
3. For large repositories, consider using pagination to reduce response size

### 3. CLI Command Failures

**Symptoms:**
- "tea command not found" errors
- "Failed to execute tea" messages
- Empty or malformed output

**Solutions:**
1. Ensure the `tea` CLI tool is installed and in your PATH:
   ```bash
   which tea
   tea --version
   ```
2. Verify `tea` is properly configured with your Forgejo instance:
   ```bash
   tea login add --name myforgejo --url https://your.forgejo.instance --token YOUR_TOKEN
   tea login list
   ```
3. Check the `FORGEJO_MCP_TEA_PATH` environment variable points to the correct executable

### 4. Performance Issues

**Symptoms:**
- Slow response times
- High CPU or memory usage
- Frequent timeouts

**Solutions:**
1. Enable caching for frequently accessed resources:
   ```yaml
   # config.yaml
   cache:
     enabled: true
     size: 1000
     ttl: 5m
   ```
2. Use batch processing for multiple related requests
3. Implement proper pagination for large result sets
4. Monitor and optimize your Forgejo instance performance

### 5. Rate Limiting

**Symptoms:**
- "429 Too Many Requests" errors
- Intermittent failures
- Requests being rejected

**Solutions:**
1. Implement exponential backoff in your client applications
2. Check your Forgejo instance rate limiting configuration
3. Use caching to reduce the number of API calls
4. Consider using the Gitea SDK client which provides better rate limit handling

## Debugging

### Enable Debug Logging

Set the log level to debug for more detailed information:

```bash
FORGEJO_MCP_LOG_LEVEL=debug forgejo-mcp serve
```

Or in your configuration file:
```yaml
log_level: debug
```

### Test Individual Components

1. **Test the tea CLI directly:**
   ```bash
   tea pr list --repo owner/repo --state open
   ```

2. **Test the Gitea SDK client:**
   ```go
   client, err := client.New("https://your.forgejo.instance", "your-token")
   if err != nil {
       log.Fatal(err)
   }
   
   prs, err := client.ListPRs("owner", "repo", &client.PullRequestFilters{State: client.StateOpen})
   ```

3. **Test the MCP server endpoints:**
   ```bash
   curl -X POST http://localhost:3000 \
     -H "Content-Type: application/json" \
     -d '{"method": "pr_list", "params": {"state": "open"}}'
   ```

## Configuration Issues

### Environment Variables vs Configuration File

Ensure consistency between environment variables and configuration files:

```bash
# Environment variables take precedence over config file
export FORGEJO_MCP_FORGEJO_URL="https://your.forgejo.instance"
export FORGEJO_MCP_AUTH_TOKEN="your-token"
```

### Configuration File Location

The server looks for configuration files in this order:
1. Path specified by `--config` flag
2. `./config.yaml` in the current directory
3. `~/.forgejo-mcp/config.yaml` in the user's home directory

## Version Compatibility

### Forgejo Version Support

This MCP server is tested with Forgejo versions 1.20+. Older versions may work but are not officially supported.

### Go Version Requirements

- Go 1.24.6 or later is required for building from source
- Pre-built binaries are available for most platforms

### Tea CLI Compatibility

The server is compatible with tea CLI version 0.7.0 and later. Check your version with:
```bash
tea --version
```

## Getting Help

If you're still experiencing issues:

1. Check the [GitHub Issues](https://github.com/Kunde21/forgejo-mcp/issues) for similar problems
2. Enable debug logging and include relevant log output in your issue report
3. Provide details about your environment:
   - Forgejo version
   - Go version
   - Operating system
   - Configuration settings (without sensitive information)

## Performance Optimization

### Caching Strategy

The Gitea SDK client includes built-in caching:

```go
// Create a cache with 1000 items and 5 minute TTL
cache, err := tea.NewCache(1000, 5*time.Minute)
```

### Batch Processing

For handling multiple requests efficiently:

```go
// Process multiple requests concurrently
processor := tea.NewBatchProcessor(10)
responses, err := processor.ProcessBatch(context.Background(), requests)
```

### Connection Pooling

The Gitea SDK client uses HTTP connection pooling by default for better performance.