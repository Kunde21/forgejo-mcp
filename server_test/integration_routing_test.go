package servertest

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Integration tests to ensure all functionality still works after the refactoring

type integrationTestCase struct {
	name           string
	method         string
	path           string
	headers        map[string]string
	queryParams    map[string]string
	body           string
	setupMock      func(*MockGiteaServer)
	expectedStatus int
	validateBody   func(t *testing.T, body string)
}

func TestIntegrationAllFunctionalityWorks(t *testing.T) {
	t.Parallel()

	testCases := []integrationTestCase{
		// Test 1: Pull Requests - Full workflow
		{
			name:   "pull_requests_full_workflow",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/pulls",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
					{ID: 1, Number: 1, Title: "Feature A", State: "open"},
					{ID: 2, Number: 2, Title: "Bug Fix B", State: "closed"},
					{ID: 3, Number: 3, Title: "Feature C", State: "open"},
				})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var prs []map[string]interface{}
				if err := json.Unmarshal([]byte(body), &prs); err != nil {
					t.Fatalf("Failed to unmarshal pull requests: %v", err)
				}
				if len(prs) != 3 {
					t.Errorf("Expected 3 pull requests, got %d", len(prs))
				}
				// Check that all expected PRs are present
				foundFeatureA := false
				foundBugFixB := false
				foundFeatureC := false
				for _, pr := range prs {
					title, ok := pr["title"].(string)
					if !ok {
						t.Errorf("Pull request title is not a string")
						continue
					}
					switch title {
					case "Feature A":
						foundFeatureA = true
					case "Bug Fix B":
						foundBugFixB = true
					case "Feature C":
						foundFeatureC = true
					}
				}
				if !foundFeatureA || !foundBugFixB || !foundFeatureC {
					t.Errorf("Not all expected pull requests found")
				}
			},
		},

		// Test 2: Pull Requests - State filtering
		{
			name:        "pull_requests_state_filtering",
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
			validateBody: func(t *testing.T, body string) {
				var prs []map[string]interface{}
				if err := json.Unmarshal([]byte(body), &prs); err != nil {
					t.Fatalf("Failed to unmarshal pull requests: %v", err)
				}
				if len(prs) != 1 {
					t.Errorf("Expected 1 open pull request, got %d", len(prs))
				}
				title, ok := prs[0]["title"].(string)
				if !ok || title != "Open PR" {
					t.Errorf("Expected 'Open PR', got %v", prs[0]["title"])
				}
			},
		},

		// Test 3: Issues - Full workflow
		{
			name:   "issues_full_workflow",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddIssues("testuser", "testrepo", []MockIssue{
					{Index: 1, Title: "Issue 1", State: "open"},
					{Index: 2, Title: "Issue 2", State: "closed"},
				})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var issues []MockIssue
				if err := json.Unmarshal([]byte(body), &issues); err != nil {
					t.Fatalf("Failed to unmarshal issues: %v", err)
				}
				if len(issues) != 2 {
					t.Errorf("Expected 2 issues, got %d", len(issues))
				}
			},
		},

		// Test 4: Comments - Create and List workflow
		{
			name:   "comments_create_and_list_workflow",
			method: "POST",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:   `{"body": "This is a test comment"}`,
			setupMock: func(mock *MockGiteaServer) {
				// No initial setup
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body string) {
				var comment map[string]interface{}
				if err := json.Unmarshal([]byte(body), &comment); err != nil {
					t.Fatalf("Failed to unmarshal comment: %v", err)
				}
				bodyText, ok := comment["body"].(string)
				if !ok || bodyText != "This is a test comment" {
					t.Errorf("Expected comment body 'This is a test comment', got %v", comment["body"])
				}
			},
		},

		// Test 5: Comments - List after creation
		{
			name:   "comments_list_after_creation",
			method: "GET",
			path:   "/api/v1/repos/testuser/testrepo/issues/1/comments",
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "First comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
					{ID: 2, Content: "Second comment", Author: "testuser", Created: "2025-09-09T10:31:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var comments []map[string]interface{}
				if err := json.Unmarshal([]byte(body), &comments); err != nil {
					t.Fatalf("Failed to unmarshal comments: %v", err)
				}
				if len(comments) != 2 {
					t.Errorf("Expected 2 comments, got %d", len(comments))
				}
			},
		},

		// Test 6: Comments - Edit workflow
		{
			name:   "comments_edit_workflow",
			method: "PATCH",
			path:   "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:   `{"body": "Updated comment content"}`,
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var comment map[string]interface{}
				if err := json.Unmarshal([]byte(body), &comment); err != nil {
					t.Fatalf("Failed to unmarshal comment: %v", err)
				}
				bodyText, ok := comment["body"].(string)
				if !ok || bodyText != "Updated comment content" {
					t.Errorf("Expected comment body 'Updated comment content', got %v", comment["body"])
				}
			},
		},

		// Test 7: Authentication - Invalid token
		{
			name:    "authentication_invalid_token",
			method:  "PATCH",
			path:    "/api/v1/repos/testuser/testrepo/issues/comments/1",
			body:    `{"body": "Should not work"}`,
			headers: map[string]string{"Authorization": "Bearer invalid-token"},
			setupMock: func(mock *MockGiteaServer) {
				mock.AddComments("testuser", "testrepo", []MockComment{
					{ID: 1, Content: "Original comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
				})
			},
			expectedStatus: http.StatusUnauthorized,
		},

		// Test 8: Error handling - Repository not found
		{
			name:   "error_handling_repository_not_found",
			method: "GET",
			path:   "/api/v1/repos/nonexistent/repo/issues",
			setupMock: func(mock *MockGiteaServer) {
				mock.SetNotFoundRepo("nonexistent", "repo")
			},
			expectedStatus: http.StatusNotFound,
		},

		// Test 9: Error handling - Invalid JSON
		{
			name:           "error_handling_invalid_json",
			method:         "POST",
			path:           "/api/v1/repos/testuser/testrepo/issues/1/comments",
			body:           `{"body": "missing quote}`,
			expectedStatus: http.StatusBadRequest,
		},

		// Test 10: Version endpoint - Still works
		{
			name:           "version_endpoint_still_works",
			method:         "GET",
			path:           "/api/v1/version",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var version map[string]string
				if err := json.Unmarshal([]byte(body), &version); err != nil {
					t.Fatalf("Failed to unmarshal version: %v", err)
				}
				if version["version"] != "1.20.0" {
					t.Errorf("Expected version '1.20.0', got %v", version["version"])
				}
			},
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

			// Check response status
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Validate response body if validator is provided
			if tc.validateBody != nil {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				tc.validateBody(t, string(body))
			}
		})
	}
}

