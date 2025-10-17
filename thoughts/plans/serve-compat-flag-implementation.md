# Serve --compat Flag Implementation Plan

## Overview

Add a `--compat` flag to the `forgejo-mcp serve` command that controls whether detailed text responses are included alongside structured data in tool responses. When enabled (compatibility mode), tools return detailed human-readable text summaries. When disabled (default), tools return simple summary messages, reducing duplicate data for modern MCP clients.

## Current State Analysis

The codebase currently has 13+ tools that all return both detailed human-readable text and structured data objects. This creates duplication where modern MCP clients receive the same information in both formats. The text formatting logic is embedded directly within each tool handler after the remote API call.

### Key Discoveries:
- Response helpers in `server/common.go:13-27` create text-only responses
- Tool handlers follow pattern: remote call → text formatting → return both formats
- Debug flag implementation provides clear pattern for CLI flag addition
- Server struct in `server/server.go:20-25` stores configuration state
- No MCP client capability detection exists for structured response preference

## Desired End State

- Default: Tools return simple summary messages in text + full structured data
- Compatibility mode (--compat): Tools return detailed text summaries + structured data
- CLI-only flag (no config file or environment variable support)
- All 13+ tools updated with conditional response building
- Comprehensive test coverage for both modes

## What We're NOT Doing

- Not adding environment variable or config file support for the flag
- Not changing error responses (they remain text-only)
- Not affecting tools without structured data (hello tool unchanged)
- Not implementing automatic client capability detection

## Implementation Approach

Following the established debug flag pattern: CLI flag definition → constructor chain → server struct storage → conditional behavior in handlers.

## Phase 1: CLI Flag and Server Infrastructure

### Overview
Add the --compat flag to CLI, update server struct to store the flag, and pass it through the constructor chain.

### Changes Required:

#### 1. cmd/serve.go
**File**: `cmd/serve.go`
**Changes**: Add --compat flag definition and extraction

```go
// Add after line 34
cmd.Flags().Bool("compat", false, "Enable compatibility mode (detailed text responses)")

// Add after line 60
compat, err := cmd.Flags().GetBool("compat")
if err != nil {
    return fmt.Errorf("failed to get compat flag: %w", err)
}

// Update line 65
srv, err := server.NewWithDebugAndCompat(debug, compat)
```

#### 2. server/server.go
**File**: `server/server.go`
**Changes**: Add compatMode field to Server struct and update constructors

```go
// Add to Server struct around line 25
type Server struct {
    mcpServer          *mcp.Server
    config             *config.Config
    remote             remote.ClientInterface
    repositoryResolver *RepositoryResolver
    compatMode         bool  // New field
}

// Add new constructor after line 47
func NewWithDebugAndCompat(debug, compat bool) (*Server, error) {
    cfg, err := config.Load()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    return NewFromConfigWithDebugAndCompat(cfg, debug, compat)
}

// Update NewFromConfigWithDebug around line 61
func NewFromConfigWithDebug(cfg *config.Config, debug bool) (*Server, error) {
    return NewFromConfigWithDebugAndCompat(cfg, debug, false)
}

// Add new constructor
func NewFromConfigWithDebugAndCompat(cfg *config.Config, debug, compat bool) (*Server, error) {
    // ... existing validation and client creation ...
    return NewFromServiceWithDebugAndCompat(client, cfg, debug, compat)
}

// Update NewFromServiceWithDebug around line 107
func NewFromServiceWithDebug(service remote.ClientInterface, cfg *config.Config, debug bool) (*Server, error) {
    return NewFromServiceWithDebugAndCompat(service, cfg, debug, false)
}

// Add new constructor
func NewFromServiceWithDebugAndCompat(service remote.ClientInterface, cfg *config.Config, debug, compat bool) (*Server, error) {
    s := &Server{
        config:             cfg,
        remote:             service,
        repositoryResolver: NewRepositoryResolver(),
        compatMode:         compat,  // Store the flag
    }
    // ... rest of existing constructor ...
}
```

### Success Criteria:

#### Automated Verification:
- [x] Server compiles with new constructors
- [x] CLI flag parsing works without errors
- [x] Server initializes correctly with both flag states

#### Manual Verification:
- [x] Help text shows --compat flag with description
- [x] Server starts with --compat flag
- [x] Server starts without --compat flag

---

## Phase 2: Response Helper Extraction

### Overview
Extract text formatting logic from each tool handler into separate helper functions that take the remote response and return formatted strings.

### Changes Required:

#### 1. server/response_formatters.go (new file)
**File**: `server/response_formatters.go`
**Changes**: Create helper functions for text formatting

