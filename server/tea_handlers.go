// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TeaPRListHandler handles pr_list tool requests with Gitea SDK integration
type TeaPRListHandler struct {
	logger *logrus.Logger
	client GiteaClientInterface
}

// NewTeaPRListHandler creates a new tea PR list handler
func NewTeaPRListHandler(logger *logrus.Logger, client GiteaClientInterface) *TeaPRListHandler {
	return &TeaPRListHandler{
		logger: logger,
		client: client,
	}
}

// HandlePRListRequest handles a pr_list request with Gitea SDK integration
func (h *TeaPRListHandler) HandlePRListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
	State  string `json:"state,omitempty"`
	Author string `json:"author,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling pr_list request with Gitea SDK")

	// Build SDK options from parameters
	opts := gitea.ListPullRequestsOptions{}

	if args.State != "" {
		switch args.State {
		case "open":
			opts.State = gitea.StateOpen
		case "closed":
			opts.State = gitea.StateClosed
		case "all":
			opts.State = gitea.StateAll
		default:
			opts.State = gitea.StateOpen // default to open
		}
	} else {
		opts.State = gitea.StateOpen // default to open
	}

	// Note: Gitea SDK doesn't have direct author filtering in ListPullRequestsOptions
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	// For this example, we'll use placeholder owner/repo
	// In a real implementation, this would come from context or configuration
	owner := "example-owner"
	repo := "example-repo"

	prs, _, err := h.client.ListRepoPullRequests(owner, repo, opts)
	if err != nil {
		h.logger.Errorf("Gitea SDK ListRepoPullRequests failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK pr list: %v", err),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"pullRequests": h.transformPRsToResponse(prs),
		"total":        len(prs),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d pull requests", len(prs)),
			},
		},
	}, result, nil
}

// transformPRsToResponse transforms Gitea SDK PR data to MCP response format
func (h *TeaPRListHandler) transformPRsToResponse(prs []*gitea.PullRequest) []map[string]interface{} {
	result := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		// Transform to MCP-compatible format
		prData := map[string]interface{}{
			"number": pr.Index,
			"title":  pr.Title,
			"state":  h.normalizePRState(string(pr.State)),
		}

		// Add author if available
		if pr.Poster != nil {
			prData["author"] = pr.Poster.UserName
		} else {
			prData["author"] = ""
		}

		// Add dates if available
		if pr.Created != nil {
			prData["createdAt"] = pr.Created.Format("2006-01-02T15:04:05Z")
		}
		if pr.Updated != nil {
			prData["updatedAt"] = pr.Updated.Format("2006-01-02T15:04:05Z")
		}

		// Add additional metadata for MCP compatibility
		prData["type"] = "pull_request"
		prData["url"] = pr.HTMLURL

		result[i] = prData
	}
	return result
}

// normalizePRState normalizes PR state to standard values
func (h *TeaPRListHandler) normalizePRState(state string) string {
	switch strings.ToLower(state) {
	case "open":
		return "open"
	case "closed":
		return "closed"
	case "merged":
		return "merged"
	default:
		return "unknown"
	}
}

// formatTimestamp formats timestamp to ISO 8601 format
func (h *TeaPRListHandler) formatTimestamp(ts string) string {
	if ts == "" {
		return ""
	}

	// If already in ISO format, return as-is
	if strings.Contains(ts, "T") && strings.Contains(ts, "Z") {
		return ts
	}

	// Try to parse and reformat if needed
	// For now, return as-is since tea should provide proper ISO format
	return ts
}

// TeaIssueListHandler handles issue_list tool requests with Gitea SDK integration
type TeaIssueListHandler struct {
	logger *logrus.Logger
	client GiteaClientInterface
}

// NewTeaIssueListHandler creates a new tea issue list handler
func NewTeaIssueListHandler(logger *logrus.Logger, client GiteaClientInterface) *TeaIssueListHandler {
	return &TeaIssueListHandler{
		logger: logger,
		client: client,
	}
}

// HandleIssueListRequest handles an issue_list request with Gitea SDK integration
func (h *TeaIssueListHandler) HandleIssueListRequest(ctx context.Context, req *mcp.CallToolRequest, args struct {
	State  string   `json:"state,omitempty"`
	Author string   `json:"author,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Handling issue_list request with Gitea SDK")

	// Build SDK options from parameters
	opts := gitea.ListIssueOption{}

	if args.State != "" {
		switch args.State {
		case "open":
			opts.State = gitea.StateOpen
		case "closed":
			opts.State = gitea.StateClosed
		case "all":
			opts.State = gitea.StateAll
		default:
			opts.State = gitea.StateOpen // default to open
		}
	} else {
		opts.State = gitea.StateOpen // default to open
	}

	// Note: Gitea SDK doesn't have direct author/label filtering in ListIssueOption
	// This would need to be handled by filtering results after fetching

	if args.Limit > 0 {
		opts.ListOptions.PageSize = args.Limit
	}

	issues, _, err := h.client.ListIssues(opts)
	if err != nil {
		h.logger.Errorf("Gitea SDK ListIssues failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error executing SDK issue list: %v", err),
				},
			},
		}, nil, nil
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"issues": h.transformIssuesToResponse(issues),
		"total":  len(issues),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d issues", len(issues)),
			},
		},
	}, result, nil
}

