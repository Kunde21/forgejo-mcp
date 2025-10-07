# Forgejo MCP

Model Context Protocol server for interacting with Forgejo repositories.

## Description

This server provides MCP (Model Context Protocol) access to Forgejo and Gitea repository features using the **official Model Context Protocol SDK** (`github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`). It enables AI agents to interact with remote repositories for common development tasks like managing pull requests and issues with direct API integration for improved performance and reliability.

## Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo/Gitea instance
- Authentication token for Forgejo/Gitea API access
- Official MCP SDK (`github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`)

## Installation

```bash
go install github.com/Kunde21/forgejo-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/Kunde21/forgejo-mcp.git
cd forgejo-mcp
go build -o bin/forgejo-mcp cmd/main.go
```

## Configuration

The application can be configured through environment variables.

### Environment Variables

- `FORGEJO_REMOTE_URL` - URL of your Forgejo/Gitea instance (required)
- `FORGEJO_AUTH_TOKEN` - Authentication token for Forgejo/Gitea API (required)
- `FORGEJO_CLIENT_TYPE` - Client type: "gitea", "forgejo", or "auto" (default: "auto")

### Configuration for OpenCode

To use this MCP server with OpenCode, add the following to your `opencode.json` configuration file:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "forgejo": {
      "type": "local",
      "command": ["/path/to/forgejo-mcp", "serve"],
      "enabled": true,
      "environment": {
        "FORGEJO_REMOTE_URL": "https://your.forgejo.instance",
        "FORGEJO_AUTH_TOKEN": "your-api-token",
        "FORGEJO_CLIENT_TYPE": "auto"
      }
    }
  }
}
```

Replace `/path/to/forgejo-mcp` with the actual path to your built binary, and update the environment variables with your Forgejo/Gitea instance details.

The server will be automatically available to OpenCode once configured. You can temporarily disable it by setting `"enabled": false`.

### Client Type Configuration

The server supports three client types for connecting to different Git hosting platforms:

- **`gitea`** - Use the Gitea SDK for Gitea instances
- **`forgejo`** - Use the Forgejo SDK for Forgejo instances
- **`auto`** - Automatically detect the platform by querying `/api/v1/version` (default)

#### Automatic Detection

When using `auto` mode (the default), the server will:

1. Query the `/api/v1/version` endpoint of your instance
2. Parse the version string to determine if it's Forgejo or Gitea
3. Select the appropriate SDK automatically
4. Fall back to Gitea if detection fails

This provides seamless compatibility with both platforms without manual configuration.

#### Manual Configuration

For specific requirements, you can explicitly set the client type:

```bash
# Force Gitea client
export FORGEJO_CLIENT_TYPE="gitea"

# Force Forgejo client
export FORGEJO_CLIENT_TYPE="forgejo"

# Use automatic detection (default)
export FORGEJO_CLIENT_TYPE="auto"
```

## Usage

Set the required environment variables:

```bash
export FORGEJO_REMOTE_URL="https://your.forgejo.instance"
export FORGEJO_AUTH_TOKEN="your-api-token"
```

### Running the Server

Start the MCP server using the CLI:

```bash
# Build the binary
go build -o forgejo-mcp .

# Start the server
./forgejo-mcp serve
```

Or run directly with Go:

```bash
go run main.go serve
```

For backward compatibility, you can also run:

```bash
go run main.go
```

The server will start and listen for MCP protocol messages on stdin/stdout.

### Directory Parameter Support

All tools support an optional `directory` parameter that automatically resolves to repository information from local git repositories. When you provide a `directory` parameter, the server will:

1. Validate the directory exists and contains a `.git` folder
2. Extract the remote repository URL from `.git/config`
3. Parse the owner/repository information
4. Use this information for API calls

**Benefits:**
- Work directly with file system paths
- No need to manually specify repository names
- Automatic repository detection from git configuration
- Backward compatible with existing `repository` parameter

**Usage Pattern:**
```bash
# Instead of manually specifying repository
"repository": "myorg/myrepo"

# You can now use directory path
"directory": "/path/to/your/local/repo"
```

**Note**: The `directory` parameter takes precedence if both `repository` and `directory` are provided.

### CLI Commands

The application provides several CLI commands:

- `serve`: Start the MCP server (default command)
- `version`: Show version information
- `config`: Validate configuration and test connectivity

Example usage:

```bash
# Show version
./forgejo-mcp version

