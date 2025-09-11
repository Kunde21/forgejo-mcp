# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-11-mock-handler-refactor/spec.md

## Technical Requirements

- **Handler Function Separation**: Create 5 individual handler functions to replace the monolithic `handleRepoRequests` function:
  - `handlePullRequests(w http.ResponseWriter, r *http.Request)` - GET /api/v1/repos/{owner}/{repo}/pulls
  - `handleIssues(w http.ResponseWriter, r *http.Request)` - GET /api/v1/repos/{owner}/{repo}/issues  
  - `handleCreateComment(w http.ResponseWriter, r *http.Request)` - POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments
  - `handleListComments(w http.ResponseWriter, r *http.Request)` - GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments
  - `handleEditComment(w http.ResponseWriter, r *http.Request)` - PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}

- **Modern Routing Registration**: Replace current handler registration with method + path patterns using Go 1.22+ `http.ServeMux`:
  ```go
  handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", mock.handlePullRequests)
  handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", mock.handleIssues)
  handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleCreateComment)
  handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleListComments)
  handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", mock.handleEditComment)
  ```

- **Helper Function Extraction**: Create reusable utility functions to reduce code duplication:
  - `getRepoKeyFromRequest(r *http.Request) string` - Extract repository key from path values
  - `validateRepository(repoKey string) bool` - Check if repository exists and is accessible
  - `parsePagination(r *http.Request) (limit, offset int)` - Parse pagination parameters from query
  - `validateAuthToken(r *http.Request) bool` - Validate authentication token from request
  - `writeJSONResponse(w http.ResponseWriter, data any, statusCode int)` - Write standardized JSON responses

- **Path Parameter Modernization**: Replace manual string parsing with `r.PathValue()`:
  ```go
  // Before: Manual string parsing
  path := r.URL.Path
  if strings.Contains(path, "/pulls") && r.Method == "GET" {
      // Complex string manipulation to extract owner/repo
  }
  
  // After: Automatic parameter extraction
  owner := r.PathValue("owner")
  repo := r.PathValue("repo")
  repoKey := owner + "/" + repo
  ```

- **Error Handling Preservation**: Maintain all existing error handling logic including:
  - Repository not found responses (404)
  - Authentication validation failures (401)
  - Invalid request parameter handling (400)
  - Rate limiting and throttling responses

- **Performance Optimization**: Leverage built-in `http.ServeMux` routing for improved performance over manual string matching operations

## External Dependencies

No new external dependencies are required. This refactoring uses built-in Go 1.22+ `http.ServeMux` features that are already available in the project's Go version (1.24.6+).

**Justification**: Using standard library features maintains project simplicity, reduces dependency overhead, and follows Go best practices for HTTP handler implementation.