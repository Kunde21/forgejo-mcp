# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-09-validation-layer-migration/spec.md

> Created: 2025-09-09
> Version: 1.0.0

## Technical Requirements

### Core Validation Layer Requirements

1. **Validation Framework Integration**
   - Integrate ozzo-validation v4 as the primary validation library
   - Create reusable validation patterns for common data types
   - Implement validation middleware for request processing

2. **Service Layer Validation**
   - Add validation to all service methods before business logic execution
   - Create domain-specific validation rules for each service
   - Implement structured error responses with field-level details

3. **Handler Layer Validation**
   - Add request validation before service calls
   - Implement response validation for outgoing data
   - Create validation middleware for HTTP handlers

4. **Error Handling Enhancement**
   - Standardize error response formats
   - Add validation error codes and messages
   - Implement error localization support

### Validation Rules Implementation

#### Common Validation Patterns

```go
// User validation rules
var UserValidationRules = validation.Errors{
    "username": validation.Validate(
        &user.Username,
        validation.Required,
        validation.Length(3, 50),
        validation.Match(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)),
    ),
    "email": validation.Validate(
        &user.Email,
        validation.Required,
        validation.By(isEmail),
    ),
    "password": validation.Validate(
        &user.Password,
        validation.Required,
        validation.Length(8, 100),
        validation.By(isStrongPassword),
    ),
}

// Repository validation rules
var RepositoryValidationRules = validation.Errors{
    "name": validation.Validate(
        &repo.Name,
        validation.Required,
        validation.Length(1, 100),
        validation.Match(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)),
    ),
    "description": validation.Validate(
        &repo.Description,
        validation.Length(0, 255),
    ),
    "private": validation.Validate(
        &repo.Private,
        validation.In(true, false),
    ),
}

// Issue validation rules
var IssueValidationRules = validation.Errors{
    "title": validation.Validate(
        &issue.Title,
        validation.Required,
        validation.Length(1, 255),
    ),
    "body": validation.Validate(
        &issue.Body,
        validation.Length(0, 65535),
    ),
    "state": validation.Validate(
        &issue.State,
        validation.In("open", "closed"),
    ),
}
```

#### Custom Validation Functions

```go
// Email validation
func isEmail(value interface{}) error {
    email, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }
    
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    
    return nil
}

// Strong password validation
func isStrongPassword(value interface{}) error {
    password, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
    
    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number, and special character")
    }
    
    return nil
}

// Repository name validation
func isValidRepoName(value interface{}) error {
    name, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }
    
    if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
        return errors.New("repository name cannot start with . or -")
    }
    
    if strings.HasSuffix(name, ".git") {
        return errors.New("repository name cannot end with .git")
    }
    
    return nil
}
```

## Approach

### Phase 1: Validation Framework Setup

1. **Add ozzo-validation v4 dependency**
   ```bash
   go get github.com/go-ozzo/ozzo-validation/v4
   ```

2. **Create validation package structure**
   ```
   validation/
   ├── validators.go      # Custom validation functions
   ├── rules.go           # Common validation rules
   ├── middleware.go      # Validation middleware
   └── errors.go          # Error handling utilities
   ```

3. **Implement base validation interfaces**
   ```go
   package validation

   type Validator interface {
       Validate() error
   }

   type RequestValidator interface {
       ValidateRequest(req interface{}) error
   }

   type ResponseValidator interface {
       ValidateResponse(resp interface{}) error
   }
   ```

### Phase 2: Service Layer Integration

#### Service Layer Changes

**File: `remote/gitea/service.go`**
```go
// Add validation to service methods
func (s *GiteaService) CreateRepository(ctx context.Context, req *CreateRepositoryRequest) (*Repository, error) {
    // Validate request before processing
    if err := validation.Validate(req, RepositoryValidationRules); err != nil {
        return nil, NewValidationError("invalid repository data", err)
    }
    
    // Existing business logic...
}

func (s *GiteaService) CreateIssue(ctx context.Context, req *CreateIssueRequest) (*Issue, error) {
    // Validate request before processing
    if err := validation.Validate(req, IssueValidationRules); err != nil {
        return nil, NewValidationError("invalid issue data", err)
    }
    
    // Existing business logic...
}

func (s *GiteaService) AddIssueComment(ctx context.Context, req *AddIssueCommentRequest) (*IssueComment, error) {
    // Validate request before processing
    if err := validation.Validate(req, IssueCommentValidationRules); err != nil {
        return nil, NewValidationError("invalid comment data", err)
    }
    
    // Existing business logic...
}
```

**File: `remote/gitea/interface.go`**
```go
// Add validation to request/response structs
type CreateRepositoryRequest struct {
    Name        string `json:"name" validate:"required"`
    Description string `json:"description" validate:"max=255"`
    Private     bool   `json:"private" validate:"boolean"`
}

type CreateIssueRequest struct {
    Title string `json:"title" validate:"required,max=255"`
    Body  string `json:"body" validate:"max=65535"`
}

type AddIssueCommentRequest struct {
    Body string `json:"body" validate:"required,max=65535"`
}
```

### Phase 3: Handler Layer Integration

#### Handler Layer Changes

