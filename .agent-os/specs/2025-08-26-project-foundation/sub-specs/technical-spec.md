# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-08-26-project-foundation/spec.md

## Technical Requirements

### Go Module Initialization
- Execute `go mod init github.com/Kunde21/forgejo-mcp` to create the module
- Set Go version requirement to 1.21 or higher in go.mod
- Configure module proxy settings if behind corporate firewall

### Dependency Installation
- Install Cobra CLI framework: `go get github.com/spf13/cobra@v1.8.0`
- Install MCP Go SDK: `go get github.com/modelcontextprotocol/go-sdk/mcp@latest`
- Install Viper for configuration: `go get github.com/spf13/viper@v1.18.0`
- Install Logrus for logging: `go get github.com/sirupsen/logrus@v1.9.3`
- Run `go mod tidy` to clean up and verify dependencies
- Create go.sum file for dependency verification

### Directory Structure Implementation
```
forgejo-mcp/
├── cmd/              # CLI commands
│   ├── main.go      # Application entry point
│   ├── root.go      # Root command definition
│   └── serve.go     # Server command
├── server/          # MCP server implementation
│   ├── server.go    # Server struct and initialization
│   ├── tools.go     # Tool registration
│   ├── transport.go # Transport layer
│   └── handlers.go  # Tool handlers
├── tea/             # Tea CLI wrapper
│   ├── wrapper.go   # Wrapper interface
│   ├── tea.go       # Tea CLI implementation
│   ├── commands.go  # Command builders
│   └── parser.go    # Output parsers
├── context/         # Repository context
│   ├── git.go       # Git detection
│   ├── forgejo.go   # Forgejo validation
│   └── context.go   # Context management
├── auth/            # Authentication
│   ├── auth.go      # Auth interface
│   ├── provider.go  # Token providers
│   └── tea.go       # Tea authentication
├── config/          # Configuration
│   └── config.go    # Config struct and loading
├── types/           # Data types
│   ├── pr.go        # Pull request types
│   ├── issue.go     # Issue types
│   └── responses.go # Response types
└── test/            # Test files
    ├── integration/ # Integration tests
    └── e2e/         # End-to-end tests
```

### Configuration System
- Create `config/config.go` with the following struct:
```go
type Config struct {
    ForgejoURL   string `mapstructure:"forgejo_url"`
    AuthToken    string `mapstructure:"auth_token"`
    TeaPath      string `mapstructure:"tea_path"`
    Debug        bool   `mapstructure:"debug"`
    LogLevel     string `mapstructure:"log_level"`
}
```
- Implement `Load() (*Config, error)` function that:
  - Reads from environment variables (prefix: FORGEJO_MCP_)
  - Supports config file in YAML/JSON format
  - Validates required fields (ForgejoURL, AuthToken)
  - Sets sensible defaults (TeaPath: "tea", LogLevel: "info")
- Implement `Validate() error` method for config validation
- Support config file locations:
  - Current directory: `.forgejo-mcp.yaml`
  - Home directory: `~/.forgejo-mcp/config.yaml`
  - System-wide: `/etc/forgejo-mcp/config.yaml`

### Project Files
- Create `.gitignore` with Go-specific patterns:
  - Binary output files
  - Vendor directory
  - IDE-specific files (.idea, .vscode)
  - Test coverage reports
  - Environment files (.env)
- Add LICENSE file (MIT or Apache 2.0 based on preference)
- Create initial README.md with:
  - Project description
  - Installation instructions placeholder
  - Basic usage examples placeholder
  - Contributing guidelines reference

### Placeholder File Creation
- Create minimal placeholder files in each package directory:
  - Add package declaration
  - Add brief package documentation comment
  - Add basic imports where applicable
- Ensure all files compile without errors

### Performance Criteria
- Module initialization should complete in under 5 seconds
- Dependency download should use module proxy for speed
- Directory structure creation should be atomic (all-or-nothing)
- Configuration loading should fail fast with clear error messages

## External Dependencies

- **github.com/spf13/cobra@v1.8.0** - CLI framework for building modern command-line applications
  - **Justification:** Industry standard for Go CLI applications, provides command structure, flag parsing, and help generation
  
- **github.com/modelcontextprotocol/go-sdk/mcp@latest** - Official MCP SDK for Go
  - **Justification:** Required for implementing the Model Context Protocol server functionality
  
- **github.com/spf13/viper@v1.18.0** - Configuration management library
  - **Justification:** Provides flexible configuration with support for multiple formats, environment variables, and config file watching
  
- **github.com/sirupsen/logrus@v1.9.3** - Structured logging library
  - **Justification:** Mature logging solution with structured logging, multiple output formats, and extensive ecosystem support