// transformIssuesToResponse transforms Gitea SDK issue data to MCP response format
func (h *TeaIssueListHandler) transformIssuesToResponse(issues []*gitea.Issue) []map[string]interface{} {
	result := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		// Transform to MCP-compatible format
		issueData := map[string]interface{}{
			"number": issue.Index,
			"title":  issue.Title,
			"state":  h.normalizeIssueState(string(issue.State)),
		}

		// Add author if available
		if issue.Poster != nil {
			issueData["author"] = issue.Poster.UserName
		} else {
			issueData["author"] = ""
		}

		// Add dates if available
		issueData["createdAt"] = issue.Created.Format("2006-01-02T15:04:05Z")
		issueData["updatedAt"] = issue.Updated.Format("2006-01-02T15:04:05Z")

		// Add additional metadata for MCP compatibility
		issueData["type"] = "issue"
		issueData["url"] = issue.HTMLURL

		result[i] = issueData
	}
	return result
}

// normalizeIssueState normalizes Gitea SDK issue state to standard values
func (h *TeaIssueListHandler) normalizeIssueState(state string) string {
	switch strings.ToLower(state) {
	case "open":
		return "open"
	case "closed":
		return "closed"
	default:
		return "unknown"
	}
}

// TeaExecutor handles execution of tea CLI commands
type TeaExecutor struct {
	logger *logrus.Logger
}

// NewTeaExecutor creates a new tea executor
func NewTeaExecutor(logger *logrus.Logger) *TeaExecutor {
	return &TeaExecutor{logger: logger}
}

// ExecuteCommand executes a tea command and returns the output
func (te *TeaExecutor) ExecuteCommand(ctx context.Context, cmd []string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	te.logger.Debugf("Executing command: %v", cmd)

	// Create exec command
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

	// Execute command and capture output
	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

// TeaOutputParser handles parsing of tea CLI output
type TeaOutputParser struct {
	logger *logrus.Logger
}

// NewTeaOutputParser creates a new tea output parser
func NewTeaOutputParser() *TeaOutputParser {
	return &TeaOutputParser{}
}

// ParsePRList parses tea pr list JSON output
func (top *TeaOutputParser) ParsePRList(data []byte) ([]PR, error) {
	// First try to parse as JSON
	var prs []PR
	if err := json.Unmarshal(data, &prs); err != nil {
		// If JSON parsing fails, try to parse as text format
		return top.parsePRListText(data)
	}
	return prs, nil
}

// ParseIssueList parses tea issue list JSON output
func (top *TeaOutputParser) ParseIssueList(data []byte) ([]Issue, error) {
	// First try to parse as JSON
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		// If JSON parsing fails, try to parse as text format
		return top.parseIssueListText(data)
	}
	return issues, nil
}

// parsePRListText parses tea pr list text output (fallback)
func (top *TeaOutputParser) parsePRListText(data []byte) ([]PR, error) {
	text := string(data)
	lines := strings.Split(text, "\n")

	var prs []PR
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse line as PR format
		// This is a basic implementation - real tea text format may vary
		if strings.Contains(line, "#") && strings.Contains(line, "by") {
			// Example format: "#42 Add dark mode support by developer1"
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				var pr PR
				// Extract number from #42
				if len(parts[0]) > 1 && parts[0][0] == '#' {
					if num, err := fmt.Sscanf(parts[0][1:], "%d", &pr.Number); err == nil && num == 1 {
						// Find "by" keyword
						byIndex := -1
						for i, part := range parts {
							if part == "by" && i+1 < len(parts) {
								byIndex = i
								break
							}
						}

						if byIndex > 0 {
							// Extract title (parts between number and "by")
							titleParts := parts[1:byIndex]
							pr.Title = strings.Join(titleParts, " ")
							pr.Author = parts[byIndex+1]
							pr.State = "unknown" // Text format doesn't include state
							pr.CreatedAt = ""    // Text format doesn't include dates
							pr.UpdatedAt = ""
							prs = append(prs, pr)
						}
					}
				}
			}
		}
	}

	if len(prs) == 0 {
		return nil, fmt.Errorf("no PRs found in text output")
	}

	return prs, nil
}

