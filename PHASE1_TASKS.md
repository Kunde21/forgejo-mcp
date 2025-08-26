# Phase 1 Implementation Tasks

## 1. Project Setup and Go Modules

### 1.1 Initialize Go Module
- [x] Run `go mod init github.com/kunde21/forgejo-mcp`
- [x] Create `.gitignore` with Go-specific patterns
- [x] Add LICENSE file (MIT or Apache 2.0)

### 1.2 Add Core Dependencies
- [x] Add Cobra: `go get github.com/spf13/cobra@v1.8.0`
- [x] Add MCP Go SDK: `go get github.com/modelcontextprotocol/go-sdk/mcp@latest`
- [x] Add Viper for config: `go get github.com/spf13/viper@v1.18.0`
- [x] Add logging library: `go get github.com/sirupsen/logrus@v1.9.3`
- [x] Run `go mod tidy` to clean up dependencies

### 1.3 Project Structure Setup
- [ ] Create `cmd/` directory with main.go, root.go, serve.go
- [ ] Create `server/` directory for MCP server implementation
- [ ] Create `tea/` directory for tea CLI wrapper
- [ ] Create `context/` directory for repository context detection
- [ ] Create `auth/` directory for authentication logic
- [ ] Create `config/` directory for configuration management
- [ ] Create `types/` directory for data types and models
- [ ] Create `test/` directory for test files
- [ ] Create `scripts/` directory for build and deployment scripts

## 2. Cobra CLI Structure

### 2.1 Initialize Root Command
- [ ] Create `cmd/root.go` with root command setup
- [ ] Implement `NewRootCommand() *cobra.Command` function
- [ ] Add global flags: `--config`, `--debug`, `--log-level`
- [ ] Set up logrus logging configuration
- [ ] Set up command descriptions and usage examples

### 2.2 Create Server Command
- [ ] Create `cmd/serve.go` for server command
- [ ] Implement `NewServeCommand() *cobra.Command` function
- [ ] Add server-specific flags: `--host`, `--port`
- [ ] Add command aliases and examples

### 2.3 Main Entry Point
- [ ] Create `main.go` as main entry point
- [ ] Initialize Cobra app
- [ ] Setup signal handling for graceful shutdown
- [ ] Configure logrus with appropriate formatter and levels

## 3. MCP Server Implementation

### 3.1 Core Server Structure
- [ ] Create `server/server.go` with main server struct
- [ ] Define `type Server struct` with mcp.Server, config, tea wrapper, and logger
- [ ] Implement `New(cfg *config.Config) (*Server, error)`
- [ ] Implement `(s *Server) Start() error`
- [ ] Implement `(s *Server) Stop() error`
- [ ] Configure logrus logger instance for server

### 3.2 Transport Layer
- [ ] Create `server/transport.go` for transport handling
- [ ] Implement stdio transport using MCP SDK
- [ ] Implement `NewStdioTransport() mcp.Transport`
- [ ] Handle connection lifecycle
- [ ] Set up request routing based on tool name

### 3.3 Tool Registration
- [ ] Create `server/tools.go` for tool definitions
- [ ] Implement `(s *Server) registerTools() error`
- [ ] Implement `(s *Server) toolManifest() []mcp.Tool`
- [ ] Register PR list tool: `pr_list`
- [ ] Register issue list tool: `issue_list`

### 3.4 Tool Handlers
- [ ] Create `server/handlers.go` for tool handler implementations
- [ ] Implement `(s *Server) handlePRList(params map[string]interface{}) (interface{}, error)`
- [ ] Implement `(s *Server) handleIssueList(params map[string]interface{}) (interface{}, error)`
- [ ] Implement error response formatting

## 4. Tea CLI Wrapper

### 4.1 Tea Wrapper Interface
- [ ] Create `tea/wrapper.go` with wrapper interface
- [ ] Define `type Wrapper interface` with required methods
- [ ] Define `type TeaCLI struct` with execPath and timeout
- [ ] Implement `New(path string) (*TeaCLI, error)`

### 4.2 Tea Implementation
- [ ] Create `tea/tea.go` for tea CLI execution
- [ ] Implement `(t *TeaCLI) Execute(args ...string) ([]byte, error)`
- [ ] Implement `(t *TeaCLI) ListPRs(filters ...string) ([]PR, error)`
- [ ] Implement `(t *TeaCLI) ListIssues(filters ...string) ([]Issue, error)`

### 4.3 Command Builders
- [ ] Create `tea/commands.go` for command building
- [ ] Implement `buildPRListCommand(filters ...string) []string`
- [ ] Implement `buildIssueListCommand(filters ...string) []string`
- [ ] Add support for filter parameters

### 4.4 Output Parsers
- [ ] Create `tea/parser.go` for output parsing
- [ ] Implement `parsePRList(output []byte) ([]PR, error)`
- [ ] Implement `parseIssueList(output []byte) ([]Issue, error)`
- [ ] Add JSON parsing for tea output
- [ ] Handle text format fallback

## 5. Repository Context Detection

### 5.1 Git Repository Detection
- [ ] Create `context/git.go` for git repository detection
- [ ] Implement `IsGitRepository(path string) bool`
- [ ] Implement `GetRemoteURL(name string) (string, error)`
- [ ] Validate `.git` directory exists

### 5.2 Forgejo Remote Validation
- [ ] Create `context/forgejo.go` for Forgejo validation
- [ ] Implement `IsForgejoRemote(url string) bool`
- [ ] Implement `ParseRepository(url string) (owner, repo string, err error)`
- [ ] Support both SSH and HTTPS URLs
- [ ] Extract owner and repository name from remote URL

