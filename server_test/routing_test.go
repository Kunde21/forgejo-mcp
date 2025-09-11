package servertest

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// Test cases for new routing registration patterns

type routingTestCase struct {
	name           string
	method         string
	path           string
	headers        map[string]string
	queryParams    map[string]string
	body           string
	setupMock      func(*MockGiteaServer)
	expectedStatus int
	contains       string
}

func TestNewRoutingPatterns(t *testing.T) {
	t.Parallel()

	testCases := []routingTestCase{
		// Pull Requests routing
		{
			name:   "pull_requests_get_routing",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/pulls",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Test PR", State: "open"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test PR",
		},
		{
			name:           "pull_requests_post_routing_not_found",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/pulls",
			expectedStatus: http.StatusMethodNotAllowed,
		},

		// Issues routing
		{
			name:   "issues_get_routing",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Test Issue", State: "open"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test Issue",
		},
		{
			name:           "issues_post_routing_not_found",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues",
			expectedStatus: http.StatusMethodNotAllowed,
		},

		// Create Comment routing
		{
			name:   "create_comment_post_routing",
			method: "POST",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:   `{"body": "Test comment"}`,
			setupMock: func(mock *MockGiteaServer) {
				// No setup needed for successful creation
			},
			expectedStatus: http.StatusCreated,
			contains:       "Test comment",
		},
		{
			name:           "create_comment_get_routing_not_found",
			method:         "GET",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			expectedStatus: http.StatusOK, // This should work because GET is handled by handleListComments
		},

		// List Comments routing
		{
			name:   "list_comments_get_routing",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Test comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Test comment",
		},
		{
			name:           "create_comment_routing_works",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:           `{"body": "test"}`,
			expectedStatus: http.StatusCreated,
		},

		// Edit Comment routing
		{
			name:   "edit_comment_patch_routing",
			method: "PATCH",
			path:   "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:   `{"body": "Updated comment"}`,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			contains:       "Updated comment",
		},
		{
			name:           "edit_comment_get_routing_not_found",
			method:         "GET",
			path:           "/api/v1/repos/testuser/testrepo/issues/comments/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},

		// Invalid path routing
		{
			name:           "invalid_path_routing",
			method:         "GET",
			path:           "/api/v1/repos/testuser/testrepo/invalid",
			expectedStatus: http.StatusNotFound,
		},

		// Version endpoint should still work
		{
			name:           "version_endpoint_still_works",
			method:         "GET",
			path:           "/api/v1/version",
			expectedStatus: http.StatusOK,
			contains:       "1.20.0",
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
			var req *http.Request
			var err error

			if tc.body == "" {
				req, err = http.NewRequest(tc.method, mock.URL()+tc.path, nil)
			} else {
				req, err = http.NewRequest(tc.method, mock.URL()+tc.path, strings.NewReader(tc.body))
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

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

			// Send request to the server
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Check response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.contains != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				if !strings.Contains(string(body), tc.contains) {
					t.Errorf("Expected body to contain %q, got %q", tc.contains, string(body))
				}
			}
		})
	}
}

// Test that the new routing will work (this test will be updated after implementing new routing)
func TestRoutingImplementation(t *testing.T) {
	t.Parallel()

	// This test will verify that the new routing patterns work correctly
	// For now, it just tests that the current setup works
	mock := NewMockGiteaServer(t)

	// Add test data
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})

	// Test pull requests endpoint
	resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "Test PR") {
		t.Errorf("Expected body to contain 'Test PR', got %q", string(body))
	}
}
