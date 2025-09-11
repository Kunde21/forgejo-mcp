# pr-list - Lite Summary

Add a new MCP tool "pr_list" to list pull requests in Forgejo/Gitea repositories, following the same patterns as the existing issue_list functionality. The tool will support filtering by state (open/closed/all) and include pagination with limit/offset parameters.

## Key Points
- New MCP tool "pr_list" for listing pull requests in repositories
- Parameters: repository (required), limit (optional, default 15), offset (optional, default 0), state (optional, default "open")
- Follows existing architectural patterns with interface, client, service, and handler layers
- Uses ozzo-validation for input validation and proper error handling
- Integrates with existing MCP server registration and Forgejo client patterns
- Includes comprehensive unit and integration tests