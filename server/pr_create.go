package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestCreateArgs represents the arguments for creating a pull request with validation tags
type PullRequestCreateArgs struct {
	Repository string `json:"repository,omitzero"`       // Repository path in "owner/repo" format
	Directory  string `json:"directory,omitzero"`        // Local directory path containing a git repository for automatic resolution
	Head       string `json:"head,omitzero"`             // Source branch (auto-detected if not provided)
	Base       string `json:"base,omitzero"`             // Target branch (default if not provided)
	Title      string `json:"title" validate:"required"` // PR title
	Body       string `json:"body,omitzero"`             // PR description
	Draft      bool   `json:"draft,omitzero"`            // Create as draft PR
	Assignee   string `json:"assignee,omitzero"`         // Single reviewer
}

// PullRequestCreateResult represents the result data for the pr_create tool
type PullRequestCreateResult struct {
	PullRequest *remote.PullRequest `json:"pull_request,omitempty"`
}

// handlePullRequestCreate handles the "pr_create" tool request.
// It creates a new pull request in a Forgejo/Gitea repository.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - head: Source branch (auto-detected if not provided)
//   - base: Target branch (default if not provided)
//   - title: PR title (required)
//   - body: PR description (optional)
//   - draft: Create as draft PR (optional)
//   - assignee: Single reviewer (optional)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
//
// Returns:
//   - Success: Pull request creation confirmation with metadata
//   - Error: Validation errors or API failures
//
// Migration Note: Implements MCP SDK v0.4.0 handler signature with ozzo-validation
// for parameter validation and structured error responses.
// Added directory parameter support with automatic repository resolution.
func (s *Server) handlePullRequestCreate(ctx context.Context, request *mcp.CallToolRequest, args PullRequestCreateArgs) (*mcp.CallToolResult, *PullRequestCreateResult, error) {
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
		v.Field(&args.Title, v.Required.Error("title is required"), v.Length(1, 255).Error("title must be between 1 and 255 characters")),
		v.Field(&args.Body, v.When(args.Body != "",
			v.Length(1, 65535).Error("body must be between 1 and 65535 characters"),
		)),
		v.Field(&args.Head, v.When(args.Head != "",
			v.Length(1, 255).Error("head branch must be between 1 and 255 characters"),
		)),
		v.Field(&args.Base, v.When(args.Base != "",
			v.Length(1, 255).Error("base branch must be between 1 and 255 characters"),
		)),
		v.Field(&args.Assignee, v.When(args.Assignee != "",
			v.Length(1, 255).Error("assignee must be between 1 and 255 characters"),
		)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	repository := args.Repository
	var forkInfo *ForkInfo
	if args.Directory != "" {
		// Resolve directory to repository with fork detection (takes precedence if both provided)
		resolution, detectedForkInfo, err := s.repositoryResolver.ResolveWithForkInfo(args.Directory)
		if err != nil {
			return enhanceRepositoryResolutionError(err, args.Directory), nil, nil
		}
		repository = resolution.Repository
		forkInfo = detectedForkInfo

		// If this is a fork, adjust the repository to target the original repo
		if forkInfo.IsFork {
			repository = fmt.Sprintf("%s/%s", forkInfo.OriginalOwner, strings.Split(repository, "/")[1])
		}
	}

	// Auto-detect current branch if not provided
	head := args.Head
	if head == "" && args.Directory != "" {
		detectedHead, err := GetCurrentBranch(args.Directory)
		if err != nil {
			return TextErrorf("Failed to detect current branch in '%s': %v. Please ensure you're in a git repository and on a valid branch.", args.Directory, err), nil, nil
		}
		head = detectedHead
	}

	// Use default base branch if not provided
	base := args.Base
	if base == "" {
		base = "main" // Default to main branch
	}

	// Validate that head branch exists if directory is provided
	if args.Directory != "" && head != "" {
		exists, err := BranchExists(args.Directory, head)
		if err != nil {
			return enhanceBranchValidationError(err, args.Directory, head), nil, nil
		}
		if !exists {
			return TextErrorf("Source branch '%s' does not exist in directory '%s'. Available branches:\n• Use 'git branch' to list local branches\n• Use 'git checkout -b %s' to create it\n• Use 'git fetch origin' to update remote branches", head, args.Directory, head), nil, nil
		}
	}

	// Check for conflicts if directory is provided
	if args.Directory != "" && head != "" && base != "" {
		conflictReport, err := GetConflictReport(args.Directory, base, head)
		if err != nil {
			return enhanceConflictErrorMessage(err, base, head), nil, nil
		}
		if conflictReport.HasConflicts {
			conflictDetails := fmt.Sprintf("\n\nConflict Analysis:\n")
			conflictDetails += fmt.Sprintf("- Total conflicts: %d\n", conflictReport.TotalConflicts)
			conflictDetails += fmt.Sprintf("- Files affected: %d\n", len(conflictReport.ConflictFiles))

			if len(conflictReport.ConflictFiles) > 0 {
				conflictDetails += "- Conflicting files:\n"
				for _, file := range conflictReport.ConflictFiles {
					conflictDetails += fmt.Sprintf("  • %s\n", file)
				}
			}

			if len(conflictReport.SuggestedActions) > 0 {
				conflictDetails += "- Suggested actions:\n"
				for _, action := range conflictReport.SuggestedActions {
					conflictDetails += fmt.Sprintf("  • %s\n", action)
				}
			}

			return TextErrorf("Branch '%s' has %d conflict(s) with '%s'.%s",
				head, conflictReport.TotalConflicts, base, conflictDetails), nil, nil
		}
	}

	// Check if head branch is behind base branch
	if args.Directory != "" && head != "" && base != "" {
		isBehind, err := IsBranchBehind(args.Directory, base, head)
		if err != nil {
			return TextErrorf("Failed to check branch status: %v", err), nil, nil
		}
		if isBehind {
			return TextErrorf("Source branch '%s' is behind target branch '%s'. Please sync your branch.", head, base), nil, nil
		}
	}

	// Load PR template if no body is provided
	body := args.Body
	if body == "" && args.Directory != "" {
		// Try to load template from the repository
		if fileContentFetcher, ok := s.remote.(remote.FileContentFetcher); ok {
			owner, repoName, ok := strings.Cut(repository, "/")
			if ok {
				template, err := LoadPRTemplate(ctx, fileContentFetcher, owner, repoName, base)
				if err == nil && template != "" {
					body = template
				}
			}
		}
	}

	// Merge template with user-provided content if both exist
	if args.Body != "" && body != "" {
		body = MergeTemplateContent(body, args.Body)
	}

	// Create the pull request using the service layer
	createArgs := remote.CreatePullRequestArgs{
		Repository: repository,
		Head:       head,
		Base:       base,
		Title:      args.Title,
		Body:       body,
		Draft:      args.Draft,
		Assignee:   args.Assignee,
	}
	pr, err := s.remote.CreatePullRequest(ctx, createArgs)
	if err != nil {
		return enhancePullRequestCreationError(err, repository, head, base), nil, nil
	}

	// Format success response with pull request metadata
	responseText := fmt.Sprintf("Pull request created successfully. Number: %d, Title: %s, State: %s",
		pr.Number, pr.Title, pr.State)
	if pr.CreatedAt != "" {
		responseText += fmt.Sprintf(", Created: %s", pr.CreatedAt)
	}
	responseText += "\n"

	// Add fork information if applicable
	if forkInfo != nil && forkInfo.IsFork {
		responseText += fmt.Sprintf("Fork Information: Created from fork '%s' targeting original repository '%s'\n",
			forkInfo.ForkOwner, forkInfo.OriginalOwner)
	}

	// Add template usage information
	if body != "" && args.Body == "" {
		responseText += "Template: Used repository PR template for description\n"
	} else if body != "" && args.Body != "" {
		responseText += "Template: Merged repository template with user-provided content\n"
	}

	if pr.Body != "" {
		responseText += fmt.Sprintf("Body: %s\n", pr.Body)
	}

	if args.Draft {
		responseText += "Note: Created as draft PR with [DRAFT] prefix in title\n"
	}

	return TextResult(responseText), &PullRequestCreateResult{PullRequest: pr}, nil
}

// enhanceRepositoryResolutionError provides detailed error messages for repository resolution failures
func enhanceRepositoryResolutionError(err error, directory string) *mcp.CallToolResult {
	baseMsg := fmt.Sprintf("Failed to resolve directory '%s'", directory)

	// Check for specific error types and provide helpful guidance
	switch e := err.(type) {
	case *DirectoryNotFoundError:
		return TextErrorf("%s: Directory does not exist. Please check the path and ensure the directory exists.", baseMsg)
	case *NotGitRepositoryError:
		return TextErrorf("%s: Not a git repository (%s). Please run 'git init' in this directory or navigate to a valid git repository.", baseMsg, e.Reason)
	case *NoRemotesConfiguredError:
		return TextErrorf("%s: No git remotes configured. Please add a remote with 'git remote add origin <url>'.", baseMsg)
	case *InvalidRemoteURLError:
		return TextErrorf("%s: Invalid remote URL format '%s'. Please check your remote configuration with 'git remote -v'.", baseMsg, e.URL)
	default:
		// Check for common patterns in error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "permission denied") {
			return TextErrorf("%s: Permission denied. Please check directory permissions.", baseMsg)
		}
		if strings.Contains(errMsg, "no such file") {
			return TextErrorf("%s: Directory not found. Please verify the path exists.", baseMsg)
		}
		return TextErrorf("%s: %v", baseMsg, err)
	}
}

