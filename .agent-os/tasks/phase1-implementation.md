# Phase 1 Implementation Tasks: Core MCP Server

## 1. Project Foundation (Week 1)

### 1.1 Initialize Go Module
```bash
go mod init github.com/Kunde21/forgejo-mcp
```
- Add dependencies:
  - `github.com/spf13/cobra@v1.8.0`
  - `github.com/modelcontextprotocol/go-sdk/mcp@latest`
  - `github.com/spf13/viper@v1.18.0` (for config)
  - `github.com/sirupsen/logrus@v1.9.3` (for logging)

### 1.2 Create Project Structure
```
forgejo-mcp/
├── cmd/
│   ├── main.go
│   ├── root.go
│   └── serve.go
├── server/
│   ├── server.go
│   ├── tools.go
│   ├── transport.go
│   └── handlers.go
├── tea/
│   ├── wrapper.go
│   ├── tea.go
│   ├── commands.go
│   └── parser.go
├── context/
│   ├── git.go
│   ├── forgejo.go
│   └── context.go
├── auth/
│   ├── auth.go
│   ├── provider.go
│   └── tea.go
├── config/
│   └── config.go
├── types/
│   ├── pr.go
│   ├── issue.go
│   └── responses.go
└── test/
```

### 1.3 Setup Configuration Management
- Create `config/config.go`:
  ```go
  type Config struct {
      ForgejoURL   string
      AuthToken    string
      TeaPath      string
      Debug        bool
      LogLevel     string
  }
  ```
- Implement `Load()` function for env vars and config file
- Add validation for required fields

## 2. Cobra CLI Implementation (Week 2)

### 2.1 Root Command Setup
- File: `cmd/root.go`
- Implement:
  ```go
  func NewRootCommand() *cobra.Command
  ```
- Add global flags: `--config`, `--debug`, `--log-level`
- Add version information

### 2.2 Serve Command
- File: `cmd/serve.go`
- Implement:
  ```go
  func NewServeCommand() *cobra.Command
  func runServe(cmd *cobra.Command, args []string) error
  ```
- Add server-specific flags: `--port`, `--host`

### 2.3 Main Entry Point
- File: `cmd/main.go`
- Initialize Cobra app
- Setup signal handling for graceful shutdown
- Configure logrus logging:
  ```go
  func setupLogging(debug bool, logLevel string) {
      if debug {
          logrus.SetLevel(logrus.DebugLevel)
      } else {
          level, _ := logrus.ParseLevel(logLevel)
          logrus.SetLevel(level)
      }
      logrus.SetFormatter(&logrus.TextFormatter{
          FullTimestamp: true,
      })
  }
  ```

## 3. MCP Server Core (Week 3-4)

### 3.1 Server Initialization
- File: `server/server.go`
- Implement:
  ```go
  type Server struct {
      mcp    *mcp.Server
      config *config.Config
      tea    tea.Wrapper
      logger *logrus.Logger
  }
  
  func New(cfg *config.Config) (*Server, error)
  func (s *Server) Start() error
  func (s *Server) Stop() error
  ```
- Configure logrus with structured logging

### 3.2 Tool Registration
- File: `server/tools.go`
- Implement:
  ```go
  func (s *Server) registerTools() error
  func (s *Server) toolManifest() []mcp.Tool
  ```
- Define tool schemas for Phase 1:
  - `pr_list`
  - `issue_list`

### 3.3 Transport Layer
- File: `server/transport.go`
- Implement stdio transport:
  ```go
  func NewStdioTransport() mcp.Transport
  ```
- Handle connection lifecycle

### 3.4 Tool Handlers
- File: `server/handlers.go`
- Implement base handlers:
  ```go
  func (s *Server) handlePRList(params map[string]interface{}) (interface{}, error)
  func (s *Server) handleIssueList(params map[string]interface{}) (interface{}, error)
  ```

## 4. Tea CLI Wrapper (Week 4-5)

### 4.1 Wrapper Interface
- File: `tea/wrapper.go`
- Define interface:
  ```go
  type Wrapper interface {
      Execute(args ...string) ([]byte, error)
      ListPRs(filters ...string) ([]PR, error)
      ListIssues(filters ...string) ([]Issue, error)
  }
  ```

### 4.2 Tea Implementation
- File: `tea/tea.go`
- Implement:
  ```go
  type TeaCLI struct {
      execPath string
      timeout  time.Duration
  }
  
  func New(path string) (*TeaCLI, error)
  func (t *TeaCLI) Execute(args ...string) ([]byte, error)
  ```

### 4.3 Command Builders
- File: `tea/commands.go`
- Implement builders for each operation:
  ```go
  func buildPRListCommand(filters ...string) []string
  func buildIssueListCommand(filters ...string) []string
  ```

