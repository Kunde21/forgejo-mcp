# Issue Edit Tool Implementation Plan

## Overview

Implement a new MCP tool `forgejo_issue_edit` that enables editing existing issues on Forgejo/Gitea repositories. The tool will support updating issue titles, bodies, and states through the MCP interface, following the established patterns from pull request editing.

## Current State Analysis

The codebase has a well-established pattern for MCP tools but lacks issue editing capability:
- Issue struct in `remote/interface.go:8-12` only contains Number, Title, and State
- No EditIssueArgs or IssueEditor interface exists
- Remote clients lack EditIssue methods
- Pull request editing provides a complete template to follow

## Desired End State

A fully functional `forgejo_issue_edit` tool that:
- Accepts optional title, body, and state parameters
- Performs partial updates (only updates provided fields)
- Validates input using ozzo-validation
- Resolves repositories from directory paths
- Returns the updated issue object
- Has comprehensive test coverage

### Key Discoveries:
- PR edit pattern in `server/pr_edit.go:14-140` provides exact implementation template
- Forgejo SDK's EditIssueOption supports all required fields
- Issue struct needs Body, Updated, and Created fields for complete representation
- Testing infrastructure exists with mock servers and table-driven tests

## What We're NOT Doing

- Not implementing labels/assignees editing in this iteration (deferred for future)
- Not supporting milestone editing
- Not adding optimistic locking for concurrent edits
- Not implementing issue attachments in edit operation

## Implementation Approach

Follow the PR edit implementation pattern exactly:
1. Extend Issue struct with missing fields
2. Add EditIssueArgs with optional fields using `omitzero`
3. Implement IssueEditor interface in both clients
4. Use partial update pattern with `hasChanges` detection
5. Apply same validation rules as issue creation
6. Return updated Issue object with all fields populated

## Phase 1: Extend Data Structures and Interfaces

### Overview
Add missing fields to Issue struct and define the interfaces needed for issue editing.

### Changes Required:

#### 1. remote/interface.go
**File**: `remote/interface.go:8-12`
**Changes**: Extend Issue struct and add EditIssueArgs/IssueEditor interface

```go
// Issue represents a Git repository issue
type Issue struct {
    Number  int    `json:"number"`
    Title   string `json:"title"`
    State   string `json:"state"`
    Body    string `json:"body,omitempty"`
    Updated string `json:"updated,omitempty"`
    Created string `json:"created,omitempty"`
}

// EditIssueArgs represents the arguments for editing an issue
type EditIssueArgs struct {
    Repository   string `json:"repository"`
    Directory    string `json:"directory"`
    IssueNumber  int    `json:"issue_number"`
    Title        string `json:"title"`
    Body         string `json:"body"`
    State        string `json:"state"`
}

// IssueEditor defines the interface for editing issues in Git repositories
type IssueEditor interface {
    EditIssue(ctx context.Context, args EditIssueArgs) (*Issue, error)
}
```

#### 2. remote/interface.go
**File**: `remote/interface.go:191-204`
**Changes**: Add IssueEditor to ClientInterface

```go
// ClientInterface combines ... and IssueEditor for complete Git operations
type ClientInterface interface {
    IssueLister
    IssueCommenter
    IssueCommentLister
    IssueCommentEditor
    IssueCreator
    IssueAttachmentCreator
    IssueEditor  // Add this line
    PullRequestLister
    PullRequestCommentLister
    PullRequestCommenter
    PullRequestCommentEditor
    PullRequestEditor
}
```

### Success Criteria:

#### Automated Verification:
- [ ] go build ./remote passes without errors
- [ ] Interface compilation succeeds
- [ ] No breaking changes to existing code

#### Manual Verification:
- [ ] Issue struct includes new fields in JSON output
- [ ] Interface definitions are correctly exported
- [ ] ClientInterface includes IssueEditor

---

## Phase 2: Implement Server Handler

### Overview
Add the handleIssueEdit function with validation, repository resolution, and error handling.

### Changes Required:

#### 1. server/issues.go
**File**: `server/issues.go` (add at end of file)
**Changes**: Add IssueEditArgs struct and handler function

