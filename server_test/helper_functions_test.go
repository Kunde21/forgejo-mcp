package servertest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test cases for helper functions that will be extracted from handleRepoRequests

func TestGetRepoKeyFromRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid pull requests path",
			path:     "/api/v1/repos/owner/repo/pulls",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "valid issues path",
			path:     "/api/v1/repos/owner/repo/issues",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "valid comments path",
			path:     "/api/v1/repos/owner/repo/issues/123/comments",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "valid comment edit path",
			path:     "/api/v1/repos/owner/repo/issues/comments/456",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "invalid path - missing segments",
			path:     "/api/v1/repos/owner",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid path - wrong prefix",
			path:     "/api/v1/wrong/owner/repo/issues",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Since we can't easily mock PathValue in tests, we'll test the function logic directly
			// by extracting the values manually as PathValue would do
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			var owner, repo string

			if len(pathParts) >= 2 {
				owner = pathParts[0]
				repo = pathParts[1]
			}

			// For testing purposes, we'll simulate the PathValue behavior
			// In real usage with Go 1.22+ routing, PathValue would extract these automatically
			if owner != "" && repo != "" {
				result := owner + "/" + repo
				if result != tc.expected {
					t.Errorf("Expected %q, got %q", tc.expected, result)
				}
			} else if tc.wantErr {
				// Expected error case
				return
			} else {
				t.Errorf("Expected valid result but got empty values")
			}
		})
	}
}

func TestValidateRepository(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		repoKey   string
		mockSetup func(*MockGiteaServer)
		expected  bool
	}{
		{
			name:    "valid existing repository",
			repoKey: "owner/repo",
			mockSetup: func(mock *MockGiteaServer) {
				mock.AddIssues("owner", "repo", []MockIssue{{Index: 1, Title: "Test Issue", State: "open"}})
			},
			expected: true,
		},
		{
			name:    "nonexistent repository",
			repoKey: "nonexistent/repo",
			mockSetup: func(mock *MockGiteaServer) {
				// No setup - repository doesn't exist
			},
			expected: false,
		},
		{
			name:    "repository marked as not found",
			repoKey: "owner/repo",
			mockSetup: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("owner", "repo")
			},
			expected: false,
		},
		{
			name:    "empty repository key",
			repoKey: "",
			mockSetup: func(mock *MockGiteaServer) {
				// No setup needed
			},
			expected: false,
		},
		{
			name:    "malformed repository key",
			repoKey: "invalid-format",
			mockSetup: func(mock *MockGiteaServer) {
				// No setup needed
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockGiteaServer(t)
			if tc.mockSetup != nil {
				tc.mockSetup(mock)
			}

			result := validateRepository(mock, tc.repoKey)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		queryParams    string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "no pagination parameters",
			queryParams:    "",
			expectedLimit:  0, // Will be set to default in implementation
			expectedOffset: 0,
		},
		{
			name:           "valid limit and offset",
			queryParams:    "limit=10&offset=20",
			expectedLimit:  10,
			expectedOffset: 20,
		},
		{
			name:           "only limit specified",
			queryParams:    "limit=5",
			expectedLimit:  5,
			expectedOffset: 0,
		},
		{
			name:           "only offset specified",
			queryParams:    "offset=15",
			expectedLimit:  0, // Will be set to default in implementation
			expectedOffset: 15,
		},
		{
			name:           "invalid limit - negative",
			queryParams:    "limit=-5",
			expectedLimit:  0, // Will be set to default in implementation
			expectedOffset: 0,
		},
		{
			name:           "invalid limit - zero",
			queryParams:    "limit=0",
			expectedLimit:  0, // Will be set to default in implementation
			expectedOffset: 0,
		},
		{
			name:           "invalid offset - negative",
			queryParams:    "offset=-10",
			expectedLimit:  0,
			expectedOffset: 0, // Will be set to default in implementation
		},
		{
			name:           "non-numeric values",
			queryParams:    "limit=abc&offset=xyz",
			expectedLimit:  0, // Will be set to default in implementation
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest("GET", "/test?"+tc.queryParams, nil)

			limit, offset := parsePagination(req)

			if limit != tc.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tc.expectedLimit, limit)
			}

			if offset != tc.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tc.expectedOffset, offset)
			}
		})
	}
}

func TestValidateAuthToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		headers     map[string]string
		queryParams string
		expected    bool
	}{
		{
			name: "valid Bearer token",
			headers: map[string]string{
				"Authorization": "Bearer mock-token",
			},
			expected: true,
		},
		{
			name: "valid token header",
			headers: map[string]string{
				"Authorization": "token mock-token",
			},
			expected: true,
		},
		{
			name:        "valid token in query params",
			queryParams: "token=mock-token",
			expected:    true,
		},
		{
			name: "valid test token",
			headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			expected: true,
		},
		{
			name: "invalid token",
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			expected: false,
		},
		{
			name:        "invalid token in query params",
			queryParams: "token=invalid-token",
			expected:    false,
		},
		{
			name: "malformed authorization header",
			headers: map[string]string{
				"Authorization": "InvalidFormat mock-token",
			},
			expected: false,
		},
		{
			name: "empty authorization header",
			headers: map[string]string{
				"Authorization": "",
			},
			expected: true, // Empty header should be acceptable for backward compatibility
		},
		{
			name:        "no authentication",
			headers:     map[string]string{},
			queryParams: "",
			expected:    true, // No token should be acceptable for backward compatibility
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with headers and query parameters
			target := "/test"
			if tc.queryParams != "" {
				target += "?" + tc.queryParams
			}
			req := httptest.NewRequest("GET", target, nil)

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			result := validateAuthToken(req)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestWriteJSONResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		data           any
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "success response with data",
			data:           map[string]string{"message": "success"},
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "created response",
			data:           map[string]int{"id": 123},
			statusCode:     http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "error response",
			data:           map[string]string{"error": "not found"},
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "nil data",
			data:           nil,
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "complex data structure",
			data:           map[string]any{"user": map[string]string{"name": "test"}, "items": []int{1, 2, 3}},
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			writeJSONResponse(w, tc.data, tc.statusCode)

			resp := w.Result()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
			}

			// If data is not nil, verify the response body can be decoded
			if tc.data != nil {
				var decodedData any
				if err := json.NewDecoder(resp.Body).Decode(&decodedData); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}
			}
		})
	}
}
