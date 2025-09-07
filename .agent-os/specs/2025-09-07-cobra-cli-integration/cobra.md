# Plan: Update to Use Cobra CLI Framework

## Current State Analysis
- **main.go**: Contains MCP server initialization and startup logic
- **Structure**: Simple single-purpose application that starts MCP server immediately
- **Dependencies**: No cobra currently in go.mod
- **CLI**: No command-line interface, just runs the server

## Proposed Structure

### 1. Root Directory Changes
- **main.go**: Initialize cobra root command and execute
- **go.mod**: Add cobra dependency

### 2. New cmd/ Directory Structure
```
cmd/
├── root.go          # Root command setup and global flags
├── serve.go         # Server subcommand (current main logic)
├── version.go       # Version subcommand
└── config.go        # Config validation subcommand
```

## Implementation Plan

### Phase 1: Dependencies
- Add `github.com/spf13/cobra` to go.mod
- Run `go mod tidy`

### Phase 2: Create cmd/ Directory Structure
- Create `cmd/` directory
- Move server logic from main.go to `cmd/serve.go`
- Create `cmd/root.go` for cobra initialization
- Create `cmd/version.go` for version command
- Create `cmd/config.go` for config validation

### Phase 3: Update main.go
- Replace server startup with cobra command execution
- Keep minimal imports and initialization

### Phase 4: Command Definitions

**Root Command (`cmd/root.go`)**:
- Global flags: `--config`, `--verbose`
- Persistent pre-run hooks for config loading
- Add subcommands

**Serve Command (`cmd/serve.go`)**:
- Move current `NewServer()` and `Start()` logic
- Add serve-specific flags: `--host`, `--port`
- Handle MCP server lifecycle

**Version Command (`cmd/version.go`)**:
- Display version information
- Show build details

**Config Command (`cmd/config.go`)**:
- Validate configuration
- Show current config values
- Test connectivity to Forgejo instance

## Benefits
1. **Extensibility**: Easy to add new subcommands
2. **Configuration**: Better CLI flag handling
3. **Help System**: Automatic help generation
4. **Testing**: Easier to test individual commands
5. **User Experience**: Standard CLI patterns

## Migration Steps
1. Add cobra dependency
2. Create cmd/ directory structure
3. Extract server logic to cmd/serve.go
4. Update main.go to use cobra
5. Add version and config commands
6. Update documentation

## Backward Compatibility
- Keep existing environment variable configuration
- Maintain same startup behavior for `go run main.go` (defaults to serve command)
- Preserve all existing functionality

This plan maintains the current functionality while providing a solid foundation for future CLI enhancements.