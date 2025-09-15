package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestCommentList represents a collection of pull request comments.
// This struct is used as the result data for the pr_comment_list tool.
type PullRequestCommentList struct {
	PullRequestComments []remote.Comment `json:"pull_request_comments,omitempty"`
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
		v.Field(&args.PullRequestNumber, v.Required.Error("must be no less than 1"), v.Min(1)),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Fetch pull request comments from the Gitea/Forgejo repository
	commentList, err := s.remote.ListPullRequestComments(ctx, args.Repository, args.PullRequestNumber, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list pull request comments: %v", err), nil, nil
	}

	responseText := "Found 0 comments"
	if len(commentList.Comments) != 0 {
		endIndex := min(args.Offset+len(commentList.Comments), commentList.Total)
		responseText = fmt.Sprintf("Found %d comments (showing %d-%d):\n",
			commentList.Total,
			args.Offset+1,
			endIndex)
		for i, comment := range commentList.Comments {
			responseText += fmt.Sprintf("Comment %d (ID: %d): %s\n", i+1, comment.ID, comment.Content)
		}
	}

	return TextResult(responseText), &PullRequestCommentList{PullRequestComments: commentList.Comments}, nil
}

// PullRequestCommentCreateArgs represents the arguments for creating a pull request comment with validation tags
type PullRequestCommentCreateArgs struct {
	Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	Comment           string `json:"comment" validate:"required,min=1"`
}

// PullRequestCommentCreateResult represents the result data for the pr_comment_create tool
type PullRequestCommentCreateResult struct {
	Comment *remote.Comment `json:"comment,omitempty"`
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
		v.Field(&args.PullRequestNumber, v.Required.Error("must be no less than 1"), v.Min(1)),
		v.Field(&args.Comment, v.Required, v.Match(emptyReg).Error("cannot be blank")),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Create the comment using the service layer
	comment, err := s.remote.CreatePullRequestComment(ctx, args.Repository, args.PullRequestNumber, args.Comment)
	if err != nil {
		return TextErrorf("Failed to create pull request comment: %v", err), nil, nil
	}

	// Format success response with comment metadata
	responseText := fmt.Sprintf("Pull request comment created successfully. ID: %d, Created: %s\nComment body: %s",
		comment.ID, comment.Created, comment.Content)

	return TextResult(responseText), &PullRequestCommentCreateResult{Comment: comment}, nil
}

// PullRequestCommentEditArgs represents the arguments for editing a pull request comment with validation tags
type PullRequestCommentEditArgs struct {
	Repository        string `json:"repository" validate:"required,regexp=^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$"`
	PullRequestNumber int    `json:"pull_request_number" validate:"required,min=1"`
	CommentID         int    `json:"comment_id" validate:"required,min=1"`
	NewContent        string `json:"new_content" validate:"required,min=1"`
}

// PullRequestCommentEditResult represents the result data for the pr_comment_edit tool
type PullRequestCommentEditResult struct {
	Comment *remote.Comment `json:"comment,omitempty"`
}

// handlePullRequestCommentEdit handles the "pr_comment_edit" tool request.
// It edits an existing comment on a specified Forgejo/Gitea pull request.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - pull_request_number: The pull request number containing the comment (must be positive)
//   - comment_id: The ID of the comment to edit (must be positive)
//   - new_content: The updated comment content (cannot be empty)
//
// Returns:
//   - Success: Comment edit confirmation with updated metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handlePullRequestCommentEdit(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCommentEditArgs) (*mcp.CallToolResult, *PullRequestCommentEditResult, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}

	// Explicit validation for integer fields (ozzo-validation v.Min doesn't work well with zero values)
	if args.PullRequestNumber < 1 {
		return TextError("Invalid request: pull_request_number: must be no less than 1."), nil, nil
	}
	if args.CommentID < 1 {
		return TextError("Invalid request: comment_id: must be no less than 1."), nil, nil
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.PullRequestNumber, v.Min(1)),
		v.Field(&args.CommentID, v.Min(1)),
		v.Field(&args.NewContent, v.Required, v.Match(emptyReg).Error("cannot be blank")),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Edit the comment using the service layer
	editArgs := remote.EditPullRequestCommentArgs{
		Repository:        args.Repository,
		PullRequestNumber: args.PullRequestNumber,
		CommentID:         args.CommentID,
		NewContent:        args.NewContent,
	}
	comment, err := s.remote.EditPullRequestComment(ctx, editArgs)
	if err != nil {
		return TextErrorf("Failed to edit pull request comment: %v", err), nil, nil
	}

	// Format success response with updated comment metadata
	responseText := fmt.Sprintf("Pull request comment edited successfully. ID: %d, Updated: %s\nComment body: %s",
		comment.ID, comment.Updated, comment.Content)

	return TextResult(responseText), &PullRequestCommentEditResult{Comment: comment}, nil
}
