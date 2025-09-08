package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Basic validation - check if request is valid
	if ctx == nil {
		return mcp.NewToolResultError("Context is required"), nil
	}

	// Return the hello world message
	return mcp.NewToolResultText("Hello, World!"), nil
}

type IssueList struct {
	Issues []gitea.Issue `json:"issues"`
}

func (s *Server) handleListIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse arguments
	type ListArgs struct {
		Repo   string
		Limit  int
		Offset int
	}
	args := ListArgs{
		Repo:   mcp.ParseString(request, "repository", ""),
		Limit:  mcp.ParseInt(request, "limit", 15),
		Offset: mcp.ParseInt(request, "offset", 0),
	}
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repo, v.Required),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid request", err), nil
	}

	issues, err := s.giteaService.ListIssues(ctx, args.Repo, args.Limit, args.Offset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list issues: %v", err), nil
	}

	return mcp.NewToolResultStructured(IssueList{Issues: issues}, fmt.Sprintf("Found %d issues", len(issues))), nil
}
