package e2e

import (
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
)

// TestE2ETestCompletionWithin5Minutes verifies that E2E tests complete within 5 minutes
func TestE2ETestCompletionWithin5Minutes(t *testing.T) {
	// Record start time
	startTime := time.Now()

	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	defer seeder.CleanupTestData(t)

	// Create client for testing
	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Run a series of operations to simulate typical E2E workflow
	t.Run("CompleteE2EWorkflow", func(t *testing.T) {
		runCompleteE2EWorkflow(t, fjClient, env)
	})

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Verify completion within 5 minutes
	maxDuration := 5 * time.Minute
	if elapsed > maxDuration {
		t.Errorf("E2E test suite took %v, which exceeds the 5-minute limit", elapsed)
	} else {
		t.Logf("E2E test suite completed in %v (within 5-minute limit)", elapsed)
	}
}

// runCompleteE2EWorkflow simulates a complete E2E workflow
func runCompleteE2EWorkflow(t *testing.T, fjClient *client.ForgejoClient, env *TestEnvironment) {
	// Step 1: Test repository operations
	t.Run("RepositoryOperations", func(t *testing.T) {
		testRepositoryOperationsPerformance(t, fjClient)
	})

	// Step 2: Test issue operations
	t.Run("IssueOperations", func(t *testing.T) {
		testIssueOperationsPerformance(t, fjClient)
	})

	// Step 3: Test PR operations
	t.Run("PROperations", func(t *testing.T) {
		testPROperationsPerformance(t, fjClient)
	})

	// Step 4: Test concurrent operations
	t.Run("ConcurrentOperations", func(t *testing.T) {
		testConcurrentOperationsPerformance(t, fjClient)
	})
}

// testRepositoryOperationsPerformance tests repository operations performance
func testRepositoryOperationsPerformance(t *testing.T, fjClient *client.ForgejoClient) {
	startTime := time.Now()

	// Test repository listing
	filters := &client.RepositoryFilters{Type: "all"}
	_, err := fjClient.ListRepositories(filters)
	if err != nil {
		t.Logf("Repository listing failed: %v", err)
	}

	// Test specific repository retrieval
	_, err = fjClient.GetRepository("testuser", "test-repo-3")
	if err != nil {
		t.Logf("Repository retrieval failed: %v", err)
	}

	elapsed := time.Since(startTime)
	t.Logf("Repository operations completed in %v", elapsed)

	// Each operation should complete within reasonable time
	if elapsed > 30*time.Second {
		t.Errorf("Repository operations took too long: %v", elapsed)
	}
}

// testIssueOperationsPerformance tests issue operations performance
func testIssueOperationsPerformance(t *testing.T, fjClient *client.ForgejoClient) {
	startTime := time.Now()

	// Test issue listing with different filters
	filters := &client.IssueFilters{State: client.StateAll}
	_, err := fjClient.ListIssues("testuser", "test-repo-3", filters)
	if err != nil {
		t.Logf("Issue listing failed: %v", err)
	}

	// Test with labels
	labelFilters := &client.IssueFilters{
		State:  client.StateAll,
		Labels: []string{"bug"},
	}
	_, err = fjClient.ListIssues("testuser", "test-repo-3", labelFilters)
	if err != nil {
		t.Logf("Issue listing with labels failed: %v", err)
	}

	elapsed := time.Since(startTime)
	t.Logf("Issue operations completed in %v", elapsed)

	if elapsed > 30*time.Second {
		t.Errorf("Issue operations took too long: %v", elapsed)
	}
}

// testPROperationsPerformance tests PR operations performance
func testPROperationsPerformance(t *testing.T, fjClient *client.ForgejoClient) {
	startTime := time.Now()

	// Test PR listing with different filters
	filters := &client.PullRequestFilters{State: client.StateAll}
	_, err := fjClient.ListPRs("testuser", "test-repo-3", filters)
	if err != nil {
		t.Logf("PR listing failed: %v", err)
	}

	// Test with pagination
	paginatedFilters := &client.PullRequestFilters{
		State:    client.StateAll,
		Page:     1,
		PageSize: 10,
	}
	_, err = fjClient.ListPRs("testuser", "test-repo-3", paginatedFilters)
	if err != nil {
		t.Logf("PR listing with pagination failed: %v", err)
	}

	elapsed := time.Since(startTime)
	t.Logf("PR operations completed in %v", elapsed)

	if elapsed > 30*time.Second {
		t.Errorf("PR operations took too long: %v", elapsed)
	}
}