# Validate configuration
./forgejo-mcp config

# Start server with custom config
./forgejo-mcp serve --config /path/to/config.yaml

# Enable verbose logging
./forgejo-mcp serve --verbose
```

### Usage Examples

#### Basic Issue Management

**List open issues using directory parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_list",
    "arguments": {
      "directory": "/home/user/projects/myapp",
      "limit": 10,
      "offset": 0
    }
  }
}
```

**Create a comment on an issue:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_comment_create",
    "arguments": {
      "directory": "/home/user/projects/myapp",
      "issue_number": 42,
      "comment": "I've investigated this issue and it appears to be related to the database connection timeout. I'll submit a fix shortly."
    }
  }
}
```

**Edit an existing issue comment:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_comment_edit",
    "arguments": {
      "repository": "myorg/myrepo",
      "issue_number": 42,
      "comment_id": 123,
      "new_content": "Update: The fix has been implemented in PR #45. Please review and merge."
    }
  }
}
```

#### Pull Request Workflow

**List open pull requests:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_list",
    "arguments": {
      "directory": "/home/user/projects/myapp",
      "state": "open",
      "limit": 5
    }
  }
}
```

**Add review comment to a PR:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_comment_create",
    "arguments": {
      "repository": "myorg/myrepo",
      "pull_request_number": 23,
      "comment": "Great work on this feature! I have a few suggestions:\n1. Consider adding error handling for edge cases\n2. The tests look comprehensive\n3. Documentation is clear and well-written"
    }
  }
}
```

**Edit a pull request (update title and close):**
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_edit",
    "arguments": {
      "directory": "/home/user/projects/myapp",
      "pull_request_number": 23,
      "title": "feat: Add user authentication with JWT tokens",
      "state": "closed"
    }
  }
}
```

#### Real-world Scenarios

**Scenario 1: Code Review Workflow**
```json
// 1. List open PRs needing review
{
  "name": "pr_list",
  "arguments": {
    "directory": "/home/user/projects/myapp",
    "state": "open"
  }
}

// 2. Add constructive feedback
{
  "name": "pr_comment_create",
  "arguments": {
    "directory": "/home/user/projects/myapp",
    "pull_request_number": 15,
    "comment": "The implementation looks solid. I noticed one potential optimization in the database query - consider adding an index on the user_id column to improve performance."
  }
}
```

**Scenario 2: Issue Triage**
```json
// 1. Check recent issues
{
  "name": "issue_list",
  "arguments": {
    "repository": "myorg/myrepo",
    "limit": 20
  }
}

// 2. Request more information on a bug report
{
  "name": "issue_comment_create",
  "arguments": {
    "repository": "myorg/myrepo",
    "issue_number": 67,
    "comment": "Thank you for reporting this issue. Could you please provide:\n- Steps to reproduce\n- Expected vs actual behavior\n- Browser/OS version\n- Any error messages from the console"
  }
}
```

**Scenario 3: Project Management**
```json
// 1. Get overview of all open issues and PRs
{
  "name": "issue_list",
  "arguments": {
    "directory": "/home/user/projects/myapp",
    "limit": 50
  }
}

// 2. Check PR status
{
  "name": "pr_list",
  "arguments": {
    "directory": "/home/user/projects/myapp",
    "state": "open"
  }
}

