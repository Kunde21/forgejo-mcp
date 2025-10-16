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

// PullRequestEditArgs represents the arguments for editing a pull request with validation tags
type PullRequestEditArgs struct {
	Repository        string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory         string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	Title             string `json:"title,omitzero"`       // New title for the pull request
	Body              string `json:"body,omitzero"`        // New description/body for the pull request
	State             string `json:"state,omitzero"`       // New state ("open" or "closed")
	BaseBranch        string `json:"base_branch,omitzero"` // New base branch for the pull request
}

// PullRequestEditResult represents the result data for the pr_edit tool
type PullRequestEditResult struct {
	PullRequest *remote.PullRequest `json:"pull_request,omitempty"`
}

// handlePullRequestEdit handles the "pr_edit" tool request.
// It edits an existing pull request in a Forgejo/Gitea repository.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - pull_request_number: The pull request number to edit (must be positive)
//   - title: New title for the pull request (optional)
//   - body: New description/body for the pull request (optional)
//   - state: New state ("open" or "closed", optional)
//   - base_branch: New base branch for the pull request (optional)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
// At least one of title, body, state, or base_branch must be provided.
//
// Returns:
//   - Success: Pull request edit confirmation with updated metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
// Added directory parameter support with automatic repository resolution.
func (s *Server) handlePullRequestEdit(ctx context.Context, request *mcp.CallToolRequest, args PullRequestEditArgs) (*mcp.CallToolResult, *PullRequestEditResult, error) {
	// Validate context - required for proper request handling
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
		v.Field(&args.PullRequestNumber, v.Required.Error("must be no less than 1"), v.Min(1)),
		v.Field(&args.State, v.When(args.State != "",
			v.In("open", "closed").Error("state must be 'open' or 'closed'"),
		)),
		v.Field(&args.Title, v.When(args.Title != "",
			v.Length(1, 255).Error("title must be between 1 and 255 characters"),
		)),
		v.Field(&args.Body, v.When(args.Body != "",
			v.Length(1, 65535).Error("body must be between 1 and 65535 characters"),
		)),
		v.Field(&args.BaseBranch, v.When(args.BaseBranch != "",
			v.Length(1, 255).Error("base branch must be between 1 and 255 characters"),
		)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Ensure at least one field is being changed
	if args.Title == "" && args.Body == "" && args.State == "" && args.BaseBranch == "" {
		return TextError("At least one of title, body, state, or base_branch must be provided"), nil, nil
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

	// Edit the pull request using the service layer
	editArgs := remote.EditPullRequestArgs{
		Repository:        repository,
		PullRequestNumber: args.PullRequestNumber,
		Title:             args.Title,
		Body:              args.Body,
		State:             args.State,
		BaseBranch:        args.BaseBranch,
	}
	pr, err := s.remote.EditPullRequest(ctx, editArgs)
	if err != nil {
		return TextErrorf("Failed to edit pull request: %v", err), nil, nil
	}

	var responseText string
	if s.compatMode {
		responseText = fmt.Sprintf("Pull request edited successfully. Number: %d, Title: %s, State: %s",
			pr.Number, pr.Title, pr.State)
		if pr.UpdatedAt != "" {
			responseText += fmt.Sprintf(", Updated: %s", pr.UpdatedAt)
		}
		responseText += "\n"

		if pr.Body != "" {
			responseText += fmt.Sprintf("Body: %s\n", pr.Body)
		}
	} else {
		responseText = FormatPullRequestEditSuccess(pr)
	}

	return TextResult(responseText), &PullRequestEditResult{PullRequest: pr}, nil
}
