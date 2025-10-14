# PR Fetch Tool Implementation Plan

## Overview

Implement a new `pr_fetch` tool to retrieve comprehensive information about a single pull request by number. This tool will provide detailed metadata for PR validation and editing, supporting all PR states (open, closed, merged, draft) and handling both same-repository and fork-based PRs.

## Current State Analysis

The codebase has a well-established pattern for PR tools but lacks a `GetPullRequest` method in the remote client interface. Both Forgejo and Gitea SDKs support single PR retrieval with rich metadata, but this functionality isn't exposed through the standardized interface.

### Key Discoveries:
- The existing `PullRequest` struct (`remote/interface.go:123-135`) contains only basic fields
- SDKs provide extensive additional metadata: reviewers, approvals, mergeable status, file statistics, labels, milestones
- The codebase follows a pattern of backward-compatible evolution using optional fields
- All PR tools use consistent validation, repository resolution, and error handling patterns

## Desired End State

A fully functional `pr_fetch` tool that:
- Returns comprehensive PR metadata including all available fields from the SDKs
- Supports both repository and directory parameters with automatic resolution
- Handles all PR states and edge cases gracefully
- Integrates seamlessly with existing PR tools
- Includes comprehensive test coverage

### Key Discoveries:
- SDKs provide identical `GetPullRequest(owner, repo, index)` methods
- Response includes URLs, user details, merge status, labels, and file statistics
- Current interface can be extended without breaking changes using optional fields

## What We're NOT Doing

- Implementing caching (out of scope per ticket)
- Modifying existing PR tools
- Changing the core `PullRequest` struct (will use composition pattern)
- Implementing bulk PR fetching

## Implementation Approach

Following the established codebase patterns:
1. Extend the remote interface with a new `PullRequestGetter` interface
2. Implement the method in both Forgejo and Gitea clients
3. Create a new `PullRequestDetails` struct for comprehensive metadata
4. Implement server handler with validation and repository resolution
5. Register the tool following established patterns
6. Add comprehensive tests from the start

## Phase 1: Interface Extension and Client Implementation

### Overview
Add the `GetPullRequest` method to the remote interface and implement it in both clients. Create acceptance tests to validate the implementation.

### Changes Required:

#### 1. Remote Interface Extension
**File**: `remote/interface.go`
**Changes**: Add new interface and result struct after line 225

```go
// PullRequestGetter defines the interface for fetching a single pull request
type PullRequestGetter interface {
    GetPullRequest(ctx context.Context, repo string, number int) (*PullRequestDetails, error)
}

// PullRequestDetails represents comprehensive pull request information
type PullRequestDetails struct {
    // Basic fields (matching PullRequest for compatibility)
    ID        int               `json:"id"`
    Number    int               `json:"number"`
    Title     string            `json:"title"`
    Body      string            `json:"body"`
    State     string            `json:"state"`
    User      string            `json:"user"`
    CreatedAt string            `json:"created"`
    UpdatedAt string            `json:"updated"`
    Head      PullRequestBranch `json:"head"`
    Base      PullRequestBranch `json:"base"`
    
    // Additional metadata fields
    HTMLURL             string                 `json:"html_url"`
    DiffURL             string                 `json:"diff_url"`
    PatchURL            string                 `json:"patch_url"`
    Labels              []Label                `json:"labels,omitempty"`
    Milestone           *Milestone             `json:"milestone,omitempty"`
    Assignee            string                 `json:"assignee,omitempty"`
    Assignees           []string               `json:"assignees,omitempty"`
    Comments            int                    `json:"comments"`
    IsLocked            bool                   `json:"is_locked"`
    Mergeable           bool                   `json:"mergeable"`
    HasMerged           bool                   `json:"has_merged"`
    MergedAt            string                 `json:"merged_at,omitempty"`
    MergeCommitSHA      string                 `json:"merge_commit_sha,omitempty"`
    MergedBy            string                 `json:"merged_by,omitempty"`
    AllowMaintainerEdit bool                   `json:"allow_maintainer_edit"`
    ClosedAt            string                 `json:"closed_at,omitempty"`
    Deadline            string                 `json:"deadline,omitempty"`
}

// Label represents a repository label
type Label struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Color       string `json:"color"`
    Description string `json:"description,omitempty"`
}

// Milestone represents a repository milestone
type Milestone struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    State       string `json:"state"`
    OpenIssues  int    `json:"open_issues"`
    ClosedIssues int   `json:"closed_issues"`
}
```

