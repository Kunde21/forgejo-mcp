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
- [x] Create `cmd/` directory with main.go, root.go, serve.go
- [x] Create `server/` directory for MCP server implementation
- [x] Create `client/` directory for Gitea SDK client
- [x] Create `context/` directory for repository context detection
- [x] Create `auth/` directory for authentication logic
- [x] Create `config/` directory for configuration management
- [x] Create `types/` directory for data types and models

## 2. Cobra CLI Structure ✅ COMPLETED

### 2.1 Initialize Root Command ✅
- [x] Create `cmd/root.go` with root command setup
- [x] Implement `NewRootCommand() *cobra.Command` function
- [x] Add global flags: `--config`, `--debug`, `--log-level`
- [x] Set up logrus logging configuration
- [x] Set up command descriptions and usage examples

### 2.2 Create Server Command ✅
- [x] Create `cmd/serve.go` for server command
- [x] Implement `NewServeCommand() *cobra.Command` function
- [x] Add server-specific flags: `--host`, `--port`
- [x] Add command aliases and examples

### 2.3 Main Entry Point ✅
- [x] Create `main.go` as main entry point
- [x] Initialize Cobra app
- [x] Setup signal handling for graceful shutdown
- [x] Configure logrus with appropriate formatter and levels

### 2.4 Additional Cobra Implementation (from recap) ✅
- [x] Extend `cmd/logging.go` with Cobra-integrated logging setup
- [x] Implement configuration loading with flag overrides
- [x] Add configuration validation in command PreRunE
- [x] Implement comprehensive error handling with proper exit codes
- [x] Write unit tests for command functions and flag parsing
- [x] Test signal handling and graceful shutdown
- [x] Add godoc comments to all exported functions
- [x] Update command help text with examples
- [x] Document CLI usage in README
- [x] Verify CLI starts and shows help with `forgejo-mcp --help`
- [x] Verify server command runs with `forgejo-mcp serve`
- [x] Verify graceful shutdown works with Ctrl+C
- [x] Verify all flags work as documented
- [x] Verify tests pass with good coverage

## 3. MCP Server Implementation ✅ COMPLETED

### 3.1 Core Server Structure ✅
- [x] Create `server/server.go` with main server struct
- [x] Define `type Server struct` with mcp.Server, config, Gitea client, and logger
- [x] Implement `New(cfg *config.Config) (*Server, error)`
- [x] Implement `(s *Server) Start() error`
- [x] Implement `(s *Server) Stop() error`
- [x] Configure logrus logger instance for server
- [x] Add server configuration to `config/config.go`
- [x] Integrate Viper configuration with environment variables
- [x] Write tests for Server struct lifecycle (New, Start, Stop)

### 3.2 Transport Layer ✅
- [x] Create `server/transport.go` for transport handling
- [x] Implement stdio transport using MCP SDK
- [x] Implement `NewStdioTransport() mcp.Transport`
- [x] Handle connection lifecycle
- [x] Set up request routing based on tool name
- [x] Implement JSON-RPC message handling over stdin/stdout
- [x] Create request dispatcher and router for tool mapping
- [x] Add connection lifecycle management
- [x] Implement timeout handling for requests
- [x] Write tests for stdio transport and request routing

### 3.3 Tool Registration ✅
- [x] Create `server/tools.go` for tool definitions
- [x] Implement `(s *Server) registerTools() error`
- [x] Implement `(s *Server) toolManifest() []mcp.Tool`
- [x] Register PR list tool: `pr_list`
- [x] Register issue list tool: `issue_list`
- [x] Define JSON schemas for pr_list and issue_list tools
- [x] Add parameter validation rules
- [x] Write tests for tool registration and manifest generation

### 3.4 Tool Handlers ✅
- [x] Create `server/handlers.go` for tool handler implementations
- [x] Implement `(s *Server) handlePRList(params map[string]interface{}) (interface{}, error)`
- [x] Implement `(s *Server) handleIssueList(params map[string]interface{}) (interface{}, error)`
- [x] Implement error response formatting
- [x] Implement parameter extraction and validation
- [x] Create Gitea API request builders with proper parameter handling
- [x] Add Gitea response parsing and transformation
- [x] Implement response transformation to MCP format
- [x] Write tests for request handlers and tea command building

