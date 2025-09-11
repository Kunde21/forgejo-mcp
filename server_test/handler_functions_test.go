package servertest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test cases for individual handler functions

type handlerTestCase struct {
	name           string
	method         string
	path           string
	headers        map[string]string
	queryParams    map[string]string
	body           string
	setupMock      func(*MockGiteaServer)
	expectedStatus int
	expectedBody   string
	contains       string
}

func TestHandlePullRequests(t *testing.T) {
	t.Parallel()

	testCases := []handlerTestCase{
		{
			name:   "get_pull_requests_success",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/pulls",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "Test PR 2", State: "closed"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test PR 1",
		},
		{
			name:   "get_pull_requests_empty",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/pulls",
			setupMock: func(mock *MockGiteaServer) {
				// No pull requests added
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]\n",
		},
		{
			name:   "get_pull_requests_not_found",
			method: "GET",
			path:   "/api/v1/repos/nonexistent/repo/pulls",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "get_pull_requests_with_state_filter",
			method:      "GET",
			path:        "/api/v1/repos/testuser/testrepo/pulls",
			queryParams: map[string]string{"state": "open"},
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Open PR", State: "open"},
					{ID: 2, Number: 2, Title: "Closed PR", State: "closed"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Open PR",
		},
		{
			name:        "get_pull_requests_with_pagination",
			method:      "GET",
			path:        "/api/v1/repos/testuser/testrepo/pulls",
			queryParams: map[string]string{"limit": "1", "offset": "1"},
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "PR 1", State: "open"},
					{ID: 2, Number: 2, Title: "PR 2", State: "open"},
					{ID: 3, Number: 3, Title: "PR 3", State: "open"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "PR 2",
		},
		{
			name:           "get_pull_requests_wrong_method",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/pulls",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}
			if tc.queryParams != nil {
				q := req.URL.Query()
				for k, v := range tc.queryParams {
					q.Set(k, v)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Set path parameters manually for direct handler testing
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			if len(pathParts) >= 2 {
				req.SetPathValue("owner", pathParts[0])
				req.SetPathValue("repo", pathParts[1])
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			mock.handlePullRequests(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body := w.Body.String()
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			}

			if tc.contains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, body)
				}
			}
		})
	}
}

func TestHandleIssues(t *testing.T) {
	t.Parallel()

	testCases := []handlerTestCase{
		{
			name:   "get_issues_success",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Test Issue 1", State: "open"},
					{Index: 2, Title: "Test Issue 2", State: "closed"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test Issue 1",
		},
		{
			name:   "get_issues_empty",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues",
			setupMock: func(mock *MockGiteaServer) {
				// No issues added
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get_issues_wrong_method",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}

			// Set path parameters manually for direct handler testing
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			if len(pathParts) >= 2 {
				req.SetPathValue("owner", pathParts[0])
				req.SetPathValue("repo", pathParts[1])
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			mock.handleIssues(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body := w.Body.String()
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			}

			if tc.contains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, body)
				}
			}
		})
	}
}

func TestHandleCreateComment(t *testing.T) {
	t.Parallel()

	testCases := []handlerTestCase{
		{
			name:   "create_comment_success",
			method: "POST",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:   `{"body": "This is a test comment"}`,
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for successful creation
			},
			expectedStatus: http.StatusCreated,
			contains:       "This is a test comment",
		},
		{
			name:           "create_comment_invalid_json",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:           `{"body": "missing quote}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "create_comment_nonexistent_repo",
			method:         "POST",
			path:           "/api/v1/repos/nonexistent/repo/issues/1/comments",
			body:           `{"body": "This is a test comment"}`,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "create_comment_wrong_method",
			method:         "GET",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}

			// Set path parameters manually for direct handler testing
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			if len(pathParts) >= 4 {
				req.SetPathValue("owner", pathParts[0])
				req.SetPathValue("repo", pathParts[1])
				req.SetPathValue("number", pathParts[3])
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			mock.handleCreateComment(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body := w.Body.String()
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			}

			if tc.contains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, body)
				}
			}
		})
	}
}

func TestHandleListComments(t *testing.T) {
	t.Parallel()

	testCases := []handlerTestCase{
		{
			name:   "list_comments_success",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment 1", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
					{ID: 2, Content: "Test comment 2", Author: "testuser", Created: "2025-09-09T10:31:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test comment 1",
		},
		{
			name:   "list_comments_empty",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			setupMock: func(mock *MockGiteaServer) {
				// No comments added
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]\n",
		},
		{
			name:   "list_comments_not_found_repo",
			method: "GET",
			path:   "/api/v1/repos/nonexistent/repo/issues/1/comments",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "list_comments_wrong_method",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}

			// Set path parameters manually for direct handler testing
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			if len(pathParts) >= 4 {
				req.SetPathValue("owner", pathParts[0])
				req.SetPathValue("repo", pathParts[1])
				req.SetPathValue("number", pathParts[3])
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			mock.handleListComments(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body := w.Body.String()
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			}

			if tc.contains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, body)
				}
			}
		})
	}
}

func TestHandleEditComment(t *testing.T) {
	t.Parallel()

	testCases := []handlerTestCase{
		{
			name:   "edit_comment_success",
			method: "PATCH",
			path:   "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:   `{"body": "Updated comment content"}`,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Updated comment content",
		},
		{
			name:           "edit_comment_invalid_token",
			method:         "PATCH",
			path:           "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:           `{"body": "Updated comment content"}`,
			headers:        map[string]string{"Authorization": "Bearer invalid-token"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "edit_comment_invalid_json",
			method:         "PATCH",
			path:           "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:           `{"body": "missing quote}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "edit_comment_nonexistent_repo",
			method:         "PATCH",
			path:           "/api/v1/repos/nonexistent/repo/issues/comments/1",
			body:           `{"body": "Updated comment content"}`,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "edit_comment_wrong_method",
			method:         "GET",
			path:           "/api/v1/repos/testuser/testrepo/issues/comments/1",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			mock := NewMockGiteaServer(t)
			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}

			// Set path parameters manually for direct handler testing
			pathParts := strings.Split(strings.TrimPrefix(tc.path, "/api/v1/repos/"), "/")
			if len(pathParts) >= 5 {
				req.SetPathValue("owner", pathParts[0])
				req.SetPathValue("repo", pathParts[1])
				req.SetPathValue("id", pathParts[4])
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			mock.handleEditComment(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body := w.Body.String()
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			}

			if tc.contains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, body)
				}
			}
		})
	}
}

// Helper function to validate JSON response structure
func validateJSONResponse(t *testing.T, body string, expectedFields []string) {
	t.Helper()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		t.Fatalf("Response should be valid JSON: %v", err)
	}

	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("Response should contain field: %s", field)
		}
	}
}

// Helper function to validate JSON array response structure
func validateJSONArrayResponse(t *testing.T, body string, expectedFields []string) {
	t.Helper()

	var data []interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		t.Fatalf("Response should be valid JSON array: %v", err)
	}

	if len(data) > 0 {
		firstItem, ok := data[0].(map[string]interface{})
		if !ok {
			t.Fatalf("First item should be an object")
		}

		for _, field := range expectedFields {
			if _, exists := firstItem[field]; !exists {
				t.Errorf("Response items should contain field: %s", field)
			}
		}
	}
}
