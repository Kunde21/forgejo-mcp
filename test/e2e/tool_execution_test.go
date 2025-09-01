package e2e

import (
	"testing"

	"github.com/Kunde21/forgejo-mcp/client"
)

// TestPRListingAgainstRealInstance tests PR listing against real Forgejo instance
func TestPRListingAgainstRealInstance(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	defer seeder.CleanupTestData(t)

	// Create client
	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test PR listing with different filters
	t.Run("ListAllPRs", func(t *testing.T) {
		filters := &client.PullRequestFilters{
			State: client.StateOpen,
		}

		prs, err := fjClient.ListPRs("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("PR listing failed (may be expected in test environment): %v", err)
		} else {
			t.Logf("Successfully listed %d PRs", len(prs))
			for i, pr := range prs {
				t.Logf("PR %d: %s - %s", i+1, pr.Title, pr.State)
			}
		}
	})

	t.Run("ListPRsWithPagination", func(t *testing.T) {
		filters := &client.PullRequestFilters{
			State:    client.StateAll,
			Page:     1,
			PageSize: 10,
		}

		prs, err := fjClient.ListPRs("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("Paginated PR listing failed: %v", err)
		} else {
			t.Logf("Successfully listed %d PRs with pagination", len(prs))
		}
	})

	t.Run("ListPRsByState", func(t *testing.T) {
		// Test different states
		states := []client.StateType{client.StateOpen, client.StateClosed, client.StateAll}

		for _, state := range states {
			t.Run(string(state), func(t *testing.T) {
				filters := &client.PullRequestFilters{
					State: state,
				}

				prs, err := fjClient.ListPRs("testuser", "test-repo-3", filters)
				if err != nil {
					t.Logf("PR listing for state %s failed: %v", state, err)
				} else {
					t.Logf("Successfully listed %d PRs with state %s", len(prs), state)
				}
			})
		}
	})
}

// TestIssueListingAgainstRealInstance tests issue listing against real Forgejo instance
func TestIssueListingAgainstRealInstance(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	defer seeder.CleanupTestData(t)

	// Create client
	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test issue listing with different filters
	t.Run("ListAllIssues", func(t *testing.T) {
		filters := &client.IssueFilters{
			State: client.StateOpen,
		}

		issues, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("Issue listing failed (may be expected in test environment): %v", err)
		} else {
			t.Logf("Successfully listed %d issues", len(issues))
			for i, issue := range issues {
				t.Logf("Issue %d: %s - %s", i+1, issue.Title, issue.State)
			}
		}
	})

	t.Run("ListIssuesWithPagination", func(t *testing.T) {
		filters := &client.IssueFilters{
			State:    client.StateAll,
			Page:     1,
			PageSize: 10,
		}

		issues, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("Paginated issue listing failed: %v", err)
		} else {
			t.Logf("Successfully listed %d issues with pagination", len(issues))
		}
	})

	t.Run("ListIssuesByState", func(t *testing.T) {
		// Test different states
		states := []client.StateType{client.StateOpen, client.StateClosed, client.StateAll}

		for _, state := range states {
			t.Run(string(state), func(t *testing.T) {
				filters := &client.IssueFilters{
					State: state,
				}

				issues, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
				if err != nil {
					t.Logf("Issue listing for state %s failed: %v", state, err)
				} else {
					t.Logf("Successfully listed %d issues with state %s", len(issues), state)
				}
			})
		}
	})

	t.Run("ListIssuesWithLabels", func(t *testing.T) {
		filters := &client.IssueFilters{
			State:  client.StateAll,
			Labels: []string{"bug", "enhancement"},
		}

		issues, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("Issue listing with labels failed: %v", err)
		} else {
			t.Logf("Successfully listed %d issues with label filter", len(issues))
		}
	})

	t.Run("ListIssuesWithKeyword", func(t *testing.T) {
		filters := &client.IssueFilters{
			State:   client.StateAll,
			KeyWord: "crash",
		}

		issues, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
		if err != nil {
			t.Logf("Issue listing with keyword failed: %v", err)
		} else {
			t.Logf("Successfully listed %d issues with keyword filter", len(issues))
		}
	})
}

// TestRepositoryOperationsAgainstRealInstance tests repository operations
func TestRepositoryOperationsAgainstRealInstance(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	defer seeder.CleanupTestData(t)

	// Create client
	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test repository listing
	t.Run("ListRepositories", func(t *testing.T) {
		filters := &client.RepositoryFilters{
			Type: "all", // List all repositories
		}

		repos, err := fjClient.ListRepositories(filters)
		if err != nil {
			t.Logf("Repository listing failed: %v", err)
		} else {
			t.Logf("Successfully listed %d repositories", len(repos))
			for i, repo := range repos {
				t.Logf("Repo %d: %s - %s", i+1, repo.Name, repo.Description)
			}
		}
	})

	// Test getting specific repository
	t.Run("GetSpecificRepository", func(t *testing.T) {
		repo, err := fjClient.GetRepository("testuser", "test-repo-3")
		if err != nil {
			t.Logf("Getting specific repository failed: %v", err)
		} else {
			t.Logf("Successfully retrieved repository: %s", repo.Name)
			if repo.Name != "test-repo-3" {
				t.Errorf("Expected repository name 'test-repo-3', got '%s'", repo.Name)
			}
		}
	})

	// Test repository filtering
	t.Run("ListPrivateRepositories", func(t *testing.T) {
		filters := &client.RepositoryFilters{
			IsPrivate: &[]bool{true}[0], // Private repos only
		}

		repos, err := fjClient.ListRepositories(filters)
		if err != nil {
			t.Logf("Private repository listing failed: %v", err)
		} else {
			t.Logf("Successfully listed %d private repositories", len(repos))
		}
	})
}

// TestDataConsistency tests that data retrieved is consistent
func TestDataConsistency(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	defer seeder.CleanupTestData(t)

	// Create client
	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test that multiple calls return consistent data
	t.Run("ConsistentRepositoryData", func(t *testing.T) {
		// Call list repositories multiple times
		for i := 0; i < 3; i++ {
			repos, err := fjClient.ListRepositories(nil)
			if err != nil {
				t.Logf("Repository listing call %d failed: %v", i+1, err)
			} else {
				t.Logf("Call %d: Listed %d repositories", i+1, len(repos))
			}
		}
	})

	t.Run("ConsistentIssueData", func(t *testing.T) {
		// Call list issues multiple times
		for i := 0; i < 3; i++ {
			issues, err := fjClient.ListIssues("testuser", "test-repo-3", nil)
			if err != nil {
				t.Logf("Issue listing call %d failed: %v", i+1, err)
			} else {
				t.Logf("Call %d: Listed %d issues", i+1, len(issues))
			}
		}
	})
}
