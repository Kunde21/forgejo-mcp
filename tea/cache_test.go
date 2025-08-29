package tea

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestNewCache tests cache creation
func TestNewCache(t *testing.T) {
	tests := []struct {
		name        string
		maxSize     int
		ttl         time.Duration
		wantMaxSize int
		wantTTL     time.Duration
		wantErr     bool
	}{
		{
			name:        "valid cache parameters",
			maxSize:     100,
			ttl:         5 * time.Minute,
			wantMaxSize: 100,
			wantTTL:     5 * time.Minute,
			wantErr:     false,
		},
		{
			name:    "zero max size",
			maxSize: 0,
			ttl:     5 * time.Minute,
			wantErr: true,
		},
		{
			name:    "negative max size",
			maxSize: -1,
			ttl:     5 * time.Minute,
			wantErr: true,
		},
		{
			name:    "zero TTL",
			maxSize: 100,
			ttl:     0,
			wantErr: true,
		},
		{
			name:    "negative TTL",
			maxSize: 100,
			ttl:     -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			cache, err := NewCache(tst.maxSize, tst.ttl)

			if tst.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cache.MaxSize() != tst.wantMaxSize {
				t.Errorf("Expected MaxSize=%d, got %d", tst.wantMaxSize, cache.MaxSize())
			}

			if cache.TTL() != tst.wantTTL {
				t.Errorf("Expected TTL=%v, got %v", tst.wantTTL, cache.TTL())
			}
		})
	}
}

// TestCacheSetAndGet tests basic cache operations
func TestCacheSetAndGet(t *testing.T) {
	cache, err := NewCache(100, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := "test-key"
	value := map[string]interface{}{
		"data": "test-value",
		"id":   123,
	}

	// Test Set operation
	cache.Set(key, value)

	// Test Get operation
	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}

	if !cmp.Equal(value, retrieved) {
		t.Error(cmp.Diff(value, retrieved))
	}
}

// TestCacheGetMiss tests cache miss behavior
func TestCacheGetMiss(t *testing.T) {
	cache, err := NewCache(100, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test Get on non-existent key
	value, found := cache.Get("non-existent-key")
	if found {
		t.Error("Expected cache miss but found value")
	}

	if value != nil {
		t.Errorf("Expected nil value on cache miss, got %v", value)
	}
}

// TestCacheTTLExpiration tests that cached items expire after TTL
func TestCacheTTLExpiration(t *testing.T) {
	shortTTL := 10 * time.Millisecond
	cache, err := NewCache(100, shortTTL)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := "expire-key"
	value := "expire-value"

	// Set value
	cache.Set(key, value)

	// Verify it exists immediately
	retrieved, found := cache.Get(key)
	if !found || retrieved != value {
		t.Error("Expected to find cached value immediately after set")
	}

	// Wait for TTL to expire
	time.Sleep(shortTTL + 5*time.Millisecond)

	// Verify it's expired
	retrieved, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be expired")
	}

	if retrieved != nil {
		t.Errorf("Expected nil value after expiration, got %v", retrieved)
	}
}

// TestCacheMaxSize tests that cache respects maximum size limit
func TestCacheMaxSize(t *testing.T) {
	maxSize := 3
	cache, err := NewCache(maxSize, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Fill cache to max capacity
	for i := 0; i < maxSize; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		cache.Set(key, value)
	}

	// Verify all items are cached
	for i := 0; i < maxSize; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		retrieved, found := cache.Get(key)
		if !found || retrieved != value {
			t.Errorf("Expected to find %s=%s", key, value)
		}
	}

	// Add one more item to exceed capacity
	cache.Set("key-overflow", "value-overflow")

	// Verify the overflow item exists
	retrieved, found := cache.Get("key-overflow")
	if !found || retrieved != "value-overflow" {
		t.Error("Expected overflow item to be cached")
	}

	// Verify that one of the original items was evicted
	foundCount := 0
	for i := 0; i < maxSize; i++ {
		key := fmt.Sprintf("key-%d", i)
		if _, found := cache.Get(key); found {
			foundCount++
		}
	}

	if foundCount >= maxSize {
		t.Errorf("Expected at least one original item to be evicted, but found %d out of %d", foundCount, maxSize)
	}
}

