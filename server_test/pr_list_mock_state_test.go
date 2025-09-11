package servertest

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServerPullRequestStateFilteringComprehensive tests comprehensive state filtering scenarios
func TestMockServerPullRequestStateFilteringComprehensive(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add test pull requests with various states
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Open PR 1", State: "open"},
		{ID: 2, Number: 2, Title: "Open PR 2", State: "open"},
		{ID: 3, Number: 3, Title: "Open PR 3", State: "open"},
		{ID: 4, Number: 4, Title: "Closed PR 1", State: "closed"},
		{ID: 5, Number: 5, Title: "Closed PR 2", State: "closed"},
		{ID: 6, Number: 6, Title: "Merged PR 1", State: "merged"},
	})

	testCases := []struct {
		name           string
		state          string
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "Filter by open state",
			state:          "open",
			expectedCount:  3,
			expectedTitles: []string{"Open PR 1", "Open PR 2", "Open PR 3"},
		},
		{
			name:           "Filter by closed state",
			state:          "closed",
			expectedCount:  2,
			expectedTitles: []string{"Closed PR 1", "Closed PR 2"},
		},
		{
			name:           "Filter by merged state",
			state:          "merged",
			expectedCount:  1,
			expectedTitles: []string{"Merged PR 1"},
		},
		{
			name:           "Filter by all state",
			state:          "all",
			expectedCount:  6,
			expectedTitles: []string{"Open PR 1", "Open PR 2", "Open PR 3", "Closed PR 1", "Closed PR 2", "Merged PR 1"},
		},
		{
			name:           "No state filter",
			state:          "",
			expectedCount:  6,
			expectedTitles: []string{"Open PR 1", "Open PR 2", "Open PR 3", "Closed PR 1", "Closed PR 2", "Merged PR 1"},
		},
		{
			name:           "Filter by unknown state",
			state:          "unknown",
			expectedCount:  0,
			expectedTitles: []string{},
		},
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

			assert.Len(t, prs, tc.expectedCount)

			// Verify titles if expected
			if len(tc.expectedTitles) > 0 {
				actualTitles := make([]string, len(prs))
				for i, pr := range prs {
					actualTitles[i] = pr["title"].(string)
				}
				assert.ElementsMatch(t, tc.expectedTitles, actualTitles)
			}
		})
	}
}

// TestMockServerPullRequestStateCaseSensitivity tests case sensitivity in state filtering
func TestMockServerPullRequestStateCaseSensitivity(t *testing.T) {
	mock := NewMockGiteaServer(t)

	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Open PR", State: "open"},
		{ID: 2, Number: 2, Title: "Closed PR", State: "closed"},
	})

	testCases := []struct {
		name          string
		state         string
		expectedCount int
	}{
		{"Lowercase open", "open", 1},
		{"Uppercase OPEN", "OPEN", 0},  // Should be case sensitive
		{"Mixed case Open", "Open", 0}, // Should be case sensitive
		{"Lowercase closed", "closed", 1},
		{"Uppercase CLOSED", "CLOSED", 0},  // Should be case sensitive
		{"Mixed case Closed", "Closed", 0}, // Should be case sensitive
		{"Lowercase all", "all", 2},
		{"Uppercase ALL", "ALL", 0},  // Should be case sensitive
		{"Mixed case All", "All", 0}, // Should be case sensitive
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/pulls?state=" + tc.state)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var prs []map[string]any
			err = json.NewDecoder(resp.Body).Decode(&prs)
			require.NoError(t, err)

			assert.Len(t, prs, tc.expectedCount)
		})
	}
}

// TestMockServerPullRequestStateEdgeCases tests edge cases in state filtering
func TestMockServerPullRequestStateEdgeCases(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add pull requests with edge case states
	mock.AddPullRequests("testuser", "testrepo", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Empty State PR", State: ""},
		{ID: 2, Number: 2, Title: "Whitespace State PR", State: "whitespace"},
		{ID: 3, Number: 3, Title: "Normal Open PR", State: "open"},
	})

	testCases := []struct {
		name          string
		state         string
		expectedCount int
	}{
		{
			name:          "Filter by empty string state",
			state:         "",
			expectedCount: 3, // Should return all when no state filter
		},
		{
			name:          "Filter by whitespace state",
			state:         "whitespace",
			expectedCount: 1, // One PR has "whitespace" as state
		},
		{
			name:          "Filter by open state",
			state:         "open",
			expectedCount: 1,
		},
		{
			name:          "Filter by all state",
			state:         "all",
			expectedCount: 3,
		},
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

			assert.Len(t, prs, tc.expectedCount)
		})
	}
}

// TestMockServerPullRequestStateMultipleRepos tests state filtering across multiple repositories
func TestMockServerPullRequestStateMultipleRepos(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add pull requests to different repositories
	mock.AddPullRequests("user1", "repo1", []MockPullRequest{
		{ID: 1, Number: 1, Title: "Repo1 Open PR", State: "open"},
		{ID: 2, Number: 2, Title: "Repo1 Closed PR", State: "closed"},
	})

	mock.AddPullRequests("user2", "repo2", []MockPullRequest{
		{ID: 3, Number: 1, Title: "Repo2 Open PR", State: "open"},
		{ID: 4, Number: 2, Title: "Repo2 Closed PR", State: "closed"},
		{ID: 5, Number: 3, Title: "Repo2 Merged PR", State: "merged"},
	})

	testCases := []struct {
		name          string
		repository    string
		state         string
		expectedCount int
		expectedTitle string
	}{
		{
			name:          "Repo1 open state",
			repository:    "user1/repo1",
			state:         "open",
			expectedCount: 1,
			expectedTitle: "Repo1 Open PR",
		},
		{
			name:          "Repo1 closed state",
			repository:    "user1/repo1",
			state:         "closed",
			expectedCount: 1,
			expectedTitle: "Repo1 Closed PR",
		},
		{
			name:          "Repo2 open state",
			repository:    "user2/repo2",
			state:         "open",
			expectedCount: 1,
			expectedTitle: "Repo2 Open PR",
		},
		{
			name:          "Repo2 all state",
			repository:    "user2/repo2",
			state:         "all",
			expectedCount: 3,
			expectedTitle: "", // Don't check specific title
		},
		{
			name:          "Repo2 merged state",
			repository:    "user2/repo2",
			state:         "merged",
			expectedCount: 1,
			expectedTitle: "Repo2 Merged PR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/" + tc.repository + "/pulls?state=" + tc.state

			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var prs []map[string]any
			err = json.NewDecoder(resp.Body).Decode(&prs)
			require.NoError(t, err)

			assert.Len(t, prs, tc.expectedCount)

			// Verify specific title if provided
			if tc.expectedTitle != "" && tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedTitle, prs[0]["title"])
			}
		})
	}
}

// TestMockServerPullRequestStateEmptyRepository tests state filtering on empty repository
func TestMockServerPullRequestStateEmptyRepository(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Don't add any pull requests for this repository

	testCases := []struct {
		name          string
		state         string
		expectedCount int
	}{
		{"Empty repo open state", "open", 0},
		{"Empty repo closed state", "closed", 0},
		{"Empty repo all state", "all", 0},
		{"Empty repo no state", "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/empty/repo/pulls"
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

			assert.Len(t, prs, tc.expectedCount)
		})
	}
}
