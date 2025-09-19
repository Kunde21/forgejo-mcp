# Directory Parameter Support Implementation Plan

## Overview

This plan outlines the implementation of directory parameter support for all tools in the forgejo-mcp server. The goal is to add a consistent `directory` parameter to all server tools that allows users to specify which directory/repository they want to work with, similar to how other MCP tools operate.

## Current State Analysis

### Existing Tools
The forgejo-mcp server currently implements 8 tools:

1. **hello** - Simple greeting tool (no repository operations)
2. **issue_list** - Lists issues from a repository
3. **issue_comment_create** - Creates comments on issues
4. **issue_comment_list** - Lists comments from issues
5. **issue_comment_edit** - Edits existing issue comments
6. **pr_list** - Lists pull requests from a repository
7. **pr_comment_list** - Lists comments from pull requests
8. **pr_comment_create** - Creates comments on pull requests
9. **pr_comment_edit** - Edits existing pull request comments

### Current Parameter Structure
All repository-specific tools currently use a `repository` parameter with the format "owner/repo". This parameter is:
- Required for all repository operations
- Validated using regex `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
- Passed directly to the Gitea client interface

### Architecture Overview
- **Server Layer**: Handles MCP protocol and tool registration
- **Common Layer**: Provides validation utilities and helper functions
- **Remote Layer**: Gitea client interface and implementation
- **Test Layer**: Comprehensive test suite with mock harness

## Design Approach

### Directory Parameter Interface

#### Parameter Design
- **Name**: `directory` (consistent with MCP conventions)
- **Type**: `string`
- **Format**: Absolute file system path to a git repository
- **Validation**: Must be a valid directory path containing a `.git` directory
- **Default**: Current working directory if not specified

#### Backward Compatibility Strategy
1. **Phase 1**: Support both `repository` and `directory` parameters
2. **Phase 2**: Deprecate `repository` parameter with warnings
3. **Phase 3**: Remove `repository` parameter in future major version

#### Parameter Resolution Logic
```go
type RepositoryResolution struct {
    Directory   string // Directory path
    Repository  string // Resolved "owner/repo" format
    IsRemote    bool   // Whether this is a remote repository
}
```

### Implementation Architecture

#### 1. Enhanced Common Layer
**New File**: `server/repository_resolver.go`

```go
package server

import (
    "os"
    "path/filepath"
    "strings"
    "regexp"
    
    v "github.com/go-ozzo/ozzo-validation/v4"
)

// RepositoryResolver handles directory-to-repository resolution
type RepositoryResolver struct {
    // Configuration for resolution behavior
}

// ResolveRepository resolves a directory path to repository information
func (r *RepositoryResolver) ResolveRepository(directory string) (*RepositoryResolution, error)

// ValidateDirectory validates that a directory contains a git repository
func (r *RepositoryResolver) ValidateDirectory(directory string) error

// ExtractRemoteInfo extracts owner/repo from git remote configuration
func (r *RepositoryResolver) ExtractRemoteInfo(directory string) (string, error)
```

#### 2. Updated Tool Argument Structures
Each tool will be enhanced to support both parameter types:

```go
// Example for issue_list
type IssueListArgs struct {
    Repository  string `json:"repository"`  // Legacy support
    Directory   string `json:"directory"`   // New parameter
    Limit       int    `json:"limit"`
    Offset      int    `json:"offset"`
}

