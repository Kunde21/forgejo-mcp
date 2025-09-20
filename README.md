# Forgejo MCP

Model Context Protocol server for interacting with Forgejo repositories.

## Description

This server provides MCP (Model Context Protocol) access to Forgejo and Gitea repository features using the **official Model Context Protocol SDK** (`github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`). It enables AI agents to interact with remote repositories for common development tasks like managing pull requests and issues with direct API integration for improved performance and reliability.

**Migration Note**: This project has been updated to use the official MCP SDK instead of the third-party `mark3labs/mcp-go` library for better protocol compliance, long-term stability, and official support.

## Prerequisites

- Go 1.24.6 or later
- Access to a Forgejo/Gitea instance
- Authentication token for Forgejo/Gitea API access
- Official MCP SDK (`github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`)

## Migration from Previous Versions

If you're upgrading from a version using the third-party `mark3labs/mcp-go` SDK:

### What Changed
- **SDK**: Migrated from `mark3labs/mcp-go` to official `github.com/modelcontextprotocol/go-sdk/mcp v0.4.0`
- **Protocol Compliance**: Improved adherence to MCP protocol specifications
- **API Stability**: Official SDK provides long-term stability and support
- **Tool Registration**: Updated tool registration methods (`mcp.AddTool()`)
- **Handler Signatures**: New handler function signatures for better type safety

### Upgrade Steps
1. **Update Dependencies**: Run `go mod tidy` to fetch the new SDK
2. **Rebuild**: Clean rebuild with `go build ./...`
3. **Test**: Run your test suite to ensure compatibility
4. **Configuration**: No configuration changes required - fully backward compatible

### Breaking Changes
- None for end users - all existing functionality preserved
- Internal API changes only affect custom integrations

## Migration from Gitea to Forgejo

If you're migrating from a Gitea instance to Forgejo, the server provides seamless support through automatic detection:

### Automatic Migration (Recommended)

1. **Update Environment Variables**: Change your `FORGEJO_REMOTE_URL` to point to your new Forgejo instance
2. **No Configuration Changes**: The server will automatically detect Forgejo and use the appropriate SDK
3. **Test Connection**: Use the `config` command to verify connectivity:

```bash
./forgejo-mcp config
```

### Manual Migration

If you prefer explicit control:

1. **Update Environment Variables**:
   ```bash
   export FORGEJO_REMOTE_URL="https://your-new-forgejo.instance"
   export FORGEJO_CLIENT_TYPE="forgejo"  # Explicitly use Forgejo SDK
   ```

2. **Update Configuration File** (if using config.yaml):
   ```yaml
   remote_url: "https://your-new-forgejo.instance"
   client_type: "forgejo"
   ```

### Migration Benefits

- **Zero Downtime**: Automatic detection ensures compatibility during transition
- **Feature Parity**: All existing tools work identically with both platforms
- **Performance**: Forgejo SDK may provide better performance for Forgejo instances
- **Future-Proof**: Ready for any future platform-specific optimizations

### Verification Steps

After migration, verify everything works:

1. **Test Tool Calls**: Run any existing MCP tool calls - they should work unchanged
2. **Check Logs**: Look for "Detected Forgejo" in server logs during startup
3. **Validate Responses**: Ensure all API responses maintain the same structure

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
- `FORGEJO_CLIENT_TYPE` - Client type: "gitea", "forgejo", or "auto" (default: "auto")
- `MCP_HOST` - Host for MCP server (default: localhost)
- `MCP_PORT` - Port for MCP server (default: 3000)

### Configuration File

The server uses environment variables for configuration. No configuration file is currently supported.

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

**New Feature**: All tools now support an optional `directory` parameter that automatically resolves to the repository information. This provides a more intuitive interface for working with local git repositories.

When you provide a `directory` parameter pointing to a local git repository, the server will:
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

### Migration Guide: From Repository to Directory Parameter

#### Why Migrate?

The `directory` parameter provides several advantages over manually specifying `repository`:

1. **Automatic Resolution**: No need to remember or look up repository names
2. **File System Integration**: Work directly with local project directories
3. **Reduced Errors**: Eliminates typos in repository name specification
4. **Git Integration**: Leverages existing git remote configuration

#### Migration Steps

1. **Identify Current Usage**: Find all places where you're using the `repository` parameter
2. **Locate Local Repository**: Ensure you have a local clone of the repository
3. **Replace Parameter**: Change `repository` to `directory` and provide the local path
4. **Test**: Verify the tool still works as expected

#### Before and After Examples

**Before (Repository Parameter):**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_list",
    "arguments": {
      "repository": "myorg/myproject",
      "limit": 10
    }
  }
}
```

**After (Directory Parameter):**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_list",
    "arguments": {
      "directory": "/home/user/projects/myproject",
      "limit": 10
    }
  }
}
```

#### Backward Compatibility

The `repository` parameter continues to work exactly as before. You can migrate gradually:

