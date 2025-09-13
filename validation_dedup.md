# Validation Deduplication Plan

## Overview

This plan outlines the strategy to deduplicate input validation happening between `server/*` and `remote/gitea/*` files. All input validation should be consolidated in `server/[tool handler].go` using the ozzo-validation pattern with inline validation rules.

## Current State Analysis

### Server Layer Validation (already using correct patterns)
- Repository: `v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'"))`
- Pagination: `v.Field(&args.Limit, v.Min(1), v.Max(100))` and `v.Field(&args.Offset, v.Min(0))`
- Positive integers: `v.Field(&args.IssueNumber, v.Min(1))`
- Non-empty strings: `v.Field(&args.Comment, v.Required, v.Length(1, 0))`

### Service Layer Duplication (to be removed)
- `validateRepository()` - duplicates `v.Match(repoReg)`
- `validatePagination()` - duplicates `v.Min(1), v.Max(100)` and `v.Min(0)`
- `validateIssueNumber()`, `validatePullRequestNumber()` - duplicates `v.Min(1)`
- `validateCommentContent()` - duplicates `v.Required, v.Length(1, 0)`

## Implementation Plan

### Phase 1: Keep Server Layer Validation Inline

Server handlers will continue using inline validation patterns without helper functions:

```go
// Keep existing inline validation in all handlers
if err := v.ValidateStruct(&args,
    v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
    v.Field(&args.Limit, v.Min(1), v.Max(100)),
    v.Field(&args.Offset, v.Min(0)),
    v.Field(&args.IssueNumber, v.Min(1)),
    v.Field(&args.Comment, v.Required, v.Length(1, 0)),
); err != nil {
    return TextErrorf("Invalid request: %v", err), nil, nil
}
```

**No changes to server/common.go:**
- Keep `repoReg` regex pattern
- Keep existing helper functions (`TextResult`, `TextError`, etc.)
- Do NOT add any new validation helper functions

### Phase 2: Remove Service Layer Validation

#### 1. Remove validation methods from remote/gitea/service.go
Delete the following functions:
- `validateRepository()` function (lines 156-168)
- `validatePagination()` function (lines 171-179)
- `validateIssueNumber()` function (lines 182-187)
- `validatePullRequestNumber()` function (lines 190-195)
- `validateCommentContent()` function (lines 198-207)
- `validateCommentID()` function (lines 210-215)
- `validatePullRequestOptions()` function (lines 218-230)
- `validatePullRequestState()` function (lines 233-245)

#### 2. Remove validation calls from all service methods

**ListIssues method:**
```go
// Before:
func (s *Service) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
    if err := s.validateRepository(repo); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validatePagination(limit, offset); err != nil {
        return nil, fmt.Errorf("pagination validation failed: %w", err)
    }
    return s.client.ListIssues(ctx, repo, limit, offset)
}

// After:
func (s *Service) ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error) {
    return s.client.ListIssues(ctx, repo, limit, offset)
}
```

**CreateIssueComment method:**
```go
// Before:
func (s *Service) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*Comment, error) {
    if err := s.validateRepository(repo); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validateIssueNumber(issueNumber); err != nil {
        return nil, fmt.Errorf("issue number validation failed: %w", err)
    }
    if err := s.validateCommentContent(comment); err != nil {
        return nil, fmt.Errorf("comment content validation failed: %w", err)
    }
    return s.client.CreateIssueComment(ctx, repo, issueNumber, comment)
}

// After:
func (s *Service) CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*Comment, error) {
    return s.client.CreateIssueComment(ctx, repo, issueNumber, comment)
}
```

**ListIssueComments method:**
```go
// Before:
func (s *Service) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
    if err := s.validateRepository(repo); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validateIssueNumber(issueNumber); err != nil {
        return nil, fmt.Errorf("issue number validation failed: %w", err)
    }
    if err := s.validatePagination(limit, offset); err != nil {
        return nil, fmt.Errorf("pagination validation failed: %w", err)
    }
    return s.client.ListIssueComments(ctx, repo, issueNumber, limit, offset)
}

// After:
func (s *Service) ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error) {
    return s.client.ListIssueComments(ctx, repo, issueNumber, limit, offset)
}
```

**EditIssueComment method:**
```go
// Before:
func (s *Service) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*Comment, error) {
    if err := s.validateRepository(args.Repository); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validateIssueNumber(args.IssueNumber); err != nil {
        return nil, fmt.Errorf("issue number validation failed: %w", err)
    }
    if err := s.validateCommentID(args.CommentID); err != nil {
        return nil, fmt.Errorf("comment ID validation failed: %w", err)
    }
    if err := s.validateCommentContent(args.NewContent); err != nil {
        return nil, fmt.Errorf("new content validation failed: %w", err)
    }
    return s.client.EditIssueComment(ctx, args)
}

// After:
func (s *Service) EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*Comment, error) {
    return s.client.EditIssueComment(ctx, args)
}
```

