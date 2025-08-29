package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestContextString(t *testing.T) {
	ctx := &Context{
		Owner:      "testuser",
		Repository: "testrepo",
		RemoteURL:  "https://codeberg.org/testuser/testrepo.git",
	}

	expected := "testuser/testrepo (https://codeberg.org/testuser/testrepo.git)"
	if ctx.String() != expected {
		t.Errorf("Context.String() = %q, expected %q", ctx.String(), expected)
	}
}

func TestNewContextManager(t *testing.T) {
	cm := NewContextManager()

	if cm == nil {
		t.Fatal("NewContextManager() returned nil")
	}

	if cm.cache == nil {
		t.Error("ContextManager cache should be initialized")
	}

	if cm.ttl != 5*time.Minute {
		t.Errorf("Expected TTL to be 5 minutes, got %v", cm.ttl)
	}

	if cm.maxSize != 100 {
		t.Errorf("Expected maxSize to be 100, got %d", cm.maxSize)
	}
}

func TestNewContextManagerWithTTL(t *testing.T) {
	customTTL := 10 * time.Minute
	cm := NewContextManagerWithTTL(customTTL)

	if cm.ttl != customTTL {
		t.Errorf("Expected TTL to be %v, got %v", customTTL, cm.ttl)
	}
}

func TestDetectContextUncached(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "context_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock git repository
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repository
	initCmd := exec.Command("git", "init")
	initCmd.Dir = repoDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	tests := []struct {
		name        string
		setup       func()
		expectedErr bool
	}{
		{
			name: "non-git directory",
			setup: func() {
				// Test with temp directory (not a git repo)
			},
			expectedErr: true,
		},
		{
			name: "git repo without remote",
			setup: func() {
				// Already in git repo, but no remote
			},
			expectedErr: true,
		},
		{
			name: "git repo with forgejo remote",
			setup: func() {
				remoteCmd := exec.Command("git", "remote", "add", "origin", "https://codeberg.org/testuser/testrepo.git")
				remoteCmd.Dir = repoDir
				remoteCmd.Run()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			var testDir string
			if tt.name == "non-git directory" {
				testDir = tempDir
			} else {
				testDir = repoDir
			}

			ctx, err := detectContextUncached(testDir)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("detectContextUncached() expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("detectContextUncached() unexpected error: %v", err)
				}
				if ctx == nil {
					t.Error("detectContextUncached() returned nil context")
				} else {
					if ctx.Owner != "testuser" {
						t.Errorf("Expected owner 'testuser', got %q", ctx.Owner)
					}
					if ctx.Repository != "testrepo" {
						t.Errorf("Expected repository 'testrepo', got %q", ctx.Repository)
					}
				}
			}
		})
	}
}

func TestContextManagerDetectContext(t *testing.T) {
	cm := NewContextManager()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "context_manager_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-git directory
	_, err = cm.DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}

	// Verify cache is empty
	if cm.CacheSize() != 0 {
		t.Errorf("Expected cache size 0, got %d", cm.CacheSize())
	}
}

func TestContextManagerCaching(t *testing.T) {
	// Create manager with short TTL for testing
	cm := NewContextManagerWithTTL(100 * time.Millisecond)

	// Test cache operations
	cm.ClearCache()
	if cm.CacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cm.CacheSize())
	}

	// Test with non-existent path (should not cache errors)
	tempDir, _ := os.MkdirTemp("", "cache_test_*")
	defer os.RemoveAll(tempDir)

	_, err := cm.DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}

	// Cache should still be empty for failed detections
	if cm.CacheSize() != 0 {
		t.Errorf("Expected cache size 0 for failed detection, got %d", cm.CacheSize())
	}
}

func TestContextManagerCacheExpiration(t *testing.T) {
	// Create manager with very short TTL
	cm := NewContextManagerWithTTL(1 * time.Millisecond)

	tempDir, _ := os.MkdirTemp("", "expiration_test_*")
	defer os.RemoveAll(tempDir)

	// First call should fail and not cache
	_, err := cm.DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}

	// Wait for cache entry to expire (if it existed)
	time.Sleep(10 * time.Millisecond)

	// Second call should also fail (no caching of errors)
	_, err = cm.DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory on second call")
	}
}