#### 2. Update ClientInterface
**File**: `remote/interface.go`
**Changes**: Add `PullRequestGetter` to the ClientInterface composition at line 232

```go
// ClientInterface combines ... and PullRequestGetter for complete Git operations
type ClientInterface interface {
    IssueLister
    IssueCommenter
    IssueCommentLister
    IssueCommentEditor
    IssueCreator
    IssueAttachmentCreator
    IssueEditor
    PullRequestLister
    PullRequestCommentLister
    PullRequestCommenter
    PullRequestCommentEditor
    PullRequestEditor
    PullRequestCreator
    PullRequestGetter  // Add this line
    FileContentFetcher
}
```

#### 3. Forgejo Client Implementation
**File**: `remote/forgejo/pull_requests.go`
**Changes**: Add GetPullRequest method after line 495

```go
// GetPullRequest fetches a single pull request with full metadata
func (c *ForgejoClient) GetPullRequest(ctx context.Context, repo string, number int) (*remote.PullRequestDetails, error) {
    // Check if client is initialized
    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }

    // Parse repository string (format: "owner/repo")
    owner, repoName, ok := strings.Cut(repo, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
    }

    if number <= 0 {
        return nil, fmt.Errorf("invalid pull request number: %d, must be positive", number)
    }

    // Fetch pull request using Forgejo SDK
    forgejoPR, _, err := c.client.GetPullRequest(owner, repoName, int64(number))
    if err != nil {
        return nil, fmt.Errorf("failed to get pull request: %w", err)
    }

    // Convert to PullRequestDetails
    return c.convertToPullRequestDetails(forgejoPR), nil
}

// convertToPullRequestDetails converts Forgejo PR to our detailed format
func (c *ForgejoClient) convertToPullRequestDetails(fpr *forgejo.PullRequest) *remote.PullRequestDetails {
    // Extract user information
    user := "unknown"
    if fpr.Poster != nil {
        user = fpr.Poster.UserName
    }

    // Format timestamps
    createdAt := ""
    if fpr.Created != nil {
        createdAt = fpr.Created.Format("2006-01-02T15:04:05Z")
    }

    updatedAt := ""
    if fpr.Updated != nil {
        updatedAt = fpr.Updated.Format("2006-01-02T15:04:05Z")
    }

    mergedAt := ""
    if fpr.Merged != nil {
        mergedAt = fpr.Merged.Format("2006-01-02T15:04:05Z")
    }

    closedAt := ""
    if fpr.Closed != nil {
        closedAt = fpr.Closed.Format("2006-01-02T15:04:05Z")
    }

    deadline := ""
    if fpr.Deadline != nil {
        deadline = fpr.Deadline.Format("2006-01-02T15:04:05Z")
    }

    // Extract assignee information
    var assignee string
    if fpr.Assignee != nil {
        assignee = fpr.Assignee.UserName
    }

    var assignees []string
    for _, a := range fpr.Assignees {
        if a != nil {
            assignees = append(assignees, a.UserName)
        }
    }

    // Extract merged by information
    var mergedBy string
    if fpr.MergedBy != nil {
        mergedBy = fpr.MergedBy.UserName
    }

    // Extract merge commit SHA
    var mergeCommitSHA string
    if fpr.MergedCommitID != nil {
        mergeCommitSHA = *fpr.MergedCommitID
    }

    // Convert labels
    var labels []remote.Label
    for _, l := range fpr.Labels {
        if l != nil {
            labels = append(labels, remote.Label{
                ID:          int(l.ID),
                Name:        l.Name,
                Color:       l.Color,
                Description: l.Description,
            })
        }
    }

    // Convert milestone
    var milestone *remote.Milestone
    if fpr.Milestone != nil {
        milestone = &remote.Milestone{
            ID:           int(fpr.Milestone.ID),
            Title:        fpr.Milestone.Title,
            Description:  fpr.Milestone.Description,
            State:        string(fpr.Milestone.State),
            OpenIssues:   int(fpr.Milestone.OpenIssues),
            ClosedIssues: int(fpr.Milestone.ClosedIssues),
        }
    }

    // Convert head branch
    var head remote.PullRequestBranch
    if fpr.Head != nil {
        head = remote.PullRequestBranch{
            Ref: fpr.Head.Ref,
            Sha: fpr.Head.Sha,
        }
    }

    // Convert base branch
    var base remote.PullRequestBranch
    if fpr.Base != nil {
        base = remote.PullRequestBranch{
            Ref: fpr.Base.Ref,
            Sha: fpr.Base.Sha,
        }
    }

    return &remote.PullRequestDetails{
        ID:                  int(fpr.ID),
        Number:              int(fpr.Index),
        Title:               fpr.Title,
        Body:                fpr.Body,
        State:               string(fpr.State),
        User:                user,
        CreatedAt:           createdAt,
        UpdatedAt:           updatedAt,
        Head:                head,
        Base:                base,
        HTMLURL:             fpr.HTMLURL,
        DiffURL:             fpr.DiffURL,
        PatchURL:            fpr.PatchURL,
        Labels:              labels,
        Milestone:           milestone,
        Assignee:            assignee,
        Assignees:           assignees,
        Comments:            fpr.Comments,
        IsLocked:            fpr.IsLocked,
        Mergeable:           fpr.Mergeable,
        HasMerged:           fpr.HasMerged,
        MergedAt:            mergedAt,
        MergeCommitSHA:      mergeCommitSHA,
        MergedBy:            mergedBy,
        AllowMaintainerEdit: fpr.AllowMaintainerEdit,
        ClosedAt:            closedAt,
        Deadline:            deadline,
    }
}
```

