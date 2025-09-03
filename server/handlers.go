// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"fmt"
	"time"

	ctxt "github.com/Kunde21/forgejo-mcp/context"
	"github.com/Kunde21/forgejo-mcp/types"
	"github.com/sirupsen/logrus"
)

// PRListHandler handles pr_list tool requests
type PRListHandler struct {
	logger *logrus.Logger
}

// NewPRListHandler creates a new PR list handler
func NewPRListHandler(logger *logrus.Logger) *PRListHandler {
	return &PRListHandler{
		logger: logger,
	}
}

// HandleRequest handles a pr_list request
func (h *PRListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with params: %v", method, params)

	// TODO: Implement actual tea CLI integration
	// For now, return mock data using new types

	createdAt1, _ := time.Parse(time.RFC3339, "2025-08-26T10:00:00Z")
	updatedAt1, _ := time.Parse(time.RFC3339, "2025-08-26T15:30:00Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2025-08-25T09:15:00Z")
	updatedAt2, _ := time.Parse(time.RFC3339, "2025-08-25T14:45:00Z")

	mockPRs := []types.PullRequest{
		{
			ID:         42,
			Number:     42,
			Title:      "Add dark mode support",
			State:      types.PRStateOpen,
			Author:     &types.PRAuthor{Username: "developer1", AvatarURL: "https://example.com/avatar1.jpg", URL: "https://example.com/user/developer1"},
			HeadBranch: "feature/dark-mode",
			BaseBranch: "main",
			CreatedAt:  types.Timestamp{Time: createdAt1},
			UpdatedAt:  types.Timestamp{Time: updatedAt1},
			URL:        "https://example.com/pr/42",
			DiffURL:    "https://example.com/pr/42.diff",
		},
		{
			ID:         41,
			Number:     41,
			Title:      "Fix authentication bug",
			State:      types.PRStateMerged,
			Author:     &types.PRAuthor{Username: "developer2", AvatarURL: "https://example.com/avatar2.jpg", URL: "https://example.com/user/developer2"},
			HeadBranch: "fix/auth-bug",
			BaseBranch: "main",
			CreatedAt:  types.Timestamp{Time: createdAt2},
			UpdatedAt:  types.Timestamp{Time: updatedAt2},
			MergedAt:   &types.Timestamp{Time: updatedAt2},
			URL:        "https://example.com/pr/41",
			DiffURL:    "https://example.com/pr/41.diff",
		},
	}

	// Return typed response
	return types.NewPaginatedResponse(mockPRs, types.NewPagination(1, 10, len(mockPRs))), nil
}

// IssueListHandler handles issue_list tool requests
type IssueListHandler struct {
	logger *logrus.Logger
}

// NewIssueListHandler creates a new issue list handler
func NewIssueListHandler(logger *logrus.Logger) *IssueListHandler {
	return &IssueListHandler{
		logger: logger,
	}
}

// HandleRequest handles an issue_list request
func (h *IssueListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with params: %v", method, params)

	// TODO: Implement actual tea CLI integration
	// For now, return mock data using new types

	createdAt1, _ := time.Parse(time.RFC3339, "2025-08-24T08:30:00Z")
	updatedAt1, _ := time.Parse(time.RFC3339, "2025-08-24T10:15:00Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2025-08-23T14:20:00Z")
	updatedAt2, _ := time.Parse(time.RFC3339, "2025-08-23T16:45:00Z")
	closedAt2, _ := time.Parse(time.RFC3339, "2025-08-23T16:45:00Z")

	mockIssues := []types.Issue{
		{
			ID:        123,
			Number:    123,
			Title:     "UI responsiveness issue on mobile",
			State:     types.IssueStateOpen,
			Author:    &types.User{ID: 1, Username: "user1", Email: "user1@example.com"},
			Labels:    []types.PRLabel{{Name: "bug"}, {Name: "ui"}, {Name: "mobile"}},
			CreatedAt: types.Timestamp{Time: createdAt1},
			UpdatedAt: types.Timestamp{Time: updatedAt1},
			URL:       "https://example.com/issue/123",
		},
		{
			ID:        122,
			Number:    122,
			Title:     "Documentation update needed",
			State:     types.IssueStateClosed,
			Author:    &types.User{ID: 2, Username: "user2", Email: "user2@example.com"},
			Labels:    []types.PRLabel{{Name: "documentation"}},
			CreatedAt: types.Timestamp{Time: createdAt2},
			UpdatedAt: types.Timestamp{Time: updatedAt2},
			ClosedAt:  &types.Timestamp{Time: closedAt2},
			URL:       "https://example.com/issue/122",
		},
	}

	// Return typed response
	return types.NewPaginatedResponse(mockIssues, types.NewPagination(1, 10, len(mockIssues))), nil
}

