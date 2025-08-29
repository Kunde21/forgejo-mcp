// Package context provides functionality for detecting and managing Git repository context
package context

import (
	"fmt"
	"sync"
	"time"
)

// Context represents repository context information extracted from a git repository.
// It contains the repository owner, name, and remote URL information.
type Context struct {
	Owner      string // Repository owner or organization name
	Repository string // Repository name
	RemoteURL  string // Full remote URL (HTTPS or SSH)
}

// String returns a human-readable string representation of the context
// in the format "owner/repository (remote-url)".
func (c *Context) String() string {
	return fmt.Sprintf("%s/%s (%s)", c.Owner, c.Repository, c.RemoteURL)
}

// ContextManager manages repository context detection with caching for performance.
// It caches context results to avoid repeated git operations and provides thread-safe access.
type ContextManager struct {
	cache   map[string]*cacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
}

type cacheEntry struct {
	context   *Context
	timestamp time.Time
}

// NewContextManager creates a new context manager with default settings:
// - TTL: 5 minutes
// - Max cache size: 100 entries
func NewContextManager() *ContextManager {
	return &ContextManager{
		cache:   make(map[string]*cacheEntry),
		ttl:     5 * time.Minute,
		maxSize: 100,
	}
}

// NewContextManagerWithTTL creates a new context manager with custom TTL.
// The max cache size remains at the default of 100 entries.
func NewContextManagerWithTTL(ttl time.Duration) *ContextManager {
	cm := NewContextManager()
	cm.ttl = ttl
	return cm
}

// DetectContext detects repository context for the given path.
// It returns a cached result if available and not expired, otherwise performs
// fresh context detection and caches the result.
//
// The function performs these steps:
// 1. Checks if the path is a git repository (including worktrees)
// 2. Extracts the remote URL from the git configuration
// 3. Validates that the remote points to a Forgejo instance
// 4. Parses the owner and repository name from the URL
//
// Returns an error if:
// - The path is not a git repository
// - No remote is configured
// - The remote doesn't point to a Forgejo instance
// - The repository URL cannot be parsed
func (cm *ContextManager) DetectContext(path string) (*Context, error) {
	if cm == nil {
		return detectContextUncached(path)
	}

	// Check cache first
	cm.mutex.RLock()
	if entry, exists := cm.cache[path]; exists {
		if time.Since(entry.timestamp) < cm.ttl {
			cm.mutex.RUnlock()
			return entry.context, nil
		}
	}
	cm.mutex.RUnlock()

	// Not in cache or expired, detect context
	context, err := detectContextUncached(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Clean up expired entries if cache is getting full
	if len(cm.cache) >= cm.maxSize {
		cm.cleanupExpired()
	}

	// If still at max size, remove oldest entry
	if len(cm.cache) >= cm.maxSize {
		cm.removeOldest()
	}

	cm.cache[path] = &cacheEntry{
		context:   context,
		timestamp: time.Now(),
	}

	return context, nil
}

// DetectContext is a convenience function that uses a default context manager.
// It provides the same functionality as calling DetectContext on a ContextManager
// but uses a shared default instance for simple use cases.
//
// For more control over caching behavior, create a custom ContextManager instance.
func DetectContext(path string) (*Context, error) {
	return defaultManager.DetectContext(path)
}

// detectContextUncached performs context detection without caching
func detectContextUncached(path string) (*Context, error) {
	// Check if it's a git repository
	if !IsGitRepository(path) {
		return nil, fmt.Errorf("not a git repository: %s", path)
	}

	// Get remote URL from the specified path
	remoteURL, err := GetRemoteURLInDir("", path)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Validate it's a Forgejo remote
	if !IsForgejoRemote(remoteURL) {
		return nil, fmt.Errorf("remote is not a Forgejo instance: %s", remoteURL)
	}

	// Parse repository information
	owner, repo, err := ParseRepository(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository from URL: %w", err)
	}

	return &Context{
		Owner:      owner,
		Repository: repo,
		RemoteURL:  remoteURL,
	}, nil
}

// cleanupExpired removes expired cache entries
func (cm *ContextManager) cleanupExpired() {
	now := time.Now()
	for path, entry := range cm.cache {
		if now.Sub(entry.timestamp) >= cm.ttl {
			delete(cm.cache, path)
		}
	}
}

// removeOldest removes the oldest cache entry
func (cm *ContextManager) removeOldest() {
	var oldestPath string
	var oldestTime time.Time
	first := true

	for path, entry := range cm.cache {
		if first || entry.timestamp.Before(oldestTime) {
			oldestPath = path
			oldestTime = entry.timestamp
			first = false
		}
	}

	if oldestPath != "" {
		delete(cm.cache, oldestPath)
	}
}

// ClearCache clears all cached entries from the context manager.
// This can be useful when you want to force fresh context detection
// for all subsequent calls.
func (cm *ContextManager) ClearCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.cache = make(map[string]*cacheEntry)
}

// CacheSize returns the current number of cached entries.
// This can be used for monitoring cache usage and performance tuning.
func (cm *ContextManager) CacheSize() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.cache)
}

// Default context manager instance
var defaultManager = NewContextManager()