- **Phase 1**: Continue using `repository` parameter
- **Phase 2**: Start using `directory` parameter for new operations
- **Phase 3**: Gradually migrate existing operations
- **Phase 4**: Remove `repository` parameter usage (optional)

#### Error Handling

When using `directory` parameter, be aware of these validation rules:

- Directory must exist
- Directory must contain a `.git` folder
- Git repository must have at least one remote configured
- Remote URL must be parseable to extract owner/repo information

If validation fails, you'll receive a clear error message indicating the specific issue.

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

**Using repository parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_list",
    "arguments": {
      "repository": "myorg/myrepo",
      "limit": 10,
      "offset": 0
    }
  }
}
```

**Using directory parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_list",
    "arguments": {
      "directory": "/path/to/your/local/repo",
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

#### Create Issue Comment

**Using repository parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_comment_create",
    "arguments": {
      "repository": "myorg/myrepo",
      "issue_number": 42,
      "comment": "This is a helpful comment on the issue."
    }
  }
}
```

**Using directory parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_comment_create",
    "arguments": {
      "directory": "/path/to/your/local/repo",
      "issue_number": 42,
      "comment": "This is a helpful comment on the issue."
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
         "text": "Comment created successfully. ID: 123, Created: 2025-09-09T10:30:00Z\nComment body: This is a helpful comment on the issue."
       }
     ],
     "structured": {
       "comment": {
         "id": 123,
         "content": "This is a helpful comment on the issue.",
         "author": "testuser",
         "created": "2025-09-09T10:30:00Z"
       }
     }
   }
 }
 ```

 #### List Issue Comments

 ```json
 {
   "method": "tools/call",
   "params": {
     "name": "issue_comment_list",
     "arguments": {
       "repository": "myorg/myrepo",
       "issue_number": 42,
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
         "text": "Found 3 comments (showing 1-3)"
       }
     ],
     "structured": {
       "comments": [
         {
           "id": 1,
           "content": "First comment on this issue",
           "author": "developer1",
           "created": "2025-09-10T09:15:00Z"
         },
         {
           "id": 2,
           "content": "Thanks for the update",
           "author": "manager",
           "created": "2025-09-10T10:30:00Z"
         },
         {
           "id": 3,
           "content": "I've implemented the requested changes",
           "author": "developer1",
           "created": "2025-09-10T14:45:00Z"
         }
       ],
       "total": 3,
       "limit": 10,
       "offset": 0
     }
   }
  }
  ```

#### Edit Issue Comment

```json
{
  "method": "tools/call",
  "params": {
    "name": "issue_comment_edit",
    "arguments": {
      "repository": "myorg/myrepo",
      "issue_number": 42,
      "comment_id": 123,
      "new_content": "Updated comment with additional information and corrections."
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
        "text": "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:30:00Z\nComment body: Updated comment with additional information and corrections."
      }
    ],
    "structured": {
      "comment": {
        "id": 123,
        "content": "Updated comment with additional information and corrections.",
        "author": "testuser",
        "created": "2025-09-10T10:30:00Z"
      }
    }
  }
}
```

#### List Pull Requests

**Using repository parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_list",
    "arguments": {
      "repository": "myorg/myrepo",
      "limit": 10,
      "offset": 0,
      "state": "open"
    }
  }
}
```

