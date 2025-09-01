package types

import (
	"fmt"
	"time"
)

// ErrorCodes for standard API error responses
const (
	ErrorCodeValidation    = "VALIDATION_ERROR"
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeForbidden     = "FORBIDDEN"
	ErrorCodeInternalError = "INTERNAL_ERROR"
	ErrorCodeBadRequest    = "BAD_REQUEST"
)

// ResponseMetadata contains metadata for API responses
type ResponseMetadata struct {
	RequestID string    `json:"requestId"`
	Timestamp Timestamp `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorDetails contains detailed error information
type ErrorDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success   bool              `json:"success"`
	Data      interface{}       `json:"data,omitempty"`
	Metadata  *ResponseMetadata `json:"metadata,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool          `json:"success"`
	Error   *ErrorDetails `json:"error"`
}

// Pagination contains pagination information
type Pagination struct {
	Page    int  `json:"page"`
	PerPage int  `json:"perPage"`
	Total   int  `json:"total"`
	HasNext bool `json:"hasNext"`
	HasPrev bool `json:"hasPrev"`
}

// PaginatedResponse extends SuccessResponse with pagination info
type PaginatedResponse struct {
	Success    bool              `json:"success"`
	Data       interface{}       `json:"data,omitempty"`
	Pagination *Pagination       `json:"pagination,omitempty"`
	Metadata   *ResponseMetadata `json:"metadata,omitempty"`
	RequestID  string            `json:"requestId,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMetadata creates a new success response with metadata
func NewSuccessResponseWithMetadata(data interface{}, metadata *ResponseMetadata) *SuccessResponse {
	return &SuccessResponse{
		Success:  true,
		Data:     data,
		Metadata: metadata,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetails{
			Code:    code,
			Message: message,
		},
	}
}

// NewErrorResponseWithDetails creates a new error response with additional details
func NewErrorResponseWithDetails(code, message string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, pagination *Pagination) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}
}

// NewPaginatedResponseWithMetadata creates a new paginated response with metadata
func NewPaginatedResponseWithMetadata(data interface{}, pagination *Pagination, metadata *ResponseMetadata) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Metadata:   metadata,
	}
}

// NewPagination creates pagination info from parameters
func NewPagination(page, perPage, total int) *Pagination {
	return &Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   total,
		HasNext: (page * perPage) < total,
		HasPrev: page > 1,
	}
}

// NewResponseMetadata creates response metadata
func NewResponseMetadata(requestID, version string) *ResponseMetadata {
	return &ResponseMetadata{
		RequestID: requestID,
		Timestamp: Timestamp{Time: time.Now()},
		Version:   version,
	}
}

// Validate checks if ErrorDetails has required fields
func (e *ErrorDetails) Validate() error {
	if e.Code == "" {
		return fmt.Errorf("error code cannot be empty")
	}
	if e.Message == "" {
		return fmt.Errorf("error message cannot be empty")
	}
	return nil
}

// Validate checks if ResponseMetadata has required fields
func (r *ResponseMetadata) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("request ID cannot be empty")
	}
	if r.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	return nil
}

// Validate checks if Pagination has valid values
func (p *Pagination) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be positive")
	}
	if p.PerPage < 1 {
		return fmt.Errorf("perPage must be positive")
	}
	if p.Total < 0 {
		return fmt.Errorf("total cannot be negative")
	}
	return nil
}
