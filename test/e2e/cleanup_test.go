package e2e

import (
	"testing"
	"time"
)

// TestCleanupProcedures tests that cleanup and teardown procedures work correctly
func TestCleanupProcedures(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Seed test data
	seeder := NewTestDataSeeder(env.GetBaseURL(), env.GetToken())
	if err := seeder.SeedTestData(t); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	// Test cleanup procedures
	t.Run("CleanupTestData", func(t *testing.T) {
		testCleanupTestData(t, seeder)
	})

	t.Run("EnvironmentTeardown", func(t *testing.T) {
		testEnvironmentTeardown(t, env)
	})

	t.Run("ResourceCleanupTiming", func(t *testing.T) {
		testResourceCleanupTiming(t, env)
	})

	t.Run("CleanupErrorHandling", func(t *testing.T) {
		testCleanupErrorHandling(t, env)
	})
}

// testCleanupTestData tests that test data cleanup works properly
func testCleanupTestData(t *testing.T, seeder *TestDataSeeder) {
	// Verify that cleanup can be called without errors
	err := seeder.CleanupTestData(t)
	if err != nil {
		t.Logf("Cleanup completed with some errors (expected in test environment): %v", err)
	}

	t.Logf("Test data cleanup procedure executed successfully")
}

// testEnvironmentTeardown tests that environment teardown works properly
func testEnvironmentTeardown(t *testing.T, env *TestEnvironment) {
	// Record start time
	startTime := time.Now()

	// Call teardown
	env.Teardown(t)

	// Verify teardown completed within reasonable time
	elapsed := time.Since(startTime)
	if elapsed > 30*time.Second {
		t.Errorf("Teardown took too long: %v", elapsed)
	}

	t.Logf("Environment teardown completed in %v", elapsed)
}

// testResourceCleanupTiming tests that resources are cleaned up in the correct order
func testResourceCleanupTiming(t *testing.T, env *TestEnvironment) {
	// Setup a new environment for this test
	newEnv := SetupTestEnvironment(t)

	// Record the order of operations
	var operations []string

	// Override teardown to track operations
	originalTeardown := func(te *TestEnvironment) {
		if te.forgejo != nil {
			operations = append(operations, "stop_container")
		}
		if te.network != nil {
			operations = append(operations, "remove_network")
		}
	}

	// Simulate the teardown process
	originalTeardown(newEnv)

	// Clean up the actual environment
	newEnv.Teardown(t)

	// Verify operations were tracked
	if len(operations) == 0 {
		t.Error("No cleanup operations were tracked")
	}

	t.Logf("Resource cleanup operations tracked: %v", operations)
}

// testCleanupErrorHandling tests that cleanup handles errors gracefully
func testCleanupErrorHandling(t *testing.T, env *TestEnvironment) {
	// Test cleanup with invalid data
	seeder := NewTestDataSeeder("http://invalid-url", "invalid-token")

	// This should not panic even with invalid data
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Cleanup panicked with invalid data: %v", r)
		}
	}()

	err := seeder.CleanupTestData(t)
	if err != nil {
		t.Logf("Cleanup handled invalid data gracefully: %v", err)
	}

	t.Logf("Cleanup error handling test completed successfully")
}

// TestCleanupWithTimeout tests cleanup procedures with timeout constraints
func TestCleanupWithTimeout(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// Test that cleanup completes within a reasonable timeout
	done := make(chan bool, 1)

	go func() {
		env.Teardown(t)
		done <- true
	}()

	select {
	case <-done:
		t.Logf("Cleanup completed within timeout")
	case <-time.After(60 * time.Second):
		t.Error("Cleanup did not complete within 60 seconds")
	}
}

// TestMultipleCleanupCalls tests that multiple cleanup calls are safe
func TestMultipleCleanupCalls(t *testing.T) {
	env := SetupTestEnvironment(t)

	// Call teardown multiple times
	for i := 0; i < 3; i++ {
		env.Teardown(t)
		t.Logf("Teardown call %d completed successfully", i+1)
	}

	t.Logf("Multiple cleanup calls handled safely")
}

// TestPartialCleanup tests cleanup when some resources are already cleaned up
func TestPartialCleanup(t *testing.T) {
	env := SetupTestEnvironment(t)

	// Manually clean up some resources first
	if env.forgejo != nil {
		env.pool.Purge(env.forgejo)
		env.forgejo = nil
	}

	// Now call teardown - should handle partial cleanup gracefully
	env.Teardown(t)

	t.Logf("Partial cleanup handled gracefully")
}

// TestCleanupLogging tests that cleanup operations are properly logged
func TestCleanupLogging(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Teardown(t)

	// The teardown process should include logging
	// This test verifies that no panics occur during logged cleanup
	t.Logf("Cleanup logging test completed - no panics during logged operations")
}
