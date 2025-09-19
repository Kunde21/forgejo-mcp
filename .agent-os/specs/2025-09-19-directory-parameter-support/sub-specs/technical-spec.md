# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-19-directory-parameter-support/spec.md

## Technical Requirements

### Repository Resolver Component
- **File**: `server/repository_resolver.go`
- **Purpose**: Handle directory-to-repository resolution with git detection and remote parsing
- **Functions**: 
  - `ResolveRepository(directory string) (*RepositoryResolution, error)` - Main resolution logic
  - `ValidateDirectory(directory string) error` - Directory existence and git repository validation
  - `ExtractRemoteInfo(directory string) (string, error)` - Parse git remote to extract owner/repo
- **Validation Rules**: Directory must exist, contain `.git` directory, and have at least one configured remote

### Parameter Structure Updates
- **Files**: All server tool files (`server/*.go`)
- **Purpose**: Add directory parameter to all tool argument structures with mutual exclusivity validation
- **Implementation Pattern**:
  ```go
  type ToolArgs struct {
      Repository  string `json:"repository"`  // Legacy support
      Directory   string `json:"directory"`   // New parameter
      // ... existing parameters
  }
  ```
- **Validation**: Exactly one of repository or directory must be provided, with appropriate format validation for each

### Server Architecture Enhancements
- **File**: `server/server.go`
- **Purpose**: Integrate repository resolver into server structure
- **Changes**: Add `RepositoryResolver` field to Server struct, update constructor to initialize resolver
- **Configuration**: Add resolver configuration options for timeout, caching, and default behavior

### Error Handling Strategy
- **Validation Errors**: Clear, actionable messages for invalid directory paths, missing git repositories, and remote parsing failures
- **Mutual Exclusivity**: Specific error when both repository and directory parameters are provided
- **Deprecation Warnings**: Logged warnings when repository parameter is used, with migration guidance
- **Resolution Failures**: Detailed error context including which validation step failed and suggestions for resolution

### Performance Considerations
- **Caching**: Optional caching mechanism for directory resolution results to improve performance for repeated calls
- **Timeout**: Configurable timeout for directory resolution operations to handle slow network drives
- **Concurrent Access**: Thread-safe implementation to handle multiple simultaneous resolution requests

## External Dependencies (Conditional)

No new external dependencies are required for this implementation. The feature will be built using existing Go standard library packages (`os`, `path/filepath`, `os/exec`) and current project dependencies (`github.com/go-ozzo/ozzo-validation/v4`).

**Justification**: The directory parameter implementation relies on standard Go file system operations and git command-line tools that are already available in the target environment. Adding new dependencies would increase complexity without providing significant benefits for this specific feature.