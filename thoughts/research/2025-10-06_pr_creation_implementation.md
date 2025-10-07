---
date: 2025-10-06T00:00:00Z
git_commit: c356fd630f008fb870059ae5251e9f4517c761f7
branch: mcp-go
repository: forgejo-mcp
topic: "PR Creation Tool Implementation Research"
tags: [research, codebase, pr, create, forgejo, gitea, mcp]
last_updated: 2025-10-06T00:00:00Z
---

## Ticket Synopsis

The ticket requests implementation of a PR creation tool (`forgejo_pr_create`) that allows creating pull requests from the current branch to a target branch. Key requirements include:
- Auto-detection of current branch and remote
- Support for both same-repo and fork-to-repo PRs
- PR template loading from `.gitea/PULL_REQUEST_TEMPLATE.md`
- Draft PR support with single reviewer assignment
- Branch validation (conflicts, behind status)
- Following existing MCP tool patterns

## Summary

The codebase has comprehensive PR management functionality (listing, editing, comments) but lacks PR creation capabilities. Implementation requires adding new API endpoints, git command execution utilities, template loading, and fork detection while following established patterns from issue creation and existing PR tools.

## Detailed Findings

### Existing PR Infrastructure

The codebase already has robust PR functionality in place:

**PR Data Model** (`remote/interface.go:123-135`)
- Core `PullRequest` struct with ID, Number, Title, Body, State, User, timestamps
- `PullRequestBranch` struct for head/base branch info (Ref, Sha)
- Consistent JSON serialization throughout

**PR Operations Available**
- `pr_list` - List PRs with state filtering and pagination (`server/pr_list.go`)
- `pr_edit` - Edit PR title, body, state, base branch (`server/pr_edit.go`)
- `pr_comment_*` - Create, edit, list PR comments (`server/pr_comments.go`)

**Remote Client Implementations**
- Forgejo: `remote/forgejo/pull_requests.go` - Complete SDK integration
- Gitea: `remote/gitea/gitea_client.go` - Parallel implementation
- Both follow identical patterns for consistency

### Tool Implementation Patterns

**Registration Pattern** (`server/server.go:104-162`)
```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_create",
    Description: "Create a new pull request...",
}, s.handlePullRequestCreate)
```

**Handler Signature Pattern**
All handlers follow: `func (s *Server) handleTool(ctx context.Context, request *mcp.CallToolRequest, args ToolArgs) (*mcp.CallToolResult, *ToolResult, error)`

**Argument Struct Pattern** (`server/pr_edit.go:14-23`)
```go
type PullRequestEditArgs struct {
    Repository        string `json:"repository,omitzero"`
    Directory         string `json:"directory,omitzero"`
    PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
    Title             string `json:"title,omitzero"`
    Body              string `json:"body,omitzero"`
    State             string `json:"state,omitzero"`
    BaseBranch        string `json:"base_branch,omitzero"`
}
```

**Validation Pattern** (`server/pr_edit.go:60-96`)
- Uses ozzo-validation with conditional rules
- Repository/directory mutual exclusivity
- Length validations with descriptive errors
- Custom validation functions for complex checks

**Repository Resolution Pattern** (`server/issues.go:137-145`)
- Directory takes precedence over repository
- Uses `repositoryResolver.ResolveRepository()` for git config parsing
- Supports HTTPS, SSH, and git protocol URLs

### Issue Creation Patterns to Follow

**Creation Flow** (`server/issues.go:107-192`)
1. Input validation with ozzo-validation
2. Repository resolution (directory → repository)
3. Attachment processing (if any)
4. Remote API call
5. Response formatting with structured result

**Remote Interface Pattern** (`remote/interface.go:72-100`)
- Separate `CreateIssueArgs` and `CreateIssueWithAttachmentsArgs` structs
- `IssueCreator` interface for basic creation
- Consistent error wrapping with context

**Response Pattern**
```go
responseText := fmt.Sprintf("Issue created successfully. Number: %d, Title: %s", issue.Number, issue.Title)
return TextResult(responseText), &IssueCreateResult{Issue: issue}, nil
```

### Missing Components for PR Creation

**1. Git Command Execution**
- Current codebase only parses `.git/config` files
- No git command execution utilities exist
- Need new `server/git_utils.go` for:
  - Current branch detection: `git rev-parse --abbrev-ref HEAD`
  - Branch existence: `git rev-parse --verify <branch>`
  - Commit comparison: `git rev-list --count <base>..<head>`
  - Conflict detection: `git merge-tree <base> <head> <base>`

