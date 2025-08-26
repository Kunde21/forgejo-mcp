# Forgejo MCP

Model Context Protocol server for interacting with Forgejo repositories.

## Description

This server wraps the functionality of the `tea` cli tool, originally from the `gitea` project, to provide MCP (Model Context Protocol) access to Forgejo repository features. It enables AI agents to interact with Forgejo repositories for common development tasks like managing pull requests and issues.

## Prerequisites

- Go 1.24.6 or later
- `tea` CLI tool installed and configured
- Access to a Forgejo instance

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

- `FORGEJO_MCP_FORGEJO_URL` - URL of your Forgejo instance
- `FORGEJO_MCP_AUTH_TOKEN` - Authentication token for Forgejo API
- `FORGEJO_MCP_TEA_PATH` - Path to the tea CLI executable (default: "tea")
- `FORGEJO_MCP_DEBUG` - Enable debug logging (default: false)
- `FORGEJO_MCP_LOG_LEVEL` - Log level (default: "info")

### Configuration File

Create a `config.yaml` file in the current directory or in `~/.forgejo-mcp/config.yaml`:

```yaml
forgejo_url: "https://your.forgejo.instance"
auth_token: "your-auth-token"
tea_path: "/path/to/tea"
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

### CLI interactions

This server wraps the functionality of the `tea` cli tool, originally from the `gitea` project.
Authentication to the server must happen outside of the mcp calls, before the agent connects.

All calls are expected to come from a git repository with a forgejo server remote.

### PR interactions

Manage Pull Requests opened on your forgejo repository

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
- `tea` CLI tool for testing

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