// testConcurrentOperationsPerformance tests concurrent operations performance
func testConcurrentOperationsPerformance(t *testing.T, fjClient *client.ForgejoClient) {
	startTime := time.Now()

	// Run multiple operations concurrently
	const numGoroutines = 5
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine performs a different operation
			switch id % 3 {
			case 0:
				// Repository operation
				_, _ = fjClient.ListRepositories(nil)
			case 1:
				// Issue operation
				_, _ = fjClient.ListIssues("testuser", "test-repo-3", nil)
			case 2:
				// PR operation
				_, _ = fjClient.ListPRs("testuser", "test-repo-3", nil)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	elapsed := time.Since(startTime)
	t.Logf("Concurrent operations completed in %v", elapsed)

	if elapsed > 45*time.Second {
		t.Errorf("Concurrent operations took too long: %v", elapsed)
	}
}

// TestE2ETestSuiteTimingBreakdown provides detailed timing breakdown
func TestE2ETestSuiteTimingBreakdown(t *testing.T) {
	// This test provides detailed timing information for each phase
	t.Run("SetupPhase", func(t *testing.T) {
		startTime := time.Now()
		env := SetupTestEnvironment(t)
		setupTime := time.Since(startTime)
		t.Logf("Environment setup took %v", setupTime)

		// Clean up
		env.Teardown(t)
	})

	t.Run("DataSeedingPhase", func(t *testing.T) {
		env := SetupTestEnvironment(t)
		defer env.Teardown(t)

		startTime := time.Now()
		seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
		err := seeder.SeedTestData(t)
		seedingTime := time.Since(startTime)

		if err != nil {
			t.Logf("Data seeding failed: %v", err)
		}

		t.Logf("Data seeding took %v", seedingTime)

		// Cleanup
		seeder.CleanupTestData(t)
	})

	t.Run("OperationsPhase", func(t *testing.T) {
		env := SetupTestEnvironment(t)
		defer env.Teardown(t)

		seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
		seeder.SeedTestData(t)
		defer seeder.CleanupTestData(t)

		fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		startTime := time.Now()

		// Perform various operations
		fjClient.ListRepositories(nil)
		fjClient.ListIssues("testuser", "test-repo-3", nil)
		fjClient.ListPRs("testuser", "test-repo-3", nil)

		operationsTime := time.Since(startTime)
		t.Logf("Operations took %v", operationsTime)
	})
}

// TestPerformanceRegressionDetection detects performance regressions
func TestPerformanceRegressionDetection(t *testing.T) {
	// This test can be used to detect if operations take significantly longer than expected
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	fjClient, err := client.New(env.GetBaseURL(), env.GetToken())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Define expected maximum times for operations
	maxRepoListTime := 10 * time.Second
	maxIssueListTime := 10 * time.Second
	maxPRListTime := 10 * time.Second

	// Test repository listing performance
	startTime := time.Now()
	_, err = fjClient.ListRepositories(nil)
	repoTime := time.Since(startTime)

	if err != nil {
		t.Logf("Repository listing failed: %v", err)
	} else if repoTime > maxRepoListTime {
		t.Errorf("Repository listing took %v, expected < %v", repoTime, maxRepoListTime)
	} else {
		t.Logf("Repository listing performance: %v", repoTime)
	}

	// Test issue listing performance
	startTime = time.Now()
	_, err = fjClient.ListIssues("testuser", "test-repo-3", nil)
	issueTime := time.Since(startTime)

	if err != nil {
		t.Logf("Issue listing failed: %v", err)
	} else if issueTime > maxIssueListTime {
		t.Errorf("Issue listing took %v, expected < %v", issueTime, maxIssueListTime)
	} else {
		t.Logf("Issue listing performance: %v", issueTime)
	}

	// Test PR listing performance
	startTime = time.Now()
	_, err = fjClient.ListPRs("testuser", "test-repo-3", nil)
	prTime := time.Since(startTime)

	if err != nil {
		t.Logf("PR listing failed: %v", err)
	} else if prTime > maxPRListTime {
		t.Errorf("PR listing took %v, expected < %v", prTime, maxPRListTime)
	} else {
		t.Logf("PR listing performance: %v", prTime)
	}
}
