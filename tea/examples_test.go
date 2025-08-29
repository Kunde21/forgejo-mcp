package tea_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/tea"
)

// ExampleListPRs demonstrates listing pull requests
func ExampleForgejoClient_ListPRs() {
	c, err := client.New("https://your.forgejo.instance", "your-access-token")
	if err != nil {
		log.Fatal(err)
	}

	// List open pull requests
	filters := &client.PullRequestFilters{
		State:    client.StateOpen,
		Page:     1,
		PageSize: 10,
	}
	prs, err := c.ListPRs("owner", "repo", filters)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d open pull requests\n", len(prs))
	for _, pr := range prs {
		fmt.Printf("#%d: %s by %s\n", pr.Index, pr.Title, pr.Poster.UserName)
	}
}

// ExampleListIssues demonstrates listing issues with filters
func ExampleForgejoClient_ListIssues() {
	c, err := client.New("https://your.forgejo.instance", "your-access-token")
	if err != nil {
		log.Fatal(err)
	}

	// List issues with specific labels
	filters := &client.IssueFilters{
		State:      client.StateOpen,
		Labels:     []string{"bug", "high-priority"},
		AssignedBy: "developer",
		Page:       1,
		PageSize:   20,
	}
	issues, err := c.ListIssues("owner", "repo", filters)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d matching issues\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("#%d: %s (created %s)\n", issue.Index, issue.Title, issue.Created.Format("2006-01-02"))
	}
}

// ExampleListRepositories demonstrates listing repositories
func ExampleForgejoClient_ListRepositories() {
	c, err := client.New("https://your.forgejo.instance", "your-access-token")
	if err != nil {
		log.Fatal(err)
	}

	// Search for repositories
	filters := &client.RepositoryFilters{
		Query:    "go microservice",
		Page:     1,
		PageSize: 15,
	}
	repos, err := c.ListRepositories(filters)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d repositories\n", len(repos))
	for _, repo := range repos {
		fmt.Printf("%s: %s\n", repo.FullName, repo.Description)
	}
}

// ExampleGetRepository demonstrates getting a specific repository
func ExampleForgejoClient_GetRepository() {
	c, err := client.New("https://your.forgejo.instance", "your-access-token")
	if err != nil {
		log.Fatal(err)
	}

	// Get a specific repository
	repo, err := c.GetRepository("owner", "repo-name")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Repository: %s\n", repo.FullName)
	fmt.Printf("Description: %s\n", repo.Description)
	fmt.Printf("URL: %s\n", repo.HTMLURL)
}

// ExampleCache demonstrates using the caching functionality
func ExampleCache() {
	// Create a cache with 1000 items and 5 minute TTL
	cache, err := tea.NewCache(1000, 5*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	// Store a value
	cache.Set("key1", "value1")

	// Retrieve a value
	if value, found := cache.Get("key1"); found {
		fmt.Printf("Found value: %v\n", value)
	}

	// Check cache statistics
	stats := cache.Stats()
	fmt.Printf("Cache hits: %d, misses: %d\n", stats.Hits, stats.Misses)
	// Output:
	// Found value: value1
	// Cache hits: 1, misses: 0
}

// ExampleBatchProcessor demonstrates using batch processing
func ExampleBatchProcessor() {
	// Create a batch processor with 5 concurrent workers
	processor := tea.NewBatchProcessor(5)

	// Create batch requests
	requests := []tea.BatchRequest{
		{ID: "1", Method: "listPRs", Owner: "gitea", Repo: "gitea"},
		{ID: "2", Method: "listIssues", Owner: "gitea", Repo: "gitea"},
		{ID: "3", Method: "listRepositories", Owner: "gitea", Repo: "gitea"},
	}

	// Process batch with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	responses, err := processor.ProcessBatch(ctx, requests)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Processed %d requests\n", len(responses))
	for _, response := range responses {
		if response.Error != nil {
			fmt.Printf("Request %s failed: %v\n", response.ID, response.Error)
		} else {
			fmt.Printf("Request %s succeeded in %v\n", response.ID, response.Duration)
		}
	}
}

// ExampleQueryBuilder demonstrates using the query builder
func ExampleQueryBuilder() {
	qb := tea.NewQueryBuilder()

	// Build repository query
	repoFilters := &tea.RepositoryFilters{
		Query: "microservice language:go",
	}
	repoQuery := qb.BuildRepositoryQuery(repoFilters)
	fmt.Printf("Repository query: %s\n", repoQuery)

	// Build issue query
	issueFilters := &tea.IssueFilters{
		KeyWord:   "performance",
		State:     "open",
		Labels:    []string{"bug", "critical"},
		CreatedBy: "developer1",
	}
	issueQuery := qb.BuildIssueQuery(issueFilters)
	fmt.Printf("Issue query: %s\n", issueQuery)

	// Build pull request query
	prFilters := &tea.PullRequestFilters{
		State: "open",
	}
	prQuery := qb.BuildPullRequestQuery(prFilters)
	fmt.Printf("PR query: %s\n", prQuery)

	// Output:
	// Repository query: microservice language:go
	// Issue query: performance state:open label:bug label:critical author:developer1
	// PR query: state:open
}