// parseIssueListText parses tea issue list text output (fallback)
func (top *TeaOutputParser) parseIssueListText(data []byte) ([]Issue, error) {
	text := string(data)
	lines := strings.Split(text, "\n")

	var issues []Issue
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse line as issue format
		// This is a basic implementation - real tea text format may vary
		if strings.Contains(line, "#") && strings.Contains(line, "by") {
			// Example format: "#123 UI responsiveness issue by user1"
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				var issue Issue
				// Extract number from #123
				if len(parts[0]) > 1 && parts[0][0] == '#' {
					if num, err := fmt.Sscanf(parts[0][1:], "%d", &issue.Number); err == nil && num == 1 {
						// Find "by" keyword
						byIndex := -1
						for i, part := range parts {
							if part == "by" && i+1 < len(parts) {
								byIndex = i
								break
							}
						}

						if byIndex > 0 {
							// Extract title (parts between number and "by")
							titleParts := parts[1:byIndex]
							issue.Title = strings.Join(titleParts, " ")
							issue.Author = parts[byIndex+1]
							issue.State = "unknown"   // Text format doesn't include state
							issue.Labels = []string{} // Text format doesn't include labels
							issue.CreatedAt = ""      // Text format doesn't include dates
							issues = append(issues, issue)
						}
					}
				}
			}
		}
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("no issues found in text output")
	}

	return issues, nil
}

// TeaCommandBuilder represents a builder for tea CLI commands
type TeaCommandBuilder struct {
	logger *logrus.Logger
}

// NewTeaCommandBuilder creates a new tea command builder
func NewTeaCommandBuilder() *TeaCommandBuilder {
	return &TeaCommandBuilder{}
}

// BuildPRListCommand builds a tea pr list command from parameters
func (tcb *TeaCommandBuilder) BuildPRListCommand(params map[string]interface{}) []string {
	cmd := []string{"tea", "pr", "list"}

	if state, ok := params["state"].(string); ok && state != "" {
		// Validate state parameter
		if tcb.isValidState(state) {
			cmd = append(cmd, "--state", tcb.escapeShellArg(state))
		}
	}

	if author, ok := params["author"].(string); ok && author != "" {
		// Validate and escape author parameter
		if tcb.isValidUsername(author) {
			cmd = append(cmd, "--author", tcb.escapeShellArg(author))
		}
	}

	if limit, ok := params["limit"].(float64); ok && limit > 0 {
		// Validate limit parameter
		if limit >= 1 && limit <= 100 {
			cmd = append(cmd, "--limit", fmt.Sprintf("%.0f", limit))
		}
	}

	cmd = append(cmd, "--output", "json")
	return cmd
}

// BuildIssueListCommand builds a tea issue list command from parameters
func (tcb *TeaCommandBuilder) BuildIssueListCommand(params map[string]interface{}) []string {
	cmd := []string{"tea", "issue", "list"}

	if state, ok := params["state"].(string); ok && state != "" {
		// Validate state parameter
		if tcb.isValidState(state) {
			cmd = append(cmd, "--state", tcb.escapeShellArg(state))
		}
	}

	if labels, ok := params["labels"].([]interface{}); ok && len(labels) > 0 {
		// Validate and process labels
		validLabels := tcb.validateAndProcessLabels(labels)
		if len(validLabels) > 0 {
			cmd = append(cmd, "--labels", strings.Join(validLabels, ","))
		}
	}

	if author, ok := params["author"].(string); ok && author != "" {
		// Validate and escape author parameter
		if tcb.isValidUsername(author) {
			cmd = append(cmd, "--author", tcb.escapeShellArg(author))
		}
	}

	if limit, ok := params["limit"].(float64); ok && limit > 0 {
		// Validate limit parameter
		if limit >= 1 && limit <= 100 {
			cmd = append(cmd, "--limit", fmt.Sprintf("%.0f", limit))
		}
	}

	cmd = append(cmd, "--output", "json")
	return cmd
}

// isValidState validates if the state parameter is valid
func (tcb *TeaCommandBuilder) isValidState(state string) bool {
	validStates := []string{"open", "closed", "merged", "all"}
	for _, validState := range validStates {
		if state == validState {
			return true
		}
	}
	return false
}

// isValidUsername validates if the username is valid
func (tcb *TeaCommandBuilder) isValidUsername(username string) bool {
	if len(username) == 0 || len(username) > 255 {
		return false
	}
	// Basic validation - allow alphanumeric, hyphens, underscores
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// escapeShellArg escapes shell arguments to prevent command injection
func (tcb *TeaCommandBuilder) escapeShellArg(arg string) string {
	// For now, return the argument as-is since we're using exec.Command
	// which handles argument passing safely. In a real shell context,
	// we would need proper shell escaping.
	return arg
}

// validateAndProcessLabels validates and processes label parameters
func (tcb *TeaCommandBuilder) validateAndProcessLabels(labels []interface{}) []string {
	if len(labels) > 10 {
		// Limit to first 10 labels
		labels = labels[:10]
	}

	validLabels := make([]string, 0, len(labels))
	for _, label := range labels {
		if labelStr, ok := label.(string); ok {
			if labelStr != "" && len(labelStr) <= 50 {
				validLabels = append(validLabels, tcb.escapeShellArg(labelStr))
			}
		}
	}

	return validLabels
}

// PR represents a pull request from tea CLI output
type PR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Issue represents an issue from tea CLI output
type Issue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	State     string   `json:"state"`
	Labels    []string `json:"labels"`
	CreatedAt string   `json:"created_at"`
}
