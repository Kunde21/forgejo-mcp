package servertest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServerPullRequestPaginationBasic tests basic pagination functionality
func TestMockServerPullRequestPaginationBasic(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add 25 pull requests for pagination testing
	var prs []MockPullRequest
	for i := 1; i <= 25; i++ {
		prs = append(prs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  "Pull Request " + strconv.Itoa(i),
			State:  "open",
		})
	}
	mock.AddPullRequests("testuser", "testrepo", prs)

	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "First page with limit 10",
			limit:          10,
			offset:         0,
			expectedCount:  10,
			expectedTitles: []string{"Pull Request 1", "Pull Request 2", "Pull Request 3", "Pull Request 4", "Pull Request 5", "Pull Request 6", "Pull Request 7", "Pull Request 8", "Pull Request 9", "Pull Request 10"},
		},
		{
			name:           "Second page with limit 10",
			limit:          10,
			offset:         10,
			expectedCount:  10,
			expectedTitles: []string{"Pull Request 11", "Pull Request 12", "Pull Request 13", "Pull Request 14", "Pull Request 15", "Pull Request 16", "Pull Request 17", "Pull Request 18", "Pull Request 19", "Pull Request 20"},
		},
		{
			name:           "Third page with limit 10",
			limit:          10,
			offset:         20,
			expectedCount:  5, // Only 5 left
			expectedTitles: []string{"Pull Request 21", "Pull Request 22", "Pull Request 23", "Pull Request 24", "Pull Request 25"},
		},
		{
			name:           "Single item per page",
			limit:          1,
			offset:         0,
			expectedCount:  1,
			expectedTitles: []string{"Pull Request 1"},
		},
		{
			name:           "Large limit",
			limit:          100,
			offset:         0,
			expectedCount:  25,
			expectedTitles: []string{}, // Don't check all titles
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/testuser/testrepo/pulls"
			url += "?limit=" + strconv.Itoa(tc.limit)
			url += "&offset=" + strconv.Itoa(tc.offset)

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
				assert.Equal(t, tc.expectedTitles, actualTitles)
			}
		})
	}
}

// TestMockServerPullRequestPaginationEdgeCases tests edge cases in pagination
func TestMockServerPullRequestPaginationEdgeCases(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add 10 pull requests for edge case testing
	var prs []MockPullRequest
	for i := 1; i <= 10; i++ {
		prs = append(prs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  "Pull Request " + strconv.Itoa(i),
			State:  "open",
		})
	}
	mock.AddPullRequests("testuser", "testrepo", prs)

	testCases := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "Zero limit",
			limit:         0,
			offset:        0,
			expectedCount: 10, // Should return all when limit is 0 (not implemented in mock yet)
		},
		{
			name:          "Negative offset",
			limit:         10,
			offset:        -1,
			expectedCount: 10, // Should treat negative offset as 0
		},
		{
			name:          "Offset beyond total count",
			limit:         10,
			offset:        100,
			expectedCount: 0, // Should return empty array
		},
		{
			name:          "Offset at exact end",
			limit:         10,
			offset:        10,
			expectedCount: 0, // Should return empty array
		},
		{
			name:          "Large offset",
			limit:         10,
			offset:        1000,
			expectedCount: 0, // Should return empty array
		},
		{
			name:          "Limit larger than remaining items",
			limit:         20,
			offset:        5,
			expectedCount: 5, // Should return remaining 5 items
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/testuser/testrepo/pulls"
			url += "?limit=" + strconv.Itoa(tc.limit)
			url += "&offset=" + strconv.Itoa(tc.offset)

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

// TestMockServerPullRequestPaginationWithState tests pagination combined with state filtering
func TestMockServerPullRequestPaginationWithState(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add pull requests with different states
	var openPRs, closedPRs []MockPullRequest
	for i := 1; i <= 15; i++ {
		openPRs = append(openPRs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  "Open PR " + strconv.Itoa(i),
			State:  "open",
		})
	}
	for i := 16; i <= 25; i++ {
		closedPRs = append(closedPRs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  "Closed PR " + strconv.Itoa(i),
			State:  "closed",
		})
	}

	allPRs := append(openPRs, closedPRs...)
	mock.AddPullRequests("testuser", "testrepo", allPRs)

	testCases := []struct {
		name           string
		state          string
		limit          int
		offset         int
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "Open state first page",
			state:          "open",
			limit:          5,
			offset:         0,
			expectedCount:  5,
			expectedTitles: []string{"Open PR 1", "Open PR 2", "Open PR 3", "Open PR 4", "Open PR 5"},
		},
		{
			name:           "Open state second page",
			state:          "open",
			limit:          5,
			offset:         5,
			expectedCount:  5,
			expectedTitles: []string{"Open PR 6", "Open PR 7", "Open PR 8", "Open PR 9", "Open PR 10"},
		},
		{
			name:           "Open state third page",
			state:          "open",
			limit:          5,
			offset:         10,
			expectedCount:  5,
			expectedTitles: []string{"Open PR 11", "Open PR 12", "Open PR 13", "Open PR 14", "Open PR 15"},
		},
		{
			name:           "Open state fourth page (empty)",
			state:          "open",
			limit:          5,
			offset:         15,
			expectedCount:  0,
			expectedTitles: []string{},
		},
		{
			name:           "Closed state first page",
			state:          "closed",
			limit:          5,
			offset:         0,
			expectedCount:  5,
			expectedTitles: []string{"Closed PR 16", "Closed PR 17", "Closed PR 18", "Closed PR 19", "Closed PR 20"},
		},
		{
			name:           "Closed state second page",
			state:          "closed",
			limit:          5,
			offset:         5,
			expectedCount:  5,
			expectedTitles: []string{"Closed PR 21", "Closed PR 22", "Closed PR 23", "Closed PR 24", "Closed PR 25"},
		},
		{
			name:           "All state first page",
			state:          "all",
			limit:          10,
			offset:         0,
			expectedCount:  10,
			expectedTitles: []string{"Open PR 1", "Open PR 2", "Open PR 3", "Open PR 4", "Open PR 5", "Open PR 6", "Open PR 7", "Open PR 8", "Open PR 9", "Open PR 10"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/testuser/testrepo/pulls"
			url += "?state=" + tc.state
			url += "&limit=" + strconv.Itoa(tc.limit)
			url += "&offset=" + strconv.Itoa(tc.offset)

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
				assert.Equal(t, tc.expectedTitles, actualTitles)
			}
		})
	}
}

