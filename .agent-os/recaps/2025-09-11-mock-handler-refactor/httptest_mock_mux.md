# Plan: Split Mock HTTP Handler Using http.ServeMux with Path Patterns

## Current Problem
The `handleRepoRequests` function in `server_test/harness.go` is 310+ lines long and handles multiple endpoints with complex string matching logic. This monolithic function is difficult to maintain, test, and extend.

## Current Implementation Issues
- Single function handles 5 different API endpoints
- Complex string parsing and manual method checking
- Code duplication for common operations
- Difficult to test individual endpoint logic
- Hard to add new endpoints without affecting existing ones

## Proposed Solution Using http.ServeMux

### 1. Create Individual Handler Functions
Split the monolithic handler into separate, focused handlers:

```go
// New handler functions to be created:
func (m *MockGiteaServer) handlePullRequests(w http.ResponseWriter, r *http.Request)
func (m *MockGiteaServer) handleIssues(w http.ResponseWriter, r *http.Request)
func (m *MockGiteaServer) handleCreateComment(w http.ResponseWriter, r *http.Request)
func (m *MockGiteaServer) handleListComments(w http.ResponseWriter, r *http.Request)
func (m *MockGiteaServer) handleEditComment(w http.ResponseWriter, r *http.Request)
```

### 2. Update Handler Registration with ServeMux Patterns
Replace the current handler registration with specific method + path patterns:

```go
// In NewMockGiteaServer():
handler := http.NewServeMux()
handler.HandleFunc("/api/v1/version", mock.handleVersion)

// Register specific endpoint patterns
handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", mock.handlePullRequests)
handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", mock.handleIssues)
handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleCreateComment)
handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleListComments)
handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", mock.handleEditComment)
```

### 3. Extract Common Helper Functions
Create shared utilities to reduce code duplication:

```go
// Helper functions to extract:
func (m *MockGiteaServer) getRepoKeyFromRequest(r *http.Request) string
func (m *MockGiteaServer) validateRepository(repoKey string) bool
func (m *MockGiteaServer) parsePagination(r *http.Request) (limit, offset int)
func (m *MockGiteaServer) validateAuthToken(r *http.Request) bool
func (m *MockGiteaServer) writeJSONResponse(w http.ResponseWriter, data any, statusCode int)
```

### 4. Refactor Each Handler to Use Path Values
Each handler will use `r.PathValue()` to extract path parameters:

```go
// Example for handlePullRequests:
func (m *MockGiteaServer) handlePullRequests(w http.ResponseWriter, r *http.Request) {
    owner := r.PathValue("owner")
    repo := r.PathValue("repo")
    repoKey := owner + "/" + repo
    
    // Check if repository is marked as not found
    if m.notFoundRepos[repoKey] {
        http.NotFound(w, r)
        return
    }
    
    // Rest of the pull request logic...
}
```

## Benefits of This Approach

1. **Clean Pattern Matching**: `http.ServeMux` handles method + path matching automatically
2. **Automatic Path Parameter Extraction**: Use `r.PathValue()` instead of manual string parsing
3. **Better Performance**: Built-in routing is more efficient than manual string matching
4. **Standard Go Idioms**: Uses modern Go 1.22+ `http.ServeMux` features
5. **Maintainability**: Clear separation of concerns with focused handlers
6. **Testability**: Individual handlers can be tested in isolation
7. **Extensibility**: New endpoints can be added without modifying existing logic

## Implementation Steps

### Phase 1: Create Helper Functions
- Extract common logic from `handleRepoRequests` into reusable helper functions
- Focus on operations like repository validation, pagination parsing, and response writing

### Phase 2: Implement Individual Handler Functions
- Create each handler function by copying relevant logic from `handleRepoRequests`
- Update each handler to use `r.PathValue()` for parameter extraction
- Ensure each handler is self-contained and focused on its specific endpoint

### Phase 3: Update Handler Registration
- Replace `http.NewServeMux()` usage with specific method + path patterns
- Remove the old `handleRepoRequests` registration
- Test that the new routing works correctly

### Phase 4: Remove Legacy Code
- Delete the old `handleRepoRequests` function
- Clean up any unused code or imports
- Ensure all functionality is preserved

### Phase 5: Testing and Validation
- Run all existing tests to ensure no functionality is broken
- Verify that all endpoints still work as expected
- Check that error handling and edge cases are preserved

### Phase 6: Documentation and Cleanup
- Update any relevant documentation
- Add comments to new handler functions if needed
- Ensure code follows project style guidelines

## Key Changes from Original Implementation

### Before (Current):
```go
handler.HandleFunc("/api/v1/repos/", mock.handleRepoRequests)

func (m *MockGiteaServer) handleRepoRequests(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    
    // Handle pull requests endpoint
    if strings.Contains(path, "/pulls") && r.Method == "GET" {
        // 60+ lines of pull request logic with manual string parsing
    }
    
    // Handle issues endpoint
    if strings.Contains(path, "/issues") && !strings.Contains(path, "/comments") && r.Method == "GET" {
        // 20+ lines of issues logic with manual string parsing
    }
    
    // ... more endpoints with complex string matching
}
```

### After (Proposed):
```go
handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", mock.handlePullRequests)

func (m *MockGiteaServer) handlePullRequests(w http.ResponseWriter, r *http.Request) {
    owner := r.PathValue("owner")
    repo := r.PathValue("repo")
    repoKey := owner + "/" + repo
    
    // Clean, focused pull request logic
}
```

## Files to Modify
- `server_test/harness.go` - Main refactoring (only file that needs changes)
- No changes needed to test files as the public API remains the same

## Risk Assessment
- **Low Risk**: The refactoring is internal to the test harness and doesn't change the public API
- **Backward Compatibility**: All existing tests should continue to work without modification
- **Testing Coverage**: Existing tests provide good coverage for the refactored functionality

## Success Criteria
- All existing tests pass without modification
- Code is more maintainable and easier to understand
- Individual handlers are focused and single-purpose
- Performance is maintained or improved
- New endpoints can be added easily in the future

## Notes
- This refactoring leverages Go 1.22+ `http.ServeMux` features
- The approach follows modern Go HTTP handling best practices
- No breaking changes to the test harness public API