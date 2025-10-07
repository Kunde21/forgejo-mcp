# PR Creation Tool Implementation Plan

## Overview

Implement a new `forgejo_pr_create` MCP tool that allows users to create pull requests from branches to target branches with auto-detection, template support, and fork handling. This implementation follows a hybrid approach with 3 phases to deliver core value quickly while managing complexity.

## Current State Analysis

The codebase has comprehensive PR management functionality (listing, editing, comments) but lacks PR creation capabilities. Existing patterns include:
- Tool registration using MCP SDK v0.4.0 in `server/server.go:104-162`
- Validation with ozzo-validation following conditional rules
- Repository resolution with directory → repository precedence
- Remote client interfaces in `remote/interface.go` with Forgejo/Gitea implementations
- No git command execution utilities exist
- No template loading infrastructure

## Desired End State

A fully functional `forgejo_pr_create` tool that can:
- Create PRs from current or specified branch to target branch
- Auto-detect current git branch and remote configuration
- Support both same-repo and fork-to-repo PR creation
- Load and use PR templates when available
- Support draft PRs and single reviewer assignment
- Validate branch status and report conflicts with file details
- Return structured errors for user action

### Key Discoveries:
- Forgejo SDK `CreatePullRequestOption` lacks dedicated draft field - needs title prefix approach
- Fork detection requires enhanced repository resolution beyond current URL parsing
- Git utilities must be built from scratch using `os/exec`
- Template loading can leverage Forgejo SDK `GetFile` method
- Conflict detection requires `git merge-tree` command execution

## What We're NOT Doing

- Multiple reviewer assignment (single reviewer only)
- Attachment support in PR creation
- Preview/dry-run mode
- Local git operations (branch must be pushed)
- PR title deduplication
- Variable substitution in PR templates
- Issue references in PR titles (only descriptions)

## Implementation Approach

Hybrid 3-phase approach to balance speed of delivery with comprehensive functionality:
- **Phase 1**: Core PR creation with git utilities for immediate value
- **Phase 2**: Fork detection and template loading for enhanced workflows  
- **Phase 3**: Advanced conflict detection for robust error handling

## Phase 1: Core PR Creation + Git Utilities

### Overview
Implement basic same-repository PR creation with essential git command utilities for branch detection and validation.

### Changes Required:

#### 1. Git Utilities Infrastructure
**File**: `server/git_utils.go` (create)
**Changes**: New git command execution utilities

```go
// GetCurrentBranch returns the current git branch name
func GetCurrentBranch(directory string) (string, error)

// BranchExists checks if a branch exists locally
func BranchExists(directory, branch string) (bool, error)

// GetCommitCount returns number of commits between base and head
func GetCommitCount(directory, base, head string) (int, error)

// HasConflicts detects if merging base into head would cause conflicts
func HasConflicts(directory, base, head string) (bool, []string, error)

// IsBranchBehind checks if head branch is behind base branch
func IsBranchBehind(directory, base, head string) (bool, error)
```

#### 2. Remote Interface Extension
**File**: `remote/interface.go:210-225`
**Changes**: Add PR creation interface definitions

```go
// CreatePullRequestArgs represents arguments for creating a pull request
type CreatePullRequestArgs struct {
    Repository string `json:"repository"`
    Head       string `json:"head"`       // Source branch
    Base       string `json:"base"`       // Target branch
    Title      string `json:"title"`
    Body       string `json:"body"`
    Draft      bool   `json:"draft"`
    Assignee   string `json:"assignee"`   // Single reviewer
}

// PullRequestCreator defines the interface for creating pull requests
type PullRequestCreator interface {
    CreatePullRequest(ctx context.Context, args CreatePullRequestArgs) (*PullRequest, error)
}

// Add PullRequestCreator to ClientInterface
type ClientInterface interface {
    // ... existing interfaces
    PullRequestCreator
}
```

#### 3. Forgejo Client Implementation
**File**: `remote/forgejo/pull_requests.go`
**Changes**: Add CreatePullRequest method

```go
// CreatePullRequest creates a new pull request in the repository
func (c *ForgejoClient) CreatePullRequest(ctx context.Context, args remote.CreatePullRequestArgs) (*remote.PullRequest, error) {
    owner, repoName, ok := strings.Cut(args.Repository, "/")
    if !ok {
        return nil, fmt.Errorf("invalid repository format: %s, expected 'owner/repo'", args.Repository)
    }

    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }

    // Handle draft PRs with title prefix since SDK lacks draft field
    title := args.Title
    if args.Draft {
        title = "[DRAFT] " + title
    }

    opts := forgejo.CreatePullRequestOption{
        Head:     args.Head,
        Base:     args.Base,
        Title:    title,
        Body:     args.Body,
        Assignee: args.Assignee,
    }

    fpr, _, err := c.client.CreatePullRequest(owner, repoName, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to create pull request: %w", err)
    }

    // Transform Forgejo SDK response to remote interface format
    return transformPullRequest(fpr), nil
}
```