#### 4. Gitea Client Implementation
**File**: `remote/gitea/gitea_client.go`
**Changes**: Add GetPullRequest method after line 690

```go
// GetPullRequest fetches a single pull request with full metadata
func (c *GiteaClient) GetPullRequest(ctx context.Context, repo string, number int) (*remote.PullRequestDetails, error) {
    // Check if client is initialized
    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }

    // Parse repository string (format: "owner/repo")
    owner, repoName, ok := strings.Cut(repo, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", repo)
    }

    if number <= 0 {
        return nil, fmt.Errorf("invalid pull request number: %d, must be positive", number)
    }

    // Fetch pull request using Gitea SDK
    giteaPR, _, err := c.client.GetPullRequest(owner, repoName, int64(number))
    if err != nil {
        return nil, fmt.Errorf("failed to get pull request: %w", err)
    }

    // Convert to PullRequestDetails
    return c.convertToPullRequestDetails(giteaPR), nil
}

// convertToPullRequestDetails converts Gitea PR to our detailed format
func (c *GiteaClient) convertToPullRequestDetails(gpr *gitea.PullRequest) *remote.PullRequestDetails {
    // Implementation identical to Forgejo version but using gitea types
    // ... (similar conversion logic as Forgejo)
}
```

#### 5. Acceptance Tests
**File**: `server_test/pr_fetch_test.go`
**Changes**: Create comprehensive test file

```go
package server_test

import (
    "context"
    "testing"
    
    "github.com/kunde21/forgejo-mcp/server"
    "github.com/kunde21/forgejo-mcp/remote"
)

func TestPullRequestFetch(t *testing.T) {
    // Test cases for various scenarios
    testCases := []struct {
        name           string
        args           server.PullRequestFetchArgs
        setupMock      func(*MockRemoteClient)
        wantErr        bool
        errContains    string
        validateResult func(*testing.T, *remote.PullRequestDetails)
    }{
        {
            name: "successful fetch with repository",
            args: server.PullRequestFetchArgs{
                Repository:        "owner/repo",
                PullRequestNumber: 123,
            },
            setupMock: func(m *MockRemoteClient) {
                m.GetPullRequestFunc = func(ctx context.Context, repo string, number int) (*remote.PullRequestDetails, error) {
                    return &remote.PullRequestDetails{
                        ID:      123,
                        Number:  123,
                        Title:   "Test PR",
                        State:   "open",
                        User:    "testuser",
                        HTMLURL: "https://example.com/repo/pull/123",
                    }, nil
                }
            },
            wantErr: false,
            validateResult: func(t *testing.T, pr *remote.PullRequestDetails) {
                if pr.ID != 123 {
                    t.Errorf("Expected ID 123, got %d", pr.ID)
                }
                if pr.Title != "Test PR" {
                    t.Errorf("Expected title 'Test PR', got '%s'", pr.Title)
                }
            },
        },
        // More test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Interface compiles without errors: `go build ./remote`
- [ ] Both client implementations compile: `go build ./remote/forgejo && go build ./remote/gitea`
- [ ] Acceptance tests compile: `go build ./server_test`
- [ ] Mock implementations follow existing patterns

#### Manual Verification:
- [ ] Interface methods match SDK signatures
- [ ] Conversion logic handles all SDK fields
- [ ] Error messages are consistent with existing patterns

---

## Phase 2: Server Handler Implementation

### Overview
Create the server handler following established patterns for validation, repository resolution, and error handling.

### Changes Required:

#### 1. Server Handler File
**File**: `server/pr_fetch.go`
**Changes**: Create new file with handler implementation

```go
package server

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    v "github.com/go-ozzo/ozzo-validation/v4"
    "github.com/kunde21/forgejo-mcp/remote"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestFetchArgs represents the arguments for fetching a pull request
