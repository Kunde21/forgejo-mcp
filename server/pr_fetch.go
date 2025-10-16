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

// PullRequestFetchArgs represents the arguments for fetching a pull request
type PullRequestFetchArgs struct {
	Repository        string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory         string `json:"directory,omitzero"`  // Local directory path for automatic resolution
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
}

// PullRequestFetchResult represents the result data for the pr_fetch tool
type PullRequestFetchResult struct {
	PullRequest *remote.PullRequestDetails `json:"pull_request,omitempty"`
}

// handlePullRequestFetch handles the "pr_fetch" tool request.
// It retrieves detailed information about a single pull request from a Forgejo/Gitea repository.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - pull_request_number: The pull request number to fetch (must be positive)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
//
// Returns:
//   - Success: Detailed pull request information including metadata, labels, assignees, etc.
//   - Error: Validation errors or API failures
func (s *Server) handlePullRequestFetch(ctx context.Context, request *mcp.CallToolRequest, args PullRequestFetchArgs) (*mcp.CallToolResult, *PullRequestFetchResult, error) {
	// Validate context
	if ctx == nil {
		return TextError("Context is required"), nil, nil
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
		v.Field(&args.PullRequestNumber, v.Required.Error("pull request number is required"), v.Min(1)),
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

	// Fetch the pull request
	pr, err := s.remote.GetPullRequest(ctx, repository, args.PullRequestNumber)
	if err != nil {
		return TextErrorf("Failed to fetch pull request: %v", err), nil, nil
	}

	var responseText string
	if s.compatMode {
		responseText = FormatPullRequestDetails(pr)
	} else {
		responseText = fmt.Sprintf("Pull request #%d: %s", pr.Number, pr.Title)
	}

	return TextResult(responseText), &PullRequestFetchResult{PullRequest: pr}, nil
}
