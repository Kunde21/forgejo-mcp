# Forgejo MCP API Documentation

This document describes the API endpoints and tools available through the Forgejo MCP server.

## Overview

The Forgejo MCP server provides a Model Context Protocol interface that allows AI agents to interact with Forgejo repositories through standardized CLI tools. All interactions are performed within the context of a git repository that has a Forgejo remote configured.

## Authentication

Authentication to the server must happen outside of the MCP calls, before the agent connects. The server expects a valid Forgejo authentication token to be configured either through environment variables or a configuration file.

## Available Tools

### PR List Tool

**Name:** `pr_list`

**Description:** Lists all open pull requests on the current repository.

**Parameters:** None

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
      "updated_at": "2025-08-26T10:30:00Z"
    }
  ]
}
```

### Issue List Tool

**Name:** `issue_list`

**Description:** Lists all open issues on the current repository.

**Parameters:** None

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
      "created_at": "2025-08-26T09:15:00Z"
    }
  ]
}
```

## Error Handling

All tools will return appropriate error messages when operations fail. Common error scenarios include:

- Invalid repository context
- Authentication failures
- Network connectivity issues
- Invalid parameters

Error responses follow the MCP standard error format.