// Test that the refactoring maintains backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	t.Parallel()

	// This test ensures that all the existing functionality still works
	// exactly as it did before the refactoring
	mock := NewMockGiteaServer(t)

	// Set up test data
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue", State: "open"},
	})
	mock.AddComments("testuser", "testrepo", []MockComment{
		{ID: 1, Content: "Test Comment", Author: "testuser", Created: "2025-09-09T10:30:00Z"},
	})

	// Test all endpoints that should work
	endpoints := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/v1/repos/testuser/testrepo/pulls", http.StatusOK},
		{"GET", "/api/v1/repos/testuser/testrepo/issues", http.StatusOK},
		{"GET", "/api/v1/repos/testuser/testrepo/issues/1/comments", http.StatusOK},
		{"GET", "/api/v1/version", http.StatusOK},
		{"POST", "/api/v1/repos/testuser/testrepo/issues/1/comments", http.StatusCreated}, // Will fail without body but should return 400, not 404
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+"_"+endpoint.path, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			var err error

			if endpoint.method == "POST" {
				req, err = http.NewRequest(endpoint.method, mock.URL()+endpoint.path, strings.NewReader(`{"body": "test"}`))
			} else {
				req, err = http.NewRequest(endpoint.method, mock.URL()+endpoint.path, nil)
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != endpoint.status {
				t.Errorf("Expected status %d, got %d", endpoint.status, resp.StatusCode)
			}
		})
	}
}