**File: `server/handlers.go`**
```go
// Add validation middleware
func (s *Server) validateRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Decode request body
        var req interface{}
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            s.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
            return
        }
        
        // Validate request based on endpoint
        validator := s.getValidator(r.URL.Path)
        if validator != nil {
            if err := validator.ValidateRequest(req); err != nil {
                s.writeValidationError(w, err)
                return
            }
        }
        
        // Call next handler
        next.ServeHTTP(w, r)
    })
}

// Add validation to handler methods
func (s *Server) handleCreateRepository(w http.ResponseWriter, r *http.Request) {
    var req CreateRepositoryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    // Validate request
    if err := validation.Validate(&req, RepositoryValidationRules); err != nil {
        s.writeValidationError(w, err)
        return
    }
    
    // Call service
    repo, err := s.giteaService.CreateRepository(r.Context(), &req)
    if err != nil {
        s.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    s.writeJSONResponse(w, http.StatusCreated, repo)
}

func (s *Server) handleCreateIssue(w http.ResponseWriter, r *http.Request) {
    var req CreateIssueRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    // Validate request
    if err := validation.Validate(&req, IssueValidationRules); err != nil {
        s.writeValidationError(w, err)
        return
    }
    
    // Call service
    issue, err := s.giteaService.CreateIssue(r.Context(), &req)
    if err != nil {
        s.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    s.writeJSONResponse(w, http.StatusCreated, issue)
}
```

### Phase 4: Error Handling Enhancement

**File: `validation/errors.go`**
```go
package validation

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-ozzo/ozzo-validation/v4"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ValidationErrorResponse struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Errors  []ValidationError `json:"errors"`
}

func NewValidationError(message string, err error) *ValidationErrorResponse {
    resp := &ValidationErrorResponse{
        Code:    "VALIDATION_ERROR",
        Message: message,
    }
    
    if validationErr, ok := err.(validation.Errors); ok {
        for field, fieldErr := range validationErr {
            resp.Errors = append(resp.Errors, ValidationError{
                Field:   field,
                Message: fieldErr.Error(),
            })
        }
    } else {
        resp.Errors = append(resp.Errors, ValidationError{
            Field:   "general",
            Message: err.Error(),
        })
    }
    
    return resp
}

func (e *ValidationErrorResponse) Error() string {
    data, _ := json.Marshal(e)
    return string(data)
}

func (e *ValidationErrorResponse) StatusCode() int {
    return http.StatusUnprocessableEntity
}
```

**File: `server/server.go`**
```go
// Add validation error response handler
func (s *Server) writeValidationError(w http.ResponseWriter, err error) {
    if validationErr, ok := err.(*validation.ValidationErrorResponse); ok {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(validationErr.StatusCode())
        json.NewEncoder(w).Encode(validationErr)
        return
    }
    
    s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
}
```

## External Dependencies

### Required Dependencies

1. **ozzo-validation v4**
   - Package: `github.com/go-ozzo/ozzo-validation/v4`
   - Purpose: Core validation framework
   - Version: v4.0.0 or later

2. **Existing Dependencies**
   - `github.com/stretchr/testify` - For testing validation rules
   - `github.com/gorilla/mux` - For route-specific validation

### Integration Points

1. **Service Layer Integration**
   - Modify all service methods to include validation
   - Update request/response structs with validation tags
   - Add validation error handling

2. **Handler Layer Integration**
   - Add validation middleware
   - Update handler methods to validate requests
   - Implement validation error responses

3. **Testing Integration**
   - Add validation tests for all service methods
   - Add validation tests for all handler methods
   - Create validation test utilities

## File-Specific Changes

### New Files to Create

1. **`validation/validators.go`**
   - Custom validation functions
   - Domain-specific validation rules

2. **`validation/rules.go`**
   - Common validation rule sets
   - Reusable validation patterns

3. **`validation/middleware.go`**
   - HTTP validation middleware
   - Request/response validation utilities

4. **`validation/errors.go`**
   - Validation error types
   - Error response formatting

### Existing Files to Modify

1. **`remote/gitea/service.go`**
   - Add validation to all service methods
   - Update error handling

2. **`remote/gitea/interface.go`**
   - Add validation tags to request/response structs
   - Update method signatures

3. **`server/handlers.go`**
   - Add validation to handler methods
   - Implement validation middleware

4. **`server/server.go`**
   - Add validation error response handling
   - Update server initialization

5. **`go.mod`**
   - Add ozzo-validation v4 dependency

6. **`config/config.go`**
   - Add validation configuration options
   - Update config validation

### Test Files to Update

1. **`remote/gitea/service_test.go`**
   - Add validation test cases
   - Update existing tests

2. **`server/handlers_test.go`**
   - Add validation test cases
   - Update existing tests

3. **`validation/validators_test.go`** (new)
   - Test custom validation functions
   - Test validation rules

4. **`validation/middleware_test.go`** (new)
   - Test validation middleware
   - Test error responses

## Implementation Timeline

### Week 1: Foundation
- Add ozzo-validation v4 dependency
- Create validation package structure
- Implement custom validation functions
- Add common validation rules

### Week 2: Service Layer
- Update service interfaces with validation
- Add validation to all service methods
- Update service error handling
- Add service layer tests

### Week 3: Handler Layer
- Implement validation middleware
- Update handler methods with validation
- Add validation error responses
- Add handler layer tests

### Week 4: Testing and Refinement
- Comprehensive testing
- Performance optimization
- Documentation updates
- Final integration testing