type PullRequestFetchArgs struct {
    Repository        string `json:"repository,omitzero"` // Repository path in "owner/repo" format
    Directory         string `json:"directory,omitzero"`  // Local directory path for automatic resolution
    PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
}

// PullRequestFetchResult represents the result data for the pr_fetch tool
type PullRequestFetchResult struct {
    PullRequest *remote.PullRequestDetails `json:"pull_request,omitempty"`
}

// handlePullRequestFetch handles the "pr_fetch" tool request.
// It retrieves detailed information about a single pull request from a Forgejo/Gitea repository.
func (s *Server) handlePullRequestFetch(ctx context.Context, request *mcp.CallToolRequest, args PullRequestFetchArgs) (*mcp.CallToolResult, *PullRequestFetchResult, error) {
    // Validate context
    if ctx == nil {
        return TextError("Context is required"), nil, nil
    }

    // Validate input arguments using ozzo-validation
    if err := v.ValidateStruct(&args,
        v.Field(&args.Repository, v.When(args.Directory == "",
            v.Required.Error("at least one of directory or repository must be provided"),
            v.Match(repoReg).Error("repository must be in format 'owner/repo'"),
        )),
        v.Field(&args.Directory, v.When(args.Repository == "",
            v.Required.Error("at least one of directory or repository must be provided"),
            v.By(func(any) error {
                if !filepath.IsAbs(args.Directory) {
                    return v.NewError("abs_dir", "directory must be an absolute path")
                }
                stat, err := os.Stat(args.Directory)
                if err != nil {
                    return v.NewError("abs_dir", "invalid directory")
                }
                if !stat.IsDir() {
                    return v.NewError("abs_dir", "does not exist")
                }
                return nil
            }),
        )),
        v.Field(&args.PullRequestNumber, v.Required.Error("pull request number is required"), v.Min(1)),
    ); err != nil {
        return TextErrorf("Invalid request: %v", err), nil, nil
    }

    repository := args.Repository
    if args.Directory != "" {
        // Resolve directory to repository (takes precedence if both provided)
        resolution, err := s.repositoryResolver.ResolveRepository(args.Directory)
        if err != nil {
            return TextErrorf("Failed to resolve directory: %v", err), nil, nil
        }
        repository = resolution.Repository
    }

    // Check if remote client supports GetPullRequest
    prGetter, ok := s.remote.(remote.PullRequestGetter)
    if !ok {
        return TextError("Remote client does not support pull request fetching"), nil, nil
    }

    // Fetch the pull request
    pr, err := prGetter.GetPullRequest(ctx, repository, args.PullRequestNumber)
    if err != nil {
        return TextErrorf("Failed to fetch pull request: %v", err), nil, nil
    }

    // Format detailed response
    responseText := fmt.Sprintf("Pull Request #%d: %s\n", pr.Number, pr.Title)
    responseText += fmt.Sprintf("State: %s\n", pr.State)
    responseText += fmt.Sprintf("Author: %s\n", pr.User)
    responseText += fmt.Sprintf("Created: %s\n", pr.CreatedAt)
    responseText += fmt.Sprintf("Updated: %s\n", pr.UpdatedAt)
    
    if pr.Assignee != "" {
        responseText += fmt.Sprintf("Assignee: %s\n", pr.Assignee)
    }
    
    if len(pr.Assignees) > 0 {
        responseText += fmt.Sprintf("Assignees: %v\n", pr.Assignees)
    }
    
    if len(pr.Labels) > 0 {
        responseText += "Labels: "
        for i, label := range pr.Labels {
            if i > 0 {
                responseText += ", "
            }
            responseText += label.Name
        }
        responseText += "\n"
    }
    
    responseText += fmt.Sprintf("Comments: %d\n", pr.Comments)
    responseText += fmt.Sprintf("Mergeable: %t\n", pr.Mergeable)
    
    if pr.HasMerged {
        responseText += fmt.Sprintf("Merged: %s", pr.MergedAt)
        if pr.MergedBy != "" {
            responseText += fmt.Sprintf(" by %s", pr.MergedBy)
        }
        responseText += "\n"
    }
    
    responseText += fmt.Sprintf("URL: %s\n", pr.HTMLURL)

    return TextResult(responseText), &PullRequestFetchResult{PullRequest: pr}, nil
}
```

#### 2. Tool Registration
**File**: `server/server.go`
**Changes**: Add tool registration after line 196

```go
// Register pr_fetch tool
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_fetch",
    Description: "Fetch detailed information about a single pull request from a Forgejo/Gitea repository",
}, s.handlePullRequestFetch)
```

### Success Criteria:

#### Automated Verification:
- [ ] Handler compiles without errors: `go build ./server`
- [ ] Tool registration compiles: `go build ./server`
- [ ] Validation rules work correctly
- [ ] Repository resolution works as expected

#### Manual Verification:
- [ ] Handler follows established patterns
- [ ] Error messages are consistent with other tools
- [ ] Response format is comprehensive and readable

---

## Phase 3: Test Implementation and Integration

### Overview
Implement comprehensive tests and ensure full integration with the existing system.

### Changes Required:

#### 1. Complete Test Suite
**File**: `server_test/pr_fetch_test.go`
**Changes**: Complete the test implementation started in Phase 1

```go
// Add comprehensive test cases:
// - Repository parameter validation
// - Directory parameter validation
// - Repository resolution
// - PR number validation
// - Various PR states (open, closed, merged, draft)
// - Fork-based PRs
// - Error handling (invalid repo, not found, access denied)
// - Full metadata validation
// - Integration with mock server
```

#### 2. Mock Client Extension
**File**: `server_test/harness.go`
**Changes**: Extend MockRemoteClient to support GetPullRequest

```go
type MockRemoteClient struct {
    // Existing fields...
    GetPullRequestFunc func(ctx context.Context, repo string, number int) (*remote.PullRequestDetails, error)
}