// TestMockServerPullRequestPaginationInvalidParameters tests invalid pagination parameters
func TestMockServerPullRequestPaginationInvalidParameters(t *testing.T) {
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
			name:           "Negative limit",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=-1&offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should return all PRs for invalid limit
		},
		{
			name:           "Non-numeric limit",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=abc&offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should treat invalid limit as default
		},
		{
			name:           "Non-numeric offset",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=10&offset=abc",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should treat invalid offset as default
		},
		{
			name:           "Empty limit parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=&offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should treat empty limit as default
		},
		{
			name:           "Empty offset parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=10&offset=",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should treat empty offset as default
		},
		{
			name:           "Missing limit parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should use default limit
		},
		{
			name:           "Missing offset parameter",
			url:            "/api/v1/repos/testuser/testrepo/pulls?limit=10",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Should use default offset
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

// TestMockServerPullRequestPaginationEmptyRepository tests pagination on empty repository
func TestMockServerPullRequestPaginationEmptyRepository(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Don't add any pull requests

	testCases := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{"Empty repo first page", 10, 0, 0},
		{"Empty repo second page", 10, 10, 0},
		{"Empty repo large offset", 10, 100, 0},
		{"Empty repo zero limit", 0, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := mock.URL() + "/api/v1/repos/empty/repo/pulls"
			url += "?limit=" + strconv.Itoa(tc.limit)
			url += "&offset=" + strconv.Itoa(tc.offset)

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

// TestMockServerPullRequestPaginationConsistency tests consistency across multiple pages
func TestMockServerPullRequestPaginationConsistency(t *testing.T) {
	mock := NewMockGiteaServer(t)

	// Add 50 pull requests for consistency testing
	var prs []MockPullRequest
	for i := 1; i <= 50; i++ {
		prs = append(prs, MockPullRequest{
			ID:     i,
			Number: i,
			Title:  "Pull Request " + strconv.Itoa(i),
			State:  "open",
		})
	}
	mock.AddPullRequests("testuser", "testrepo", prs)

	// Collect all PRs through pagination
	var allPRs []map[string]any
	limit := 10
	offset := 0

	for {
		url := mock.URL() + "/api/v1/repos/testuser/testrepo/pulls"
		url += "?limit=" + strconv.Itoa(limit)
		url += "&offset=" + strconv.Itoa(offset)

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var pagePRs []map[string]any
		err = json.NewDecoder(resp.Body).Decode(&pagePRs)
		require.NoError(t, err)

		if len(pagePRs) == 0 {
			break
		}

		allPRs = append(allPRs, pagePRs...)
		offset += limit
	}

	// Verify we got all 50 PRs
	assert.Len(t, allPRs, 50)

	// Verify no duplicates
	titleMap := make(map[string]bool)
	for _, pr := range allPRs {
		title := pr["title"].(string)
		assert.False(t, titleMap[title], "Duplicate PR found: "+title)
		titleMap[title] = true
	}

	// Verify correct order
	for i, pr := range allPRs {
		expectedTitle := "Pull Request " + strconv.Itoa(i+1)
		assert.Equal(t, expectedTitle, pr["title"].(string))
	}
}
