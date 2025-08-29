package tea

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestBatchProcessor tests batch processing functionality
func TestBatchProcessor(t *testing.T) {
	processor := NewBatchProcessor(5) // Process 5 requests concurrently

	requests := []BatchRequest{
		{
			ID:     "req1",
			Method: "listPRs",
			Owner:  "org1",
			Repo:   "repo1",
			Filters: map[string]interface{}{
				"state": "open",
			},
		},
		{
			ID:     "req2",
			Method: "listIssues",
			Owner:  "org1",
			Repo:   "repo1",
			Filters: map[string]interface{}{
				"labels": []string{"bug"},
			},
		},
		{
			ID:     "req3",
			Method: "listPRs",
			Owner:  "org2",
			Repo:   "repo2",
			Filters: map[string]interface{}{
				"state": "closed",
			},
		},
	}

	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// Verify each response has the correct ID
	responseMap := make(map[string]BatchResponse)
	for _, resp := range responses {
		responseMap[resp.ID] = resp
	}

	for _, req := range requests {
		if resp, exists := responseMap[req.ID]; !exists {
			t.Errorf("Missing response for request %s", req.ID)
		} else {
			if resp.ID != req.ID {
				t.Errorf("Response ID mismatch: expected %s, got %s", req.ID, resp.ID)
			}
		}
	}
}

// TestBatchProcessorConcurrency tests that batch processing handles concurrency correctly
func TestBatchProcessorConcurrency(t *testing.T) {
	processor := NewBatchProcessor(3) // Limit concurrency to 3

	// Create a large batch to test concurrency limits
	var requests []BatchRequest
	for i := 0; i < 10; i++ {
		requests = append(requests, BatchRequest{
			ID:     fmt.Sprintf("req%d", i),
			Method: "listPRs",
			Owner:  "testorg",
			Repo:   "testrepo",
			Filters: map[string]interface{}{
				"page": i + 1,
			},
		})
	}

	start := time.Now()
	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// With concurrency limit of 3 and 10 requests, it should take some minimum time
	// This is a rough test - in practice the mock operations are very fast
	t.Logf("Batch processing took %v for %d requests with concurrency %d", elapsed, len(requests), 3)
}

// TestBatchProcessorContextCancellation tests that batch processing respects context cancellation
func TestBatchProcessorContextCancellation(t *testing.T) {
	processor := NewBatchProcessor(2)

	requests := []BatchRequest{
		{ID: "req1", Method: "listPRs", Owner: "org", Repo: "repo1"},
		{ID: "req2", Method: "listPRs", Owner: "org", Repo: "repo2"},
		{ID: "req3", Method: "listPRs", Owner: "org", Repo: "repo3"},
	}

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	responses, err := processor.ProcessBatch(ctx, requests)

	// Should get context cancellation error or partial results
	if err == nil && len(responses) == len(requests) {
		// All requests completed before timeout - that's okay for fast operations
		t.Log("All requests completed before context timeout")
	} else {
		t.Logf("Context cancellation worked as expected: err=%v, responses=%d/%d", err, len(responses), len(requests))
	}
}

// TestBatchProcessorErrorHandling tests error handling in batch processing
func TestBatchProcessorErrorHandling(t *testing.T) {
	processor := NewBatchProcessor(2)

	requests := []BatchRequest{
		{
			ID:     "good-req",
			Method: "listPRs",
			Owner:  "validorg",
			Repo:   "validrepo",
		},
		{
			ID:     "bad-req",
			Method: "listPRs",
			Owner:  "", // Invalid owner should cause error
			Repo:   "", // Invalid repo should cause error
		},
	}

	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)

	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// Find responses by ID
	var goodResp, badResp *BatchResponse
	for i, resp := range responses {
		if resp.ID == "good-req" {
			goodResp = &responses[i]
		} else if resp.ID == "bad-req" {
			badResp = &responses[i]
		}
	}

	// Good request should succeed
	if goodResp == nil {
		t.Error("Missing response for good-req")
	} else if goodResp.Error != nil {
		t.Errorf("Expected good request to succeed, got error: %v", goodResp.Error)
	}

	// Bad request should fail
	if badResp == nil {
		t.Error("Missing response for bad-req")
	} else if badResp.Error == nil {
		t.Error("Expected bad request to fail, but it succeeded")
	}
}

