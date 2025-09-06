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

- `FORGEJO_MCP_REMOTE_URL` - URL of your Forgejo instance
- `FORGEJO_MCP_AUTH_TOKEN` - Authentication token for Forgejo API
- `FORGEJO_MCP_DEBUG` - Enable debug logging (default: false)
- `FORGEJO_MCP_LOG_LEVEL` - Log level (default: "info")

### Configuration File

Create a `config.yaml` file in the current directory or in `~/.forgejo-mcp/config.yaml`:

```yaml
forgejo_url: "https://your.forgejo.instance"
auth_token: "your-auth-token"
debug: false
log_level: "info"
```

## Usage

All calls are expected to come from a git repository with a forgejo server remote.

Start the MCP server:

```bash
forgejo-mcp serve
```

For more options:

```bash
forgejo-mcp --help
forgejo-mcp serve --help
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
