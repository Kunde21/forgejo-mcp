package server

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	// Basic validation - check if request is valid
	if ctx == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Context is required"},
			},
			IsError: true,
		}, nil, nil
	}

	// Return the hello world message
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello, World!"},
		},
	}, nil, nil
}

type IssueList struct {
	Issues []gitea.Issue `json:"issues"`
}

func (s *Server) handleListIssues(ctx context.Context, request *mcp.CallToolRequest, args struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}) (*mcp.CallToolResult, any, error) {
	// Set defaults
	if args.Limit == 0 {
		args.Limit = 15
	}

	// Validate arguments
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
