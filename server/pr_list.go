package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestList represents a collection of repository pull requests.
// This struct is used as the result data for the pr_list tool.
type PullRequestList struct {
	PullRequests []remote.PullRequest `json:"pull_requests,omitempty"`
}

// PullRequestListArgs represents the arguments for listing pull requests with validation tags
type PullRequestListArgs struct {
	Repository string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory  string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
	Limit      int    `json:"limit,omitzero"`
	Offset     int    `json:"offset,omitzero"`
	State      string `json:"state"`
}

// handlePullRequestList handles the "pr_list" tool request.
// It retrieves pull requests from a specified Forgejo/Gitea repository with optional pagination and state filtering.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - limit: Maximum number of pull requests to return (1-100, default 15)
//   - offset: Number of pull requests to skip for pagination (default 0)
//   - state: State of pull requests to filter by ("open", "closed", "all", default "open")
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
//
// Migration Note: Updated to use official SDK's handler signature and
// result construction patterns. Error handling follows new SDK's conventions.
// Added directory parameter support with automatic repository resolution.
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
		v.Field(&args.Repository, v.When(args.Directory == "",
			v.Required.Error("at least one of directory or repository must be provided"),
			v.Match(repoReg).Error("repository must be in format 'owner/repo'"),
		)),
		v.Field(&args.Directory, v.When(args.Repository == "",
			v.Required.Error("at least one of directory or repository must be provided"),
			v.By(func(any) error {
				if !filepath.IsAbs(args.Directory) {
					return v.NewError("abs_dir", "directory must be an absolute path")
				}
				stat, err := os.Stat(args.Directory)
				if err != nil {
					return v.NewError("abs_dir", "invalid directory")
				}
				if !stat.IsDir() {
					return v.NewError("abs_dir", "does not exist")
				}
				return nil
			}),
		)),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
		v.Field(&args.State, v.In("open", "closed", "all").Error("state must be one of: open, closed, all")),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	repository := args.Repository
	if args.Directory != "" {
		// Resolve directory to repository (takes precedence if both provided)
		resolution, err := s.repositoryResolver.ResolveRepository(args.Directory)
		if err != nil {
			return TextErrorf("Failed to resolve directory: %v", err), nil, nil
		}
		repository = resolution.Repository
	}

	// Create pull request options
	options := remote.ListPullRequestsOptions{
		State:  args.State,
		Limit:  args.Limit,
		Offset: args.Offset,
	}

	// Fetch pull requests from the Gitea/Forgejo repository
	pullRequests, err := s.remote.ListPullRequests(ctx, repository, options)
	if err != nil {
		return TextErrorf("Failed to list pull requests: %v", err), nil, nil
	}

	// Build detailed text result for backwards compatibility
	var resultText string
	if len(pullRequests) == 0 {
		resultText = "No pull requests found"
	} else {
		resultText = fmt.Sprintf("Found %d pull requests:\n", len(pullRequests))
		for _, pr := range pullRequests {
			resultText += fmt.Sprintf("- #%d: %s (%s)\n", pr.Number, pr.Title, pr.State)
		}
	}

	return TextResult(resultText), &PullRequestList{PullRequests: pullRequests}, nil
}
