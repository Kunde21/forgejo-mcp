package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TextResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}

func TextResultf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}
}

func TextError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}, IsError: true}
}

func TextErrorf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}, IsError: true}
}

// handleHello handles the "hello" tool request.
// This is a simple demonstration tool that returns a hello world message.
//
// Migration Note: Updated to use the official SDK's handler signature:
// (context.Context, *mcp.CallToolRequest, args) (*mcp.CallToolResult, any, error)
// instead of the previous SDK's handler pattern.
func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}
	// Return successful response with hello message
	// Migration: Uses official SDK's CallToolResult structure
	return TextResult("Hello, World!"), nil, nil
}

// IssueList represents a collection of repository issues.
// This struct is used as the result data for the list_issues tool.
type IssueList struct {
	Issues []gitea.Issue `json:"issues"`
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
func (s *Server) handleIssueList(ctx context.Context, request *mcp.CallToolRequest, args struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}) (*mcp.CallToolResult, any, error) {
	// Set default limit if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required),
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
	return TextResultf("Found %d issues", len(issues)), IssueList{Issues: issues}, nil
}

// CommentResult represents the result data for the create_issue_comment tool.
type CommentResult struct {
	Comment gitea.IssueComment `json:"comment"`
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
func (s *Server) handleIssueCommentCreate(ctx context.Context, request *mcp.CallToolRequest, args struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Comment     string `json:"comment"`
}) (*mcp.CallToolResult, any, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required),
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

	return TextResult(responseText), CommentResult{Comment: *comment}, nil
}

// CommentListResult represents the result data for the list_issue_comments tool.
type CommentListResult struct {
	Comments []gitea.IssueComment `json:"comments"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
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
func (s *Server) handleIssueCommentList(ctx context.Context, request *mcp.CallToolRequest, args struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}) (*mcp.CallToolResult, any, error) {
	// Set default limit if not provided
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.Required),
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
	endIndex := args.Offset + len(commentList.Comments)
	if endIndex > commentList.Total {
		endIndex = commentList.Total
	}
	responseText := fmt.Sprintf("Found %d comments (showing %d-%d)",
		commentList.Total,
		args.Offset+1,
		endIndex)

	return TextResult(responseText), CommentListResult{
		Comments: commentList.Comments,
		Total:    commentList.Total,
		Limit:    commentList.Limit,
		Offset:   commentList.Offset,
	}, nil
}
