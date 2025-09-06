package server

import (
	"context"
	"fmt"
	"math/rand/v2"
	"path"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

const (
	testRepoName = "test-repo"
	testUser     = "test-user"
	testRepo     = testUser + "/" + testRepoName
)

// TestDataSeeder provides comprehensive test data seeding for SDK scenarios
type TestDataSeeder struct {
	baseTime time.Time
	userPool []*gitea.User
}

// NewTestDataSeeder creates a new test data seeder with default configuration
func NewTestDataSeeder() *TestDataSeeder {
	return &TestDataSeeder{
		baseTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		userPool: []*gitea.User{
			{ID: 1, UserName: "alice", Email: "alice@example.com"},
			{ID: 2, UserName: "bob", Email: "bob@example.com"},
			{ID: 3, UserName: "charlie", Email: "charlie@example.com"},
			{ID: 4, UserName: "diana", Email: "diana@example.com"},
			{ID: 5, UserName: "eve", Email: "eve@example.com"},
		},
	}
}

// SeedPRs generates test PR data with realistic scenarios
func (s *TestDataSeeder) SeedPRs(count int, options SeedOptions) []*gitea.PullRequest {
	prs := make([]*gitea.PullRequest, count)
	states := []gitea.StateType{gitea.StateOpen, gitea.StateClosed}

	for _, user := range s.userPool {
		for i := range count {
			user := user
			state := states[i%len(states)]
			createdTime := s.baseTime.Add(time.Duration(i) * time.Hour)
			pr := &gitea.PullRequest{
				ID:      int64(i + 1),
				Index:   int64(i + 1),
				Title:   fmt.Sprintf("%s PR %d", options.Prefix, i+1),
				State:   state,
				Body:    fmt.Sprintf("Description for %s PR %d", options.Prefix, i+1),
				Poster:  user,
				Created: &createdTime,
				Updated: &createdTime,
				HTMLURL: fmt.Sprintf("https://%s.com/pr/%d", options.Domain, i+1),
			}

			if options.IncludeLabels && i%3 == 0 {
				pr.Labels = []*gitea.Label{
					{Name: "enhancement", Color: "84cc16"},
					{Name: "documentation", Color: "10b981"},
				}
			}

			prs[i] = pr
		}
	}
	return prs
}

// SeedIssues generates test issue data with realistic scenarios
func (s *TestDataSeeder) SeedIssues(count int, options SeedOptions) []*gitea.Issue {
	issues := make([]*gitea.Issue, 0, count)
	states := []string{"open", "closed"}

	for _, user := range s.userPool {
		for i := range count {
			user := user
			state := states[i%len(states)]
			createdTime := s.baseTime.Add(time.Duration(i) * time.Hour)

			issue := &gitea.Issue{
				ID:      int64(i + 1),
				Index:   int64(i + 1),
				Title:   fmt.Sprintf("%s Issue %d", options.Prefix, i+1),
				State:   gitea.StateType(state),
				Body:    fmt.Sprintf("Description for %s issue %d", options.Prefix, i+1),
				Poster:  user,
				Created: createdTime,
				Updated: createdTime,
				HTMLURL: fmt.Sprintf("https://%s.com/issue/%d", options.Domain, i+1),
			}

			if options.IncludeLabels && i%2 == 0 {
				issue.Labels = []*gitea.Label{
					{Name: "bug", Color: "ef4444"},
					{Name: "help wanted", Color: "f59e0b"},
				}
			}

			issues = append(issues, issue)
		}
	}
	return issues
}

// SeedRepos generates test repository data with realistic scenarios
func (s *TestDataSeeder) SeedRepos(count int, options SeedOptions) []*gitea.Repository {
	repos := make([]*gitea.Repository, 0, count)

	for _, user := range s.userPool {
		for i := range count {
			user := user
			repo := &gitea.Repository{
				ID:          int64(i + 1),
				Name:        fmt.Sprintf("%s-repo-%d", options.Prefix, i+1),
				FullName:    fmt.Sprintf("%s/%s-repo-%d", user.UserName, options.Prefix, i+1),
				Description: fmt.Sprintf("Test repository %d for %s", i+1, options.Prefix),
				Private:     i%5 == 0, // Every 5th repo is private
				Owner:       user,
				HTMLURL:     fmt.Sprintf("https://%s.com/%s/%s-repo-%d", options.Domain, user.UserName, options.Prefix, i+1),
			}

			repos = append(repos, repo)
		}
	}
	return repos
}

// SeedOptions configures test data seeding behavior
type SeedOptions struct {
	Prefix        string
	Domain        string
	IncludeLabels bool
}

// DefaultSeedOptions returns default seeding options
func DefaultSeedOptions() SeedOptions {
	return SeedOptions{
		Prefix:        "test",
		Domain:        "example",
		IncludeLabels: true,
	}
}

// TestSDKDataSeeding tests the comprehensive test data seeding system
func TestSDKDataSeeding(t *testing.T) {
	seeder := NewTestDataSeeder()
	options := DefaultSeedOptions()

	// Test PR seeding
	prs := seeder.SeedPRs(5, options)
	if len(prs) != 5 {
		t.Errorf("Expected 5 PRs, got %d", len(prs))
	}
	if prs[0].Title != "test PR 1" {
		t.Errorf("Expected PR title 'test PR 1', got '%s'", prs[0].Title)
	}

	// Test Issue seeding
	issues := seeder.SeedIssues(3, options)
	if len(issues) != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}
	if issues[0].Title != "test Issue 1" {
		t.Errorf("Expected issue title 'test Issue 1', got '%s'", issues[0].Title)
	}

	// Test Repository seeding
	repos := seeder.SeedRepos(4, options)
	if len(repos) != 4*len(seeder.userPool) {
		t.Errorf("Expected 4 repos, got %d", len(repos))
	}
	if repos[0].Name != "test-repo-1" {
		t.Errorf("Expected repo name 'test-repo-1', got '%s'", repos[0].Name)
	}
}