**2. Template Loading**
- No existing template infrastructure
- Need to fetch `.gitea/PULL_REQUEST_TEMPLATE.md` from target branch
- Support fallback to `.github/PULL_REQUEST_TEMPLATE.md`
- Merge template with user-provided content

**3. Fork Detection**
- Current URL parsing (`server/repository_resolver.go:221-242`) doesn't detect forks
- Need to compare source repository with target repository
- Handle cross-repository branch references: `fork-owner:branch-name`

**4. Branch Validation**
- Validate source branch exists and has commits
- Detect conflicts with target branch
- Ensure source branch is not behind target
- Check repository protection rules

### API Implementation Requirements

**New Interface Methods** (`remote/interface.go`)
```go
type CreatePullRequestArgs struct {
    Repository string `json:"repository"`
    Head       string `json:"head"`       // Source branch
    Base       string `json:"base"`       // Target branch
    Title      string `json:"title"`
    Body       string `json:"body"`
    Draft      bool   `json:"draft"`
    Assignee   string `json:"assignee"`   // Single reviewer
}

type PullRequestCreator interface {
    CreatePullRequest(ctx context.Context, args CreatePullRequestArgs) (*PullRequest, error)
}
```

**Forgejo SDK Integration**
The Forgejo SDK provides `CreatePullRequestOption` with:
- Head, Base, Title, Body fields
- Assignee, Assignees, Milestone, Labels
- Deadline for due dates

**Fork-to-Repo Support**
- Head branch format: `fork-owner:branch-name` for cross-repository PRs
- Validate permissions for cross-repository operations
- Detect fork relationships from remote URLs

## Code References

### Core Files to Modify
- `server/server.go:115` - Add tool registration
- `server/pr_create.go` - New handler file (create)
- `remote/interface.go:210-225` - Add interface definitions
- `remote/forgejo/pull_requests.go` - Add Forgejo implementation
- `remote/gitea/gitea_client.go` - Add Gitea implementation
- `server/git_utils.go` - New git utilities file (create)

### Reference Implementations
- `server/issues.go:107-192` - Issue creation handler pattern
- `server/pr_edit.go:53-140` - PR editing handler pattern
- `server/pr_list.go:45-100` - PR listing validation pattern
- `server/repository_resolver.go:244-307` - Repository resolution logic

### Test Patterns
- `server_test/pr_edit_test.go` - Table-driven test structure
- `server_test/issue_list_test.go:23-48` - Git repo setup helpers
- `server_test/harness.go` - Mock server patterns

## Architecture Insights

### Layered Architecture
1. **MCP Protocol Layer** (`server/`) - Tool registration, validation, response formatting
2. **Business Logic Layer** (`remote/interface.go`) - Domain models and service interfaces
3. **SDK Integration Layer** (`remote/forgejo/`, `remote/gitea/`) - Platform-specific implementations

### Design Patterns
- **Interface Segregation**: Separate interfaces for different operations
- **Repository Pattern**: Consistent repository resolution across tools
- **Validation Chain**: Structured validation with descriptive errors
- **Error Wrapping**: Context preservation through error chains

### Configuration Management
- Repository format validation via regex (`server/common.go:30`)
- Default values for pagination (limit=15, max=100)
- State defaults (open/closed/all)

## Historical Context (from thoughts/)

### Related Implementations
- `thoughts/reviews/issue-edit-tool-implementation-review.md` - References PR edit patterns
- `thoughts/research/2025-10-05_issue_creation_implementation.md` - Issue creation research
- `thoughts/research/2025-10-06_issue_edit_implementation.md` - Issue edit patterns

### Planning Documents
- `thoughts/tickets/feature_pr_create.md` - Complete requirements and research context
- `pr_edit_tool.md` - PR edit implementation documentation

## Open Questions

1. **Git Command Execution**: Should we use `os/exec` directly or a git library?
2. **Template Variables**: Should we support variable substitution in PR templates?
3. **Branch Detection Limits**: How to handle detached HEAD or no remote scenarios?
4. **Fork Detection**: Should we auto-detect forks or require explicit specification?
5. **Conflict Details**: How much detail to provide in conflict error messages?

## Implementation Priority

1. **High**: Basic PR creation (same repo, no template)
2. **High**: Git utilities for branch detection
3. **Medium**: Template loading support
4. **Medium**: Fork-to-repo PR creation
5. **Low**: Advanced conflict detection details