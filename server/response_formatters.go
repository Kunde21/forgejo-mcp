package server

import (
	"fmt"
	"strings"

	"github.com/kunde21/forgejo-mcp/remote"
)

// FormatIssueList creates a human-readable summary of issues
func FormatIssueList(issues []remote.Issue) string {
	if len(issues) == 0 {
		return "No issues found"
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "Found %d issues:\n", len(issues))
	for _, issue := range issues {
		fmt.Fprintf(&builder, "- #%d: %s (%s)\n", issue.Number, issue.Title, issue.State)
	}
	return builder.String()
}

// FormatPullRequestList creates a human-readable summary of pull requests
func FormatPullRequestList(pullRequests []remote.PullRequest) string {
	if len(pullRequests) == 0 {
		return "No pull requests found"
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "Found %d pull requests:\n", len(pullRequests))
	for _, pr := range pullRequests {
		fmt.Fprintf(&builder, "- #%d: %s (%s)\n", pr.Number, pr.Title, pr.State)
	}
	return builder.String()
}

// FormatPullRequestDetails creates detailed PR information
func FormatPullRequestDetails(pr *remote.PullRequestDetails) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Pull Request #%d: %s\n", pr.Number, pr.Title)
	fmt.Fprintf(&builder, "State: %s\n", pr.State)
	fmt.Fprintf(&builder, "Author: %s\n", pr.User)
	fmt.Fprintf(&builder, "Created: %s\n", pr.CreatedAt)
	fmt.Fprintf(&builder, "Updated: %s\n", pr.UpdatedAt)
	if pr.Body != "" {
		fmt.Fprintf(&builder, "Body:\n%s\n", pr.Body)
	}

	// Assignee information
	if pr.Assignee != "" {
		fmt.Fprintf(&builder, "Assignee: %s\n", pr.Assignee)
	}

	if len(pr.Assignees) > 0 {
		fmt.Fprintf(&builder, "Assignees: %v\n", pr.Assignees)
	}

	// Labels
	if len(pr.Labels) > 0 {
		builder.WriteString("Labels: ")
		for i, label := range pr.Labels {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(label.Name)
		}
		builder.WriteString("\n")
	}

	// PR metadata
	fmt.Fprintf(&builder, "Comments: %d\n", pr.Comments)
	fmt.Fprintf(&builder, "Mergeable: %t\n", pr.Mergeable)

	// Merge information (if merged)
	if pr.HasMerged {
		fmt.Fprintf(&builder, "Merged: %s", pr.MergedAt)
		if pr.MergedBy != "" {
			fmt.Fprintf(&builder, " by %s", pr.MergedBy)
		}
		builder.WriteString("\n")
	}

	// URL
	fmt.Fprintf(&builder, "URL: %s\n", pr.HTMLURL)

	return builder.String()
}

// FormatIssueDetails creates detailed issue information
func FormatIssueDetails(issue *remote.Issue) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Fprintf(&builder, "State: %s\n", issue.State)
	fmt.Fprintf(&builder, "Author: %s\n", issue.User)
	if issue.Created != "" {
		fmt.Fprintf(&builder, "Created: %s\n", issue.Created)
	}
	if issue.Updated != "" {
		fmt.Fprintf(&builder, "Updated: %s\n", issue.Updated)
	}
	if issue.Body != "" {
		fmt.Fprintf(&builder, "Body:\n%s\n", issue.Body)
	}

	return builder.String()
}

// FormatIssueCreateSuccess creates success message for issue creation
func FormatIssueCreateSuccess(issue *remote.Issue) string {
	return fmt.Sprintf("Issue created successfully. Number: %d, Title: %s", issue.Number, issue.Title)
}

// FormatIssueEditSuccess creates success message for issue editing
func FormatIssueEditSuccess(issue *remote.Issue) string {
	responseText := fmt.Sprintf("Issue edited successfully. Number: %d, Title: %s, State: %s",
		issue.Number, issue.Title, issue.State)
	if issue.Updated != "" {
		responseText += fmt.Sprintf(", Updated: %s", issue.Updated)
	}
	responseText += "\n"

	if issue.Body != "" {
		responseText += fmt.Sprintf("Body: %s\n", issue.Body)
	}

	return responseText
}

// FormatPullRequestCreateSuccess creates success message for PR creation
func FormatPullRequestCreateSuccess(pr *remote.PullRequest) string {
	return fmt.Sprintf("Pull request created successfully. Number: %d, Title: %s", pr.Number, pr.Title)
}

// FormatPullRequestEditSuccess creates success message for PR editing
func FormatPullRequestEditSuccess(pr *remote.PullRequest) string {
	responseText := fmt.Sprintf("Pull request edited successfully. Number: %d, Title: %s, State: %s",
		pr.Number, pr.Title, pr.State)
	if pr.UpdatedAt != "" {
		responseText += fmt.Sprintf(", Updated: %s", pr.UpdatedAt)
	}
	responseText += "\n"

	if pr.Body != "" {
		responseText += fmt.Sprintf("Body: %s\n", pr.Body)
	}

	return responseText
}

// FormatCommentList creates a human-readable summary of comments
func FormatCommentList(comments []remote.Comment) string {
	if len(comments) == 0 {
		return "No comments found"
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "Found %d comments:\n", len(comments))
	for _, comment := range comments {
		fmt.Fprintf(&builder, "- Comment by %s on %s\n", comment.Author, comment.Created)
		if len(comment.Content) > 100 {
			fmt.Fprintf(&builder, "  %s...\n", comment.Content[:100])
		} else {
			fmt.Fprintf(&builder, "  %s\n", comment.Content)
		}
	}
	return builder.String()
}

// FormatCommentDetails creates detailed comment information
func FormatCommentDetails(comment *remote.Comment) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Comment by %s on %s\n", comment.Author, comment.Created)
	fmt.Fprintf(&builder, "Body:\n%s\n", comment.Content)
	return builder.String()
}

// FormatCommentCreateSuccess creates success message for comment creation
func FormatCommentCreateSuccess(comment *remote.Comment) string {
	return fmt.Sprintf("Comment created successfully by %s", comment.Author)
}

// FormatCommentEditSuccess creates success message for comment editing
func FormatCommentEditSuccess(comment *remote.Comment) string {
	return fmt.Sprintf("Comment edited successfully by %s", comment.Author)
}
