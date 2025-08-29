package tea

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Cache provides an in-memory cache with TTL (Time To Live) support
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*cacheItem
	maxSize int
	ttl     time.Duration
	stats   CacheStats
}

// cacheItem represents a single cached item with expiration time
type cacheItem struct {
	value   interface{}
	expires time.Time
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits   int64 // Number of successful cache hits
	Misses int64 // Number of cache misses
	Size   int64 // Current number of items in cache
}

// NewCache creates a new cache instance with specified maximum size and TTL
func NewCache(maxSize int, ttl time.Duration) (*Cache, error) {
	if maxSize <= 0 {
		return nil, errors.New("maxSize must be positive")
	}

	if ttl <= 0 {
		return nil, errors.New("TTL must be positive")
	}

	return &Cache{
		items:   make(map[string]*cacheItem),
		maxSize: maxSize,
		ttl:     ttl,
		stats:   CacheStats{},
	}, nil
}

// MaxSize returns the maximum number of items the cache can hold
func (c *Cache) MaxSize() int {
	return c.maxSize
}

// TTL returns the time-to-live duration for cache items
func (c *Cache) TTL() time.Duration {
	return c.ttl
}

// Set stores a value in the cache with the given key
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items to make room
	if len(c.items) >= c.maxSize {
		// If key already exists, we'll replace it
		if _, exists := c.items[key]; !exists {
			// Need to evict an item first
			c.evictLRU()
		}
	}

	c.items[key] = &cacheItem{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}

	c.stats.Size = int64(len(c.items))
}

// Get retrieves a value from the cache by key
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.expires) {
		// Item expired, remove it
		delete(c.items, key)
		c.stats.Size = int64(len(c.items))
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return item.value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	c.stats.Size = int64(len(c.items))
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.stats.Size = 0
}

// Stats returns current cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update size in case any items have been evicted due to expiration
	c.stats.Size = int64(len(c.items))
	return c.stats
}

// evictLRU evicts the least recently used item (simple implementation)
// In a more sophisticated implementation, we would track access times
func (c *Cache) evictLRU() {
	// For simplicity, evict the first item we find
	for key := range c.items {
		delete(c.items, key)
		break
	}
}

// GenerateCacheKey creates a consistent cache key from method parameters
func GenerateCacheKey(method, owner, repo string, filters map[string]interface{}) string {
	// Pre-allocate string builder for better performance
	var keyBuilder strings.Builder
	keyBuilder.Grow(64) // Pre-allocate reasonable capacity

	keyBuilder.WriteString(method)
	keyBuilder.WriteString(":")
	keyBuilder.WriteString(owner)
	keyBuilder.WriteString(":")
	keyBuilder.WriteString(repo)
	keyBuilder.WriteString(":")

	// Convert filters to JSON for consistent key generation
	if filters != nil && len(filters) > 0 {
		if jsonBytes, err := json.Marshal(filters); err == nil {
			keyBuilder.Write(jsonBytes)
		} else {
			keyBuilder.WriteString("{}")
		}
	} else {
		keyBuilder.WriteString("{}")
	}

	return keyBuilder.String()
}

// CachedClient wraps a Gitea client with caching capabilities
type CachedClient struct {
	client interface{} // The underlying client (client.Client interface)
	cache  *Cache
}

// NewCachedClient creates a new cached client wrapper
func NewCachedClient(client interface{}, maxCacheSize int, cacheTTL time.Duration) (*CachedClient, error) {
	cache, err := NewCache(maxCacheSize, cacheTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &CachedClient{
		client: client,
		cache:  cache,
	}, nil
}

// Cache returns the underlying cache instance for direct access
func (cc *CachedClient) Cache() *Cache {
	return cc.cache
}

// CacheResponse caches a response for a given key
func (cc *CachedClient) CacheResponse(key string, response interface{}) {
	cc.cache.Set(key, response)
}

// GetCachedResponse retrieves a cached response for a given key
func (cc *CachedClient) GetCachedResponse(key string) (interface{}, bool) {
	return cc.cache.Get(key)
}