#### 4. Gitea Client Implementation
**File**: `remote/gitea/gitea_client.go`
**Changes**: Add CreatePullRequest method (parallel to Forgejo)

```go
// CreatePullRequest creates a new pull request in the repository
func (c *GiteaClient) CreatePullRequest(ctx context.Context, args remote.CreatePullRequestArgs) (*remote.PullRequest, error) {
    // Similar implementation using Gitea SDK
}
```

#### 5. Server Handler Implementation
**File**: `server/pr_create.go` (create)
**Changes**: New MCP tool handler

```go
// PullRequestCreateArgs represents the arguments for creating a pull request
type PullRequestCreateArgs struct {
    Repository string `json:"repository,omitzero"`     // Repository path in "owner/repo" format
    Directory  string `json:"directory,omitzero"`      // Local directory path for auto-resolution
    Head       string `json:"head,omitzero"`           // Source branch (auto-detected if not provided)
    Base       string `json:"base,omitzero"`           // Target branch (default if not provided)
    Title      string `json:"title" validate:"required"` // PR title
    Body       string `json:"body,omitzero"`            // PR description
    Draft      bool   `json:"draft,omitzero"`          // Create as draft PR
    Assignee   string `json:"assignee,omitzero"`       // Single reviewer
}

// PullRequestCreateResult represents the result data for the pr_create tool
type PullRequestCreateResult struct {
    PullRequest *remote.PullRequest `json:"pull_request,omitempty"`
}

func (s *Server) handlePullRequestCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCreateArgs) (*mcp.CallToolResult, *PullRequestCreateResult, error) {
    // Implementation with validation, git operations, and API call
}
```

#### 6. Tool Registration
**File**: `server/server.go:162`
**Changes**: Add tool registration

```go
mcp.AddTool(mcpServer, &mcp.Tool{
    Name:        "pr_create",
    Description: "Create a new pull request in a Forgejo/Gitea repository",
}, s.handlePullRequestCreate)
```

### Success Criteria:

#### Automated Verification:
- [ ] Tool creates PRs successfully from same repo: `go test ./server_test -run TestPullRequestCreate`
- [ ] Branch detection works correctly: `go test ./server_test -run TestGitUtils`
- [ ] Input validation passes: `go test ./server_test -run TestPullRequestCreateValidation`
- [ ] All unit tests pass: `go test ./...`
- [ ] Type checking passes: `go vet ./...`

#### Manual Verification:
- [ ] PR appears on repository with correct title and body
- [ ] PR is created from correct source to target branch
- [ ] Draft PRs have "[DRAFT]" prefix in title
- [ ] Reviewer assignment works when specified
- [ ] Auto-detection works when directory parameter used

---

## Phase 2: Fork Detection + Template Loading

### Overview
Enhance repository resolution to detect fork relationships and add template loading from `.gitea/PULL_REQUEST_TEMPLATE.md`.

### Changes Required:

#### 1. Enhanced Repository Resolution
**File**: `server/repository_resolver.go`
**Changes**: Add fork detection capabilities

```go
// ForkInfo contains information about fork relationships
type ForkInfo struct {
    IsFork        bool   `json:"is_fork"`
    ForkOwner     string `json:"fork_owner,omitempty"`
    OriginalOwner string `json:"original_owner,omitempty"`
    ForkRemote    string `json:"fork_remote,omitempty"`
}

// ResolveWithForkInfo performs repository resolution with fork detection
func (r *RepositoryResolver) ResolveWithForkInfo(directory string) (*RepositoryResolution, *ForkInfo, error)

// ExtractAllRemotes extracts all remote configurations from git config
func (r *RepositoryResolver) ExtractAllRemotes(directory string) (map[string]string, error)

// DetectForkRelationship analyzes remotes to detect fork relationships
func (r *RepositoryResolver) DetectForkRelationship(remotes map[string]string, targetRepo string) (*ForkInfo, error)
```

#### 2. Template Loading Interface
**File**: `remote/interface.go`
**Changes**: Add file content fetching interface

