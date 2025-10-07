package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/kunde21/forgejo-mcp/remote"
)

// LoadPRTemplate attempts to load PR template from repository
func LoadPRTemplate(ctx context.Context, client remote.FileContentFetcher, owner, repo, branch string) (string, error) {
	// Try to load PR template from .gitea directory
	templatePath := ".gitea/PULL_REQUEST_TEMPLATE.md"
	content, err := client.GetFileContent(ctx, owner, repo, branch, templatePath)
	if err == nil && len(content) > 0 {
		return string(content), nil
	}

	// Try alternative locations
	alternativePaths := []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		"PULL_REQUEST_TEMPLATE.md",
		"docs/PULL_REQUEST_TEMPLATE.md",
	}

	for _, path := range alternativePaths {
		content, err := client.GetFileContent(ctx, owner, repo, branch, path)
		if err == nil && len(content) > 0 {
			return string(content), nil
		}
	}

	return "", fmt.Errorf("no PR template found")
}

// MergeTemplateContent merges template with user-provided content
func MergeTemplateContent(template, userContent string) string {
	if userContent == "" {
		return template
	}

	// If template contains placeholders, try to fill them
	if strings.Contains(template, "{{") && strings.Contains(template, "}}") {
		// Simple placeholder replacement - in a real implementation,
		// this could be more sophisticated with variable extraction
		result := template

		// Replace common placeholders with user content
		result = strings.ReplaceAll(result, "{{title}}", extractFirstLine(userContent))
		result = strings.ReplaceAll(result, "{{description}}", userContent)
		result = strings.ReplaceAll(result, "{{body}}", userContent)

		return result
	}

	// If no placeholders, append user content to template
	if template != "" && userContent != "" {
		return template + "\n\n" + userContent
	}

	// Return whichever is non-empty
	if template != "" {
		return template
	}
	return userContent
}

// extractFirstLine extracts the first line from text, useful for title placeholders
func extractFirstLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}
