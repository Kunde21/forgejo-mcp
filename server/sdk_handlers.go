package server

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// GiteaClientInterface defines the interface for Gitea client operations
type GiteaClientInterface interface {
	ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
	ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error)
}

// SDKPRListHandler handles pr_list tool requests with Gitea SDK integration
type SDKPRListHandler struct {
	logger *logrus.Logger
	client GiteaClientInterface
}

// NewSDKPRListHandler creates a new SDK PR list handler
func NewSDKPRListHandler(logger *logrus.Logger, client GiteaClientInterface) *SDKPRListHandler {
	return &SDKPRListHandler{
		logger: logger,
		client: client,
	}
}

// HandlePRListRequest handles a pr_list request with Gitea SDK integration
func (h *SDKPRListHandler) HandlePRListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
	State  string `json:"state,omitempty"`
	Author string `json:"author,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling pr_list request with Gitea SDK")

	// Build SDK options from parameters
	opts := gitea.ListPullRequestsOptions{}

	if args.State != "" {
		switch args.State {
		case "open":
			opts.State = gitea.StateOpen
		case "closed":
			opts.State = gitea.StateClosed
		case "all":
			opts.State = gitea.StateAll
		default:
			opts.State = gitea.StateOpen // default to open
		}
	} else {
		opts.State = gitea.StateOpen // default to open
	}

	// Note: Gitea SDK doesn't have direct author filtering in ListPullRequestsOptions
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	// For this example, we'll use placeholder owner/repo
	// In a real implementation, this would come from context or configuration
	owner := "example-owner"
	repo := "example-repo"

	prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)
	if err != nil {
		h.logger.Errorf("Gitea SDK ListRepoPullRequests failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK pr list: %v", err),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"pullRequests": h.transformPRsToResponse(prs),
		"total":        len(prs),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d pull requests", len(prs)),
			},
		},
	}, result, nil
}

// transformPRsToResponse transforms Gitea SDK PR data to MCP response format
func (h *SDKPRListHandler) transformPRsToResponse(prs []*gitea.PullRequest) []map[string]interface{} {
	result := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		// Transform to MCP-compatible format
		prData := map[string]interface{}{
			"number": pr.Index,
			"title":  pr.Title,
			"state":  h.normalizePRState(pr.State),
		}

		// Add author if available
		if pr.Poster != nil {
			prData["author"] = pr.Poster.UserName
		} else {
			prData["author"] = ""
		}

		// Add dates if available
		if pr.Created != nil {
			prData["createdAt"] = pr.Created.Format("2006-01-02T15:04:05Z")
		}
		if pr.Updated != nil {
			prData["updatedAt"] = pr.Updated.Format("2006-01-02T15:04:05Z")
		}

		// Add additional metadata for MCP compatibility
		prData["type"] = "pull_request"
		prData["url"] = pr.HTMLURL

		result[i] = prData
	}
	return result
}

// normalizePRState normalizes Gitea SDK PR state to standard values
func (h *SDKPRListHandler) normalizePRState(state gitea.StateType) string {
	switch state {
	case gitea.StateOpen:
		return "open"
	case gitea.StateClosed:
		if state == gitea.StateClosed {
			return "closed"
		}
		return "closed"
	default:
		return "unknown"
	}
}

// SDKRepositoryHandler handles repository operations with Gitea SDK integration
type SDKRepositoryHandler struct {
	logger *logrus.Logger
	client GiteaClientInterface
}

// NewSDKRepositoryHandler creates a new SDK repository handler
func NewSDKRepositoryHandler(logger *logrus.Logger, client GiteaClientInterface) *SDKRepositoryHandler {
	return &SDKRepositoryHandler{
		logger: logger,
		client: client,
	}
}

// ListRepositories handles a repository list request with Gitea SDK integration
func (h *SDKRepositoryHandler) ListRepositories(ctx context.Context, req *mcp.CallToolRequest, args struct {
	Limit int `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling repository list request with Gitea SDK")

	// Build SDK options from parameters
	opts := gitea.ListReposOptions{}

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	repos, _, err := h.client.ListMyRepos(opts)
	if err != nil {
		h.logger.Errorf("Gitea SDK ListMyRepos failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK repository list: %v", err),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"repositories": h.transformReposToResponse(repos),
		"total":        len(repos),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d repositories", len(repos)),
			},
		},
	}, result, nil
}