```go
package server

import (
    "fmt"
    "strings"
)

// FormatIssueList creates a human-readable summary of issues
func FormatIssueList(issues []Issue) string {
    if len(issues) == 0 {
        return "No issues found"
    }
    var builder strings.Builder
    fmt.Fprintf(&builder, "Found %d issues:\n", len(issues))
    for _, issue := range issues {
        fmt.Fprintf(&builder, "- #%d: %s (%s)\n", issue.Number, issue.Title, issue.State)
    }
    return builder.String()
}

// FormatPullRequestList creates a human-readable summary of pull requests
func FormatPullRequestList(pullRequests []PullRequest) string {
    if len(pullRequests) == 0 {
        return "No pull requests found"
    }
    var builder strings.Builder
    fmt.Fprintf(&builder, "Found %d pull requests:\n", len(pullRequests))
    for _, pr := range pullRequests {
        fmt.Fprintf(&builder, "- #%d: %s (%s)\n", pr.Number, pr.Title, pr.State)
    }
    return builder.String()
}

// FormatPullRequestDetails creates detailed PR information
func FormatPullRequestDetails(pr PullRequest) string {
    var builder strings.Builder
    fmt.Fprintf(&builder, "Pull Request #%d: %s\n", pr.Number, pr.Title)
    fmt.Fprintf(&builder, "State: %s\n", pr.State)
    fmt.Fprintf(&builder, "Author: %s\n", pr.User)
    fmt.Fprintf(&builder, "Created: %s\n", pr.CreatedAt)
    fmt.Fprintf(&builder, "Updated: %s\n", pr.UpdatedAt)
    if pr.Body != "" {
        fmt.Fprintf(&builder, "Body:\n%s\n", pr.Body)
    }
    return builder.String()
}

// Add similar formatters for all other response types...
```

### Success Criteria:

#### Automated Verification:
- [x] New file compiles without errors
- [x] All formatter functions have correct signatures
- [x] Formatters handle empty results correctly

#### Manual Verification:
- [x] Formatters produce expected output format
- [x] Output matches current tool text responses

---

## Phase 3: Tool Handler Updates

### Overview
Update all 13+ tool handlers to use conditional response building based on the compatMode flag.

### Changes Required:

#### 1. server/issues.go
**File**: `server/issues.go`
**Changes**: Update handleIssueList to use conditional formatting

```go
// Around line 92, replace existing return
var responseText string
if s.compatMode {
    responseText = FormatIssueList(issues)
} else {
    responseText = fmt.Sprintf("Found %d issues", len(issues))
}

return TextResult(responseText), &IssueList{Issues: issues}, nil
```

#### 2. server/pr_list.go
**File**: `server/pr_list.go`
**Changes**: Update handlePullRequestList

```go
// Around line 106-117, replace existing logic
var responseText string
if s.compatMode {
    responseText = FormatPullRequestList(pullRequests)
} else {
    responseText = fmt.Sprintf("Found %d pull requests", len(pullRequests))
}

return TextResult(responseText), &PullRequestList{PullRequests: pullRequests}, nil
```

#### 3. server/pr_fetch.go
**File**: `server/pr_fetch.go`
**Changes**: Update handlePullRequestFetch

```go
// Around line 144, replace existing return
var responseText string
if s.compatMode {
    responseText = FormatPullRequestDetails(*pr)
} else {
    responseText = fmt.Sprintf("Pull request #%d: %s", pr.Number, pr.Title)
}

return TextResult(responseText), &PullRequestFetchResult{PullRequest: pr}, nil
```

#### 4. Update remaining handlers
Apply similar pattern to:
- `server/issue_create.go`
- `server/issue_edit.go`
- `server/issue_comment_create.go`
- `server/issue_comment_list.go`
- `server/issue_comment_edit.go`
- `server/pr_create.go`
- `server/pr_edit.go`
- `server/pr_comment_create.go`
- `server/pr_comment_list.go`
- `server/pr_comment_edit.go`

### Success Criteria:

#### Automated Verification:
- [x] All handlers compile without errors
- [x] Tools return structured data in both modes
- [x] No regression in error handling

#### Manual Verification:
- [x] With --compat: Detailed text responses match current behavior
- [x] Without --compat: Simple summary messages only
- [x] Structured data unchanged in both modes

---

## Phase 4: Testing

### Overview
Add comprehensive tests to verify both compatibility modes work correctly.

### Changes Required:

#### 1. server_test/compat_mode_test.go (new file)
**File**: `server_test/compat_mode_test.go`
**Changes**: Test compat flag behavior