### 3.5 Integration Testing and Validation ✅
- [x] Write integration tests for complete request/response flow
- [x] Test server startup and MCP connection acceptance
- [x] Test tool discovery through manifest
- [x] Test pr_list with mocked Gitea API responses
- [x] Test issue_list with mocked Gitea API responses
- [x] Test error handling and timeout scenarios
- [x] Verify all tests pass with >80% coverage

## 4. Gitea SDK Client ✅ COMPLETED

### 4.1 Client Interface and Core Structure ✅
- [x] Create `client/client.go` with client interface
- [x] Define `type Client interface` with required methods
- [x] Define `type ForgejoClient struct` with Gitea SDK integration
- [x] Implement `New(baseURL, token string) (*ForgejoClient, error)`
- [x] Add comprehensive error handling with custom error types
- [x] Implement client configuration with defaults
- [x] Add validation for client creation
- [x] Write comprehensive tests for client interface and validation

### 4.2 Gitea SDK Integration ✅
- [x] Create `tea/wrapper.go` for Gitea SDK wrapper
- [x] Implement Gitea SDK client initialization and configuration
- [x] Add authentication support (token-based, basic auth, fallback)
- [x] Implement connection validation and health checks
- [x] Add comprehensive error handling and transformation
- [x] Write tests for authentication methods and wrapper functionality

### 4.3 Request Building and Filtering ✅
- [x] Create `tea/repositories.go` for repository operations
- [x] Implement `buildRepoListOptions(filters *RepositoryFilters) *gitea.ListReposOptions`
- [x] Implement `buildSearchRepoOptions(filters *RepositoryFilters) *gitea.SearchRepoOptions`
- [x] Add support for advanced filtering (pagination, search, ownership, visibility, sorting)
- [x] Replace `map[string]interface{}` with type-safe `RepositoryFilters` struct
- [x] Implement `ListRepositories()` method with comprehensive filtering
- [x] Implement `GetRepository()` method for individual repository retrieval
- [x] Write comprehensive tests for all filtering functionality

### 4.4 Response Transformation ✅
- [x] Implement `transformRepository(giteaRepo *gitea.Repository) Repository`
- [x] Handle field mapping and metadata preservation
- [x] Add proper error handling for transformation failures
- [x] Write tests for transformation functions

### 4.5 Performance and Caching ✅
- [x] Create `tea/cache.go` with in-memory cache implementation
- [x] Implement TTL-based expiration with automatic cleanup
- [x] Add size-limited cache with LRU eviction policy
- [x] Implement cache statistics tracking
- [x] Create `tea/batch.go` with batch processing capabilities
- [x] Add concurrency control with configurable limits
- [x] Implement request deduplication and optimization
- [x] Add performance benchmarks for cache and batch operations
- [x] Write comprehensive tests for caching and batch processing

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
- [ ] Create `client/client_test.go`
- [ ] Create `context/context_test.go`
- [ ] Create `auth/auth_test.go`
- [ ] Add logrus logging tests
- [ ] Achieve minimum 80% code coverage

### 8.2 Integration Tests
- [ ] Create `test/integration/server_test.go`
- [ ] Test MCP server startup/shutdown
- [ ] Test tool registration
- [ ] Test basic tool execution with mocked Gitea client
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
- [ ] Document Gitea SDK client configuration
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

- **Week 1**: Project foundation and setup ✅ COMPLETED
- **Week 2**: Cobra CLI implementation ✅ COMPLETED
- **Week 3-4**: MCP server core implementation (with logrus logging) ✅ COMPLETED
- **Week 4-5**: Gitea SDK client implementation
- **Week 5**: Repository context detection
- **Week 6**: Authentication system
- **Week 7**: Testing suite
- **Week 8**: Documentation and release preparation

Total estimated time: 8 weeks for full Phase 1 implementation