func (m *MockRemoteClient) GetPullRequest(ctx context.Context, repo string, number int) (*remote.PullRequestDetails, error) {
    if m.GetPullRequestFunc != nil {
        return m.GetPullRequestFunc(ctx, repo, number)
    }
    return nil, fmt.Errorf("GetPullRequest not implemented")
}
```

#### 3. Integration Tests
**File**: `server_test/tool_discovery_test.go`
**Changes**: Add pr_fetch to tool discovery tests

```go
func TestPullRequestFetchToolDiscovery(t *testing.T) {
    // Test that pr_fetch tool is properly registered and discoverable
}
```

### Success Criteria:

#### Automated Verification:
- [ ] All unit tests pass: `go test -run TestPullRequestFetch ./server_test`
- [ ] Integration tests pass: `go test -run Integration ./...`
- [ ] Tool discovery tests pass: `go test -run TestToolDiscovery ./server_test`
- [ ] Test coverage is adequate (>90%)

#### Manual Verification:
- [ ] Tool appears in tool list
- [ ] All test scenarios pass
- [ ] Mock server integration works correctly
- [ ] Error scenarios are properly tested

---

## Testing Strategy

### Unit Tests:
- Argument validation with various inputs
- Repository resolution scenarios
- PR number validation
- Error handling for all failure modes
- Response formatting

### Integration Tests:
- End-to-end PR fetching with mock servers
- Directory-based repository resolution
- Fork detection and handling
- Different PR states and metadata

### Manual Testing Steps:
1. Test with open PR: Verify all fields are populated
2. Test with closed PR: Verify closed state and timestamps
3. Test with merged PR: Verify merge information
4. Test with draft PR: Verify draft status
5. Test with fork-based PR: Verify fork information
6. Test invalid PR number: Verify appropriate error
7. Test invalid repository: Verify appropriate error
8. Test directory resolution: Verify automatic detection

## Performance Considerations

- Single API call per fetch (no additional queries)
- Efficient conversion from SDK format to interface format
- Minimal memory allocation for response construction
- Timeout handling for network requests

## Migration Notes

- No breaking changes to existing interfaces
- New functionality is additive only
- Existing PR tools continue to work unchanged
- ClientInterface composition ensures backward compatibility

## References

- Original ticket: `thoughts/tickets/feature_pr_fetch.md`
- Research document: `thoughts/research/2025-10-14_pr_fetch_implementation.md`
- PR list implementation: `server/pr_list.go:45-118`
- PR edit implementation: `server/pr_edit.go:53-140`
- Remote interface: `remote/interface.go:123-248`
- Forgejo client: `remote/forgejo/pull_requests.go`
- Gitea client: `remote/gitea/gitea_client.go`
- Test patterns: `server_test/pr_list_test.go`