```go
// IssueEditArgs represents the arguments for editing an issue with validation tags
type IssueEditArgs struct {
    Repository   string `json:"repository,omitzero"` // Repository path in "owner/repo" format
    Directory    string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
    IssueNumber  int    `json:"issue_number" validate:"required,min=1"`
    Title        string `json:"title,omitzero"`       // New title for the issue
    Body         string `json:"body,omitzero"`        // New description/body for the issue
    State        string `json:"state,omitzero"`       // New state ("open" or "closed")
}

// IssueEditResult represents the result data for the issue_edit tool
type IssueEditResult struct {
    Issue *remote.Issue `json:"issue,omitempty"`
}

// handleIssueEdit handles the "issue_edit" tool request.
// It edits an existing issue in a Forgejo/Gitea repository.
func (s *Server) handleIssueEdit(ctx context.Context, request *mcp.CallToolRequest, args IssueEditArgs) (*mcp.CallToolResult, *IssueEditResult, error) {
    // Validate context - required for proper request handling
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
        v.Field(&args.IssueNumber, v.Required.Error("must be no less than 1"), v.Min(1)),
        v.Field(&args.State, v.When(args.State != "",
            v.In("open", "closed").Error("state must be 'open' or 'closed'"),
        )),
        v.Field(&args.Title, v.When(args.Title != "",
            v.Length(1, 255).Error("title must be between 1 and 255 characters"),
        )),
        v.Field(&args.Body, v.When(args.Body != "",
            v.Length(1, 65535).Error("body must be between 1 and 65535 characters"),
        )),
    ); err != nil {
        return TextErrorf("Invalid request: %v", err), nil, nil
    }

    // Ensure at least one field is being changed
    if args.Title == "" && args.Body == "" && args.State == "" {
        return TextError("At least one of title, body, or state must be provided"), nil, nil
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

    // Edit the issue using the service layer
    editArgs := remote.EditIssueArgs{
        Repository:  repository,
        IssueNumber: args.IssueNumber,
        Title:       args.Title,
        Body:        args.Body,
        State:       args.State,
    }
    issue, err := s.remote.EditIssue(ctx, editArgs)
    if err != nil {
        return TextErrorf("Failed to edit issue: %v", err), nil, nil
    }

    // Format success response with updated issue metadata
    responseText := fmt.Sprintf("Issue edited successfully. Number: %d, Title: %s, State: %s",
        issue.Number, issue.Title, issue.State)
    if issue.Updated != "" {
        responseText += fmt.Sprintf(", Updated: %s", issue.Updated)
    }
    responseText += "\n"

    if issue.Body != "" {
        responseText += fmt.Sprintf("Body: %s\n", issue.Body)
    }

    return TextResult(responseText), &IssueEditResult{Issue: issue}, nil
}
```

#### 2. server_test/issue_edit_test.go
**File**: `server_test/issue_edit_test.go` (new file)
**Changes**: Create comprehensive test suite

