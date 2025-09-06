package server

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"

	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
)

// SDKPRListHandler handles pr_list tool requests with Gitea SDK integration
type SDKPRListHandler struct {
	logger *logrus.Logger
	client giteasdk.GiteaClientInterface
}

// NewSDKPRListHandler creates a new SDK PR list handler
func NewSDKPRListHandler(logger *logrus.Logger, client giteasdk.GiteaClientInterface) *SDKPRListHandler {
	return &SDKPRListHandler{
		logger: logger,
		client: client,
	}
}

// HandlePRListRequest handles a pr_list request with Gitea SDK integration
func (h *SDKPRListHandler) HandlePRListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
	Repository string `json:"repository,omitempty"`
	CWD        string `json:"cwd,omitempty"`
	State      string `json:"state,omitempty"`
	Author     string `json:"author,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling pr_list request with Gitea SDK")

	// Validate repository parameter
	var repoParam string
	if args.Repository != "" {
		repoParam = args.Repository
	} else if args.CWD != "" {
		// Resolve CWD to repository identifier
		var err error
		repoParam, err = giteasdk.ResolveCWDToRepository(args.CWD)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error resolving repository from CWD: %v", err),
					},
				},
			}, nil, nil
		}
	} else {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: repository parameter or cwd parameter is required",
				},
			},
		}, nil, nil
	}

	// Validate repository access
	if valid, err := ValidateRepositoryAccess(h.client, repoParam); !valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	} else if args.CWD != "" {
		// Resolve CWD to repository identifier
		var err error
		repoParam, err = giteasdk.ResolveCWDToRepository(args.CWD)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error resolving repository from CWD: %v", err),
					},
				},
			}, nil, nil
		}
	}

	// Validate repository access
	if valid, err := ValidateRepositoryAccess(h.client, repoParam); !valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	// Parse repository identifier
	owner, repo, _ := strings.Cut(repoParam, "/")

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

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)
	if err != nil {
		sdkErr := giteasdk.NewSDKError("ListRepoPullRequests", err, fmt.Sprintf("owner=%s, repo=%s", owner, repo))
		h.logger.Errorf("%v", sdkErr)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK pr list: %v", sdkErr),
				},
			},
		}, nil, nil
	}

	// Extract repository metadata
	repoMetadata, err := giteasdk.ExtractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]any{}
	}

	// Transform to MCP response format
	result := map[string]any{
		"pullRequests": h.transformPRsToResponse(prs, repoMetadata),
		"total":        len(prs),
		"repository":   repoMetadata,
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
func (h *SDKPRListHandler) transformPRsToResponse(prs []*gitea.PullRequest, repoMetadata map[string]any) []map[string]any {
	result := make([]map[string]any, len(prs))
	for i, pr := range prs {
		// Transform to MCP-compatible format
		prData := map[string]any{
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
		prData["url"] = pr.URL

		// Add repository metadata to individual PR object
		prData["repository"] = repoMetadata

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
	client giteasdk.GiteaClientInterface
}

// NewSDKRepositoryHandler creates a new SDK repository handler
func NewSDKRepositoryHandler(logger *logrus.Logger, client giteasdk.GiteaClientInterface) *SDKRepositoryHandler {
	return &SDKRepositoryHandler{
		logger: logger,
		client: client,
	}
}

// ListRepositories handles a repository list request with Gitea SDK integration
func (h *SDKRepositoryHandler) ListRepositories(ctx context.Context, req *mcp.CallToolRequest, args RepositoryListArgs) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling repository list request with Gitea SDK")

	// Validate arguments
	if err := ValidateRepositoryListArgs(args); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	// Build SDK options from parameters
	opts := gitea.ListReposOptions{}

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	repos, _, err := h.client.ListMyRepos(opts)
	if err != nil {
		sdkErr := giteasdk.NewSDKError("ListMyRepos", err, fmt.Sprintf("limit=%d", args.Limit))
		h.logger.Errorf("%v", sdkErr)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK repository list: %v", sdkErr),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]any{
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
func (h *SDKRepositoryHandler) transformReposToResponse(repos []*gitea.Repository) []map[string]any {
	result := make([]map[string]any, len(repos))
	for i, repo := range repos {
		// Transform to MCP-compatible format
		repoData := map[string]any{
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
	client giteasdk.GiteaClientInterface
}

// NewSDKIssueListHandler creates a new SDK issue list handler
func NewSDKIssueListHandler(logger *logrus.Logger, client giteasdk.GiteaClientInterface) *SDKIssueListHandler {
	return &SDKIssueListHandler{
		logger: logger,
		client: client,
	}
}

// HandleIssueListRequest handles an issue_list request with Gitea SDK integration
func (h *SDKIssueListHandler) HandleIssueListRequest(ctx context.Context, req *mcp.CallToolRequest, args IssueListArgs) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling issue_list request with Gitea SDK")

	// Validate arguments
	if err := ValidateIssueListArgs(args); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	// Validate repository parameter
	var repoParam string
	if args.Repository != "" {
		repoParam = args.Repository
	} else if args.CWD != "" {
		// Resolve CWD to repository identifier
		var err error
		repoParam, err = giteasdk.ResolveCWDToRepository(args.CWD)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error resolving repository from CWD: %v", err),
					},
				},
			}, nil, err
		}
	}

	// Validate repository access
	if valid, err := ValidateRepositoryAccess(h.client, repoParam); !valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	// Parse repository identifier
	owner, repo, _ := strings.Cut(repoParam, "/")

	// Build SDK options from parameters
	opts := gitea.ListIssueOption{}
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

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	issues, _, err := h.client.ListRepoIssues(owner, repo, opts)
	if err != nil {
		sdkErr := giteasdk.NewSDKError("ListIssues", err, fmt.Sprintf("state=%s, limit=%d", args.State, args.Limit))
		h.logger.Errorf("%v", sdkErr)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK issue list: %v", sdkErr),
				},
			},
		}, nil, nil
	}

	// Extract repository metadata
	repoMetadata, err := giteasdk.ExtractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]any{}
	}

	// Transform to MCP response format
	result := map[string]any{
		"issues":     h.transformIssuesToResponse(issues, repoMetadata),
		"total":      len(issues),
		"repository": repoMetadata,
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
func (h *SDKIssueListHandler) transformIssuesToResponse(issues []*gitea.Issue, repoMetadata map[string]any) []map[string]any {
	result := make([]map[string]any, len(issues))
	for i, issue := range issues {
		// Transform to MCP-compatible format
		issueData := map[string]any{
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
		issueData["url"] = issue.URL

		// Add repository metadata to individual issue object
		issueData["repository"] = repoMetadata

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