func TestContextManagerCacheSizeLimit(t *testing.T) {
	// Create manager with small cache size
	cm := &ContextManager{
		cache:   make(map[string]*cacheEntry),
		ttl:     time.Hour, // Long TTL so entries don't expire
		maxSize: 2,         // Only allow 2 entries
	}

	// Fill cache with dummy entries
	cm.cache["/path1"] = &cacheEntry{
		context:   &Context{Owner: "user1", Repository: "repo1", RemoteURL: "https://example.com"},
		timestamp: time.Now(),
	}
	cm.cache["/path2"] = &cacheEntry{
		context:   &Context{Owner: "user2", Repository: "repo2", RemoteURL: "https://example.com"},
		timestamp: time.Now(),
	}

	// Verify initial size
	if cm.CacheSize() != 2 {
		t.Errorf("Expected cache size 2, got %d", cm.CacheSize())
	}

	// Adding a third entry should trigger cleanup
	tempDir, _ := os.MkdirTemp("", "size_test_*")
	defer os.RemoveAll(tempDir)

	_, _ = cm.DetectContext(tempDir) // This will fail but test cache management

	// Cache size should still be manageable
	if cm.CacheSize() > cm.maxSize {
		t.Errorf("Cache size %d exceeds maxSize %d", cm.CacheSize(), cm.maxSize)
	}
}

func TestDetectContextConvenienceFunction(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "convenience_test_*")
	defer os.RemoveAll(tempDir)

	// Test the convenience function
	_, err := DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}
}

func TestContextManagerNilReceiver(t *testing.T) {
	var cm *ContextManager

	tempDir, _ := os.MkdirTemp("", "nil_test_*")
	defer os.RemoveAll(tempDir)

	// Should not panic and should call detectContextUncached
	_, err := cm.DetectContext(tempDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}
}

// Integration Tests

func TestCompleteContextDetectionFlow(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up a complete git repository with Forgejo remote
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repository
	initCmd := exec.Command("git", "init")
	initCmd.Dir = repoDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = repoDir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	// Add Forgejo remote
	remoteCmd := exec.Command("git", "remote", "add", "origin", "https://codeberg.org/testuser/testrepo.git")
	remoteCmd.Dir = repoDir
	if err := remoteCmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// Test 1: Direct context detection
	t.Run("DirectContextDetection", func(t *testing.T) {
		ctx, err := detectContextUncached(repoDir)
		if err != nil {
			t.Fatalf("detectContextUncached failed: %v", err)
		}

		if ctx.Owner != "testuser" {
			t.Errorf("Expected owner 'testuser', got %q", ctx.Owner)
		}
		if ctx.Repository != "testrepo" {
			t.Errorf("Expected repository 'testrepo', got %q", ctx.Repository)
		}
		if ctx.RemoteURL != "https://codeberg.org/testuser/testrepo.git" {
			t.Errorf("Expected remote URL 'https://codeberg.org/testuser/testrepo.git', got %q", ctx.RemoteURL)
		}
	})

	// Test 2: Context manager with caching
	t.Run("ContextManagerWithCaching", func(t *testing.T) {
		cm := NewContextManager()

		// First call should detect and cache
		ctx1, err := cm.DetectContext(repoDir)
		if err != nil {
			t.Fatalf("First DetectContext failed: %v", err)
		}

		// Second call should return cached result
		ctx2, err := cm.DetectContext(repoDir)
		if err != nil {
			t.Fatalf("Second DetectContext failed: %v", err)
		}

		// Results should be identical
		if ctx1.Owner != ctx2.Owner || ctx1.Repository != ctx2.Repository || ctx1.RemoteURL != ctx2.RemoteURL {
			t.Error("Cached result differs from original")
		}

		// Cache should have one entry
		if cm.CacheSize() != 1 {
			t.Errorf("Expected cache size 1, got %d", cm.CacheSize())
		}
	})

	// Test 3: Multiple repositories
	t.Run("MultipleRepositories", func(t *testing.T) {
		cm := NewContextManager()

		// Create second repository
		repoDir2 := filepath.Join(tempDir, "test-repo2")
		if err := os.MkdirAll(repoDir2, 0755); err != nil {
			t.Fatalf("Failed to create repo2 dir: %v", err)
		}

		initCmd := exec.Command("git", "init")
		initCmd.Dir = repoDir2
		if err := initCmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo2: %v", err)
		}

		remoteCmd := exec.Command("git", "remote", "add", "origin", "https://codeberg.org/otheruser/otherrepo.git")
		remoteCmd.Dir = repoDir2
		if err := remoteCmd.Run(); err != nil {
			t.Fatalf("Failed to add remote to repo2: %v", err)
		}

		// Detect context for both repositories
		ctx1, err := cm.DetectContext(repoDir)
		if err != nil {
			t.Fatalf("Failed to detect context for repo1: %v", err)
		}

		ctx2, err := cm.DetectContext(repoDir2)
		if err != nil {
			t.Fatalf("Failed to detect context for repo2: %v", err)
		}

		// Contexts should be different
		if ctx1.Owner == ctx2.Owner && ctx1.Repository == ctx2.Repository {
			t.Error("Expected different contexts for different repositories")
		}

		// Cache should have two entries
		if cm.CacheSize() != 2 {
			t.Errorf("Expected cache size 2, got %d", cm.CacheSize())
		}
	})

	// Test 4: Error scenarios
	t.Run("ErrorScenarios", func(t *testing.T) {
		cm := NewContextManager()

		// Non-git directory
		_, err := cm.DetectContext(tempDir)
		if err == nil {
			t.Error("Expected error for non-git directory")
		}

		// Git directory without remote
		gitOnlyDir := filepath.Join(tempDir, "git-only")
		if err := os.MkdirAll(filepath.Join(gitOnlyDir, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create git-only dir: %v", err)
		}

		_, err = cm.DetectContext(gitOnlyDir)
		if err == nil {
			t.Error("Expected error for git directory without remote")
		}

		// Git directory with non-Forgejo remote
		nonForgejoDir := filepath.Join(tempDir, "non-forgejo")
		if err := os.MkdirAll(nonForgejoDir, 0755); err != nil {
			t.Fatalf("Failed to create non-forgejo dir: %v", err)
		}

		initCmd := exec.Command("git", "init")
		initCmd.Dir = nonForgejoDir
		if err := initCmd.Run(); err != nil {
			t.Fatalf("Failed to init non-forgejo git repo: %v", err)
		}

		remoteCmd := exec.Command("git", "remote", "add", "origin", "https://github.com/user/repo.git")
		remoteCmd.Dir = nonForgejoDir
		if err := remoteCmd.Run(); err != nil {
			t.Fatalf("Failed to add non-forgejo remote: %v", err)
		}

		_, err = cm.DetectContext(nonForgejoDir)
		if err == nil {
			t.Error("Expected error for non-Forgejo remote")
		}
	})
}

