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

// IssueList represents a collection of repository issues.
// This struct is used as the result data for the list_issues tool.
type IssueList struct {
	Issues []remote.Issue `json:"issues,omitempty"`
}

type IssueListArgs struct {
	Repository string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory  string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
	Limit      int    `json:"limit,omitzero"`
	Offset     int    `json:"offset,omitzero"`
}

// handleIssueList handles the "issue_list" tool request.
// It retrieves issues from a specified Forgejo/Gitea repository with optional pagination.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - limit: Maximum number of issues to return (1-100, default 15)
//   - offset: Number of issues to skip for pagination (default 0)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
//
// Migration Note: Updated to use the official SDK's handler signature and
// result construction patterns. Error handling follows the new SDK's conventions.
// Added directory parameter support with automatic repository resolution.
func (s *Server) handleIssueList(ctx context.Context, request *mcp.CallToolRequest, args IssueListArgs) (*mcp.CallToolResult, *IssueList, error) {
	// Set default limit if not provided
	if args.Limit == 0 {
		args.Limit = 15
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
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
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

	// Fetch issues from the Gitea/Forgejo repository
	issues, err := s.remote.ListIssues(ctx, repository, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list issues: %v", err), nil, nil
	}

	return TextResultf("Found %d issues", len(issues)), &IssueList{Issues: issues}, nil
}

type IssueCreateArgs struct {
	Repository  string        `json:"repository,omitzero"`
	Directory   string        `json:"directory,omitzero"`
	Title       string        `json:"title"`
	Body        string        `json:"body,omitzero"`
	Attachments []interface{} `json:"attachments,omitzero"` // MCP Content objects
}

type IssueCreateResult struct {
	Issue *remote.Issue `json:"issue,omitempty"`
}

// handleIssueCreate handles the "issue_create" tool request
func (s *Server) handleIssueCreate(ctx context.Context, request *mcp.CallToolRequest, args IssueCreateArgs) (*mcp.CallToolResult, *IssueCreateResult, error) {
	// Validation using ozzo-validation (follow issue_comments.go pattern)
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
		v.Field(&args.Title, v.Required, v.Length(1, 255).Error("title must be between 1 and 255 characters")),
		v.Field(&args.Body, v.Length(0, 65535).Error("body must be less than 65535 characters")),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Repository resolution (follow existing pattern)
	repository := args.Repository
	if args.Directory != "" {
		resolution, err := s.repositoryResolver.ResolveRepository(args.Directory)
		if err != nil {
			return TextErrorf("Failed to resolve directory: %v", err), nil, nil
		}
		repository = resolution.Repository
	}

	// Process attachments
	var processedAttachments []remote.ProcessedAttachment
	for _, content := range args.Attachments {
		attachment, err := s.processAttachment(content)
		if err != nil {
			return TextErrorf("Invalid attachment: %v", err), nil, nil
		}
		processedAttachments = append(processedAttachments, *attachment)
	}

	// Create issue
	var issue *remote.Issue
	if len(processedAttachments) > 0 {
		// Use attachment-enabled method
		createArgs := remote.CreateIssueWithAttachmentsArgs{
			CreateIssueArgs: remote.CreateIssueArgs{
				Repository: repository,
				Title:      args.Title,
				Body:       args.Body,
			},
			Attachments: processedAttachments,
		}

		var err error
		issue, err = s.remote.CreateIssueWithAttachments(ctx, createArgs)
		if err != nil {
			return TextErrorf("Failed to create issue with attachments: %v", err), nil, nil
		}
	} else {
		// Use regular method
		createArgs := remote.CreateIssueArgs{
			Repository: repository,
			Title:      args.Title,
			Body:       args.Body,
		}

		var err error
		issue, err = s.remote.CreateIssue(ctx, createArgs)
		if err != nil {
			return TextErrorf("Failed to create issue: %v", err), nil, nil
		}
	}

	// Success response
	responseText := fmt.Sprintf("Issue created successfully. Number: %d, Title: %s", issue.Number, issue.Title)
	return TextResult(responseText), &IssueCreateResult{Issue: issue}, nil
}

func (s *Server) processAttachment(content interface{}) (*remote.ProcessedAttachment, error) {
	switch c := content.(type) {
	case *mcp.ImageContent:
		// Handle image content
		data := []byte(c.Data) // MCP SDK handles base64 decoding
		filename := generateFilename(c.MIMEType)
		return &remote.ProcessedAttachment{
			Data:     data,
			Filename: filename,
			MIMEType: c.MIMEType,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %T", content)
	}
}

func generateFilename(mimeType string) string {
	// Simple filename generation based on MIME type
	// In a real implementation, this might use more sophisticated logic
	switch mimeType {
	case "image/jpeg":
		return "attachment.jpg"
	case "image/png":
		return "attachment.png"
	case "image/gif":
		return "attachment.gif"
	default:
		return "attachment.bin"
	}
}

// IssueEditArgs represents the arguments for editing an issue with validation tags
type IssueEditArgs struct {
	Repository  string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory   string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
	IssueNumber int    `json:"issue_number" validate:"required,min=1"`
	Title       string `json:"title,omitzero"` // New title for the issue
	Body        string `json:"body,omitzero"`  // New description/body for the issue
	State       string `json:"state,omitzero"` // New state ("open" or "closed")
}

// IssueEditResult represents the result data for the issue_edit tool
type IssueEditResult struct {
	Issue *remote.Issue `json:"issue,omitempty"`
}

// handleIssueEdit handles the "issue_edit" tool request.
// It edits an existing issue in a Forgejo/Gitea repository.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - issue_number: The issue number to edit (must be positive)
//   - title: New title for the issue (optional)
//   - body: New description/body for the issue (optional)
//   - state: New state ("open" or "closed", optional)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
// At least one of title, body, or state must be provided.
//
// Returns:
//   - Success: Issue edit confirmation with updated metadata
//   - Error: Validation errors or API failures
func (s *Server) handleIssueEdit(ctx context.Context, request *mcp.CallToolRequest, args IssueEditArgs) (*mcp.CallToolResult, *IssueEditResult, error) {
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
		v.Field(&args.IssueNumber, v.Required.Error("must be no less than 1"), v.Min(1)),
		v.Field(&args.State, v.When(args.State != "",
			v.In("open", "closed").Error("state must be 'open' or 'closed'"),
		)),
		v.Field(&args.Title, v.When(args.Title != "",
			v.Length(1, 255).Error("title must be between 1 and 255 characters"),
		)),
		v.Field(&args.Body, v.When(args.Body != "",
			v.Length(1, 65535).Error("body must be between 1 and 65535 characters"),
		)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	// Ensure at least one field is being changed
	if args.Title == "" && args.Body == "" && args.State == "" {
		return TextError("At least one of title, body, or state must be provided"), nil, nil
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

	// Edit the issue using the service layer
	editArgs := remote.EditIssueArgs{
		Repository:  repository,
		IssueNumber: args.IssueNumber,
		Title:       args.Title,
		Body:        args.Body,
		State:       args.State,
	}
	issue, err := s.remote.EditIssue(ctx, editArgs)
	if err != nil {
		return TextErrorf("Failed to edit issue: %v", err), nil, nil
	}

	// Format success response with updated issue metadata
	responseText := fmt.Sprintf("Issue edited successfully. Number: %d, Title: %s, State: %s",
		issue.Number, issue.Title, issue.State)
	if issue.Updated != "" {
		responseText += fmt.Sprintf(", Updated: %s", issue.Updated)
	}
	responseText += "\n"

	if issue.Body != "" {
		responseText += fmt.Sprintf("Body: %s\n", issue.Body)
	}

	return TextResult(responseText), &IssueEditResult{Issue: issue}, nil
}