// transformReposToResponse transforms Gitea SDK repository data to MCP response format
func (h *SDKRepositoryHandler) transformReposToResponse(repos []*gitea.Repository) []map[string]interface{} {
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		// Transform to MCP-compatible format
		repoData := map[string]interface{}{
			"id":       repo.ID,
			"name":     repo.Name,
			"fullName": repo.FullName,
			"private":  repo.Private,
		}

		// Add owner if available
		if repo.Owner != nil {
			repoData["owner"] = repo.Owner.UserName
		} else {
			repoData["owner"] = ""
		}

		// Add description if available
		if repo.Description != "" {
			repoData["description"] = repo.Description
		}

		// Add additional metadata for MCP compatibility
		repoData["type"] = "repository"
		repoData["url"] = repo.HTMLURL

		result[i] = repoData
	}
	return result
}

// SDKIssueListHandler handles issue operations with Gitea SDK integration
type SDKIssueListHandler struct {
	logger *logrus.Logger
	client GiteaClientInterface
}

// NewSDKIssueListHandler creates a new SDK issue list handler
func NewSDKIssueListHandler(logger *logrus.Logger, client GiteaClientInterface) *SDKIssueListHandler {
	return &SDKIssueListHandler{
		logger: logger,
		client: client,
	}
}

// HandleIssueListRequest handles an issue_list request with Gitea SDK integration
func (h *SDKIssueListHandler) HandleIssueListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
	State  string   `json:"state,omitempty"`
	Author string   `json:"author,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling issue_list request with Gitea SDK")

	// Build SDK options from parameters
	opts := gitea.ListIssueOption{}

	if args.State != "" {
		switch args.State {
		case "open":
			opts.State = gitea.StateOpen
		case "closed":
			opts.State = gitea.StateClosed
		case "all":
			opts.State = gitea.StateAll
		default:
			opts.State = gitea.StateOpen // default to open
		}
	} else {
		opts.State = gitea.StateOpen // default to open
	}

	// Note: Gitea SDK doesn't have direct author/label filtering in ListIssueOption
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	issues, _, err := h.client.ListIssues(opts)
	if err != nil {
		h.logger.Errorf("Gitea SDK ListIssues failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK issue list: %v", err),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"issues": h.transformIssuesToResponse(issues),
		"total":  len(issues),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d issues", len(issues)),
			},
		},
	}, result, nil
}

// transformIssuesToResponse transforms Gitea SDK issue data to MCP response format
func (h *SDKIssueListHandler) transformIssuesToResponse(issues []*gitea.Issue) []map[string]interface{} {
	result := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		// Transform to MCP-compatible format
		issueData := map[string]interface{}{
			"number": issue.Index,
			"title":  issue.Title,
			"state":  h.normalizeIssueState(string(issue.State)),
		}

		// Add author if available
		if issue.Poster != nil {
			issueData["author"] = issue.Poster.UserName
		} else {
			issueData["author"] = ""
		}

		// Add dates if available
		issueData["createdAt"] = issue.Created.Format("2006-01-02T15:04:05Z")
		issueData["updatedAt"] = issue.Updated.Format("2006-01-02T15:04:05Z")

		// Add additional metadata for MCP compatibility
		issueData["type"] = "issue"
		issueData["url"] = issue.HTMLURL

		result[i] = issueData
	}
	return result
}

// normalizeIssueState normalizes Gitea SDK issue state to standard values
func (h *SDKIssueListHandler) normalizeIssueState(state string) string {
	switch strings.ToLower(state) {
	case "open":
		return "open"
	case "closed":
		return "closed"
	default:
		return "unknown"
	}
}