func TestContextDetectionWithWorktrees(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "worktree_integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up main repository
	mainRepo := filepath.Join(tempDir, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatalf("Failed to create main repo dir: %v", err)
	}

	initCmd := exec.Command("git", "init")
	initCmd.Dir = mainRepo
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init main repo: %v", err)
	}

	// Configure git user
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = mainRepo
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = mainRepo
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user: %v", err)
	}

	// Add remote to main repo
	remoteCmd := exec.Command("git", "remote", "add", "origin", "https://codeberg.org/testuser/testrepo.git")
	remoteCmd.Dir = mainRepo
	if err := remoteCmd.Run(); err != nil {
		t.Fatalf("Failed to add remote to main repo: %v", err)
	}

	// Create a commit in main repo (required for worktree)
	readmePath := filepath.Join(mainRepo, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	addCmd := exec.Command("git", "add", "README.md")
	addCmd.Dir = mainRepo
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add README: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = mainRepo
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create worktree
	worktreeDir := filepath.Join(tempDir, "worktree")
	worktreeCmd := exec.Command("git", "worktree", "add", worktreeDir)
	worktreeCmd.Dir = mainRepo
	if err := worktreeCmd.Run(); err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Test context detection on worktree
	t.Run("WorktreeContextDetection", func(t *testing.T) {
		ctx, err := detectContextUncached(worktreeDir)
		if err != nil {
			t.Fatalf("detectContextUncached failed for worktree: %v", err)
		}

		if ctx.Owner != "testuser" {
			t.Errorf("Expected owner 'testuser', got %q", ctx.Owner)
		}
		if ctx.Repository != "testrepo" {
			t.Errorf("Expected repository 'testrepo', got %q", ctx.Repository)
		}
	})

	// Test context manager with worktree
	t.Run("ContextManagerWithWorktree", func(t *testing.T) {
		cm := NewContextManager()

		ctx, err := cm.DetectContext(worktreeDir)
		if err != nil {
			t.Fatalf("DetectContext failed for worktree: %v", err)
		}

		// Verify caching works for worktree
		ctx2, err := cm.DetectContext(worktreeDir)
		if err != nil {
			t.Fatalf("Second DetectContext failed for worktree: %v", err)
		}

		if ctx.Owner != ctx2.Owner || ctx.Repository != ctx2.Repository {
			t.Error("Cached worktree context differs from original")
		}
	})
}