### 5.3 Context Manager
- [ ] Create `context/context.go` for context management
- [ ] Define `type Context struct` with Owner, Repository, RemoteURL
- [ ] Implement `DetectContext(path string) (*Context, error)`
- [ ] Integrate git detection and Forgejo validation
- [ ] Cache context for performance



## 6. Authentication Validation

### 6.1 Token Validation
- [ ] Create `auth/auth.go` for authentication interface
- [ ] Define `type Authenticator interface` with Validate() and GetToken()
- [ ] Implement authentication validation logic
- [ ] Return helpful error messages for auth failures

### 6.2 Token Provider
- [ ] Create `auth/provider.go` for token providers
- [ ] Define `type TokenProvider struct` with token field
- [ ] Implement `NewFromEnv() (*TokenProvider, error)` for env vars
- [ ] Implement `NewFromFile(path string) (*TokenProvider, error)` for file-based tokens
- [ ] Support reading from environment variable `FORGEJO_TOKEN`

### 6.3 Tea Authentication
- [ ] Create `auth/tea.go` for tea-based authentication
- [ ] Implement `ValidateWithTea(token string) error`
- [ ] Use tea CLI to verify authentication works
- [ ] Cache validation results for performance



## 7. Types and Models

### 7.1 Domain Types
- [ ] Create `types/pr.go` with PullRequest struct
- [ ] Define fields: Number, Title, Author, State, CreatedAt, UpdatedAt
- [ ] Add JSON tags for serialization
- [ ] Add validation methods

### 7.2 Issue Types
- [ ] Create `types/issue.go` with Issue struct
- [ ] Define fields: Number, Title, Author, State, Labels, CreatedAt
- [ ] Add JSON tags for serialization
- [ ] Add validation methods

### 7.3 Response Types
- [ ] Create `types/responses.go` for MCP responses
- [ ] Define standard response formats
- [ ] Add error response types
- [ ] Add success response types

## 8. Integration and Testing

### 8.1 Unit Tests
- [ ] Create `server/server_test.go`
- [ ] Create `tea/wrapper_test.go`
- [ ] Create `context/context_test.go`
- [ ] Create `auth/auth_test.go`
- [ ] Add logrus logging tests
- [ ] Achieve minimum 80% code coverage

### 8.2 Integration Tests
- [ ] Create `test/integration/server_test.go`
- [ ] Test MCP server startup/shutdown
- [ ] Test tool registration
- [ ] Test basic tool execution with mocked tea
- [ ] Test error handling scenarios

### 8.3 End-to-End Tests
- [ ] Create `test/e2e/workflow_test.go`
- [ ] Test complete workflow with test Forgejo instance
- [ ] Cover authentication → context detection → tool execution
- [ ] Test actual PR listing functionality
- [ ] Test actual issue listing functionality
- [ ] Document manual testing procedures

## 9. Documentation

### 9.1 API Documentation
- [ ] Create `docs/API.md` with MCP tool specifications
- [ ] Document all tool parameters and responses
- [ ] Include example requests/responses
- [ ] Add authentication setup guide

### 9.2 Setup Guide
- [ ] Create `docs/SETUP.md` with installation instructions
- [ ] Document authentication configuration
- [ ] Document tea CLI installation steps
- [ ] Add configuration options
- [ ] Add troubleshooting section

### 9.3 Development Guide
- [ ] Create `docs/DEVELOPMENT.md` with architecture overview
- [ ] Document contribution guidelines
- [ ] Add inline godoc comments for all exported functions
- [ ] Document testing instructions

## 10. Build and Release

### 10.1 Build Scripts
- [ ] Create `Makefile` with common tasks
- [ ] Add `make build` target
- [ ] Add `make test` target
- [ ] Add `make install` target
- [ ] Add `make clean` target

### 10.2 CI/CD Setup
- [ ] Create `.github/workflows/ci.yml` for GitHub Actions
- [ ] Add build job for multiple Go versions
- [ ] Add test job with coverage reporting
- [ ] Add linting job (golangci-lint)
- [ ] Add security scanning (gosec)

### 10.3 Release Process
- [ ] Create `scripts/release.sh` for releases
- [ ] Set up semantic versioning
- [ ] Create build artifacts for Linux/Mac/Windows
- [ ] Generate checksums for releases
- [ ] Create GitHub release automation

## Success Criteria Checklist

- [ ] MCP server starts and accepts connections
- [ ] AI agents can authenticate and connect
- [ ] `pr_list` tool returns PR data from Forgejo
- [ ] `issue_list` tool returns issue data from Forgejo
- [ ] Repository context is correctly detected
- [ ] Authentication errors are clearly reported
- [ ] All unit tests pass with >80% coverage
- [ ] Documentation is complete and accurate
- [ ] Binary builds successfully for target platforms
- [ ] Manual testing confirms all Phase 1 features work

## Timeline Summary

- **Week 1**: Project foundation and setup
- **Week 2**: Cobra CLI implementation
- **Week 3-4**: MCP server core implementation (with logrus logging)
- **Week 4-5**: Tea CLI wrapper
- **Week 5**: Repository context detection
- **Week 6**: Authentication system
- **Week 7**: Testing suite
- **Week 8**: Documentation and release preparation

Total estimated time: 8 weeks for full Phase 1 implementation
