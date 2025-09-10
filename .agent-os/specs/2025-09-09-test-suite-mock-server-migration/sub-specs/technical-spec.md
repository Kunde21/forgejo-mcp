# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-09-test-suite-mock-server-migration/spec.md

## Technical Requirements

- **HTTP Mock Server Migration**: Replace all NewMockGiteaClient() calls with NewMockGiteaServer(t) in issue_comment_test.go
- **Test Setup Standardization**: Replace NewTestServerWithClient() calls with NewTestServer() using mock server URL configuration
- **Code Cleanup**: Remove MockGiteaClient struct definition and all associated methods from harness.go
- **Endpoint Verification**: Ensure MockGiteaServer properly handles POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments endpoint
- **Response Format Validation**: Verify mock server returns correct JSON responses for comment creation with proper error handling

## External Dependencies (Conditional)

No new external dependencies are required for this specification. The migration utilizes existing Go testing packages and the current MockGiteaServer implementation.

## Implementation Details

### File Changes Required

1. **server_test/issue_comment_test.go**:
   - Replace `mockClient := NewMockGiteaClient()` with `mock := NewMockGiteaServer(t)`
   - Replace `NewTestServerWithClient(t, ctx, env, mockClient)` with `NewTestServer(t, ctx, env)`
   - Update environment configuration to use `mock.URL()` for FORGEJO_REMOTE_URL
   - Remove all mock client setup and data preparation code

2. **server_test/harness.go**:
   - Remove MockGiteaClient struct definition (lines 56-61)
   - Remove NewMockGiteaClient() function (lines 63-70)
   - Remove MockGiteaClient methods: ListIssues, CreateIssueComment, AddMockIssues (lines 72-118)
   - Remove NewTestServerWithClient() function (lines 234-312)
   - Keep only NewTestServer() function that uses mock server approach

### MockGiteaServer Enhancement Requirements

The existing MockGiteaServer must support:
- **Comment Creation**: Proper handling of POST requests to comment endpoints
- **Error Scenarios**: Return appropriate HTTP status codes for invalid repositories/issue numbers
- **Response Format**: Consistent JSON response structure matching Gitea API format
- **Data Persistence**: Store created comments for verification in test assertions

### Testing Considerations

- **Backward Compatibility**: Ensure no existing functionality is broken during migration
- **Test Coverage**: Maintain or improve existing test coverage levels
- **Performance**: HTTP-based testing should not significantly impact test execution time
- **Reliability**: Mock server should be stable and not introduce flakiness in tests