// enhanceBranchValidationError provides detailed error messages for branch validation failures
func enhanceBranchValidationError(err error, directory, branch string) *mcp.CallToolResult {
	baseMsg := fmt.Sprintf("Failed to validate branch '%s' in directory '%s'", branch, directory)

	errMsg := err.Error()
	if strings.Contains(errMsg, "unknown revision") || strings.Contains(errMsg, "not found") {
		return TextErrorf("%s: Branch '%s' does not exist locally. Available branches:\n• Use 'git branch' to list local branches\n• Use 'git checkout -b %s' to create it\n• Use 'git fetch origin' to update remote branches", baseMsg, branch, branch)
	}
	if strings.Contains(errMsg, "permission denied") {
		return TextErrorf("%s: Permission denied accessing git repository. Please check directory permissions.", baseMsg)
	}
	return TextErrorf("%s: %v", baseMsg, err)
}

// enhanceConflictErrorMessage provides detailed error messages for conflict detection failures
func enhanceConflictErrorMessage(err error, base, head string) *mcp.CallToolResult {
	baseMsg := fmt.Sprintf("Failed to check conflicts between '%s' and '%s'", base, head)

	errMsg := err.Error()
	if strings.Contains(errMsg, "not a git repository") {
		return TextErrorf("%s: Not in a git repository. Please navigate to a valid git repository.", baseMsg)
	}
	if strings.Contains(errMsg, "fatal: bad revision") {
		return TextErrorf("%s: One or both branches do not exist. Please check branch names with 'git branch -a'.", baseMsg)
	}
	if strings.Contains(errMsg, "timeout") {
		return TextErrorf("%s: Git command timed out. The repository may be large or network slow. Try again.", baseMsg)
	}
	return TextErrorf("%s: %v", baseMsg, err)
}

