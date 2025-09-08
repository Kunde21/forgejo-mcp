package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// handleHello handles the "hello" tool request.
// This is a simple demonstration tool that returns a hello world message.
//
// Migration Note: Updated to use the official SDK's handler signature:
// (context.Context, *mcp.CallToolRequest, args) (*mcp.CallToolResult, any, error)
// instead of the previous SDK's handler pattern.
func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Context is required"},
			},
			IsError: true,
		}, nil, nil
	}

	// Return successful response with hello message
	// Migration: Uses official SDK's CallToolResult structure
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello, World!"},
		},
	}, nil, nil
}

// IssueList represents a collection of repository issues.
// This struct is used as the result data for the list_issues tool.
type IssueList struct {
	Issues []gitea.Issue `json:"issues"`
}

// handleListIssues handles the "list_issues" tool request.
// It retrieves issues from a specified Forgejo/Gitea repository with optional pagination.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - limit: Maximum number of issues to return (1-100, default 15)
//   - offset: Number of issues to skip for pagination (default 0)
//
// Migration Note: Updated to use the official SDK's handler signature and
// result construction patterns. Error handling follows the new SDK's conventions.
func (s *Server) handleListIssues(ctx context.Context, request *mcp.CallToolRequest, args struct {
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid request: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	// Fetch issues from the Gitea/Forgejo repository
	issues, err := s.giteaService.ListIssues(ctx, args.Repository, args.Limit, args.Offset)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list issues: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Found %d issues", len(issues))},
		},
	}, IssueList{Issues: issues}, nil
}
