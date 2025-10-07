package server

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GitCommandTimeout is the timeout for git command execution
const GitCommandTimeout = 30 * time.Second

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch(directory string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = directory

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get current branch: %w, stderr: %s", err, stderr.String())
	}

	branch := strings.TrimSpace(stdout.String())
	if branch == "" {
		return "", fmt.Errorf("empty branch name returned")
	}

	return branch, nil
}

// BranchExists checks if a branch exists locally
func BranchExists(directory, branch string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = directory

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "unknown revision or path") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check branch existence: %w, stderr: %s", err, stderr.String())
	}

	return true, nil
}

// GetCommitCount returns number of commits between base and head
func GetCommitCount(directory, base, head string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	cmd.Dir = directory

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to get commit count: %w, stderr: %s", err, stderr.String())
	}

	countStr := strings.TrimSpace(stdout.String())
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count '%s': %w", countStr, err)
	}

	return count, nil
}

// ConflictDetail represents detailed information about a conflict
type ConflictDetail struct {
	File     string   `json:"file"`     // Path to the conflicting file
	Type     string   `json:"type"`     // Type of conflict: "content", "add_add", "delete_modify"
	Lines    []int    `json:"lines"`    // Line numbers where conflicts occur (if available)
	Markers  []string `json:"markers"`  // Conflict markers found
	Severity string   `json:"severity"` // "low", "medium", "high" based on conflict complexity
}

// ConflictReport represents a comprehensive conflict analysis
type ConflictReport struct {
	HasConflicts     bool             `json:"has_conflicts"`
	ConflictFiles    []string         `json:"conflict_files"`    // Simple list of conflicting files
	ConflictDetails  []ConflictDetail `json:"conflict_details"`  // Detailed conflict information
	TotalConflicts   int              `json:"total_conflicts"`   // Total number of conflict regions
	SuggestedActions []string         `json:"suggested_actions"` // Suggested resolution steps
}

// HasConflicts detects if merging base into head would cause conflicts
func HasConflicts(directory, base, head string) (bool, []string, error) {
	report, err := GetConflictReport(directory, base, head)
	if err != nil {
		return false, nil, err
	}
	return report.HasConflicts, report.ConflictFiles, nil
}

// GetConflictReport provides detailed conflict analysis
func GetConflictReport(directory, base, head string) (*ConflictReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	// Use git merge-tree to detect conflicts
	cmd := exec.CommandContext(ctx, "git", "merge-tree", base, head)
	cmd.Dir = directory

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// merge-tree exits with non-zero status if conflicts exist
		if strings.Contains(stderr.String(), "merge conflict") || strings.Contains(stdout.String(), "<<<<<<<") {
			return analyzeConflictOutput(stdout.String()), nil
		}
		return nil, fmt.Errorf("failed to check conflicts: %w, stderr: %s", err, stderr.String())
	}

	// Check if output contains conflict markers
	output := stdout.String()
	if strings.Contains(output, "<<<<<<<") || strings.Contains(output, "=======") || strings.Contains(output, ">>>>>>>") {
		return analyzeConflictOutput(output), nil
	}

	return &ConflictReport{
		HasConflicts:     false,
		ConflictFiles:    []string{},
		ConflictDetails:  []ConflictDetail{},
		TotalConflicts:   0,
		SuggestedActions: []string{"Branch appears to be clean and ready for merge"},
	}, nil
}

// IsBranchBehind checks if head branch is behind base branch
func IsBranchBehind(directory, base, head string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	// Check if base is an ancestor of head
	cmd := exec.CommandContext(ctx, "git", "merge-base", "--is-ancestor", base, head)
	cmd.Dir = directory

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If base is not an ancestor of head, head might be behind or diverged
		if strings.Contains(stderr.String(), "is not an ancestor") {
			// Check if head is an ancestor of base (head is behind)
			cmd2 := exec.CommandContext(ctx, "git", "merge-base", "--is-ancestor", head, base)
			cmd2.Dir = directory
			if err2 := cmd2.Run(); err2 == nil {
				return true, nil
			}
		}
		return false, fmt.Errorf("failed to check branch relationship: %w, stderr: %s", err, stderr.String())
	}

	return false, nil
}