// enhancePullRequestCreationError provides detailed error messages for PR creation failures
func enhancePullRequestCreationError(err error, repository, head, base string) *mcp.CallToolResult {
	baseMsg := fmt.Sprintf("Failed to create pull request in '%s' from '%s' to '%s'", repository, head, base)

	errMsg := err.Error()

	// API-related errors
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
		return TextErrorf("%s: Authentication failed. Please check your API token is valid and has pull request permissions.", baseMsg)
	}
	if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
		return TextErrorf("%s: Permission denied. Your token may not have pull request creation permissions for this repository.", baseMsg)
	}
	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
		return TextErrorf("%s: Repository or branch not found. Please verify:\n• Repository '%s' exists and is accessible\n• Branch '%s' exists in the repository\n• Branch '%s' exists in the repository", baseMsg, repository, head, base)
	}
	if strings.Contains(errMsg, "409") || strings.Contains(errMsg, "conflict") {
		return TextErrorf("%s: Pull request already exists or there's a conflict. Check if a PR for these branches already exists.", baseMsg)
	}
	if strings.Contains(errMsg, "422") || strings.Contains(errMsg, "validation") {
		return TextErrorf("%s: Validation failed. Please check:\n• Title is not empty and within length limits\n• Branch names are valid\n• Repository format is 'owner/repo'", baseMsg)
	}
	if strings.Contains(errMsg, "500") || strings.Contains(errMsg, "internal server error") {
		return TextErrorf("%s: Server error. The Forgejo/Gitea instance is experiencing issues. Try again later.", baseMsg)
	}

	// Network-related errors
	if strings.Contains(errMsg, "connection refused") {
		return TextErrorf("%s: Cannot connect to the Forgejo/Gitea server. Please check your network connection and server URL.", baseMsg)
	}
	if strings.Contains(errMsg, "timeout") {
		return TextErrorf("%s: Request timed out. The server may be slow or the repository large. Try again.", baseMsg)
	}
	if strings.Contains(errMsg, "no such host") {
		return TextErrorf("%s: Cannot resolve the server hostname. Please check your network configuration.", baseMsg)
	}

	// Git-related errors
	if strings.Contains(errMsg, "branch not found") {
		return TextErrorf("%s: One or both branches not found. Please verify branches exist with 'git branch -a'.", baseMsg)
	}
	if strings.Contains(errMsg, "diverged") {
		return TextErrorf("%s: Branches have diverged. Consider rebasing or merging before creating PR.", baseMsg)
	}

	return TextErrorf("%s: %v", baseMsg, err)
}