// ToolManifestHandler handles tool manifest requests
type ToolManifestHandler struct {
	logger *logrus.Logger
}

// NewToolManifestHandler creates a new tool manifest handler
func NewToolManifestHandler(logger *logrus.Logger) *ToolManifestHandler {
	return &ToolManifestHandler{
		logger: logger,
	}
}

// HandleRequest handles a tools/list request
func (h *ToolManifestHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request", method)

	tools := []map[string]interface{}{
		{
			"name":        "pr_list",
			"description": "List pull requests from the Forgejo repository",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Filter by PR state (open, closed, merged)",
						"enum":        []string{"open", "closed", "merged"},
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Filter by PR author",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of PRs to return",
						"minimum":     1,
						"maximum":     100,
					},
				},
			},
		},
		{
			"name":        "issue_list",
			"description": "List issues from the Forgejo repository",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Filter by issue state (open, closed)",
						"enum":        []string{"open", "closed"},
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"description": "Filter by issue labels",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of issues to return",
						"minimum":     1,
						"maximum":     100,
					},
				},
			},
		},
		{
			"name":        "context_detect",
			"description": "Detect repository context from the current git environment",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to check for repository context (defaults to current directory)",
					},
				},
			},
		},
	}

	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// ContextDetectHandler handles context_detect tool requests
type ContextDetectHandler struct {
	logger *logrus.Logger
}

// NewContextDetectHandler creates a new context detect handler
func NewContextDetectHandler(logger *logrus.Logger) *ContextDetectHandler {
	return &ContextDetectHandler{
		logger: logger,
	}
}

// HandleRequest handles a context_detect request
func (h *ContextDetectHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with params: %v", method, params)

	// Extract path parameter, default to current directory
	path := "."
	if pathParam, exists := params["path"]; exists {
		if pathStr, ok := pathParam.(string); ok && pathStr != "" {
			path = pathStr
		}
	}

	// Detect repository context
	repoCtx, err := ctxt.DetectContext(path)
	if err != nil {
		h.logger.Errorf("Context detection failed for path %s: %v", path, err)
		return types.NewErrorResponse(types.ErrorCodeNotFound, "Failed to detect repository context"), nil
	}

	h.logger.Infof("Successfully detected context: %s", repoCtx.String())

	// Return typed repository information
	repository := types.Repository{
		Owner:    repoCtx.Owner,
		Name:     repoCtx.Repository,
		FullName: repoCtx.String(),
		URL:      repoCtx.RemoteURL,
	}

	return types.NewSuccessResponse(repository), nil
}

// HealthCheckHandler handles health check requests
type HealthCheckHandler struct {
	logger *logrus.Logger
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(logger *logrus.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{
		logger: logger,
	}
}

// HandleRequest handles a health check request
func (h *HealthCheckHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Debug("Handling health check request")

	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}, nil
}

// RegisterDefaultHandlers registers the default tool handlers
func (s *Server) RegisterDefaultHandlers() {
	// Register tool handlers
	prHandler := NewPRListHandler(s.logger)
	issueHandler := NewIssueListHandler(s.logger)
	contextHandler := NewContextDetectHandler(s.logger)
	manifestHandler := NewToolManifestHandler(s.logger)
	healthHandler := NewHealthCheckHandler(s.logger)

	s.dispatcher.RegisterHandler("tools/call", &ToolCallRouter{
		prHandler:      prHandler,
		issueHandler:   issueHandler,
		contextHandler: contextHandler,
		logger:         s.logger,
	})
	s.dispatcher.RegisterHandler("tools/list", manifestHandler)
	s.dispatcher.RegisterHandler("health/check", healthHandler)

	s.logger.Info("Registered default MCP tool handlers")
}

// ToolCallRouter routes tool calls to the appropriate handler based on the tool name
type ToolCallRouter struct {
	prHandler      *PRListHandler
	issueHandler   *IssueListHandler
	contextHandler *ContextDetectHandler
	logger         *logrus.Logger
}

// HandleRequest routes a tool call to the appropriate handler
func (r *ToolCallRouter) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	r.logger.Debugf("Routing tool call with params: %v", params)

	// Extract tool name from params
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required and must be a string")
	}

	// Extract tool arguments
	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Route to appropriate handler
	switch toolName {
	case "pr_list":
		return r.prHandler.HandleRequest(ctx, toolName, arguments)
	case "issue_list":
		return r.issueHandler.HandleRequest(ctx, toolName, arguments)
	case "context_detect":
		return r.contextHandler.HandleRequest(ctx, toolName, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}