// TestSDKDataSeedingIntegration tests data seeding with mock client integration
func TestSDKDataSeedingIntegration(t *testing.T) {
	seeder := NewTestDataSeeder()
	options := DefaultSeedOptions()

	// Seed comprehensive test data
	prs := seeder.SeedPRs(3, options)
	issues := seeder.SeedIssues(3, options)
	repos := seeder.SeedRepos(3, options)

	// Create mock client with seeded data
	mockClient := &giteasdk.MockGiteaClient{
		MockPRs:    prs,
		MockIssues: issues,
		MockRepos:  repos,
	}

	// Test integration with handlers
	logger := logrus.New()
	prHandler := &SDKPRListHandler{logger: logger, client: mockClient}
	issueHandler := &SDKIssueListHandler{logger: logger, client: mockClient}
	repoHandler := &SDKRepositoryHandler{logger: logger, client: mockClient}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	u := seeder.userPool[rand.IntN(len(seeder.userPool))]
	rp := []string{}
	for _, r := range repos {
		if r.Owner.UserName == u.UserName {
			rp = append(rp, r.Name)
		}
	}
	if len(rp) == 0 {
		rp = append(rp, "test-repo") // fallback
	}
	name := rp[rand.IntN(len(rp))]
	// Test all handlers with seeded data
	prArgs := PRListArgs{Repository: path.Join(u.UserName, name)}

	prResult, prData, prErr := prHandler.HandlePRListRequest(ctx, req, prArgs)
	if prErr != nil {
		t.Fatalf("PR handler failed: %v", prErr)
	}
	if prResult == nil || prData == nil {
		t.Fatal("PR handler returned nil results")
	}

	issueArgs := IssueListArgs{Repository: path.Join(u.UserName, name)}

	issueResult, issueData, issueErr := issueHandler.HandleIssueListRequest(ctx, req, issueArgs)
	if issueErr != nil {
		t.Fatalf("Issue handler failed: %v", issueErr)
	}
	if issueResult == nil || issueData == nil {
		t.Fatal("Issue handler returned nil results")
	}

	repoArgs := RepoListArgs{}

	repoResult, repoData, repoErr := repoHandler.ListRepositories(ctx, req, repoArgs)
	if repoErr != nil {
		t.Fatalf("Repository handler failed: %v", repoErr)
	}
	if repoResult == nil || repoData == nil {
		t.Fatal("Repository handler returned nil results")
	}

	// Verify seeded data integrity
	prDataMap := prData.(map[string]any)
	if prDataMap["total"] != 2 {
		t.Errorf("Expected 2 seeded PRs, got %v", prDataMap["total"])
	}

	issueDataMap := issueData.(map[string]any)
	if issueDataMap["total"] != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 seeded issues, got %v", issueDataMap["total"])
	}

	repoDataMap := repoData.(map[string]any)
	if repoDataMap["total"] != 3*len(seeder.userPool) {
		t.Errorf("Expected 3 seeded repos, got %v", repoDataMap["total"])
	}
}