```go
func TestCompatModeResponseFormat(t *testing.T) {
    ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
    t.Cleanup(cancel)

    mock := NewMockGiteaServer(t)
    mock.AddIssues("testuser", "testrepo", []MockIssue{
        {Index: 1, Title: "Test Issue", State: "open"},
    })

    // Test with compat mode enabled
    tsCompat := NewTestServerWithCompat(t, ctx, map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
        "FORGEJO_AUTH_TOKEN": "mock-token",
    }, true)
    if err := tsCompat.Initialize(); err != nil {
        t.Fatal(err)
    }

    resultCompat, err := tsCompat.Client().CallTool(ctx, &mcp.CallToolParams{
        Name: "issue_list",
        Arguments: map[string]any{
            "repository": "testuser/testrepo",
        },
    })
    if err != nil {
        t.Fatalf("Failed to call issue_list in compat mode: %v", err)
    }

    // Test with compat mode disabled
    tsRegular := NewTestServerWithCompat(t, ctx, map[string]string{
        "FORGEJO_REMOTE_URL": mock.URL(),
        "FORGEJO_AUTH_TOKEN": "mock-token",
    }, false)
    if err := tsRegular.Initialize(); err != nil {
        t.Fatal(err)
    }

    resultRegular, err := tsRegular.Client().CallTool(ctx, &mcp.CallToolParams{
        Name: "issue_list",
        Arguments: map[string]any{
            "repository": "testuser/testrepo",
        },
    })
    if err != nil {
        t.Fatalf("Failed to call issue_list in regular mode: %v", err)
    }

    // Verify text responses differ
    textCompat := GetTextContent(resultCompat.Content)
    textRegular := GetTextContent(resultRegular.Content)

    if !strings.Contains(textCompat, "Found 1 issues:") {
        t.Errorf("Compat mode should include detailed formatting, got: %s", textCompat)
    }

    if !strings.Contains(textRegular, "Found 1 issues") {
        t.Errorf("Regular mode should include simple summary, got: %s", textRegular)
    }

    if strings.Contains(textRegular, "#1: Test Issue (open)") {
        t.Errorf("Regular mode should not include detailed formatting")
    }

    // Verify structured data is identical
    structCompat := GetStructuredContent(resultCompat)
    structRegular := GetStructuredContent(resultRegular)
    // Compare structured data...
}
```

#### 2. server_test/harness.go
**File**: `server_test/harness.go`
**Changes**: Add NewTestServerWithCompat function

```go
// Add after NewTestServerWithDebug
func NewTestServerWithCompat(t *testing.T, ctx context.Context, env map[string]string, compat bool) *TestServer {
    // Similar to NewTestServerWithDebug but uses NewFromConfigWithDebugAndCompat
    // ...
}
```

#### 3. cmd/serve_test.go (new file)
**File**: `cmd/serve_test.go`
**Changes**: Test CLI flag parsing

```go
func TestServeCompatFlag(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected bool
    }{
        {"default", []string{"serve"}, false},
        {"compat enabled", []string{"serve", "--compat"}, true},
        {"compat disabled", []string{"serve", "--compat=false"}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewServeCmd()
            cmd.SetArgs(tt.args)
            
            err := cmd.ParseFlags(tt.args)
            if err != nil {
                t.Fatalf("Failed to parse flags: %v", err)
            }
            
            compat, err := cmd.Flags().GetBool("compat")
            if err != nil {
                t.Fatalf("Failed to get compat flag: %v", err)
            }
            
            if compat != tt.expected {
                t.Errorf("Expected compat=%v, got %v", tt.expected, compat)
            }
        })
    }
}
```

### Success Criteria:

#### Automated Verification:
- [x] All new tests pass
- [x] Existing tests continue to pass (some need updates for changed default behavior)
- [x] Coverage includes both compat modes
- [x] CLI flag parsing tests pass

#### Manual Verification:
- [x] Test output shows correct behavior differences
- [x] Integration tests work with both modes

---

## Testing Strategy

### Unit Tests:
- Test each formatter function with various inputs
- Test CLI flag parsing with different combinations
- Test server initialization with both flag states

### Integration Tests:
- End-to-end tool calls in both modes
- Verify structured data consistency
- Test with empty results

### Manual Testing Steps:
1. Start server with `--compat` flag and verify detailed responses
2. Start server without flag and verify simple responses
3. Test with various tools (issues, PRs, comments)
4. Verify error responses unchanged
5. Test hello tool unchanged

## Performance Considerations

- Minimal performance impact - simple boolean check per response
- Formatters only called when compat mode enabled
- No additional memory allocation in default mode

## Migration Notes

- Breaking change for clients relying on detailed text (default behavior changed)
- Users needing detailed text should add `--compat` flag
- No database or configuration migration required
- Version tags will communicate the breaking change

## References

- Original ticket: `thoughts/tickets/feature_serve_compat_flag.md`
- Related research: `thoughts/research/2025-10-14_serve_compat_flag_implementation.md`
- Debug flag pattern: `cmd/serve.go:34`, `server/server.go:40-47`
- Response helpers: `server/common.go:13-27`
- Tool handler pattern: `server/pr_list.go:106-117`