// Validation will check that exactly one of Repository or Directory is provided
```

#### 3. Enhanced Server Interface
**Modification**: `server/server.go`

Add repository resolver to server struct:
```go
type Server struct {
    mcpServer         *mcp.Server
    config           *config.Config
    remote           gitea.GiteaClientInterface
    repoResolver     *RepositoryResolver  // New field
}
```

#### 4. Updated Tool Handlers
Each tool handler will:
1. Accept both `repository` and `directory` parameters
2. Use the repository resolver to normalize the input
3. Pass resolved repository information to the remote client
4. Provide appropriate error messages and deprecation warnings

## Implementation Steps

### Phase 1: Foundation (Week 1)

#### 1.1 Create Repository Resolver
- **File**: `server/repository_resolver.go`
- **Tasks**:
  - Implement `RepositoryResolver` struct
  - Add directory validation logic
  - Implement git remote parsing
  - Add unit tests for resolver functionality

#### 1.2 Update Common Layer
- **File**: `server/common.go`
- **Tasks**:
  - Add validation helpers for directory parameters
  - Add parameter resolution utilities
  - Update regex patterns for directory validation

#### 1.3 Enhance Server Structure
- **File**: `server/server.go`
- **Tasks**:
  - Add `RepositoryResolver` to server struct
  - Update constructor to initialize resolver
  - Add configuration options for resolver behavior

### Phase 2: Tool Implementation (Week 2)

#### 2.1 Update Issue Tools
- **Files**: `server/issues.go`, `server/issue_comments.go`
- **Tasks**:
  - Update argument structures to include `directory` parameter
  - Implement parameter resolution logic
  - Add validation for mutually exclusive parameters
  - Update tool descriptions and documentation

#### 2.2 Update Pull Request Tools
- **Files**: `server/pr_list.go`, `server/pr_comments.go`
- **Tasks**:
  - Update argument structures to include `directory` parameter
  - Implement parameter resolution logic
  - Add validation for mutually exclusive parameters
  - Update tool descriptions and documentation

#### 2.3 Update Hello Tool
- **File**: `server/hello.go`
- **Tasks**:
  - Add directory parameter for consistency
  - Implement directory validation
  - Update tool description

### Phase 3: Testing (Week 3)

#### 3.1 Unit Tests
- **Files**: `server_test/*_test.go`
- **Tasks**:
  - Add tests for directory parameter validation
  - Add tests for repository resolution logic
  - Add tests for backward compatibility
  - Add tests for error scenarios

#### 3.2 Integration Tests
- **Files**: `server_test/tool_discovery_test.go`
- **Tasks**:
  - Update tool discovery tests to include directory parameter
  - Add end-to-end tests for directory-based operations
  - Add migration path tests

#### 3.3 Mock Server Updates
- **Files**: `server_test/harness.go`, `server_test/helpers.go`
- **Tasks**:
  - Update mock server to handle directory parameters
  - Add mock directory resolution
  - Update test helpers for new parameter structure

### Phase 4: Documentation and Migration (Week 4)

#### 4.1 Update Documentation
- **Files**: `README.md`, `mcp.json`
- **Tasks**:
  - Update tool descriptions to include directory parameter
  - Add migration guide for existing users
  - Update configuration documentation

#### 4.2 Deprecation Strategy
- **Implementation**: Add deprecation warnings for `repository` parameter
- **Tasks**:
  - Add warning messages when `repository` parameter is used
  - Log deprecation notices
  - Update documentation to reflect deprecation timeline

## Technical Specifications

### Parameter Validation Rules

#### Directory Parameter
```go
// Validation rules for directory parameter
v.Field(&args.Directory, v.When(args.Directory != "", v.And(
    v.Required,
    v.Match(dirReg).Error("directory must be a valid absolute path"),
    v.By(func(value interface{}) error {
        dir := value.(string)
        return validateGitDirectory(dir)
    }),
)))
```

#### Repository Parameter (Legacy)
```go
// Validation rules for repository parameter (with deprecation warning)
v.Field(&args.Repository, v.When(args.Repository != "", v.And(
    v.Required,
    v.Match(repoReg).Error("repository must be in format 'owner/repo'"),
    v.By(func(value interface{}) error {
        // Log deprecation warning
        logDeprecationWarning("repository parameter")
        return nil
    }),
)))
```

#### Mutual Exclusivity
```go
// Ensure exactly one of repository or directory is provided
v.Field(&args, v.By(func(value interface{}) error {
    args := value.(*IssueListArgs)
    if args.Repository != "" && args.Directory != "" {
        return errors.New("cannot specify both repository and directory parameters")
    }
    if args.Repository == "" && args.Directory == "" {
        return errors.New("must specify either repository or directory parameter")
    }
    return nil
}))
```

### Repository Resolution Logic

```go
func (r *RepositoryResolver) ResolveRepository(directory string) (*RepositoryResolution, error) {
    // Validate directory exists
    if err := r.ValidateDirectory(directory); err != nil {
        return nil, err
    }
    
    // Check if it's a git repository
    gitDir := filepath.Join(directory, ".git")
    if _, err := os.Stat(gitDir); err != nil {
        return nil, fmt.Errorf("directory is not a git repository: %w", err)
    }
    
    // Extract remote information
    remoteInfo, err := r.ExtractRemoteInfo(directory)
    if err != nil {
        return nil, fmt.Errorf("failed to extract remote information: %w", err)
    }
    
    return &RepositoryResolution{
        Directory:  directory,
        Repository: remoteInfo,
        IsRemote:   true,
    }, nil
}
```

### Error Handling Strategy

#### Validation Errors
- Clear, actionable error messages
- Consistent error formatting across all tools
- Proper error codes for different failure types

#### Resolution Errors
- Distinguish between directory validation errors and git repository errors
- Provide helpful suggestions for common issues
- Include context about which validation step failed

#### Migration Errors
- Clear deprecation warnings with migration guidance
- Links to documentation for migration steps
- Timeline information for parameter removal

## Testing Strategy

### Unit Tests
- **Repository Resolver**: Test directory validation, git detection, remote parsing
- **Parameter Validation**: Test mutual exclusivity, format validation, default handling
- **Error Handling**: Test various error scenarios and messages

### Integration Tests
- **Tool Discovery**: Verify all tools expose directory parameter correctly
- **End-to-End**: Test complete workflows using directory parameter
- **Backward Compatibility**: Verify existing repository parameter still works

### Performance Tests
- **Resolution Speed**: Test directory resolution performance
- **Memory Usage**: Verify no memory leaks with repeated resolution calls
- **Concurrent Access**: Test thread safety of resolver

## Migration Plan

### Timeline
- **Week 1-2**: Implementation and internal testing
- **Week 3**: User testing and feedback collection
- **Week 4**: Documentation updates and release preparation
- **Month 2**: Deprecation warnings activated
- **Month 6**: Repository parameter removed (major version bump)

### Communication Strategy
- **Release Notes**: Detailed explanation of new parameter
- **Blog Post**: Migration guide and best practices
- **Documentation**: Updated with examples and migration steps
- **Community Outreach**: Announcements in relevant forums

### Rollback Plan
- **Feature Flag**: Ability to disable directory parameter support
- **Configuration**: Option to force legacy behavior
- **Version Pinning**: Clear versioning for compatibility

## Success Metrics

### Technical Metrics
- **Test Coverage**: Maintain >90% test coverage
- **Performance**: Directory resolution <100ms for local repositories
- **Compatibility**: 100% backward compatibility with existing tools

### User Experience Metrics
- **Adoption Rate**: Track usage of directory vs repository parameter
- **Error Rate**: Monitor error rates for new parameter usage
- **Support Tickets**: Track migration-related support requests

### Code Quality Metrics
- **Complexity**: Maintain cyclomatic complexity <10 per function
- **Duplication**: Minimize code duplication through shared utilities
- **Documentation**: 100% API documentation coverage

## Risks and Mitigation

### Technical Risks
- **Git Repository Detection**: Different git configurations may cause detection issues
  - *Mitigation*: Comprehensive testing with various git setups
- **Performance**: Directory resolution may be slow for network drives
  - *Mitigation*: Caching and timeout mechanisms
- **Compatibility**: Breaking changes in MCP protocol
  - *Mitigation*: Version pinning and compatibility layer

### User Experience Risks
- **Migration Complexity**: Users may struggle with migration
  - *Mitigation*: Clear documentation and migration tools
- **Parameter Confusion**: Users may not understand when to use which parameter
  - *Mitigation*: Clear examples and error messages
- **Adoption Resistance**: Users may resist changing existing workflows
  - *Mitigation*: Gradual transition with extended support period

## Conclusion

This implementation plan provides a comprehensive approach to adding directory parameter support to the forgejo-mcp server. The plan emphasizes:

1. **Backward Compatibility**: Existing workflows continue to work
2. **Consistent Interface**: All tools follow the same parameter pattern
3. **Robust Validation**: Comprehensive validation and error handling
4. **Clear Migration**: Gradual transition with proper communication
5. **Comprehensive Testing**: Full test coverage for all scenarios

The implementation will be delivered in phases to ensure stability and allow for feedback at each stage. The final result will be a more flexible and user-friendly MCP server that aligns with modern MCP conventions.