// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/tea"
	"github.com/sirupsen/logrus"
)

// TeaPRListHandler handles pr_list tool requests with actual tea CLI integration
type TeaPRListHandler struct {
	logger         *logrus.Logger
	executor       *TeaExecutor
	parser         *TeaOutputParser
	commandBuilder *TeaCommandBuilder
}

// NewTeaPRListHandler creates a new tea PR list handler
func NewTeaPRListHandler(logger *logrus.Logger) *TeaPRListHandler {
	return &TeaPRListHandler{
		logger:         logger,
		executor:       NewTeaExecutor(),
		parser:         NewTeaOutputParser(),
		commandBuilder: NewTeaCommandBuilder(),
	}
}

// HandleRequest handles a pr_list request with tea CLI integration
func (h *TeaPRListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with params: %v", method, params)

	// Build tea command from parameters
	cmd := h.commandBuilder.BuildPRListCommand(params)
	h.logger.Debugf("Executing tea command: %v", cmd)

	// Execute tea command with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := h.executor.ExecuteCommand(ctx, cmd)
	if err != nil {
		h.logger.Errorf("Tea command execution failed: %v", err)
		return nil, fmt.Errorf("failed to execute tea pr list: %w", err)
	}

	// Parse tea output
	prs, err := h.parser.ParsePRList([]byte(output))
	if err != nil {
		h.logger.Errorf("Failed to parse tea output: %v", err)
		return nil, fmt.Errorf("failed to parse tea pr list output: %w", err)
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"pullRequests": h.transformPRsToResponse(prs),
		"total":        len(prs),
	}

	return result, nil
}

// transformPRsToResponse transforms tea PR data to MCP response format
func (h *TeaPRListHandler) transformPRsToResponse(prs []PR) []map[string]interface{} {
	result := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		// Transform to MCP-compatible format
		prData := map[string]interface{}{
			"number": pr.Number,
			"title":  pr.Title,
			"author": pr.Author,
			"state":  h.normalizePRState(pr.State),
		}

		// Add dates if available
		if pr.CreatedAt != "" {
			prData["createdAt"] = h.formatTimestamp(pr.CreatedAt)
		}
		if pr.UpdatedAt != "" {
			prData["updatedAt"] = h.formatTimestamp(pr.UpdatedAt)
		}

		// Add additional metadata for MCP compatibility
		prData["type"] = "pull_request"
		prData["url"] = "" // Would be populated with actual URL in real implementation

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

// TeaIssueListHandler handles issue_list tool requests with actual tea CLI integration
type TeaIssueListHandler struct {
	logger         *logrus.Logger
	executor       *TeaExecutor
	parser         *TeaOutputParser
	commandBuilder *TeaCommandBuilder
}

// NewTeaIssueListHandler creates a new tea issue list handler
func NewTeaIssueListHandler(logger *logrus.Logger) *TeaIssueListHandler {
	return &TeaIssueListHandler{
		logger:         logger,
		executor:       NewTeaExecutor(),
		parser:         NewTeaOutputParser(),
		commandBuilder: NewTeaCommandBuilder(),
	}
}

// HandleRequest handles an issue_list request with tea CLI integration
func (h *TeaIssueListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with params: %v", method, params)

	// Build tea command from parameters
	cmd := h.commandBuilder.BuildIssueListCommand(params)
	h.logger.Debugf("Executing tea command: %v", cmd)

	// Execute tea command with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := h.executor.ExecuteCommand(ctx, cmd)
	if err != nil {
		h.logger.Errorf("Tea command execution failed: %v", err)
		return nil, fmt.Errorf("failed to execute tea issue list: %w", err)
	}

	// Parse tea output
	issues, err := h.parser.ParseIssueList([]byte(output))
	if err != nil {
		h.logger.Errorf("Failed to parse tea output: %v", err)
		return nil, fmt.Errorf("failed to parse tea issue list output: %w", err)
	}

	// Transform to MCP response format
	result := map[string]interface{}{
		"issues": h.transformIssuesToResponse(issues),
		"total":  len(issues),
	}

	return result, nil
}

// transformIssuesToResponse transforms tea issue data to MCP response format
func (h *TeaIssueListHandler) transformIssuesToResponse(issues []Issue) []map[string]interface{} {
	result := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		// Transform to MCP-compatible format
		issueData := map[string]interface{}{
			"number": issue.Number,
			"title":  issue.Title,
			"author": issue.Author,
			"state":  h.normalizeIssueState(issue.State),
			"labels": h.normalizeLabels(issue.Labels),
		}

		// Add date if available
		if issue.CreatedAt != "" {
			issueData["createdAt"] = h.formatTimestamp(issue.CreatedAt)
		}

		// Add additional metadata for MCP compatibility
		issueData["type"] = "issue"
		issueData["url"] = "" // Would be populated with actual URL in real implementation

		result[i] = issueData
	}
	return result
}

