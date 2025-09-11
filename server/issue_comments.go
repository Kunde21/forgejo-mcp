package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CommentResult represents the result data for the create_issue_comment tool.
type CommentResult struct {
	Comment gitea.IssueComment `json:"comment"`
}

type IssueCommentArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Comment     string `json:"comment"`
}

// handleIssueCommentCreate handles the "issue_comment_create" tool request.
// It creates a new comment on a specified Forgejo/Gitea issue.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - issue_number: The issue number to comment on (must be positive)
//   - comment: The comment content (cannot be empty)
//
// Returns:
//   - Success: Comment creation confirmation with metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handleIssueCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args IssueCommentArgs) (*mcp.CallToolResult, *CommentResult, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.IssueNumber, v.Min(1)),
		v.Field(&args.Comment, v.Required, v.Length(1, 0)), // Non-empty string
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Create the comment using the service layer
	comment, err := s.giteaService.CreateIssueComment(ctx, args.Repository, args.IssueNumber, args.Comment)
	if err != nil {
		return TextErrorf("Failed to create comment: %v", err), nil, nil
	}

	// Format success response with comment metadata
	responseText := fmt.Sprintf("Comment created successfully. ID: %d, Created: %s\nComment body: %s",
		comment.ID, comment.Created, comment.Content)

	return TextResult(responseText), &CommentResult{Comment: *comment}, nil
}

// CommentListResult represents the result data for the list_issue_comments tool.
type CommentListResult struct {
	Comments []gitea.IssueComment `json:"comments"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

type IssueCommentListArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// handleIssueCommentList handles the "issue_comment_list" tool request.
// It retrieves comments from a specified Forgejo/Gitea issue with optional pagination.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - issue_number: The issue number to list comments from (must be positive)
//   - limit: Maximum number of comments to return (1-100, default 15)
//   - offset: Number of comments to skip for pagination (default 0)
//
// Returns:
//   - Success: Comment list with pagination metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
func (s *Server) handleIssueCommentList(ctx context.Context, request *mcp.CallToolRequest, args IssueCommentListArgs) (*mcp.CallToolResult, *CommentListResult, error) {
	// Set default limit if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required, v.Match(repoReg).Error("repository must be in format 'owner/repo'")),
		v.Field(&args.IssueNumber, v.Min(1)),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Fetch comments from the Gitea/Forgejo repository
	commentList, err := s.giteaService.ListIssueComments(ctx, args.Repository, args.IssueNumber, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list issue comments: %v", err), nil, nil
	}

	// Format success response with comment count and pagination info
	responseText := fmt.Sprintf("Found %d comments (showing %d-%d)",
		commentList.Total,
		args.Offset+1,
		min(args.Offset+len(commentList.Comments), commentList.Total))

	return TextResult(responseText), &CommentListResult{
		Comments: commentList.Comments,
		Total:    commentList.Total,
		Limit:    commentList.Limit,
		Offset:   commentList.Offset,
	}, nil
}
