package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestCommentList represents a collection of pull request comments.
// This struct is used as the result data for the pr_comment_list tool.
type PullRequestCommentList struct {
	PullRequestComments []gitea.PullRequestComment `json:"pull_request_comments"`
}

// PullRequestCommentListArgs represents the arguments for listing pull request comments with validation tags
type PullRequestCommentListArgs struct {
	Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	Limit             int    `json:"limit" validate:"min=1,max=100"`
	Offset            int    `json:"offset" validate:"min=0"`
}

// handlePullRequestCommentList handles the "pr_comment_list" tool request.
// It retrieves comments from a specified Forgejo/Gitea pull request with pagination support.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - pull_request_number: The pull request number to get comments from (must be positive)
//   - limit: Maximum number of comments to return (1-100, default 15)
//   - offset: Number of comments to skip for pagination (default 0)
//
// Returns:
//   - Success: List of pull request comments with pagination metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handlePullRequestCommentList(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCommentListArgs) (*mcp.CallToolResult, *PullRequestCommentList, error) {
	// Set default values if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.PullRequestNumber, v.Min(1)),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Fetch pull request comments from the Gitea/Forgejo repository
	commentList, err := s.giteaService.ListPullRequestComments(ctx, args.Repository, args.PullRequestNumber, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list pull request comments: %v", err), nil, nil
	}

	return TextResultf("Found %d pull request comments", len(commentList.Comments)), &PullRequestCommentList{PullRequestComments: commentList.Comments}, nil
}

// PullRequestCommentCreateArgs represents the arguments for creating a pull request comment with validation tags
type PullRequestCommentCreateArgs struct {
	Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	Comment           string `json:"comment" validate:"required,min=1"`
}

// PullRequestCommentCreateResult represents the result data for the pr_comment_create tool
type PullRequestCommentCreateResult struct {
	Comment gitea.PullRequestComment `json:"comment"`
}

// handlePullRequestCommentCreate handles the "pr_comment_create" tool request.
// It creates a new comment on a specified Forgejo/Gitea pull request.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - pull_request_number: The pull request number to comment on (must be positive)
//   - comment: The comment content (cannot be empty)
//
// Returns:
//   - Success: Comment creation confirmation with metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handlePullRequestCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCommentCreateArgs) (*mcp.CallToolResult, *PullRequestCommentCreateResult, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.PullRequestNumber, v.Min(1)),
		v.Field(&args.Comment, v.Required, v.Length(1, 0)), // Non-empty string
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Create the comment using the service layer
	comment, err := s.giteaService.CreatePullRequestComment(ctx, args.Repository, args.PullRequestNumber, args.Comment)
	if err != nil {
		return TextErrorf("Failed to create pull request comment: %v", err), nil, nil
	}

	// Format success response with comment metadata
	responseText := fmt.Sprintf("Pull request comment created successfully. ID: %d, Created: %s\nComment body: %s",
		comment.ID, comment.CreatedAt, comment.Body)

	return TextResult(responseText), &PullRequestCommentCreateResult{Comment: *comment}, nil
}
