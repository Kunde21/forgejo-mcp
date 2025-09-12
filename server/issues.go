package server

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// IssueList represents a collection of repository issues.
// This struct is used as the result data for the list_issues tool.
type IssueList struct {
	Issues []gitea.Issue `json:"issues,omitempty"`
}

type IssueListArgs struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// handleIssueList handles the "issue_list" tool request.
// It retrieves issues from a specified Forgejo/Gitea repository with optional pagination.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - limit: Maximum number of issues to return (1-100, default 15)
//   - offset: Number of issues to skip for pagination (default 0)
//
// Migration Note: Updated to use the official SDK's handler signature and
// result construction patterns. Error handling follows the new SDK's conventions.
func (s *Server) handleIssueList(ctx context.Context, request *mcp.CallToolRequest, args IssueListArgs) (*mcp.CallToolResult, *IssueList, error) {
	// Set default limit if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Fetch issues from the Gitea/Forgejo repository
	issues, err := s.giteaService.ListIssues(ctx, args.Repository, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list issues: %v", err), nil, nil
	}
	return TextResultf("Found %d issues", len(issues)), &IssueList{Issues: issues}, nil
}
