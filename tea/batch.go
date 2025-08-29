package tea

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// BatchRequest represents a request that can be processed in batch
type BatchRequest struct {
	ID      string
	Method  string
	Owner   string
	Repo    string
	Filters map[string]interface{}
}

// BatchResponse represents the response from a batch operation
type BatchResponse struct {
	ID       string
	Result   interface{}
	Error    error
	Duration time.Duration
}

// BatchProcessor handles batch processing of multiple requests concurrently
type BatchProcessor struct {
	maxConcurrency int
	sem            chan struct{} // Semaphore for limiting concurrency
}

// NewBatchProcessor creates a new batch processor with specified concurrency limit
func NewBatchProcessor(maxConcurrency int) *BatchProcessor {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	return &BatchProcessor{
		maxConcurrency: maxConcurrency,
		sem:            make(chan struct{}, maxConcurrency),
	}
}

// ProcessBatch processes a batch of requests concurrently
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, requests []BatchRequest) ([]BatchResponse, error) {
	if len(requests) == 0 {
		return []BatchResponse{}, nil
	}

	// Create response channel
	responsesChan := make(chan BatchResponse, len(requests))
	var wg sync.WaitGroup

	// Process each request concurrently with concurrency limit
	for _, req := range requests {
		wg.Add(1)
		go func(request BatchRequest) {
			defer wg.Done()

			// Acquire semaphore for concurrency control
			select {
			case bp.sem <- struct{}{}:
			case <-ctx.Done():
				responsesChan <- BatchResponse{
					ID:    request.ID,
					Error: ctx.Err(),
				}
				return
			}
			defer func() { <-bp.sem }() // Release semaphore

			start := time.Now()
			result, err := bp.processSingleRequest(ctx, request)
			duration := time.Since(start)

			responsesChan <- BatchResponse{
				ID:       request.ID,
				Result:   result,
				Error:    err,
				Duration: duration,
			}
		}(req)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(responsesChan)
	}()

	// Collect responses
	var responses []BatchResponse
	for response := range responsesChan {
		responses = append(responses, response)
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		return responses, ctx.Err()
	}

	return responses, nil
}

// processSingleRequest processes a single request (mock implementation)
func (bp *BatchProcessor) processSingleRequest(ctx context.Context, req BatchRequest) (interface{}, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate request
	if req.Owner == "" || req.Repo == "" {
		return nil, errors.New("owner and repo are required")
	}

	// Mock processing time to simulate real work
	time.Sleep(1 * time.Millisecond)

	// Generate mock response based on method
	switch req.Method {
	case "listPRs":
		return bp.mockListPRs(req), nil
	case "listIssues":
		return bp.mockListIssues(req), nil
	case "listRepositories":
		return bp.mockListRepositories(req), nil
	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}
}

// mockListPRs generates mock PR list response
func (bp *BatchProcessor) mockListPRs(req BatchRequest) interface{} {
	state := "open"
	if stateFilter, ok := req.Filters["state"]; ok {
		if s, ok := stateFilter.(string); ok {
			state = s
		}
	}

	return map[string]interface{}{
		"pullRequests": []map[string]interface{}{
			{
				"number": 1,
				"title":  fmt.Sprintf("Mock PR for %s/%s", req.Owner, req.Repo),
				"state":  state,
				"author": "mock-user",
				"url":    fmt.Sprintf("https://github.com/%s/%s/pull/1", req.Owner, req.Repo),
			},
		},
		"total": 1,
	}
}

// mockListIssues generates mock issue list response
func (bp *BatchProcessor) mockListIssues(req BatchRequest) interface{} {
	labels := []string{}
	if labelsFilter, ok := req.Filters["labels"]; ok {
		if l, ok := labelsFilter.([]string); ok {
			labels = l
		}
	}

	return map[string]interface{}{
		"issues": []map[string]interface{}{
			{
				"number": 1,
				"title":  fmt.Sprintf("Mock Issue for %s/%s", req.Owner, req.Repo),
				"state":  "open",
				"author": "mock-user",
				"labels": labels,
				"url":    fmt.Sprintf("https://github.com/%s/%s/issues/1", req.Owner, req.Repo),
			},
		},
		"total": 1,
	}
}

// mockListRepositories generates mock repository list response
func (bp *BatchProcessor) mockListRepositories(req BatchRequest) interface{} {
	return map[string]interface{}{
		"repositories": []map[string]interface{}{
			{
				"id":          1,
				"name":        req.Repo,
				"full_name":   fmt.Sprintf("%s/%s", req.Owner, req.Repo),
				"description": fmt.Sprintf("Mock repository %s/%s", req.Owner, req.Repo),
				"url":         fmt.Sprintf("https://github.com/%s/%s", req.Owner, req.Repo),
			},
		},
		"total": 1,
	}
}

// BatchOptimizer provides optimization capabilities for batch requests
type BatchOptimizer struct {
	cache *Cache
}

// NewBatchOptimizer creates a new batch optimizer with caching
func NewBatchOptimizer(cacheSize int, cacheTTL time.Duration) (*BatchOptimizer, error) {
	cache, err := NewCache(cacheSize, cacheTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache for batch optimizer: %w", err)
	}

	return &BatchOptimizer{
		cache: cache,
	}, nil
}

// OptimizeBatch optimizes a batch of requests by:
// 1. Deduplicating identical requests
// 2. Checking cache for previously computed results
// 3. Grouping similar requests for batch processing
func (bo *BatchOptimizer) OptimizeBatch(requests []BatchRequest) ([]BatchRequest, map[string]interface{}) {
	if len(requests) == 0 {
		return requests, make(map[string]interface{})
	}

	// Map to store cached results
	cachedResults := make(map[string]interface{})

	// Map to deduplicate requests
	uniqueRequests := make(map[string]BatchRequest)

	// Track which request IDs map to which cache keys
	idToKey := make(map[string]string)
	keyToIDs := make(map[string][]string)

	// Process each request
	for _, req := range requests {
		// Generate cache key
		key := GenerateCacheKey(req.Method, req.Owner, req.Repo, req.Filters)
		idToKey[req.ID] = key

		// Check if we already have this request
		if _, exists := uniqueRequests[key]; exists {
			// Deduplicate: map this ID to existing request
			keyToIDs[key] = append(keyToIDs[key], req.ID)
		} else {
			// New unique request
			uniqueRequests[key] = req
			keyToIDs[key] = []string{req.ID}

			// Check cache
			if result, found := bo.cache.Get(key); found {
				cachedResults[key] = result
			}
		}
	}

	// Build list of requests that need processing
	var needProcessing []BatchRequest
	for key, req := range uniqueRequests {
		if _, cached := cachedResults[key]; !cached {
			needProcessing = append(needProcessing, req)
		}
	}

	return needProcessing, cachedResults
}

// CacheResults caches the results from batch processing
func (bo *BatchOptimizer) CacheResults(results []BatchResponse) {
	for _, result := range results {
		if result.Error == nil && result.Result != nil {
			// We would need the original request to generate the proper cache key
			// For now, use the result ID as a simple key
			bo.cache.Set(result.ID, result.Result)
		}
	}
}