```go
// FileContentFetcher defines interface for fetching repository file contents
type FileContentFetcher interface {
    GetFileContent(ctx context.Context, owner, repo, ref, filepath string) ([]byte, error)
}

// Add FileContentFetcher to ClientInterface
type ClientInterface interface {
    // ... existing interfaces
    FileContentFetcher
}
```

#### 3. Template Loading Implementation
**File**: `server/template_loader.go` (create)
**Changes**: Template loading utilities

```go
// LoadPRTemplate attempts to load PR template from repository
func LoadPRTemplate(ctx context.Context, client remote.FileContentFetcher, owner, repo, branch string) (string, error)

// MergeTemplateContent merges template with user-provided content
func MergeTemplateContent(template, userContent string) string
```

#### 4. Enhanced PR Creation Handler
**File**: `server/pr_create.go`
**Changes**: Update handler to use fork detection and templates

```go
// Enhanced handler with fork detection and template loading
func (s *Server) handlePullRequestCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCreateArgs) (*mcp.CallToolResult, *PullRequestCreateResult, error) {
    // ... existing validation ...
    
    // Enhanced repository resolution with fork detection
    var repository string
    var forkInfo *server.ForkInfo
    if args.Directory != "" {
        resolution, err := s.repositoryResolver.ResolveWithForkInfo(args.Directory)
        if err != nil {
            return TextErrorf("Failed to resolve directory: %v", err), nil, nil
        }
        repository = resolution.Repository
        forkInfo = resolution.ForkInfo
    } else {
        repository = args.Repository
    }
    
    // Auto-detect current branch if not provided
    head := args.Head
    if head == "" && args.Directory != "" {
        detectedHead, err := git.GetCurrentBranch(args.Directory)
        if err != nil {
            return TextErrorf("Failed to detect current branch: %v", err), nil, nil
        }
        head = detectedHead
    }
    
    // Format head branch for fork-to-repo PRs
    if forkInfo != nil && forkInfo.IsFork {
        head = fmt.Sprintf("%s:%s", forkInfo.ForkOwner, head)
    }
    
    // Load template if no body provided
    body := args.Body
    if body == "" {
        template, err := LoadPRTemplate(ctx, s.remote, owner, repo, base)
        if err == nil && template != "" {
            body = MergeTemplateContent(template, body)
        }
    }
    
    // ... rest of implementation ...
}
```

#### 5. Remote Client File Content Support
**File**: `remote/forgejo/forgejo_client.go`
**Changes**: Add GetFileContent method

```go
// GetFileContent fetches file content from repository
func (c *ForgejoClient) GetFileContent(ctx context.Context, owner, repo, ref, filepath string) ([]byte, error) {
    content, _, err := c.client.GetFile(owner, repo, ref, filepath)
    return content, err
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Fork detection works correctly: `go test ./server_test -run TestForkDetection`
- [ ] Template loading works when template exists: `go test ./server_test -run TestTemplateLoading`
- [ ] Cross-repository PR creation works: `go test ./server_test -run TestCrossRepositoryPR`
- [ ] All existing tests still pass: `go test ./...`

#### Manual Verification:
- [ ] Fork-to-repo PRs created with correct head format
- [ ] Template content used when no body provided
- [ ] Template merged correctly with user content
- [ ] Auto-detection works for forked repositories
- [ ] Fallback to same-repo when no fork detected

---

## Phase 3: Advanced Conflict Detection

### Overview
Implement detailed conflict reporting with file lists and enhanced branch status validation.

### Changes Required:

#### 1. Enhanced Git Utilities
**File**: `server/git_utils.go`
**Changes**: Add detailed conflict detection

```go
// GetConflictFiles returns detailed list of conflicting files
func GetConflictFiles(directory, base, head string) ([]string, error)

// GetBranchStatus returns comprehensive branch status information
type BranchStatus struct {
    Exists       bool     `json:"exists"`
    HasCommits   bool     `json:"has_commits"`
    CommitCount  int      `json:"commit_count"`
    IsBehind     bool     `json:"is_behind"`
    HasConflicts bool     `json:"has_conflicts"`
    ConflictFiles []string `json:"conflict_files,omitempty"`
}