// TestCacheDelete tests cache item deletion
func TestCacheDelete(t *testing.T) {
	cache, err := NewCache(100, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := "delete-key"
	value := "delete-value"

	// Set value
	cache.Set(key, value)

	// Verify it exists
	retrieved, found := cache.Get(key)
	if !found || retrieved != value {
		t.Error("Expected to find cached value before deletion")
	}

	// Delete the value
	cache.Delete(key)

	// Verify it's deleted
	retrieved, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be deleted")
	}

	if retrieved != nil {
		t.Errorf("Expected nil value after deletion, got %v", retrieved)
	}
}

// TestCacheClear tests cache clearing
func TestCacheClear(t *testing.T) {
	cache, err := NewCache(100, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add multiple items
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		cache.Set(key, "value-"+key)
	}

	// Verify all items exist
	for _, key := range keys {
		if _, found := cache.Get(key); !found {
			t.Errorf("Expected to find %s before clear", key)
		}
	}

	// Clear the cache
	cache.Clear()

	// Verify all items are gone
	for _, key := range keys {
		if _, found := cache.Get(key); found {
			t.Errorf("Expected %s to be cleared", key)
		}
	}
}

// TestCacheStats tests cache statistics
func TestCacheStats(t *testing.T) {
	cache, err := NewCache(100, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Initial stats should be zero
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 || stats.Size != 0 {
		t.Errorf("Expected initial stats to be zero, got: %+v", stats)
	}

	// Set a value
	cache.Set("test-key", "test-value")

	// Check stats after set
	stats = cache.Stats()
	if stats.Size != 1 {
		t.Errorf("Expected size=1 after set, got %d", stats.Size)
	}

	// Get existing value (hit)
	cache.Get("test-key")
	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected hits=1 after successful get, got %d", stats.Hits)
	}

	// Get non-existent value (miss)
	cache.Get("non-existent")
	stats = cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected misses=1 after unsuccessful get, got %d", stats.Misses)
	}
}

// TestCacheKeyGeneration tests cache key generation for different request types
func TestCacheKeyGeneration(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		owner   string
		repo    string
		filters map[string]interface{}
		wantKey string
	}{
		{
			name:    "pr list basic",
			method:  "listPRs",
			owner:   "myorg",
			repo:    "myrepo",
			filters: map[string]interface{}{},
			wantKey: "listPRs:myorg:myrepo:{}",
		},
		{
			name:   "pr list with state",
			method: "listPRs",
			owner:  "myorg",
			repo:   "myrepo",
			filters: map[string]interface{}{
				"state": "open",
			},
			wantKey: "listPRs:myorg:myrepo:{\"state\":\"open\"}",
		},
		{
			name:   "issue list with labels",
			method: "listIssues",
			owner:  "myorg",
			repo:   "myrepo",
			filters: map[string]interface{}{
				"labels": []string{"bug", "priority-high"},
			},
			wantKey: "listIssues:myorg:myrepo:{\"labels\":[\"bug\",\"priority-high\"]}",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			key := GenerateCacheKey(tst.method, tst.owner, tst.repo, tst.filters)
			if key != tst.wantKey {
				t.Errorf("Expected key=%s, got %s", tst.wantKey, key)
			}
		})
	}
}

// TestCacheConcurrentAccess tests cache behavior under concurrent access
func TestCacheConcurrentAccess(t *testing.T) {
	cache, err := NewCache(1000, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing cache operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)

				// Set value
				cache.Set(key, value)

				// Get value
				retrieved, found := cache.Get(key)
				if found && retrieved != value {
					t.Errorf("Concurrent access: expected %s=%s, got %v", key, value, retrieved)
				}

				// Occasionally delete values
				if j%10 == 0 {
					cache.Delete(key)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Cache should still be functional after concurrent access
	cache.Set("final-test", "final-value")
	retrieved, found := cache.Get("final-test")
	if !found || retrieved != "final-value" {
		t.Error("Cache not functional after concurrent access")
	}
}

// BenchmarkCacheSet benchmarks cache set operations
func BenchmarkCacheSet(b *testing.B) {
	cache, _ := NewCache(1000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		cache.Set(key, value)
	}
}

// BenchmarkCacheGet benchmarks cache get operations
func BenchmarkCacheGet(b *testing.B) {
	cache, _ := NewCache(1000, 5*time.Minute)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		cache.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		cache.Get(key)
	}
}

