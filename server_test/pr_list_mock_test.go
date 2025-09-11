package servertest

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServerPullRequestListing tests basic mock server pull request listing functionality
func TestMockServerPullRequestListing(t *testing.T) {
	// Create mock server
	mock := NewMockGiteaServer(t)

	// Add test pull requests
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Feature: Add dark mode", State: "open"},
		{ID: 2, Number: 2, Title: "Fix: Memory leak", State: "open"},
		{ID: 3, Number: 3, Title: "Bug: Login fails", State: "closed"},
	})

	// Test direct HTTP request to mock server
	resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Parse response
	var prs []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prs)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, prs, 3)

	// Verify first PR
	assert.Equal(t, float64(1), prs[0]["id"])
	assert.Equal(t, float64(1), prs[0]["number"])
	assert.Equal(t, "Feature: Add dark mode", prs[0]["title"])
	assert.Equal(t, "open", prs[0]["state"])

	// Verify user structure
	user, ok := prs[0]["user"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "testuser", user["login"])

	// Verify branch structure
	head, ok := prs[0]["head"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "feature-branch", head["ref"])
	assert.Equal(t, "abc123", head["sha"])
}

// TestMockServerPullRequestStateFiltering tests state filtering functionality
func TestMockServerPullRequestStateFiltering(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add test pull requests with different states
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
		{ID: 3, Number: 3, Title: "Closed PR 1", State: "closed"},
		{ID: 4, Number: 4, Title: "Closed PR 2", State: "closed"},
	})

	testCases := []struct {
		name     string
		state    string
		expected int
	}{
		{"Open state", "open", 2},
		{"Closed state", "closed", 2},
		{"No state filter", "", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/testuser/testrepo/pulls"
			if tc.state != "" {
				url += "?state=" + tc.state
			}

			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var prs []map[string]any
			err = json.NewDecoder(resp.Body).Decode(&prs)
			require.NoError(t, err)

			assert.Len(t, prs, tc.expected)
		})
	}
}

// TestMockServerPullRequestAllStateFiltering tests "all" state filtering
func TestMockServerPullRequestAllStateFiltering(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Open PR", State: "open"},
		{ID: 2, Number: 2, Title: "Closed PR", State: "closed"},
	})

	resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls?state=all")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prs []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prs)
	require.NoError(t, err)

	assert.Len(t, prs, 2)
}

// TestMockServerPullRequestNonExistentRepo tests non-existent repository scenario
func TestMockServerPullRequestNonExistentRepo(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Don't add any pull requests for this repo

	resp, err := http.Get(mock.URL() + "/api/v1/repos/nonexistent/repo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prs []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prs)
	require.NoError(t, err)

	// Should return empty array, not 404
	assert.Empty(t, prs)
}

// TestMockServerPullRequestInvalidPath tests invalid path handling
func TestMockServerPullRequestInvalidPath(t *testing.T) {
	mock := NewMockGiteaServer(t)

	testCases := []string{
		"/api/v1/repos/invalid-format/pulls",
		"/api/v1/repos/testuser/pulls",
		"/api/v1/repos//pulls",
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			resp, err := http.Get(mock.URL() + path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

// TestMockServerPullRequestInvalidMethod tests invalid HTTP method handling
func TestMockServerPullRequestInvalidMethod(t *testing.T) {
	mock := NewMockGiteaServer(t)

	resp, err := http.Post(mock.URL()+"/api/v1/repos/testuser/testrepo/pulls", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMockServerPullRequestEmptyRepository tests empty repository name handling
func TestMockServerPullRequestEmptyRepository(t *testing.T) {
	mock := NewMockGiteaServer(t)

	resp, err := http.Get(mock.URL() + "/api/v1/repos//pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMockServerPullRequestSpecialCharacters tests repository names with special characters
func TestMockServerPullRequestSpecialCharacters(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("test.user", "test-repo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "PR with special chars", State: "open"},
	})

	resp, err := http.Get(mock.URL() + "/api/v1/repos/test.user/test-repo/pulls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prs []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prs)
	require.NoError(t, err)

	assert.Len(t, prs, 1)
	assert.Equal(t, "PR with special chars", prs[0]["title"])
}

// TestMockServerPullRequestCaseSensitivity tests case sensitivity in state filtering
func TestMockServerPullRequestCaseSensitivity(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Open PR", State: "open"},
		{ID: 2, Number: 2, Title: "Closed PR", State: "closed"},
	})

	testCases := []struct {
		state    string
		expected int
	}{
		{"open", 1},
		{"OPEN", 0}, // Should be case sensitive
		{"Open", 0}, // Should be case sensitive
		{"closed", 1},
		{"CLOSED", 0}, // Should be case sensitive
	}

	for _, tc := range testCases {
		t.Run("State_"+tc.state, func(t *testing.T) {
			resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls?state=" + tc.state)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var prs []map[string]any
			err = json.NewDecoder(resp.Body).Decode(&prs)
			require.NoError(t, err)

			assert.Len(t, prs, tc.expected)
		})
	}
}
