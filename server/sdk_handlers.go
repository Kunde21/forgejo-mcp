package server

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// SDKError represents an error from the Gitea SDK with additional context
type SDKError struct {
	Operation string // The operation that failed (e.g., "ListRepoPullRequests")
	Cause     error  // The original SDK error
	Context   string // Additional context about the operation
}

func (e *SDKError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("Gitea SDK %s failed (%s): %v", e.Operation, e.Context, e.Cause)
	}
	return fmt.Sprintf("Gitea SDK %s failed: %v", e.Operation, e.Cause)
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

// NewSDKError creates a new SDK error with context
func NewSDKError(operation string, cause error, context ...string) *SDKError {
	ctx := ""
	if len(context) > 0 {
		ctx = strings.Join(context, ", ")
	}
	return &SDKError{
		Operation: operation,
		Cause:     cause,
		Context:   ctx,
	}
}

// GiteaClientInterface defines the interface for Gitea client operations
type GiteaClientInterface interface {
	// Repository operations
	ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error)
	GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error)
	CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	DeleteRepo(owner, repo string) (*gitea.Response, error)

	// Pull Request operations
	ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
	GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error)
	CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
	EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error)

	// Issue operations
	ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error)
	CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error)
	EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error)

	// User operations
	GetMyUserInfo() (*gitea.User, *gitea.Response, error)
	GetUserInfo(user string) (*gitea.User, *gitea.Response, error)

	// Branch operations
	ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error)
	GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error)

	// Commit operations
	GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error)
	ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error)
}

// ValidateRepositoryFormat validates that a repository parameter follows the owner/repo format
func ValidateRepositoryFormat(repoParam string) (bool, error) {
	if repoParam == "" {
		return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
	}

	owner, repo, ok := strings.Cut(repoParam, "/")
	if !ok {
		return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
	}
	if owner == "" {
		return false, fmt.Errorf("invalid repository format: owner cannot be empty")
	}
	if repo == "" {
		return false, fmt.Errorf("invalid repository format: repository name cannot be empty")
	}

	// Basic validation - allow common special characters
	// Let the API handle more complex validation and security
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.<>@"
	for _, char := range owner + repo {
		if !strings.ContainsRune(validChars, char) {
			return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
		}
	}
	return true, nil
}

// validateRepositoryExistence checks if a repository exists via Gitea API
func validateRepositoryExistence(client GiteaClientInterface, repoParam string) (bool, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return false, err
	}
	owner, repo, _ := strings.Cut(repoParam, "/")
	if _, _, err = client.GetRepo(owner, repo); err != nil {
		return false, fmt.Errorf("failed to validate repository existence: %w", err)
	}
	return true, nil
}

// validateRepositoryAccess checks if the user has access to the repository
func validateRepositoryAccess(client GiteaClientInterface, repoParam string) (bool, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return false, err
	}
	owner, repo, _ := strings.Cut(repoParam, "/")
	_, _, err = client.GetRepo(owner, repo)
	if err != nil {
		return false, fmt.Errorf("failed to validate repository access: %w", err)
	}
	return true, nil
}

// resolveCWDToRepository attempts to resolve a CWD path to a repository identifier
// This is a basic implementation that looks for common patterns
func resolveCWDToRepository(cwd string) (string, error) {
	if cwd == "" {
		return "", fmt.Errorf("CWD is empty")
	}

	// Remove any trailing slashes
	cwd = strings.TrimSuffix(cwd, "/")

	// Look for patterns like /owner/repo or owner/repo in the path
	parts := strings.Split(cwd, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("CWD does not contain a valid repository path")
	}

	// Find the last two parts that could be owner/repo
	for i := len(parts) - 1; i >= 1; i-- {
		potentialOwner := parts[i-1]
		potentialRepo := parts[i]

		// Skip empty parts
		if potentialOwner == "" || potentialRepo == "" {
			continue
		}

		// Basic validation - should not contain spaces or special chars that aren't allowed in repo names
		if strings.ContainsAny(potentialOwner, " \t\n\r") || strings.ContainsAny(potentialRepo, " \t\n\r") {
			continue
		}

		repoParam := potentialOwner + "/" + potentialRepo
		if valid, _ := ValidateRepositoryFormat(repoParam); valid {
			return repoParam, nil
		}
	}

	return "", fmt.Errorf("could not resolve repository from CWD: %s", cwd)
}

