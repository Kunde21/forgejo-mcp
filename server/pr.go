package server

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// SDKPRListHandler handles pr_list tool requests with Gitea SDK integration
type SDKPRListHandler struct {
	logger *logrus.Logger
	client giteasdk.GiteaClientInterface
}

// NewSDKPRListHandler creates a new SDK PR list handler
func NewSDKPRListHandler(logger *logrus.Logger, client giteasdk.GiteaClientInterface) *SDKPRListHandler {
	return &SDKPRListHandler{logger: logger, client: client}
}

// HandlePRListRequest handles a pr_list request with Gitea SDK integration
func (h *SDKPRListHandler) HandlePRListRequest(ctx context.Context, req *mcp.CallToolRequest, args PRListArgs) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling pr_list request with Gitea SDK")

	if err := ValidatePRListArgs(args); err != nil {
		return TextError(err), nil, nil
	}
	// Validate repository parameter
	var repoParam string
	switch {
	case args.Repository != "":
		repoParam = args.Repository
	default:
		var err error
		repoParam, err = giteasdk.ResolveCWDToRepository(args.CWD)
		if err != nil {
			return TextError(fmt.Errorf("resolving repository from CWD: %w", err)), nil, nil
		}
	}

	if h.client == nil {
		return TextResult("Error: Gitea client not configured"), nil, nil
	}

	// Validate repository access
	if valid, err := ValidateRepositoryAccess(h.client, repoParam); !valid {
		return TextError(err), nil, err
	}
	// Parse repository identifier
	owner, repo, _ := strings.Cut(repoParam, "/")

	// Build SDK options from parameters
	opts := gitea.ListPullRequestsOptions{}

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

	prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)
	if err != nil {
		sdkErr := giteasdk.NewSDKError("ListRepoPullRequests", err, fmt.Sprintf("owner=%s, repo=%s", owner, repo))
		h.logger.Errorf("%v", sdkErr)
		return TextError(fmt.Errorf("executing SDK pr list: %w", sdkErr)), nil, nil
	}

	// Extract repository metadata
	repoMetadata, err := giteasdk.ExtractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]any{}
	}

	// Transform to MCP response format
	result := map[string]any{
		"pullRequests": h.transformPRsToResponse(prs),
		"total":        len(prs),
		"repository":   repoMetadata,
	}

	return TextResultf("Found %d pull requests", len(prs)), result, nil
}

// transformPRsToResponse transforms Gitea SDK PR data to MCP response format
func (h *SDKPRListHandler) transformPRsToResponse(prs []*gitea.PullRequest) []map[string]any {
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
		return "closed"
	default:
		return "unknown"
	}
}
