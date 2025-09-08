package server

import (
	"context"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	if err := validation.ValidateStruct(&args,
		validation.Field(&args.Repo, validation.Required),
		validation.Field(&args.Limit, validation.Min(1), validation.Max(100)),
		validation.Field(&args.Offset, validation.Min(0)),
	); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid request", err), nil
	}

	// Call the Gitea service
	issues, err := s.giteaService.ListIssues(ctx, args.Repo, args.Limit, args.Offset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list issues: %v", err), nil
	}

	// Format response
	type issueList struct {
		Issues []gitea.Issue `json:"issues"`
	}
	return mcp.NewToolResultStructured(issueList{Issues: issues}, fmt.Sprintf("Found %d issues", len(issues))), nil
}