// analyzeConflictOutput provides detailed analysis of git merge-tree conflict output
func analyzeConflictOutput(output string) *ConflictReport {
	lines := strings.Split(output, "\n")
	conflictFiles := make(map[string]bool)
	conflictDetails := []ConflictDetail{}
	totalConflicts := 0
	currentFile := ""
	conflictLines := []int{}
	markers := []string{}

	for i, line := range lines {
		// Detect file changes
		if strings.Contains(line, "diff --cc") {
			// Save previous file details if any
			if currentFile != "" && len(conflictLines) > 0 {
				detail := createConflictDetail(currentFile, conflictLines, markers)
				conflictDetails = append(conflictDetails, detail)
				totalConflicts++
			}

			// Extract new file path
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = parts[3]
				conflictFiles[currentFile] = true
				conflictLines = []int{}
				markers = []string{}
			}
		}

		// Detect conflict markers
		if strings.Contains(line, "<<<<<<<") || strings.Contains(line, "=======") || strings.Contains(line, ">>>>>>>") {
			if currentFile != "" {
				conflictLines = append(conflictLines, i+1) // +1 for 1-based line numbers
				markers = append(markers, strings.TrimSpace(line))
			}
		}

		// Detect conflict types
		if strings.Contains(line, "content conflict") {
			if currentFile != "" {
				// Content conflict marker found
			}
		}
	}

	// Don't forget the last file
	if currentFile != "" && len(conflictLines) > 0 {
		detail := createConflictDetail(currentFile, conflictLines, markers)
		conflictDetails = append(conflictDetails, detail)
		totalConflicts++
	}

	// Convert map to slice
	files := make([]string, 0, len(conflictFiles))
	for file := range conflictFiles {
		files = append(files, file)
	}

	// Generate suggested actions
	suggestions := generateConflictSuggestions(totalConflicts, len(files))

	return &ConflictReport{
		HasConflicts:     totalConflicts > 0,
		ConflictFiles:    files,
		ConflictDetails:  conflictDetails,
		TotalConflicts:   totalConflicts,
		SuggestedActions: suggestions,
	}
}

// createConflictDetail creates a ConflictDetail from parsed information
func createConflictDetail(file string, lines []int, markers []string) ConflictDetail {
	conflictType := "content"
	severity := "medium"

	// Analyze markers to determine conflict type and severity
	for _, marker := range markers {
		if strings.Contains(marker, "<<<<<<<") {
			if strings.Contains(marker, "HEAD") || strings.Contains(marker, "ours") {
				conflictType = "content"
			}
		}
	}

	// Determine severity based on number of conflict lines
	if len(lines) > 10 {
		severity = "high"
	} else if len(lines) <= 3 {
		severity = "low"
	}

	return ConflictDetail{
		File:     file,
		Type:     conflictType,
		Lines:    lines,
		Markers:  markers,
		Severity: severity,
	}
}

// generateConflictSuggestions provides actionable advice based on conflict analysis
func generateConflictSuggestions(totalConflicts, fileCount int) []string {
	suggestions := []string{}

	if totalConflicts == 0 {
		suggestions = append(suggestions, "No conflicts detected - branch is ready for merge")
		return suggestions
	}

	if fileCount == 1 {
		suggestions = append(suggestions, fmt.Sprintf("Single file has %d conflict(s) - review and resolve manually", totalConflicts))
	} else {
		suggestions = append(suggestions, fmt.Sprintf("%d files have conflicts - resolve each file before merging", fileCount))
	}

	if totalConflicts > 5 {
		suggestions = append(suggestions, "High number of conflicts - consider rebasing your branch")
	}

	suggestions = append(suggestions, "Use 'git merge-base' to find common ancestor for better context")
	suggestions = append(suggestions, "Test your changes after resolving conflicts")

	return suggestions
}

// parseConflictFiles extracts file paths from git merge-tree conflict output
func parseConflictFiles(output string) []string {
	report := analyzeConflictOutput(output)
	return report.ConflictFiles
}