**ListPullRequests method:**
```go
// Before:
func (s *Service) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
    if err := s.validateRepository(repo); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validatePullRequestOptions(options); err != nil {
        return nil, fmt.Errorf("pull request options validation failed: %w", err)
    }
    return s.client.ListPullRequests(ctx, repo, options)
}

// After:
func (s *Service) ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error) {
    return s.client.ListPullRequests(ctx, repo, options)
}
```

**ListPullRequestComments method:**
```go
// Before:
func (s *Service) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
    if err := s.validateRepository(repo); err != nil {
        return nil, fmt.Errorf("repository validation failed: %w", err)
    }
    if err := s.validatePullRequestNumber(pullRequestNumber); err != nil {
        return nil, fmt.Errorf("pull request number validation failed: %w", err)
    }
    if err := s.validatePagination(limit, offset); err != nil {
        return nil, fmt.Errorf("pagination validation failed: %w", err)
    }
    return s.client.ListPullRequestComments(ctx, repo, pullRequestNumber, limit, offset)
}

// After:
func (s *Service) ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error) {
    return s.client.ListPullRequestComments(ctx, repo, pullRequestNumber, limit, offset)
}
```

**CreatePullRequestComment and EditPullRequestComment methods:**
- Already have no validation - keep as-is

### Phase 3: Clean Up Interface Layer

#### 1. Remove unused validation tags from remote/gitea/interface.go
Remove the following validation tags from all structs:
- `validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
- `validate:"required,min=1"`
- `validate:"min=1,max=100"`
- `validate:"min=0"`
- `validate:"oneof=open closed all"`

#### 2. Simplify struct definitions
```go
// Before (with validation tags):
type ListIssueCommentsArgs struct {
    Repository  string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
    IssueNumber int    `json:"issue_number" validate:"required,min=1"`
    Limit       int    `json:"limit" validate:"min=1,max=100"`
    Offset      int    `json:"offset" validate:"min=0"`
}

// After (without validation tags):
type ListIssueCommentsArgs struct {
    Repository  string `json:"repository"`
    IssueNumber int    `json:"issue_number"`
    Limit       int    `json:"limit"`
    Offset      int    `json:"offset"`
}
```

**Apply to all structs in interface.go:**
- `ListIssueCommentsArgs`
- `EditIssueCommentArgs`
- `ListPullRequestsOptions`
- `ListPullRequestCommentsArgs`
- `CreatePullRequestCommentArgs`
- `EditPullRequestCommentArgs`

## Files to Modify

### Server Layer (minimal changes)
- `server/common.go` - No changes needed (keep existing patterns)
- `server/hello.go` - No changes needed
- `server/issue_comments.go` - No changes needed (already uses inline validation)
- `server/issues.go` - No changes needed (already uses inline validation)
- `server/pr_comments.go` - No changes needed (already uses inline validation)
- `server/pr_list.go` - No changes needed (already uses inline validation)

### Remote/Gitea Layer (major changes)
- `remote/gitea/service.go` - Remove all validation functions and calls
- `remote/gitea/interface.go` - Remove all validation tags from structs

## Expected Benefits

1. **Single Validation Source**: All validation logic remains in server handlers using inline patterns
2. **No Helper Functions**: Validation rules stay inline as requested
3. **Reduced Duplication**: Eliminates duplicate validation between server and service layers
4. **Clean Separation**: Service layer focuses purely on business logic
5. **Maintainable**: Validation logic is centralized in server handlers
6. **Performance**: Eliminates redundant validation checks

## Validation Patterns to Use

### Repository Validation
```go
v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'"))
```

### Pagination Validation
```go
v.Field(&args.Limit, v.Min(1), v.Max(100))
v.Field(&args.Offset, v.Min(0))
```

### Positive Integer Validation
```go
v.Field(&args.IssueNumber, v.Min(1))
v.Field(&args.CommentID, v.Min(1))
```

### Non-empty String Validation
```go
v.Field(&args.Comment, v.Required, v.Length(1, 0))
```

### State Validation
```go
v.Field(&args.State, v.In("open", "closed", "all").Error("state must be one of: open, closed, all"))
```

## Testing Considerations

1. **Run existing tests** to ensure functionality remains intact
2. **Verify error messages** are consistent and helpful
3. **Test edge cases** to ensure validation still works correctly
4. **Confirm service layer** no longer performs validation

## Success Criteria

- All validation logic is consolidated in server handlers
- Service layer contains no validation code
- Interface layer contains no validation tags
- All existing tests pass
- Error messages remain consistent and user-friendly
- No functional regression in tool behavior