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
		return TextError(err), nil, err
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
			return TextError(fmt.Errorf("resolving repository from CWD: %w", err)), nil, err
		}
	}

	// Validate repository access
	if valid, err := ValidateRepositoryAccess(h.client, repoParam); !valid {
		return TextError(err), nil, err
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
		return TextError(fmt.Errorf("executing SDK issue list: %w", sdkErr)), nil, nil
	}

	// Extract repository metadata
	repoMetadata, err := giteasdk.ExtractRepositoryMetadata(h.client, repoParam)
	if err != nil {
		h.logger.Warnf("Failed to extract repository metadata: %v", err)
		repoMetadata = map[string]any{}
	}

	// Transform to MCP response format
	result := map[string]any{
		"issues":     h.transformIssuesToResponse(issues),
		"total":      len(issues),
		"repository": repoMetadata,
	}

	return TextResult(fmt.Sprintf("Found %d issues", len(issues))), result, nil
}

// transformIssuesToResponse transforms Gitea SDK issue data to MCP response format
func (h *SDKIssueListHandler) transformIssuesToResponse(issues []*gitea.Issue) []map[string]any {
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
