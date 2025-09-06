package server

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"

	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
)

// SDKRepositoryHandler handles repository operations with Gitea SDK integration
type SDKRepositoryHandler struct {
	logger *logrus.Logger
	client giteasdk.GiteaClientInterface
}

// NewSDKRepositoryHandler creates a new SDK repository handler
func NewSDKRepositoryHandler(logger *logrus.Logger, client giteasdk.GiteaClientInterface) *SDKRepositoryHandler {
	return &SDKRepositoryHandler{logger: logger, client: client}
}

// ListRepositories handles a repository list request with Gitea SDK integration
func (h *SDKRepositoryHandler) ListRepositories(ctx context.Context, req *mcp.CallToolRequest, args RepoListArgs) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling repository list request with Gitea SDK")

	// Validate arguments
	if err := ValidateRepositoryListArgs(args); err != nil {
		return TextError(err), nil, err
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
		return TextError(fmt.Errorf("executing SDK repository list: %w", sdkErr)), nil, nil
	}

	// Transform to MCP response format
	result := map[string]any{
		"repositories": h.transformReposToResponse(repos),
		"total":        len(repos),
	}

	return TextResult(fmt.Sprintf("Found %d repositories", len(repos))), result, nil
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
