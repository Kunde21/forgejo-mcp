# Gitea SDK Integration Guide

This document provides guidance on using the Gitea SDK client integration for Forgejo repositories in the MCP server.

## Overview

The Gitea SDK client provides a direct integration with Forgejo repositories using the official Gitea SDK, offering better performance and more features compared to the CLI-based approach.

## Getting Started

### Prerequisites

- A Forgejo instance URL
- A personal access token with appropriate permissions

### Basic Usage

```go
import (
    "github.com/Kunde21/forgejo-mcp/client"
)

// Create a new client
config := client.DefaultConfig()
client, err := client.NewWithConfig("https://your.forgejo.instance", "your-token", config)
if err != nil {
    log.Fatal(err)
}

// List pull requests
filters := &client.PullRequestFilters{
    State: client.StateOpen,
}
prs, err := client.ListPRs("owner", "repo", filters)
if err != nil {
    log.Fatal(err)
}

// List issues
issueFilters := &client.IssueFilters{
    State: client.StateOpen,
    Labels: []string{"bug", "urgent"},
}
issues, err := client.ListIssues("owner", "repo", issueFilters)
if err != nil {
    log.Fatal(err)
}
```

## Client Interface

The client provides a clean interface for interacting with Forgejo repositories:

```go
type Client interface {
    // ListPRs retrieves pull requests for a repository with optional filters
    ListPRs(owner, repo string, filters *PullRequestFilters) ([]PullRequest, error)

    // ListIssues retrieves issues for a repository with optional filters
    ListIssues(owner, repo string, filters *IssueFilters) ([]Issue, error)

    // ListRepositories retrieves repositories with optional filters
    ListRepositories(filters *RepositoryFilters) ([]Repository, error)

    // GetRepository retrieves a specific repository by owner and name
    GetRepository(owner, name string) (*Repository, error)
}
```

## Configuration

The client can be configured with custom timeouts and user agents:

```go
config := &client.ClientConfig{
    Timeout:   60 * time.Second,
    UserAgent: "my-app/1.0.0",
}
client, err := client.NewWithConfig("https://your.forgejo.instance", "your-token", config)
```

## Filtering

### Pull Request Filters

```go
filters := &client.PullRequestFilters{
    State:     client.StateOpen,  // open, closed, or all
    Page:      1,
    PageSize:  30,
    Sort:      "created",         // or "updated", "comments"
    Milestone: 123,              // milestone ID
}
```

### Issue Filters

```go
filters := &client.IssueFilters{
    State:       client.StateOpen,     // open, closed, or all
    Page:        1,
    PageSize:    30,
    Labels:      []string{"bug", "frontend"},
    Milestones:  []string{"v1.0", "v1.1"},
    KeyWord:     "performance",
    CreatedBy:   "username",
    AssignedBy:  "assignee",
    MentionedBy: "mentioned_user",
    Owner:       "repo_owner",
    Team:        "team_name",
}
```

### Repository Filters

```go
filters := &client.RepositoryFilters{
    Page:            1,
    PageSize:        30,
    Query:           "search term",
    OwnerID:         123,
    StarredByUser:   456,
    Type:            "source",      // source, fork, mirror
    IsPrivate:       &trueValue,    // pointer to bool
    IsArchived:      &falseValue,   // pointer to bool
    Sort:            "created",     // created, updated, id, name, size
    Order:           "asc",         // asc or desc
    ExcludeTemplate: true,
}
```

## Error Handling

The client provides specific error types for better error handling:

```go
_, err := client.ListPRs("owner", "repo", filters)
if err != nil {
    var validationErr *client.ValidationError
    if errors.As(err, &validationErr) {
        // Handle validation errors
        log.Printf("Validation error in field %s: %s", validationErr.Field, validationErr.Message)
    } else {
        // Handle other errors
        log.Printf("Error listing PRs: %v", err)
    }
}
```

## Performance Features

### Caching

The client includes built-in caching capabilities:

```go
// Create a cached client
cache, err := tea.NewCache(1000, 5*time.Minute)
if err != nil {
    log.Fatal(err)
}

cachedClient := &tea.CachedClient{
    client: client,
    cache:  cache,
}
```

### Batch Processing

For handling multiple requests efficiently:

```go
// Create a batch processor
processor := tea.NewBatchProcessor(10) // max 10 concurrent requests

// Create batch requests
requests := []tea.BatchRequest{
    {ID: "1", Method: "listPRs", Owner: "owner1", Repo: "repo1"},
    {ID: "2", Method: "listIssues", Owner: "owner2", Repo: "repo2"},
}

// Process batch
responses, err := processor.ProcessBatch(context.Background(), requests)
```

## Integration with MCP Server

The Gitea SDK client is integrated into the MCP server through dedicated handlers:

- `GiteaSDKPRListHandler` for pull request operations
- `GiteaSDKIssueListHandler` for issue operations

These handlers transform the Gitea SDK responses into the MCP-compatible format automatically.

## Troubleshooting

### Common Issues

1. **Authentication failures**: Ensure your token has the correct permissions
2. **Timeout errors**: Increase the client timeout for large repositories
3. **Rate limiting**: Implement exponential backoff for high-frequency requests

### Debugging

Enable debug logging to see detailed request/response information:

```go
import "github.com/sirupsen/logrus"

logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
```