// BenchmarkCacheSetGet benchmarks mixed cache operations
func BenchmarkCacheSetGet(b *testing.B) {
	cache, _ := NewCache(1000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		if i%2 == 0 {
			cache.Set(key, "value")
		} else {
			cache.Get(key)
		}
	}
}

// BenchmarkCacheKeyGeneration benchmarks cache key generation
func BenchmarkCacheKeyGeneration(b *testing.B) {
	filters := map[string]interface{}{
		"state":  "open",
		"labels": []string{"bug", "priority-high"},
		"page":   1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey("listPRs", "myorg", "myrepo", filters)
	}
}

// BenchmarkCacheConcurrent benchmarks concurrent cache operations
func BenchmarkCacheConcurrent(b *testing.B) {
	cache, _ := NewCache(1000, 5*time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%100)
			if i%2 == 0 {
				cache.Set(key, "value")
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

// FuzzGenerateCacheKey tests the GenerateCacheKey function with fuzzing
func FuzzGenerateCacheKey(f *testing.F) {
	// Add some seed corpus
	f.Add("listPRs", "owner1", "repo1", `{"state":"open"}`)
	f.Add("listIssues", "owner2", "repo2", `{"labels":["bug"]}`)
	f.Add("", "", "", `{}`)

	f.Fuzz(func(t *testing.T, method, owner, repo, filtersJSON string) {
		// Try to parse filtersJSON as JSON
		var filters map[string]interface{}
		if filtersJSON != "" {
			// This might fail, but that's okay for fuzzing
			json.Unmarshal([]byte(filtersJSON), &filters)
		}

		// The function should not panic regardless of inputs
		key := GenerateCacheKey(method, owner, repo, filters)

		// Key should be a string (this is always true, but good to verify)
		if key == "" {
			// This is valid, empty inputs produce empty key
			return
		}
	})
}

// TestNewCachedClient tests the NewCachedClient function
func TestNewCachedClient(t *testing.T) {
	// Test successful creation
	client, err := NewCachedClient("test-client", 100, time.Minute)
	if err != nil {
		t.Errorf("NewCachedClient failed: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.cache == nil {
		t.Error("Expected cache to be initialized")
	}

	// Test error cases
	client, err = NewCachedClient("test-client", -1, time.Minute)
	if err == nil {
		t.Error("Expected error for negative cache size")
	}

	client, err = NewCachedClient("test-client", 100, -time.Minute)
	if err == nil {
		t.Error("Expected error for negative TTL")
	}
}

// TestCachedClientCache tests the Cache method
func TestCachedClientCache(t *testing.T) {
	client, err := NewCachedClient("test-client", 100, time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cached client: %v", err)
	}

	// Test that we can access the cache
	cache := client.Cache()
	if cache == nil {
		t.Error("Expected non-nil cache")
	}

	if cache.MaxSize() != 100 {
		t.Errorf("Expected max size 100, got %d", cache.MaxSize())
	}
}

// TestCacheResponse tests the CacheResponse method
func TestCacheResponse(t *testing.T) {
	client, err := NewCachedClient("test-client", 100, time.Minute)
	if err != nil {
		t.Fatalf("Failed to create cached client: %v", err)
	}

	// Cache a response
	key := "test-key"
	response := map[string]interface{}{"data": "test-value"}
	client.CacheResponse(key, response)

	// Retrieve the cached response
	cachedResponse, found := client.GetCachedResponse(key)
	if !found {
		t.Error("Expected to find cached response")
	}

	if !cmp.Equal(response, cachedResponse) {
		t.Errorf("Cached response mismatch: %v != %v", response, cachedResponse)
	}

	// Test non-existent key
	_, found = client.GetCachedResponse("non-existent-key")
	if found {
		t.Error("Expected not to find non-existent key")
	}
}
