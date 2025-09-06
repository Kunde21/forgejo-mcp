# Gitea SDK Refactor - Lite Summary

Separate MCP handlers from Gitea SDK implementation by moving Gitea-specific code to a dedicated `remote/gitea` package while keeping MCP handlers and input validation in the `server` package. This refactor addresses current structural issues where server/sdk_handlers.go contains both MCP handler logic and Gitea SDK implementation details, creating tight coupling and reduced maintainability.

## Key Points
- Separate MCP handlers from Gitea SDK implementation
- Create dedicated `remote/gitea` package for Gitea-specific code  
- Improve maintainability by reducing tight coupling