func GetBranchStatus(directory, base, head string) (*BranchStatus, error)
```

#### 2. Enhanced Validation in Handler
**File**: `server/pr_create.go`
**Changes**: Add comprehensive branch validation

```go
// Enhanced validation with detailed error reporting
func (s *Server) handlePullRequestCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCreateArgs) (*mcp.CallToolResult, *PullRequestCreateResult, error) {
    // ... existing validation ...
    
    // Comprehensive branch validation
    if args.Directory != "" {
        status, err := git.GetBranchStatus(args.Directory, base, head)
        if err != nil {
            return TextErrorf("Failed to validate branches: %v", err), nil, nil
        }
        
        if !status.Exists {
            return TextErrorf("Source branch '%s' does not exist", head), nil, nil
        }
        
        if !status.HasCommits {
            return TextErrorf("Source branch '%s' has no commits to merge", head), nil, nil
        }
        
        if status.IsBehind {
            return TextErrorf("Source branch '%s' is behind target branch '%s'. Please sync your branch.", head, base), nil, nil
        }
        
        if status.HasConflicts {
            conflictList := strings.Join(status.ConflictFiles, ", ")
            return TextErrorf("Branch '%s' has conflicts with '%s'. Conflicting files: %s", head, base, conflictList), nil, nil
        }
    }
    
    // ... rest of implementation ...
}
```

#### 3. Enhanced Error Messages
**File**: `server/pr_create.go`
**Changes**: Improve error context and user guidance

```go
// Enhanced error handling with specific guidance
func validateAndCreatePR(ctx context.Context, s *Server, args PullRequestCreateArgs) (*mcp.CallToolResult, *PullRequestCreateResult, error) {
    // ... validation logic with detailed error messages ...
    
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "permission denied"):
            return TextErrorf("Permission denied: You may not have write access to create PRs in this repository. Check your permissions or contact the repository owner."), nil, nil
        case strings.Contains(err.Error(), "branch not found"):
            return TextErrorf("Branch not found: Ensure the source branch exists and is pushed to the remote repository."), nil, nil
        case strings.Contains(err.Error(), "protected branch"):
            return TextErrorf("Protected branch: The target branch may be protected. Contact a maintainer or choose a different base branch."), nil, nil
        default:
            return TextErrorf("Failed to create pull request: %v", err), nil, nil
        }
    }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Detailed conflict detection works: `go test ./server_test -run TestConflictDetection`
- [ ] Branch status validation comprehensive: `go test ./server_test -run TestBranchStatus`
- [ ] Enhanced error messages appropriate: `go test ./server_test -run TestErrorHandling`
- [ ] All tests pass including edge cases: `go test ./...`

#### Manual Verification:
- [ ] Conflict errors list specific files in conflict
- [ ] Behind-branch errors provide clear guidance
- [ ] Permission errors suggest specific actions
- [ ] Protected branch errors provide alternatives
- [ ] All error messages are user-friendly and actionable

---

## Testing Strategy

### Unit Tests:
- **Git Utilities**: Test all git command executions with mock repositories
- **Repository Resolution**: Test fork detection with various remote configurations
- **Template Loading**: Test template fetching and merging logic
- **Handler Logic**: Test validation, error handling, and response formatting
- **Remote Clients**: Test API integration and response transformation

### Integration Tests:
- **End-to-End PR Creation**: Complete workflow from directory to PR creation
- **Fork Scenarios**: Test various fork configurations and cross-repository PRs
- **Template Scenarios**: Test template loading with and without user content
- **Error Scenarios**: Test all error conditions and recovery paths

### Manual Testing Steps:
1. **Basic PR Creation**: Create PR from current branch to main
2. **Directory Auto-Detection**: Use directory parameter without explicit repository
3. **Fork PR Creation**: Create PR from fork to original repository
4. **Template Usage**: Create PR with and without templates
5. **Draft PRs**: Create draft PRs and verify title prefix
6. **Error Conditions**: Test conflicts, behind branches, missing branches
7. **Reviewer Assignment**: Test single reviewer assignment

## Performance Considerations

- **Git Command Execution**: Use context-aware timeouts for git operations
- **Template Loading**: Cache template content when possible
- **Fork Detection**: Minimize API calls by using local git config first
- **Conflict Detection**: Use efficient git commands to avoid unnecessary work

## Migration Notes

- **Backward Compatibility**: New tool doesn't affect existing functionality
- **Configuration**: No configuration changes required
- **Dependencies**: No new external dependencies required
- **API Changes**: Only additive changes to interfaces and clients

## References

- Original ticket: `thoughts/tickets/feature_pr_create.md`
- Research document: `thoughts/research/2025-10-06_pr_creation_implementation.md`
- PR edit implementation: `server/pr_edit.go`
- Repository resolver: `server/repository_resolver.go`
- Remote interfaces: `remote/interface.go`
- Forgejo SDK patterns: `remote/forgejo/pull_requests.go`
- Test patterns: `server_test/pr_edit_test.go`