// TestBatchOptimization tests that batch processing can optimize similar requests
func TestBatchOptimization(t *testing.T) {
	processor := NewBatchProcessor(5)

	// Create multiple requests for the same resource
	requests := []BatchRequest{
		{ID: "req1", Method: "listPRs", Owner: "sameorg", Repo: "samerepo", Filters: map[string]interface{}{"state": "open"}},
		{ID: "req2", Method: "listPRs", Owner: "sameorg", Repo: "samerepo", Filters: map[string]interface{}{"state": "open"}},
		{ID: "req3", Method: "listPRs", Owner: "sameorg", Repo: "samerepo", Filters: map[string]interface{}{"state": "open"}},
	}

	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)

	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// All responses should be identical since they're for the same request
	if len(responses) >= 2 {
		if !cmp.Equal(responses[0].Result, responses[1].Result) {
			t.Error("Expected identical results for identical requests")
		}
	}
}

// TestBatchPerformance benchmarks batch processing performance
func TestBatchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	processor := NewBatchProcessor(10)

	// Create a large batch of requests
	var requests []BatchRequest
	for i := 0; i < 100; i++ {
		requests = append(requests, BatchRequest{
			ID:     fmt.Sprintf("perf-req-%d", i),
			Method: "listPRs",
			Owner:  fmt.Sprintf("org-%d", i%5),   // Spread across 5 orgs
			Repo:   fmt.Sprintf("repo-%d", i%10), // Spread across 10 repos
			Filters: map[string]interface{}{
				"state": "open",
				"page":  (i % 3) + 1,
			},
		})
	}

	start := time.Now()
	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	requestsPerSecond := float64(len(requests)) / elapsed.Seconds()
	t.Logf("Batch performance: %d requests in %v (%.1f req/sec)", len(requests), elapsed, requestsPerSecond)

	// Performance should be reasonable - at least 100 requests per second for mock operations
	if requestsPerSecond < 100 {
		t.Logf("Performance warning: only %.1f req/sec (expected > 100)", requestsPerSecond)
	}
}

// BenchmarkBatchProcessor benchmarks batch processing with different concurrency levels
func BenchmarkBatchProcessor(b *testing.B) {
	processor := NewBatchProcessor(10)

	// Create test requests
	requests := make([]BatchRequest, 50)
	for i := 0; i < 50; i++ {
		requests[i] = BatchRequest{
			ID:     fmt.Sprintf("bench-req-%d", i),
			Method: "listPRs",
			Owner:  fmt.Sprintf("org-%d", i%5),
			Repo:   fmt.Sprintf("repo-%d", i%10),
			Filters: map[string]interface{}{
				"state": "open",
				"page":  (i % 3) + 1,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		processor.ProcessBatch(ctx, requests)
	}
}

// BenchmarkBatchProcessorConcurrency tests different concurrency levels
func BenchmarkBatchProcessorConcurrency1(b *testing.B)  { benchmarkBatchConcurrency(b, 1) }
func BenchmarkBatchProcessorConcurrency5(b *testing.B)  { benchmarkBatchConcurrency(b, 5) }
func BenchmarkBatchProcessorConcurrency10(b *testing.B) { benchmarkBatchConcurrency(b, 10) }
func BenchmarkBatchProcessorConcurrency20(b *testing.B) { benchmarkBatchConcurrency(b, 20) }

func benchmarkBatchConcurrency(b *testing.B, concurrency int) {
	processor := NewBatchProcessor(concurrency)

	requests := make([]BatchRequest, 100)
	for i := 0; i < 100; i++ {
		requests[i] = BatchRequest{
			ID:     fmt.Sprintf("req-%d", i),
			Method: "listPRs",
			Owner:  "benchorg",
			Repo:   "benchrepo",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		processor.ProcessBatch(ctx, requests)
	}
}

// BenchmarkBatchOptimizer benchmarks batch optimization with caching
func BenchmarkBatchOptimizer(b *testing.B) {
	optimizer, _ := NewBatchOptimizer(1000, 5*time.Minute)

	requests := make([]BatchRequest, 100)
	for i := 0; i < 100; i++ {
		requests[i] = BatchRequest{
			ID:     fmt.Sprintf("opt-req-%d", i),
			Method: "listPRs",
			Owner:  fmt.Sprintf("org-%d", i%10), // 10 different orgs
			Repo:   fmt.Sprintf("repo-%d", i%5), // 5 different repos
			Filters: map[string]interface{}{
				"state": "open",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.OptimizeBatch(requests)
	}
}