**Using directory parameter:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_list",
    "arguments": {
      "directory": "/path/to/your/local/repo",
      "limit": 10,
      "offset": 0,
      "state": "open"
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
        "text": "Found 2 pull requests"
      }
    ],
    "structured": {
      "pull_requests": [
        {
          "id": 1,
          "number": 1,
          "title": "Feature: Add dark mode",
          "state": "open",
          "user": {
            "login": "developer1"
          },
          "created_at": "2025-09-10T09:15:00Z",
          "updated_at": "2025-09-10T09:15:00Z",
          "head": {
            "ref": "feature/dark-mode",
            "sha": "abc123def456"
          },
          "base": {
            "ref": "main",
            "sha": "def456abc123"
          }
        },
        {
          "id": 2,
          "number": 2,
          "title": "Fix: Memory leak",
          "state": "open",
          "user": {
            "login": "developer2"
          },
          "created_at": "2025-09-10T10:30:00Z",
          "updated_at": "2025-09-10T10:30:00Z",
          "head": {
            "ref": "fix/memory-leak",
            "sha": "ghi789jkl012"
          },
          "base": {
            "ref": "main",
            "sha": "def456abc123"
          }
        }
      ]
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
        "text": "Comment edited successfully. ID: 123, Updated: 2025-09-10T10:00:00Z\nComment body: Updated comment with additional information and corrections."
      }
    ],
    "structured": {
      "comment": {
        "id": 123,
        "content": "Updated comment with additional information and corrections.",
        "author": "testuser",
        "created": "2025-09-10T10:00:00Z"
      }
    }
  }
}
```

#### Create Pull Request Comment

```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_comment_create",
    "arguments": {
      "repository": "myorg/myrepo",
      "pull_request_number": 42,
      "comment": "This is a helpful comment on the pull request."
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
        "text": "Pull request comment created successfully. ID: 123, Created: 2025-09-12T14:30:45Z\nComment body: This is a helpful comment on the pull request."
      }
    ],
    "structured": {
      "comment": {
        "id": 123,
        "body": "This is a helpful comment on the pull request.",
        "user": "reviewer",
        "created_at": "2025-09-12T14:30:45Z",
        "updated_at": "2025-09-12T14:30:45Z"
      }
    }
  }
}
```

#### List Pull Request Comments

```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_comment_list",
    "arguments": {
      "repository": "myorg/myrepo",
      "pull_request_number": 42,
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
        "text": "Found 2 pull request comments"
      }
    ],
    "structured": {
      "pull_request_comments": [
        {
          "id": 1,
          "body": "This is a great PR!",
          "user": "reviewer1",
          "created_at": "2025-09-12T10:30:00Z",
          "updated_at": "2025-09-12T10:30:00Z"
        },
        {
          "id": 2,
          "body": "I agree, well done!",
          "user": "reviewer2",
          "created_at": "2025-09-12T11:15:00Z",
          "updated_at": "2025-09-12T11:15:00Z"
        }
      ]
    }
  }
}
```

#### Edit Pull Request Comment

```json
{
  "method": "tools/call",
  "params": {
    "name": "pr_comment_edit",
    "arguments": {
      "repository": "myorg/myrepo",
      "pull_request_number": 42,
      "comment_id": 123,
      "new_content": "Updated comment with additional information and corrections."
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
        "text": "Pull request comment edited successfully. ID: 123, Updated: 2025-09-12T14:30:45Z\nComment body: Updated comment with additional information and corrections."
      }
    ],
    "structured": {
      "comment": {
        "id": 123,
        "body": "Updated comment with additional information and corrections.",
        "user": "reviewer",
        "created_at": "2025-09-12T14:30:45Z",
        "updated_at": "2025-09-12T14:30:45Z"
      }
    }
  }
}
```

## Features

### SDK Integration

This server uses the **official Model Context Protocol SDK** for standardized MCP protocol implementation, combined with the official Gitea SDK for direct API integration with Gitea/Forgejo instances. This provides:

- **Official Protocol Support**: Full compliance with MCP specifications
- **Improved Performance**: Direct API integration with comprehensive error handling
- **Enhanced Reliability**: Official SDK with long-term support guarantees
- **Better Tool Management**: Standardized tool registration and lifecycle management

Authentication is handled through API tokens configured at startup. All operations are performed directly via the Gitea/Forgejo REST API using the official SDKs.

### PR interactions

Manage Pull Requests opened on your gitea repository

Tools List:
- `pr_list`: List pull requests from a repository with pagination and state filtering
  - Parameters: repository (owner/repo) OR directory (local path), limit (1-100, default 15), offset (0-based, default 0), state (open/closed/all, default "open")
  - Returns: Array of pull requests with ID, number, title, state, user, timestamps, and branch information
- `pr_comment_create`: Create a comment on a repository pull request
  - Parameters: repository (owner/repo), pull_request_number (positive integer), comment (non-empty string)
  - Returns: Comment creation confirmation with metadata
- `pr_comment_list`: List comments from a repository pull request with pagination support
  - Parameters: repository (owner/repo), pull_request_number (positive integer), limit (1-100, default 15), offset (0-based, default 0)
  - Returns: Array of comments with ID, content, author, and creation timestamp
- `pr_comment_edit`: Edit an existing comment on a repository pull request
  - Parameters: repository (owner/repo), pull_request_number (positive integer), comment_id (positive integer), new_content (non-empty string)
  - Returns: Comment edit confirmation with updated metadata
- Review PR: approve or request changes

### Issue interactions

Manage issues in your forgejo repository

 Tools List:
 - `issue_list`: List issues from a repository with pagination support
   - Parameters: repository (owner/repo) OR directory (local path), limit (1-100), offset (0-based)
   - Returns: Array of issues with number, title, and status
 - `issue_comment_create`: Create a comment on a repository issue
   - Parameters: repository (owner/repo), issue_number (positive integer), comment (non-empty string)
   - Returns: Comment creation confirmation with metadata
  - `issue_comment_list`: List comments from a repository issue with pagination support
    - Parameters: repository (owner/repo), issue_number (positive integer), limit (1-100, default 15), offset (0-based, default 0)
    - Returns: Array of comments with ID, content, author, and creation timestamp
  - `issue_comment_edit`: Edit an existing comment on a repository issue
    - Parameters: repository (owner/repo), issue_number (positive integer), comment_id (positive integer), new_content (non-empty string)
    - Returns: Comment edit confirmation with updated metadata
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
