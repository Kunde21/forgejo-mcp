package servertest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServerPullRequestErrorScenarios tests various error scenarios
func TestMockServerPullRequestErrorScenarios(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add test pull requests
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})

	testCases := []struct {
		name           string
		repository     string
		expectedStatus int
		setupFunc      func(*MockGiteaServer)
	}{
		{
			name:           "Valid repository",
			repository:     "testuser/testrepo",
			expectedStatus: http.StatusOK,
			setupFunc:      nil,
		},
		{
			name:           "Repository not found",
			repository:     "nonexistent/repo",
			expectedStatus: http.StatusNotFound,
			setupFunc: func(m *MockGiteaServer) {
				// Mark repository as not found
				m.SetNotFoundRepo("nonexistent", "repo")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				tc.setupFunc(mock)
			}

			resp, err := http.Get(mock.URL() + "/api/v1/repos/" + tc.repository + "/pulls")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestMockServerPullRequestAuthenticationError tests authentication error scenarios
func TestMockServerPullRequestAuthenticationError(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("private", "repo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Private PR", State: "open"},
	})

	// Test without authentication token
	resp, err := http.Get(mock.URL() + "/api/v1/repos/private/repo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Currently mock server doesn't enforce authentication, so it should succeed
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMockServerPullRequestServerError tests server error scenarios
func TestMockServerPullRequestServerError(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Test with a repository name that might trigger server errors
	// This is a placeholder for more sophisticated error simulation
	resp, err := http.Get(mock.URL() + "/api/v1/repos/error/repo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Currently returns empty array for non-existent repos
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMockServerPullRequestMalformedRequest tests malformed request handling
func TestMockServerPullRequestMalformedRequest(t *testing.T) {
	mock := NewMockGiteaServer(t)

	testCases := []struct {
		name           string
		url            string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST request",
			url:            "/api/v1/repos/testuser/testrepo/pulls",
			method:         "POST",
			body:           "{}",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT request",
			url:            "/api/v1/repos/testuser/testrepo/pulls",
			method:         "PUT",
			body:           "{}",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "DELETE request",
			url:            "/api/v1/repos/testuser/testrepo/pulls",
			method:         "DELETE",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PATCH request",
			url:            "/api/v1/repos/testuser/testrepo/pulls",
			method:         "PATCH",
			body:           "{}",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{}

			var req *http.Request
			var err error

			if tc.method == "POST" || tc.method == "PUT" || tc.method == "PATCH" {
				req, err = http.NewRequest(tc.method, mock.URL()+tc.url, strings.NewReader(tc.body))
			} else {
				req, err = http.NewRequest(tc.method, mock.URL()+tc.url, nil)
			}

			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestMockServerPullRequestInvalidQueryParameters tests invalid query parameter handling
func TestMockServerPullRequestInvalidQueryParameters(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Test PR", State: "open"},
	})

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Invalid state parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?state=invalid",
			expectedStatus: http.StatusOK,
			expectedCount:  0, // Should return empty array for invalid state
		},
		{
			name:           "Empty state parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?state=",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should return all PRs when state is empty
		},
		{
			name:           "Multiple state parameters",
			url:            "/api/v1/repos/testuser/testrepo/pulls?state=open&state=closed",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should use first parameter
		},
		{
			name:           "Valid state with extra parameters",
			url:            "/api/v1/repos/testuser/testrepo/pulls?state=open&sort=created&direction=desc",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(mock.URL() + tc.url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			var prs []map[string]any
			err = json.NewDecoder(resp.Body).Decode(&prs)
			require.NoError(t, err)

			assert.Len(t, prs, tc.expectedCount)
		})
	}
}

// TestMockServerPullRequestLargeDataset tests handling of large datasets
func TestMockServerPullRequestLargeDataset(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add many pull requests
	var prs []MockPullRequest
	for i := 1; i <= 100; i++ {
		prs = append(prs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  fmt.Sprintf("Pull Request %d", i),
			State:  "open",
		})
	}
	mock.AddPullRequests("testuser", "large-repo", prs)

	resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/large-repo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prsResponse []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prsResponse)
	require.NoError(t, err)

	assert.Len(t, prsResponse, 100)
}

// TestMockServerPullRequestConcurrentAccess tests concurrent access to mock server
func TestMockServerPullRequestConcurrentAccess(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Concurrent PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Concurrent PR 2", State: "closed"},
	})

	// Make concurrent requests
	const numRequests = 10
	results := make(chan *http.Response, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls")
			if err != nil {
				errors <- err
				return
			}
			results <- resp
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case resp := <-results:
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		case err := <-errors:
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