// extractRepositoryMetadata extracts and caches repository metadata
func extractRepositoryMetadata(client GiteaClientInterface, repoParam string) (map[string]interface{}, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return nil, err
	}

	owner, repo, _ := strings.Cut(repoParam, "/")
	giteaRepo, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repository metadata: %w", err)
	}

	metadata := map[string]interface{}{
		"id":          giteaRepo.ID,
		"name":        giteaRepo.Name,
		"fullName":    giteaRepo.FullName,
		"description": giteaRepo.Description,
		"private":     giteaRepo.Private,
		"fork":        giteaRepo.Fork,
		"archived":    giteaRepo.Archived,
		"stars":       giteaRepo.Stars,
		"forks":       giteaRepo.Forks,
		"size":        giteaRepo.Size,
		"url":         giteaRepo.HTMLURL,
		"sshUrl":      giteaRepo.SSHURL,
		"cloneUrl":    giteaRepo.CloneURL,
	}

	if giteaRepo.Owner != nil {
		metadata["owner"] = map[string]interface{}{
			"id":       giteaRepo.Owner.ID,
			"username": giteaRepo.Owner.UserName,
			"fullName": giteaRepo.Owner.FullName,
			"email":    giteaRepo.Owner.Email,
		}
	}

	return metadata, nil
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
		repoParam, err = resolveCWDToRepository(args.CWD)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error resolving repository from CWD: %v", err),
					},
				},
			}, nil, err
		}
	} else {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: repository parameter or cwd parameter is required",
				},
			},
		}, nil, fmt.Errorf("repository parameter or cwd parameter is required")
	}

	// Validate repository format and access
	if valid, err := ValidateRepositoryFormat(repoParam); !valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	if valid, err := validateRepositoryAccess(h.client, repoParam); !valid {
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

	// Note: Gitea SDK doesn't have direct author filtering in ListPullRequestsOptions
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)
	if err != nil {
		sdkErr := NewSDKError("ListRepoPullRequests", err, fmt.Sprintf("owner=%s, repo=%s", owner, repo))
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
	repoMetadata, err := extractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]interface{}{}
	}

	// Transform to MCP response format
	result := map[string]interface{}{
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
func (h *SDKPRListHandler) transformPRsToResponse(prs []*gitea.PullRequest, repoMetadata map[string]interface{}) []map[string]interface{} {
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
		sdkErr := NewSDKError("ListMyRepos", err, fmt.Sprintf("limit=%d", args.Limit))
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
	Repository string   `json:"repository,omitempty"`
	CWD        string   `json:"cwd,omitempty"`
	State      string   `json:"state,omitempty"`
	Author     string   `json:"author,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling issue_list request with Gitea SDK")

	// Validate repository parameter
	var repoParam string
	if args.Repository != "" {
		repoParam = args.Repository
	} else if args.CWD != "" {
		// Resolve CWD to repository identifier
		var err error
		repoParam, err = resolveCWDToRepository(args.CWD)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error resolving repository from CWD: %v", err),
					},
				},
			}, nil, err
		}
	} else {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: repository parameter or cwd parameter is required",
				},
			},
		}, nil, fmt.Errorf("repository parameter or cwd parameter is required")
	}

	// Validate repository format and access
	if valid, err := ValidateRepositoryFormat(repoParam); !valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil, err
	}

	if valid, err := validateRepositoryAccess(h.client, repoParam); !valid {
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

	// Note: Gitea SDK doesn't have direct author/label filtering in ListIssueOption
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	issues, _, err := h.client.ListRepoIssues(owner, repo, opts)
	if err != nil {
		sdkErr := NewSDKError("ListIssues", err, fmt.Sprintf("state=%s, limit=%d", args.State, args.Limit))
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
	repoMetadata, err := extractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]interface{}{}
	}

	// Transform to MCP response format
	result := map[string]interface{}{
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
func (h *SDKIssueListHandler) transformIssuesToResponse(issues []*gitea.Issue, repoMetadata map[string]interface{}) []map[string]interface{} {
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