// 3. Update PR with merge information
{
  "name": "pr_edit",
  "arguments": {
    "directory": "/home/user/projects/myapp",
    "pull_request_number": 12,
    "body": "This PR has been tested and is ready for merge. All tests pass and code review is complete."
  }
}
```

#### Response Format

All tools return responses in the MCP standard format with both human-readable text and structured data:

```json
{
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 3 issues"
      }
    ],
    "structured": [
      {
        "number": 1,
        "title": "Bug: Login fails",
        "state": "open",
        "user": {
          "login": "reporter"
        },
        "created_at": "2025-10-01T10:30:00Z"
      }
    ]
  }
}
```

The `content` field provides human-readable summaries, while `structured` contains the full data for programmatic use.

## Features

### SDK Integration

This server uses the **official Model Context Protocol SDK** for standardized MCP protocol implementation, combined with platform-specific SDKs for direct API integration with Gitea/Forgejo instances. This provides:

- **Official Protocol Support**: Full compliance with MCP specifications
- **Improved Performance**: Direct API integration with comprehensive error handling
- **Enhanced Reliability**: Official SDK with long-term support guarantees
- **Better Tool Management**: Standardized tool registration and lifecycle management
- **Platform Detection**: Automatic detection and optimal SDK selection for Gitea vs Forgejo

Authentication is handled through API tokens configured at startup. All operations are performed directly via the Gitea/Forgejo REST API using the official SDKs.

### Available Tools

#### Issue Management
- **`issue_list`**: List issues from a repository with pagination support
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `limit` (1-100, default 15), `offset` (0-based, default 0)
  - Returns: Array of issues with number, title, state, and metadata

- **`issue_comment_create`**: Create a comment on a repository issue
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `issue_number` (positive integer), `comment` (non-empty string)
  - Returns: Comment creation confirmation with metadata

- **`issue_comment_list`**: List comments from a repository issue with pagination support
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `issue_number` (positive integer), `limit` (1-100, default 15), `offset` (0-based, default 0)
  - Returns: Array of comments with ID, content, author, and creation timestamp

- **`issue_comment_edit`**: Edit an existing comment on a repository issue
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `issue_number` (positive integer), `comment_id` (positive integer), `new_content` (non-empty string)
  - Returns: Comment edit confirmation with updated metadata

#### Pull Request Management
- **`pr_list`**: List pull requests from a repository with pagination and state filtering
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `limit` (1-100, default 15), `offset` (0-based, default 0), `state` (open/closed/all, default "open")
  - Returns: Array of pull requests with ID, number, title, state, user, timestamps, and branch information

- **`pr_comment_create`**: Create a comment on a repository pull request
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `pull_request_number` (positive integer), `comment` (non-empty string)
  - Returns: Comment creation confirmation with metadata

- **`pr_comment_list`**: List comments from a repository pull request with pagination support
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `pull_request_number` (positive integer), `limit` (1-100, default 15), `offset` (0-based, default 0)
  - Returns: Array of comments with ID, content, author, and creation timestamp

- **`pr_comment_edit`**: Edit an existing comment on a repository pull request
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `pull_request_number` (positive integer), `comment_id` (positive integer), `new_content` (non-empty string)
  - Returns: Comment edit confirmation with updated metadata

- **`pr_edit`**: Edit an existing pull request
  - Parameters: `repository` (owner/repo) OR `directory` (local path), `pull_request_number` (positive integer), optional: `title` (string), `body` (string), `state` (open/closed), `base_branch` (string)
  - Returns: Pull request edit confirmation with updated metadata

#### Repository Utilities
- **`forgejo_hello`**: Simple hello world tool for testing connectivity
  - Parameters: none
  - Returns: Hello message confirming server is working

### Platform Support

The server automatically detects and optimizes for different platforms:

- **Forgejo Instances**: Uses Forgejo-specific SDK for optimal performance
- **Gitea Instances**: Uses Gitea SDK with full feature compatibility
- **Automatic Detection**: Queries `/api/v1/version` endpoint to determine platform
- **Manual Override**: Force specific client type via `FORGEJO_CLIENT_TYPE` environment variable

## Development

### Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo or Gitea instance for testing

### Building

```bash
go build ./...
```

### Testing

The project includes comprehensive test coverage with both unit and integration tests. The test suite uses a mock Gitea server built with `httptest.Server` for reliable, fast testing without external dependencies.

#### Mock Server Architecture

The test harness provides:
- **Individual Handler Functions**: Each API endpoint has a dedicated handler function for better testability and maintainability
- **Modern Routing**: Uses Go 1.22+ `http.ServeMux` with method + path patterns for precise routing
- **Helper Functions**: Reusable utilities for path parameter extraction, validation, pagination, and authentication
- **Extensible Design**: New endpoints can be easily added by registering additional handlers

#### Running Tests

Run the complete test suite:

```bash
go test ./...
```

Run specific test files:

```bash
go test -run TestName ./...
```

Run integration tests:

```bash
go test -run Integration ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

#### Test Structure

- **Unit Tests**: Test individual functions and handlers in isolation
- **Integration Tests**: Test end-to-end functionality with the mock server
- **Acceptance Tests**: Test complete workflows and user scenarios
- **Performance Tests**: Validate response times and resource usage

### Linting

```bash
go vet ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
