package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	response := NewSuccessResponse(data)

	if !response.Success {
		t.Errorf("Success should be true")
	}
	if !reflect.DeepEqual(response.Data, data) {
		t.Errorf("Data = %v, want %v", response.Data, data)
	}
}

func TestNewSuccessResponseWithMetadata(t *testing.T) {
	data := "test data"
	metadata := &ResponseMetadata{
		RequestID: "req-123",
		Version:   "1.0.0",
	}

	response := NewSuccessResponseWithMetadata(data, metadata)

	if !response.Success {
		t.Errorf("Success should be true")
	}
	if response.Data != data {
		t.Errorf("Data = %v, want %v", response.Data, data)
	}
	if response.Metadata.RequestID != metadata.RequestID {
		t.Errorf("Metadata.RequestID = %v, want %v", response.Metadata.RequestID, metadata.RequestID)
	}
}

func TestNewErrorResponse(t *testing.T) {
	response := NewErrorResponse(ErrorCodeNotFound, "Resource not found")

	if response.Success {
		t.Errorf("Success should be false")
	}
	if response.Error.Code != ErrorCodeNotFound {
		t.Errorf("Error.Code = %v, want %v", response.Error.Code, ErrorCodeNotFound)
	}
	if response.Error.Message != "Resource not found" {
		t.Errorf("Error.Message = %v, want %v", response.Error.Message, "Resource not found")
	}
}

func TestNewErrorResponseWithDetails(t *testing.T) {
	details := map[string]interface{}{"field": "username", "reason": "required"}
	response := NewErrorResponseWithDetails(ErrorCodeValidation, "Validation failed", details)

	if response.Success {
		t.Errorf("Success should be false")
	}
	if response.Error.Code != ErrorCodeValidation {
		t.Errorf("Error.Code = %v, want %v", response.Error.Code, ErrorCodeValidation)
	}
	if response.Error.Details["field"] != "username" {
		t.Errorf("Error.Details[field] = %v, want %v", response.Error.Details["field"], "username")
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"item1", "item2"}
	pagination := &Pagination{Page: 1, PerPage: 10, Total: 25}

	response := NewPaginatedResponse(data, pagination)

	if !response.Success {
		t.Errorf("Success should be true")
	}
	if !reflect.DeepEqual(response.Data, data) {
		t.Errorf("Data = %v, want %v", response.Data, data)
	}
	if response.Pagination.Page != 1 {
		t.Errorf("Pagination.Page = %v, want %v", response.Pagination.Page, 1)
	}
}

func TestNewPagination(t *testing.T) {
	pagination := NewPagination(2, 20, 100)

	if pagination.Page != 2 {
		t.Errorf("Page = %v, want %v", pagination.Page, 2)
	}
	if pagination.PerPage != 20 {
		t.Errorf("PerPage = %v, want %v", pagination.PerPage, 20)
	}
	if pagination.Total != 100 {
		t.Errorf("Total = %v, want %v", pagination.Total, 100)
	}
	if !pagination.HasNext {
		t.Errorf("HasNext should be true")
	}
	if !pagination.HasPrev {
		t.Errorf("HasPrev should be true")
	}
}

func TestNewResponseMetadata(t *testing.T) {
	metadata := NewResponseMetadata("req-456", "2.1.0")

	if metadata.RequestID != "req-456" {
		t.Errorf("RequestID = %v, want %v", metadata.RequestID, "req-456")
	}
	if metadata.Version != "2.1.0" {
		t.Errorf("Version = %v, want %v", metadata.Version, "2.1.0")
	}
	if metadata.Timestamp.Time.IsZero() {
		t.Errorf("Timestamp should not be zero")
	}
}

func TestErrorDetailsValidate(t *testing.T) {
	tests := []struct {
		name    string
		details *ErrorDetails
		wantErr bool
	}{
		{
			name: "valid error details",
			details: &ErrorDetails{
				Code:    ErrorCodeNotFound,
				Message: "Resource not found",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			details: &ErrorDetails{
				Code:    "",
				Message: "Resource not found",
			},
			wantErr: true,
		},
		{
			name: "empty message",
			details: &ErrorDetails{
				Code:    ErrorCodeNotFound,
				Message: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.details.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ErrorDetails.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseMetadataValidate(t *testing.T) {
	tests := []struct {
		name     string
		metadata *ResponseMetadata
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: &ResponseMetadata{
				RequestID: "req-123",
				Version:   "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty request ID",
			metadata: &ResponseMetadata{
				RequestID: "",
				Version:   "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			metadata: &ResponseMetadata{
				RequestID: "req-123",
				Version:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.metadata.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ResponseMetadata.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaginationValidate(t *testing.T) {
	tests := []struct {
		name       string
		pagination *Pagination
		wantErr    bool
	}{
		{
			name: "valid pagination",
			pagination: &Pagination{
				Page:    1,
				PerPage: 10,
				Total:   100,
			},
			wantErr: false,
		},
		{
			name: "zero page",
			pagination: &Pagination{
				Page:    0,
				PerPage: 10,
				Total:   100,
			},
			wantErr: true,
		},
		{
			name: "zero perPage",
			pagination: &Pagination{
				Page:    1,
				PerPage: 0,
				Total:   100,
			},
			wantErr: true,
		},
		{
			name: "negative total",
			pagination: &Pagination{
				Page:    1,
				PerPage: 10,
				Total:   -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pagination.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Pagination.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseJSONMarshal(t *testing.T) {
	// Test SuccessResponse
	successResp := NewSuccessResponse(map[string]string{"status": "ok"})
	data, err := json.Marshal(successResp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaledSuccess SuccessResponse
	err = json.Unmarshal(data, &unmarshaledSuccess)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !unmarshaledSuccess.Success {
		t.Errorf("Success should be true")
	}

	// Test ErrorResponse
	errorResp := NewErrorResponse(ErrorCodeNotFound, "Not found")
	data, err = json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaledError ErrorResponse
	err = json.Unmarshal(data, &unmarshaledError)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaledError.Success {
		t.Errorf("Success should be false")
	}
	if unmarshaledError.Error.Code != ErrorCodeNotFound {
		t.Errorf("Error.Code = %v, want %v", unmarshaledError.Error.Code, ErrorCodeNotFound)
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined
	constants := map[string]string{
		ErrorCodeValidation:    "VALIDATION_ERROR",
		ErrorCodeNotFound:      "NOT_FOUND",
		ErrorCodeUnauthorized:  "UNAUTHORIZED",
		ErrorCodeForbidden:     "FORBIDDEN",
		ErrorCodeInternalError: "INTERNAL_ERROR",
		ErrorCodeBadRequest:    "BAD_REQUEST",
	}

	for name, expected := range constants {
		if name != expected {
			t.Errorf("Constant = %v, want %v", name, expected)
		}
	}
}
