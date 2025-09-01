package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestDataSeeder handles seeding test data into Forgejo instance
type TestDataSeeder struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewTestDataSeeder creates a new test data seeder
func NewTestDataSeeder(baseURL, token string) *TestDataSeeder {
	return &TestDataSeeder{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SeedTestData creates test repositories, issues, and PRs
func (s *TestDataSeeder) SeedTestData(t *testing.T) error {
	t.Helper()

	// Create test repositories
	if err := s.createTestRepositories(t); err != nil {
		return fmt.Errorf("failed to create test repositories: %w", err)
	}

	// Create test issues
	if err := s.createTestIssues(t); err != nil {
		return fmt.Errorf("failed to create test issues: %w", err)
	}

	// Create test pull requests
	if err := s.createTestPullRequests(t); err != nil {
		return fmt.Errorf("failed to create test pull requests: %w", err)
	}

	return nil
}

// createTestRepositories creates sample repositories for testing
func (s *TestDataSeeder) createTestRepositories(t *testing.T) error {
	repositories := []map[string]interface{}{
		{
			"name":        "test-repo-1",
			"description": "First test repository",
			"private":     false,
		},
		{
			"name":        "test-repo-2",
			"description": "Second test repository",
			"private":     true,
		},
		{
			"name":        "test-repo-3",
			"description": "Third test repository with issues",
			"private":     false,
		},
	}

	for _, repo := range repositories {
		if err := s.createRepository(repo); err != nil {
			t.Logf("Failed to create repository %s: %v", repo["name"], err)
			// Continue with other repositories
		}
	}

	return nil
}

// createRepository creates a single repository
func (s *TestDataSeeder) createRepository(repoData map[string]interface{}) error {
	url := s.baseURL + "/api/v1/user/repos"

	jsonData, err := json.Marshal(repoData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create repository, status: %d", resp.StatusCode)
	}

	return nil
}

// createTestIssues creates sample issues for testing
func (s *TestDataSeeder) createTestIssues(t *testing.T) error {
	issues := []map[string]interface{}{
		{
			"title":  "Bug: Application crashes on startup",
			"body":   "The application crashes immediately when started on Windows 11.",
			"labels": []string{"bug", "critical"},
			"repo":   "test-repo-3",
		},
		{
			"title":  "Feature: Add dark mode support",
			"body":   "Users have requested dark mode support for better visibility in low light conditions.",
			"labels": []string{"enhancement", "ui"},
			"repo":   "test-repo-3",
		},
		{
			"title":  "Documentation: Update API docs",
			"body":   "The API documentation needs to be updated to reflect recent changes.",
			"labels": []string{"documentation"},
			"repo":   "test-repo-3",
		},
	}

	for _, issue := range issues {
		if err := s.createIssue(issue); err != nil {
			t.Logf("Failed to create issue %s: %v", issue["title"], err)
			// Continue with other issues
		}
	}

	return nil
}

// createIssue creates a single issue
func (s *TestDataSeeder) createIssue(issueData map[string]interface{}) error {
	repo := issueData["repo"].(string)
	url := fmt.Sprintf("%s/api/v1/repos/testuser/%s/issues", s.baseURL, repo)

	// Remove repo from data before marshaling
	issuePayload := map[string]interface{}{
		"title":  issueData["title"],
		"body":   issueData["body"],
		"labels": issueData["labels"],
	}

	jsonData, err := json.Marshal(issuePayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create issue, status: %d", resp.StatusCode)
	}

	return nil
}

// createTestPullRequests creates sample pull requests for testing
func (s *TestDataSeeder) createTestPullRequests(t *testing.T) error {
	// Note: Creating PRs requires actual git operations and branches
	// For this test implementation, we'll create some basic PR data
	// In a real implementation, you'd:
	// 1. Create branches in the repository
	// 2. Make commits to those branches
	// 3. Create PRs via the API

	prs := []map[string]interface{}{
		{
			"title": "Fix: Resolve application crash on startup",
			"body":  "This PR fixes the application crash issue reported in issue #1.",
			"head":  "fix-crash",
			"base":  "main",
			"repo":  "test-repo-3",
		},
		{
			"title": "Feature: Implement dark mode toggle",
			"body":  "Adds a dark mode toggle to the application settings.",
			"head":  "dark-mode",
			"base":  "main",
			"repo":  "test-repo-3",
		},
	}

	for _, pr := range prs {
		if err := s.createPullRequest(pr); err != nil {
			t.Logf("Failed to create PR %s: %v", pr["title"], err)
			// Continue with other PRs
		}
	}

	return nil
}

// createPullRequest creates a single pull request
func (s *TestDataSeeder) createPullRequest(prData map[string]interface{}) error {
	repo := prData["repo"].(string)
	url := fmt.Sprintf("%s/api/v1/repos/testuser/%s/pulls", s.baseURL, repo)

	// Prepare PR payload
	prPayload := map[string]interface{}{
		"title": prData["title"],
		"body":  prData["body"],
		"head":  prData["head"],
		"base":  prData["base"],
	}

	jsonData, err := json.Marshal(prPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create PR, status: %d", resp.StatusCode)
	}

	return nil
}

// CleanupTestData removes test data created during testing
func (s *TestDataSeeder) CleanupTestData(t *testing.T) error {
	t.Helper()

	// Clean up test repositories
	repositories := []string{"test-repo-1", "test-repo-2", "test-repo-3"}

	for _, repo := range repositories {
		if err := s.deleteRepository(repo); err != nil {
			t.Logf("Failed to delete repository %s: %v", repo, err)
			// Continue with other repositories
		}
	}

	return nil
}

// deleteRepository deletes a test repository
func (s *TestDataSeeder) deleteRepository(repoName string) error {
	url := fmt.Sprintf("%s/api/v1/repos/testuser/%s", s.baseURL, repoName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete repository, status: %d", resp.StatusCode)
	}

	return nil
}