```go
package server_test

import (
    "testing"
    
    "github.com/kunde21/forgejo-mcp/server"
    "github.com/kunde21/forgejo-mcp/server_test"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/stretchr/testify/assert"
)

func TestIssueEdit(t *testing.T) {
    testCases := []struct {
        name      string
        setupMock func(*server_test.MockGiteaServer)
        arguments map[string]any
        expect    *mcp.CallToolResult
    }{
        {
            name: "edit issue title",
            setupMock: func(mock *server_test.MockGiteaServer) {
                mock.AddIssue("testuser", "testrepo", server_test.MockIssue{
                    Index: 1,
                    Title: "Original Title",
                    Body:  "Original body",
                    State: "open",
                })
                mock.EditIssue("testuser", "testrepo", 1, map[string]interface{}{
                    "title": "New Title",
                })
            },
            arguments: map[string]any{
                "repository":    "testuser/testrepo",
                "issue_number":  1,
                "title":         "New Title",
            },
            expect: &mcp.CallToolResult{
                Content: []mcp.Content{
                    &mcp.TextContent{Text: "Issue edited successfully. Number: 1, Title: New Title, State: open\nBody: Original body\n"},
                },
                IsError: false,
            },
        },
        // Add more test cases for body, state, directory resolution, validation errors
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation following pr_edit_test.go pattern
        })
    }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] go build ./server passes without errors
- [ ] Unit tests for handleIssueEdit pass
- [ ] Validation tests cover all scenarios
- [ ] Mock server tests succeed

#### Manual Verification:
- [ ] Handler validates all input correctly
- [ ] Repository resolution works from directory
- [ ] Error messages are clear and helpful
- [ ] Success response includes all issue fields

---

## Phase 3: Implement Remote Clients

### Overview
Add EditIssue methods to both Forgejo and Gitea clients following the partial update pattern.

### Changes Required:

#### 1. remote/forgejo/issues.go
**File**: `remote/forgejo/issues.go` (add at end of file)
**Changes**: Add EditIssue method

```go
// EditIssue edits an existing issue in the specified repository
func (c *ForgejoClient) EditIssue(ctx context.Context, args remote.EditIssueArgs) (*remote.Issue, error) {
    // Check if client is initialized
    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }

    // Parse repository string (format: "owner/repo")
    owner, repoName, ok := strings.Cut(args.Repository, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
    }

    if args.IssueNumber <= 0 {
        return nil, fmt.Errorf("invalid issue number: %d, must be positive", args.IssueNumber)
    }

    // Prepare edit options - only include fields that are provided
    var editOptions forgejo.EditIssueOption
    hasChanges := false

    if args.Title != "" {
        editOptions.Title = args.Title
        hasChanges = true
    }

    if args.Body != "" {
        editOptions.Body = &args.Body
        hasChanges = true
    }

    if args.State != "" {
        // Convert state to Forgejo SDK format
        var state forgejo.StateType
        switch args.State {
        case "open":
            state = forgejo.StateOpen
        case "closed":
            state = forgejo.StateClosed
        default:
            return nil, fmt.Errorf("invalid state: %s, must be 'open' or 'closed'", args.State)
        }
        editOptions.State = &state
        hasChanges = true
    }

    if !hasChanges {
        return nil, fmt.Errorf("no changes specified")
    }

    // Edit the issue using Forgejo SDK
    forgejoIssue, _, err := c.client.EditIssue(owner, repoName, int64(args.IssueNumber), editOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to edit issue: %w", err)
    }

    // Convert to our Issue struct
    issue := &remote.Issue{
        Number:  int(forgejoIssue.Index),
        Title:   forgejoIssue.Title,
        State:   string(forgejoIssue.State),
        Body:    forgejoIssue.Body,
        Updated: forgejoIssue.Updated.Format("2006-01-02T15:04:05Z07:00"),
        Created: forgejoIssue.Created.Format("2006-01-02T15:04:05Z07:00"),
    }

    return issue, nil
}
```

#### 2. remote/gitea/gitea_client.go
**File**: `remote/gitea/gitea_client.go` (add at end of file)
**Changes**: Add EditIssue method following same pattern as Forgejo

```go
// EditIssue edits an existing issue in the specified repository
func (c *GiteaClient) EditIssue(ctx context.Context, args remote.EditIssueArgs) (*remote.Issue, error) {
    // Implementation similar to Forgejo client but using Gitea SDK
    // Follow same pattern with hasChanges detection and validation
}
```

#### 3. remote/forgejo/forgejo_client_test.go
**File**: `remote/forgejo/forgejo_client_test.go` (add tests)
**Changes**: Add tests for EditIssue method

```go
func TestEditIssue(t *testing.T) {
    // Test implementation following existing patterns
    // Test partial updates, validation, error handling
}
```

### Success Criteria:

#### Automated Verification:
- [ ] go build ./remote/forgejo passes
- [ ] go build ./remote/gitea passes
- [ ] Client tests pass for both implementations
- [ ] Integration tests with mock servers succeed

#### Manual Verification:
- [ ] Partial updates work correctly
- [ ] Only provided fields are updated
- [ ] State conversion works properly
- [ ] Error handling is consistent

---

## Phase 4: Register Tool and Final Integration

### Overview
Register the new tool with the MCP server and ensure full integration.

### Changes Required:

#### 1. server/server.go
**File**: `server/server.go:114-131`
**Changes**: Register the issue_edit tool

```go
// Add after existing issue tool registrations
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "issue_edit",
    Description: "Edit an existing issue on a Forgejo/Gitea repository",
}, s.handleIssueEdit)
```

#### 2. server_test/tool_discovery_test.go
**File**: `server_test/tool_discovery_test.go`
**Changes**: Add test for tool registration

```go
func TestIssueEditToolRegistration(t *testing.T) {
    // Verify tool is properly registered and discoverable
}
```

#### 3. Integration Tests
**File**: `server_test/issue_edit_integration_test.go` (new file)
**Changes**: Add end-to-end integration tests

```go
func TestIssueEditIntegration(t *testing.T) {
    // Test complete workflow with real repositories
    // Test directory resolution
    // Test concurrent access scenarios
}
```

### Success Criteria:

#### Automated Verification:
- [ ] go test ./... passes with new tests
- [ ] Tool appears in server's tool list
- [ ] Integration tests cover all scenarios
- [ ] No regressions in existing functionality

#### Manual Verification:
- [ ] Tool can be called via MCP client
- [ ] Directory resolution works in practice
- [ ] Issue updates persist correctly
- [ ] Performance is acceptable

---

## Testing Strategy

### Unit Tests:
- Test all validation scenarios (missing fields, invalid formats)
- Test partial update combinations
- Test repository resolution from directory
- Test error handling for API failures

### Integration Tests:
- End-to-end issue editing workflow
- Directory parameter functionality
- State transitions (open ↔ closed)
- Markdown content preservation

### Manual Testing Steps:
1. Create an issue via MCP
2. Edit the issue title only
3. Edit the issue body with markdown
4. Change issue state from open to closed
5. Verify partial updates don't affect other fields
6. Test with directory parameter instead of repository
7. Test error cases (non-existent issue, invalid repository)

## Performance Considerations

- Partial updates minimize API calls
- Repository resolution is cached by RepositoryResolver
- No additional memory overhead beyond existing patterns
- Response size remains consistent with other issue tools

## Migration Notes

- No breaking changes to existing API
- Backward compatible with all existing tools
- New fields in Issue struct are optional with `omitempty`
- ClientInterface addition is additive

## References

- Original ticket: `thoughts/tickets/feature_issue_edit.md`
- Related research: `thoughts/research/2025-10-06_issue_edit_implementation.md`
- PR edit pattern: `server/pr_edit.go:14-140`
- Issue creation pattern: `server/issues.go:107-193`
- Remote client pattern: `remote/forgejo/pull_requests.go:310-416`