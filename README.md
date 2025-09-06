# Forgejo MCP

Model Context Protocol server for interacting with Forgejo repositories.

## Description

This server provides MCP (Model Context Protocol) access to Forgejo repository features using the official Gitea SDK. It enables AI agents to interact with Forgejo repositories for common development tasks like managing pull requests and issues with direct API integration for improved performance and reliability.

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
go build -o bin/forgejo-mcp .
```

## Configuration

The application can be configured through environment variables or a configuration file.

### Environment Variables

- `FORGEJO_MCP_FORGEJO_URL` - URL of your Forgejo instance
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

The MCP server now supports repository-specific queries for improved performance and targeted results.

Start the MCP server:

```bash
forgejo-mcp serve
```

For more options:

```bash
forgejo-mcp --help
forgejo-mcp serve --help
```

### Repository-Based Queries

All tools now require a repository parameter to specify which repository to query:

```javascript
// List PRs for a specific repository
const prs = await mcp.callTool('pr_list', {
  repository: 'my-org/my-project'
});

// Alternative: Use current working directory
const prs = await mcp.callTool('pr_list', {
  cwd: '/path/to/repository'
});
```

### CLI Commands

The forgejo-mcp CLI provides the following commands:

- `forgejo-mcp` - Root command with global flags
- `forgejo-mcp serve` - Start the MCP server (aliases: server, start)
- `forgejo-mcp completion` - Generate autocompletion script
- `forgejo-mcp help` - Get help for any command

### Global Flags

- `--config string` - Path to configuration file (default is ./config.yaml)
- `--debug` - Enable debug mode
- `--log-level string` - Set log level (trace, debug, info, warn, error, fatal, panic) (default "info")

### Serve Command Flags

- `--host string` - Host to bind to (default "localhost")
- `--port int` - Port to listen on (default 3000)

## Features

### SDK Integration

This server uses the official Gitea SDK for direct API integration with Forgejo instances, providing improved performance, reliability, and comprehensive error handling compared to CLI-based approaches.

Authentication is handled through API tokens configured at startup. All operations are performed directly via the Forgejo REST API.

### PR interactions

Manage Pull Requests for specific repositories

Tools List:
- PR list: show all open PRs for a specified repository
- PR Comment: Add a comment to a given PR
- Review PR: approve or request changes

### Issue interactions

Manage issues in specific repositories

Tools List:
- List Issues: show all open issues for a specified repository
- Open Issue: create a new issue with details about a feature request or a bug
- Close Issue: close an issue that has been answered or completed
- Issue Comment: Add a comment to a given issue

## Development

### Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo instance for testing

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Linting

```bash
go vet ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.