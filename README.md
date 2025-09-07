# Forgejo MCP

Model Context Protocol server for interacting with Forgejo repositories.

## Description

This server provides MCP (Model Context Protocol) access to Forgejo and Gitea repository features using the official SDKs. It enables AI agents to interact with remote repositories for common development tasks like managing pull requests and issues with direct API integration for improved performance and reliability.

## Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo instance
- Authentication token for Forgejo API access

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

The application can be configured through environment variables or a configuration file.

### Environment Variables

- `FORGEJO_REMOTE_URL` - URL of your Forgejo/Gitea instance (required)
- `FORGEJO_AUTH_TOKEN` - Authentication token for Forgejo/Gitea API (required)
- `MCP_HOST` - Host for MCP server (default: localhost)
- `MCP_PORT` - Port for MCP server (default: 3000)

### Configuration File

The server uses environment variables for configuration. No configuration file is currently supported.

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

### Tool Usage Examples

#### List Issues

```json
{
  "method": "tools/call",
  "params": {
    "name": "list_issues",
    "arguments": {
      "repository": "myorg/myrepo",
      "limit": 10,
      "offset": 0
    }
  }
}
```

Response:
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
        "state": "open"
      },
      {
        "number": 2,
        "title": "Feature: Add dark mode",
        "state": "open"
      },
      {
        "number": 3,
        "title": "Fix: Memory leak",
        "state": "closed"
      }
    ]
  }
}
```

## Features

### SDK Integration

This server uses the official Gitea SDK for direct API integration with Gitea instances, providing improved performance, reliability, and comprehensive error handling compared to CLI-based approaches.

Authentication is handled through API tokens configured at startup. All operations are performed directly via the Gitea REST API.

### PR interactions

Manage Pull Requests opened on your gitea repository

Tools List:
- PR list: show all open PRs on the current repository
- PR Comment: Add a comment to a given PR
- Review PR: approve or request changes

### Issue interactions

Manage issues in your forgejo repository

Tools List:
- `list_issues`: List issues from a repository with pagination support
  - Parameters: repository (owner/repo), limit (1-100), offset (0-based)
  - Returns: Array of issues with number, title, and status
- List Issues: show all open issues on the current repository
- Open Issue: create a new issue with details about a feature request or a bug
- Close Issue: close an issue that has been answered or completed
- Issue Comment: Add a comment to a given issue

## Development

### Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo or Gitea instance for testing

### Building

```bash
go build ./...
```

### Testing

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

### Linting

```bash
go vet ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
