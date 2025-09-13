package server

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestList represents a collection of repository pull requests.
// This struct is used as the result data for the pr_list tool.
type PullRequestList struct {
	PullRequests []gitea.PullRequest `json:"pull_requests,omitempty"`
}

// PullRequestListArgs represents the arguments for listing pull requests with validation tags
type PullRequestListArgs struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	State      string `json:"state"`
}

// handlePullRequestList handles the "pr_list" tool request.
// It retrieves pull requests from a specified Forgejo/Gitea repository with optional pagination and state filtering.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - limit: Maximum number of pull requests to return (1-100, default 15)
//   - offset: Number of pull requests to skip for pagination (default 0)
//   - state: State of pull requests to filter by ("open", "closed", "all", default "open")
//
// Migration Note: Updated to use official SDK's handler signature and
// result construction patterns. Error handling follows new SDK's conventions.
func (s *Server) handlePullRequestList(ctx context.Context, request *mcp.CallToolRequest, args PullRequestListArgs) (*mcp.CallToolResult, *PullRequestList, error) {
	// Set default values if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}
	if args.State == "" {
		args.State = "open"
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
		v.Field(&args.State, v.In("open", "closed", "all").Error("state must be one of: open, closed, all")),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Create pull request options
	options := gitea.ListPullRequestsOptions{
		State:  args.State,
		Limit:  args.Limit,
		Offset: args.Offset,
	}

	// Fetch pull requests from the Gitea/Forgejo repository
	pullRequests, err := s.remote.ListPullRequests(ctx, args.Repository, options)
	if err != nil {
		return TextErrorf("Failed to list pull requests: %v", err), nil, nil
	}

	return TextResultf("Found %d pull requests", len(pullRequests)), &PullRequestList{PullRequests: pullRequests}, nil
}
