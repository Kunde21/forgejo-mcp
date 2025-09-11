# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-11-pr-list/spec.md

> Created: 2025-09-11
> Version: 1.0.0

## Technical Requirements

### Tool Definition
- **Tool Name**: `pr_list`
- **Description**: Lists pull requests in a Forgejo/Gitea repository with optional filtering and pagination
- **Parameters**:
  - `repository` (string, required): Repository name in format "owner/repo"
  - `limit` (integer, optional, default: 15): Maximum number of PRs to return
  - `offset` (integer, optional, default: 0): Offset for pagination
  - `state` (string, optional, default: "open"): PR state filter ("open", "closed", "all")

### Architecture Layers

#### 1. Interface Layer (`remote/gitea/interface.go`)
- Add `ListPullRequests` method to the `GiteaClient` interface
- Method signature: `ListPullRequests(ctx context.Context, repo string, opts ListPullRequestsOptions) ([]PullRequest, error)`
- Define `ListPullRequestsOptions` struct with pagination and filtering parameters
- Define `PullRequest` struct for PR data representation

#### 2. Client Layer (`remote/gitea/gitea_client.go`)
- Implement `ListPullRequests` method using Forgejo/Gitea API
- Use HTTP client to call `/repos/{owner}/{repo}/pulls` endpoint
- Handle API response parsing and error mapping
- Support query parameters for state, page, and limit

#### 3. Service Layer (`remote/gitea/service.go`)
- Add `ListPullRequests` method to the service
- Handle business logic and coordinate between client and handler
- Map between API response models and internal data structures
- Implement proper error handling and logging

#### 4. Handler Layer (`server/pr_list.go`)
- Create new handler file `server/pr_list.go`
- Implement MCP tool handler following existing patterns
- Validate input parameters using ozzo-validation
- Call service layer and format response for MCP
- Handle errors and return appropriate MCP responses

### Validation Rules
- `repository`: Required string, must match format "owner/repo" using regex validation
- `limit`: Optional integer, must be between 1 and 100 (inclusive)
- `offset`: Optional integer, must be >= 0
- `state`: Optional string, must be one of "open", "closed", "all"

### Error Handling
- Return structured errors with clear messages for invalid inputs
- Handle API errors from Forgejo/Gitea with proper mapping
- Use consistent error formatting across all layers
- Log errors appropriately for debugging

### MCP Integration
- Register new tool in `server/server.go` `GetTools` method
- Add tool definition with proper schema and descriptions
- Follow existing MCP tool registration patterns
- Ensure tool is properly exposed to MCP clients

## Approach

### Implementation Strategy
1. **Phase 1: Interface and Data Models**
   - Define interface methods and data structures
   - Create PullRequest and ListPullRequestsOptions structs
   - Update interface.go with new method signatures

2. **Phase 2: Client Implementation**
   - Implement ListPullRequests in gitea_client.go
   - Add HTTP client logic for Forgejo/Gitea PR API
   - Handle API response parsing and error handling

3. **Phase 3: Service Layer**
   - Implement service method in service.go
   - Add business logic and data transformation
   - Implement proper error handling and logging

4. **Phase 4: Handler and MCP Integration**
   - Create pr_list.go handler file
   - Implement MCP tool handler with validation
   - Register tool in server.go
   - Add comprehensive tests

### Code Patterns to Follow
- Use existing issue_list implementation as reference
- Follow the same error handling patterns with fmt.Errorf
- Use ozzo-validation for input validation
- Implement table-driven tests following project standards
- Use context.Context for all method signatures
- Follow Go naming conventions and code style

### Testing Strategy
- **Unit Tests**: Test each layer independently with mocked dependencies
- **Integration Tests**: Test end-to-end functionality with test harness
- **Validation Tests**: Test all validation scenarios and edge cases
- **Error Handling Tests**: Test error paths and error message formatting
- **API Response Tests**: Test various API response scenarios

## External Dependencies

### Required Dependencies
- `github.com/go-ozzo/ozzo-validation` - Input validation
- `github.com/go-ozzo/ozzo-validation/is` - Validation helpers
- Existing Forgejo/Gitea API client patterns
- MCP SDK for tool registration and handling

### API Endpoints
- `GET /repos/{owner}/{repo}/pulls` - List pull requests
- Query parameters: `state`, `page`, `limit`

### Data Models
- PullRequest: ID, Number, Title, Body, State, User, CreatedAt, UpdatedAt, Head, Base
- ListPullRequestsOptions: State, Page, Limit

### Integration Points
- Existing GiteaClient interface
- MCP server tool registration
- Error handling and logging systems
- Configuration and authentication systems