### 4.4 Output Parsers
- File: `tea/parser.go`
- Implement parsers:
  ```go
  func parsePRList(output []byte) ([]PR, error)
  func parseIssueList(output []byte) ([]Issue, error)
  ```

## 5. Repository Context Detection (Week 5)

### 5.1 Git Repository Detection
- File: `context/git.go`
- Implement:
  ```go
  func IsGitRepository(path string) bool
  func GetRemoteURL(name string) (string, error)
  ```

### 5.2 Forgejo Remote Validation
- File: `context/forgejo.go`
- Implement:
  ```go
  func IsForgejoRemote(url string) bool
  func ParseRepository(url string) (owner, repo string, err error)
  ```

### 5.3 Context Manager
- File: `context/context.go`
- Implement:
  ```go
  type Context struct {
      Owner      string
      Repository string
      RemoteURL  string
  }
  
  func DetectContext(path string) (*Context, error)
  ```

## 6. Authentication (Week 6)

### 6.1 Token Validation
- File: `auth/auth.go`
- Implement:
  ```go
  type Authenticator interface {
      Validate() error
      GetToken() string
  }
  ```

### 6.2 Token Provider
- File: `auth/provider.go`
- Implement:
  ```go
  type TokenProvider struct {
      token string
  }
  
  func NewFromEnv() (*TokenProvider, error)
  func NewFromFile(path string) (*TokenProvider, error)
  ```

### 6.3 Tea Authentication
- File: `auth/tea.go`
- Implement:
  ```go
  func ValidateWithTea(token string) error
  ```
- Use tea CLI to verify authentication works

## 7. Types and Models (Week 3-6, parallel)

### 7.1 Domain Types
- File: `types/pr.go`
- Define:
  ```go
  type PullRequest struct {
      Number      int
      Title       string
      Author      string
      State       string
      CreatedAt   time.Time
      UpdatedAt   time.Time
  }
  ```

### 7.2 Issue Types
- File: `types/issue.go`
- Define:
  ```go
  type Issue struct {
      Number      int
      Title       string
      Author      string
      State       string
      Labels      []string
      CreatedAt   time.Time
  }
  ```

### 7.3 Response Types
- File: `types/responses.go`
- Define standard response formats for MCP

## 8. Testing (Week 7)

### 8.1 Unit Tests
- Files: `*_test.go` alongside each implementation
- Cover:
  - Tea wrapper command building
  - Output parsing
  - Context detection
  - Authentication validation

### 8.2 Integration Tests
- File: `test/integration/server_test.go`
- Test MCP server startup/shutdown
- Test tool registration
- Test basic tool execution with mocked tea

### 8.3 E2E Tests
- File: `test/e2e/workflow_test.go`
- Test complete workflow with test Forgejo instance
- Cover authentication → context detection → tool execution

### 8.4 Logging Tests
- File: `test/logging_test.go`
- Test logrus configuration
- Test log levels and formatting
- Test structured logging fields

## 9. Documentation (Week 8)

### 9.1 API Documentation
- File: `docs/API.md`
- Document all MCP tools and their parameters
- Include example requests/responses

### 9.2 Setup Guide
- File: `docs/SETUP.md`
- Installation instructions
- Authentication setup
- Configuration options

### 9.3 Development Guide
- File: `docs/DEVELOPMENT.md`
- Architecture overview
- Contributing guidelines
- Testing instructions

## 10. Build and Release (Week 8)

### 10.1 Makefile
- Create comprehensive Makefile with targets:
  ```makefile
  build:     # Build the binary
  test:      # Run all tests
  lint:      # Run linters
  install:   # Install locally
  release:   # Create release artifacts
  ```

### 10.2 GitHub Actions
- File: `.github/workflows/ci.yml`
- Setup CI pipeline:
  - Run tests on PR
  - Build binaries for multiple platforms
  - Run security scanning

### 10.3 Release Process
- File: `.github/workflows/release.yml`
- Automate releases:
  - Tag-triggered builds
  - Cross-platform binaries
  - GitHub release creation

## Timeline Summary

- **Week 1**: Project foundation and setup
- **Week 2**: Cobra CLI implementation
- **Week 3-4**: MCP server core implementation (with logrus logging)
- **Week 4-5**: Tea CLI wrapper
- **Week 5**: Repository context detection
- **Week 6**: Authentication system
- **Week 7**: Testing suite
- **Week 8**: Documentation and release preparation

## Success Criteria

- [x] MCP server starts and registers tools
- [x] Authentication with Forgejo validates successfully
- [x] Repository context is correctly detected
- [x] PR list tool returns data from tea CLI
- [x] Issue list tool returns data from tea CLI
- [x] All tests pass with >80% coverage
- [x] Documentation is complete and accurate
- [x] Binary builds for Linux, macOS, Windows