# Validation Layer Architecture

This document describes the validation architecture implemented in the forgejo-mcp project following the validation layer migration.

## Overview

The validation layer has been migrated from the service layer to the handler layer to:
- Eliminate validation duplication
- Improve separation of concerns
- Establish clear validation patterns
- Enhance maintainability

## Architecture

### Layer Responsibilities

1. **Handler Layer**: Input validation and error formatting
2. **Service Layer**: Business logic only (no validation)
3. **Validation Package**: Shared validation utilities and rules

### Validation Flow

```
Client Request → Handler (Validation) → Service (Business Logic) → Client Response
```

## Validation Patterns

### Repository Validation

```go
// Using shared validation rule
v.Field(&args.Repository, v.Required, validation.RepositoryRule())

// Validates format: owner/repo
// Examples: "user/repo", "organization/project", "user123/repo_name"
```

### Issue Number Validation

```go
// Using shared validation rule
v.Field(&args.IssueNumber, validation.IssueNumberRule())

// Validates: must be positive integer (> 0)
```

### Comment Content Validation

```go
// Using shared validation rule
v.Field(&args.Comment, v.Required, validation.CommentContentRule())

// Validates: non-empty string, not only whitespace
```

### Pagination Validation

```go
// Combined limit validation (1-100)
v.Field(&args.Limit, validation.CombinedPaginationLimitRule())

// Offset validation (≥ 0)
v.Field(&args.Offset, validation.PaginationOffsetRule())
```

## Error Handling

### Consistent Error Messages

All validation errors follow the pattern:
```
"Validation failed: [specific validation error]"
```

### Handler Error Response

```go
if err := v.ValidateStruct(&args, ...); err != nil {
    return TextErrorf("Validation failed: %v", err), nil, nil
}
```

## Testing Strategy

### Unit Tests
- Individual validation functions
- Handler validation logic
- Edge cases and boundary conditions

### Integration Tests
- End-to-end validation flows
- Error response consistency
- Handler-to-service communication

### Consistency Tests
- Validation rules consistency across handlers
- Error message standardization
- Shared validation utility usage

## Usage Examples

### Adding Validation to a New Handler

```go
func (s *Server) handleNewTool(ctx context.Context, request *mcp.CallToolRequest, args struct {
    Repository string `json:"repository"`
    SomeField  string `json:"some_field"`
}) (*mcp.CallToolResult, any, error) {
    // Validate input using shared utilities
    if err := v.ValidateStruct(&args,
        v.Field(&args.Repository, v.Required, validation.RepositoryRule()),
        v.Field(&args.SomeField, v.Required), // Add custom validation as needed
    ); err != nil {
        return TextErrorf("Validation failed: %v", err), nil, nil
    }

    // Business logic (no validation)
    result, err := s.someService.DoSomething(ctx, args.Repository, args.SomeField)
    if err != nil {
        return TextErrorf("Operation failed: %v", err), nil, nil
    }

    return TextResult("Success"), result, nil
}
```

### Adding New Validation Rules

```go
// In validation/validators.go
func NewCustomRule() v.Rule {
    return v.By(func(value interface{}) error {
        // Custom validation logic
        return nil
    })
}
```

## Best Practices

1. **Always validate at handler entry points**
2. **Use shared validation utilities for consistency**
3. **Keep service layer focused on business logic**
4. **Provide clear, helpful error messages**
5. **Test validation thoroughly with edge cases**
6. **Maintain consistency across all handlers**

## Migration Benefits

- ✅ **Single Source of Truth**: Validation logic centralized
- ✅ **Separation of Concerns**: Handlers validate, services execute
- ✅ **Consistency**: Uniform validation patterns
- ✅ **Maintainability**: Easier to update validation rules
- ✅ **Testability**: Clear validation testing boundaries
- ✅ **Performance**: Validation happens early in request flow