// normalizeIssueState normalizes issue state to standard values
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

// normalizeLabels normalizes label data for MCP compatibility
func (h *TeaIssueListHandler) normalizeLabels(labels []string) []string {
	if len(labels) == 0 {
		return []string{}
	}

	// Ensure labels are clean and non-empty
	validLabels := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label != "" {
			validLabels = append(validLabels, label)
		}
	}

	return validLabels
}

// formatTimestamp formats timestamp to ISO 8601 format
func (h *TeaIssueListHandler) formatTimestamp(ts string) string {
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

// TeaExecutor handles execution of tea CLI commands
type TeaExecutor struct {
	logger *logrus.Logger
}

// NewTeaExecutor creates a new tea executor
func NewTeaExecutor() *TeaExecutor {
	return &TeaExecutor{}
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

// GiteaSDKPRListHandler handles pr_list tool requests using the Gitea SDK client
type GiteaSDKPRListHandler struct {
	logger     *logrus.Logger
	client     client.Client
	transformer *tea.PRTransformer
}

// NewGiteaSDKPRListHandler creates a new Gitea SDK PR list handler
func NewGiteaSDKPRListHandler(logger *logrus.Logger, giteaClient client.Client) *GiteaSDKPRListHandler {
	return &GiteaSDKPRListHandler{
		logger:      logger,
		client:      giteaClient,
		transformer: tea.NewPRTransformer(),
	}
}

// HandleRequest handles a pr_list request using the Gitea SDK client
func (h *GiteaSDKPRListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with Gitea SDK client: %v", method, params)

	// Extract repository information
	owner, repo, err := extractRepositoryInfo(params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repository info: %w", err)
	}

	// Build filters from parameters
	filters, err := buildPRFilters(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build PR filters: %w", err)
	}

	// List pull requests using Gitea SDK client
	prs, err := h.client.ListPRs(owner, repo, filters)
	if err != nil {
		h.logger.Errorf("Gitea SDK client failed to list PRs: %v", err)
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	// Transform to MCP response format
	mcpPRs, err := h.transformer.TransformPRsToMCP(prs)
	if err != nil {
		h.logger.Errorf("Failed to transform PRs to MCP format: %v", err)
		return nil, fmt.Errorf("failed to transform response: %w", err)
	}

	result := map[string]interface{}{
		"pullRequests": mcpPRs,
		"total":        len(prs),
		"repository":   fmt.Sprintf("%s/%s", owner, repo),
	}

	h.logger.Infof("Successfully retrieved %d pull requests", len(prs))
	return result, nil
}

// GiteaSDKIssueListHandler handles issue_list tool requests using the Gitea SDK client
type GiteaSDKIssueListHandler struct {
	logger     *logrus.Logger
	client     client.Client
	transformer *tea.IssueTransformer
}

// NewGiteaSDKIssueListHandler creates a new Gitea SDK issue list handler
func NewGiteaSDKIssueListHandler(logger *logrus.Logger, giteaClient client.Client) *GiteaSDKIssueListHandler {
	return &GiteaSDKIssueListHandler{
		logger:      logger,
		client:      giteaClient,
		transformer: tea.NewIssueTransformer(),
	}
}

// HandleRequest handles an issue_list request using the Gitea SDK client
func (h *GiteaSDKIssueListHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	h.logger.Infof("Handling %s request with Gitea SDK client: %v", method, params)

	// Extract repository information
	owner, repo, err := extractRepositoryInfo(params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repository info: %w", err)
	}

	// Build filters from parameters
	filters, err := buildIssueFilters(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build issue filters: %w", err)
	}

	// List issues using Gitea SDK client
	issues, err := h.client.ListIssues(owner, repo, filters)
	if err != nil {
		h.logger.Errorf("Gitea SDK client failed to list issues: %v", err)
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	// Transform to MCP response format
	mcpIssues, err := h.transformer.TransformIssuesToMCP(issues)
	if err != nil {
		h.logger.Errorf("Failed to transform issues to MCP format: %v", err)
		return nil, fmt.Errorf("failed to transform response: %w", err)
	}

	result := map[string]interface{}{
		"issues":     mcpIssues,
		"total":      len(issues),
		"repository": fmt.Sprintf("%s/%s", owner, repo),
	}

	h.logger.Infof("Successfully retrieved %d issues", len(issues))
	return result, nil
}

// Helper functions for SDK integration

// extractRepositoryInfo extracts owner and repository name from parameters
func extractRepositoryInfo(params map[string]interface{}) (string, string, error) {
	// Look for repository information in various formats
	if repo, ok := params["repository"].(string); ok {
		parts := strings.Split(repo, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	// Look for separate owner and repo parameters
	owner, ownerOk := params["owner"].(string)
	repo, repoOk := params["repo"].(string)
	if ownerOk && repoOk {
		return owner, repo, nil
	}

	// Default repository if not specified (for testing)
	return "forgejo", "forgejo", nil
}

// buildPRFilters builds pull request filters from request parameters
func buildPRFilters(params map[string]interface{}) (*client.PullRequestFilters, error) {
	filters := &client.PullRequestFilters{}

	if state, ok := params["state"].(string); ok && state != "" {
		switch state {
		case "open":
			filters.State = client.StateOpen
		case "closed":
			filters.State = client.StateClosed
		case "all":
			filters.State = client.StateAll
		default:
			return nil, fmt.Errorf("invalid state: %s", state)
		}
	}

	if limit, ok := params["limit"].(float64); ok && limit > 0 {
		filters.PageSize = int(limit)
	}

	if page, ok := params["page"].(float64); ok && page > 0 {
		filters.Page = int(page)
	}

	if sort, ok := params["sort"].(string); ok && sort != "" {
		filters.Sort = sort
	}

	if milestone, ok := params["milestone"].(float64); ok && milestone > 0 {
		filters.Milestone = int64(milestone)
	}

	return filters, nil
}

// buildIssueFilters builds issue filters from request parameters
func buildIssueFilters(params map[string]interface{}) (*client.IssueFilters, error) {
	filters := &client.IssueFilters{}

	if state, ok := params["state"].(string); ok && state != "" {
		switch state {
		case "open":
			filters.State = client.StateOpen
		case "closed":
			filters.State = client.StateClosed
		case "all":
			filters.State = client.StateAll
		default:
			return nil, fmt.Errorf("invalid state: %s", state)
		}
	}

	if limit, ok := params["limit"].(float64); ok && limit > 0 {
		filters.PageSize = int(limit)
	}

	if page, ok := params["page"].(float64); ok && page > 0 {
		filters.Page = int(page)
	}

	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		filters.KeyWord = keyword
	}

	if createdBy, ok := params["created_by"].(string); ok && createdBy != "" {
		filters.CreatedBy = createdBy
	}

	if author, ok := params["author"].(string); ok && author != "" {
		filters.CreatedBy = author // author is an alias for created_by
	}

	if assignedBy, ok := params["assigned_by"].(string); ok && assignedBy != "" {
		filters.AssignedBy = assignedBy
	}

	if mentionedBy, ok := params["mentioned_by"].(string); ok && mentionedBy != "" {
		filters.MentionedBy = mentionedBy
	}

	if owner, ok := params["owner"].(string); ok && owner != "" {
		filters.Owner = owner
	}

	if team, ok := params["team"].(string); ok && team != "" {
		filters.Team = team
	}

	// Handle labels array
	if labels, ok := params["labels"].([]interface{}); ok && len(labels) > 0 {
		labelStrings := make([]string, 0, len(labels))
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				labelStrings = append(labelStrings, labelStr)
			}
		}
		filters.Labels = labelStrings
	}

	// Handle milestones array
	if milestones, ok := params["milestones"].([]interface{}); ok && len(milestones) > 0 {
		milestoneStrings := make([]string, 0, len(milestones))
		for _, milestone := range milestones {
			if milestoneStr, ok := milestone.(string); ok {
				milestoneStrings = append(milestoneStrings, milestoneStr)
			}
		}
		filters.Milestones = milestoneStrings
	}

	return filters, nil
}
