package tea

import (
	"code.gitea.io/sdk/gitea"
)

// ExtractRepositoryMetadata extracts metadata from a Gitea repository
func ExtractRepositoryMetadata(repo *gitea.Repository) map[string]interface{} {
	result := make(map[string]interface{})

	if repo == nil {
		return result
	}

	result["stars_count"] = repo.Stars
	result["forks_count"] = repo.Forks
	result["open_issues"] = repo.OpenIssues
	result["open_pulls"] = repo.OpenPulls
	result["watchers_count"] = repo.Watchers
	result["size"] = repo.Size

	if !repo.Created.IsZero() {
		result["created_at"] = repo.Created
	}

	if !repo.Updated.IsZero() {
		result["updated_at"] = repo.Updated
	}

	result["private"] = repo.Private
	result["archived"] = repo.Archived
	result["has_issues"] = repo.HasIssues
	result["has_wiki"] = repo.HasWiki
	result["has_pull_requests"] = repo.HasPullRequests

	return result
}

// ExtractIssueMetadata extracts metadata from a Gitea issue
func ExtractIssueMetadata(issue *gitea.Issue) map[string]interface{} {
	result := make(map[string]interface{})

	if issue == nil {
		return result
	}

	result["id"] = issue.ID
	result["number"] = issue.Index
	result["state"] = string(issue.State)

	if issue.Poster != nil {
		result["author"] = issue.Poster.UserName
		result["author_name"] = issue.Poster.FullName
	}

	if issue.Milestone != nil {
		result["milestone"] = issue.Milestone.Title
	}

	// Extract assignees
	assignees := make([]string, len(issue.Assignees))
	for i, assignee := range issue.Assignees {
		if assignee != nil {
			assignees[i] = assignee.UserName
		}
	}
	result["assignees"] = assignees

	result["comments_count"] = issue.Comments

	if !issue.Created.IsZero() {
		result["created_at"] = issue.Created
	}

	if !issue.Updated.IsZero() {
		result["updated_at"] = issue.Updated
	}

	if issue.Closed != nil && !issue.Closed.IsZero() {
		result["closed_at"] = issue.Closed
	}

	if issue.OriginalAuthor != "" {
		result["original_author"] = issue.OriginalAuthor
	}

	if issue.OriginalAuthorID != 0 {
		result["original_author_id"] = issue.OriginalAuthorID
	}

	return result
}

// ExtractPullRequestMetadata extracts metadata from a Gitea pull request
func ExtractPullRequestMetadata(pr *gitea.PullRequest) map[string]interface{} {
	result := make(map[string]interface{})

	if pr == nil {
		return result
	}

	result["id"] = pr.ID
	result["number"] = pr.Index
	result["state"] = string(pr.State)

	if pr.Poster != nil {
		result["author"] = pr.Poster.UserName
		result["author_name"] = pr.Poster.FullName
	}

	if pr.Milestone != nil {
		result["milestone"] = pr.Milestone.Title
	}

	// Extract assignees
	assignees := make([]string, len(pr.Assignees))
	for i, assignee := range pr.Assignees {
		if assignee != nil {
			assignees[i] = assignee.UserName
		}
	}
	result["assignees"] = assignees

	result["comments_count"] = pr.Comments

	if pr.Merged != nil && !pr.Merged.IsZero() {
		result["merged_at"] = pr.Merged
	}

	result["has_merged"] = pr.HasMerged

	if pr.MergedBy != nil {
		result["merged_by"] = pr.MergedBy.UserName
		result["merged_by_name"] = pr.MergedBy.FullName
	}

	if pr.Created != nil && !pr.Created.IsZero() {
		result["created_at"] = pr.Created
	}

	if pr.Updated != nil && !pr.Updated.IsZero() {
		result["updated_at"] = pr.Updated
	}

	if pr.Closed != nil && !pr.Closed.IsZero() {
		result["closed_at"] = pr.Closed
	}

	result["allow_maintainer_edit"] = pr.AllowMaintainerEdit

	return result
}

// BuildLabelsContext builds context information from labels
func BuildLabelsContext(labels []*gitea.Label) []map[string]interface{} {
	if len(labels) == 0 {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(labels))
	for i, label := range labels {
		if label != nil {
			result[i] = map[string]interface{}{
				"name":  label.Name,
				"color": label.Color,
			}
		}
	}

	return result
}

// BuildMilestoneContext builds context information from a milestone
func BuildMilestoneContext(milestone *gitea.Milestone) map[string]interface{} {
	if milestone == nil {
		return nil
	}

	return map[string]interface{}{
		"id":          milestone.ID,
		"title":       milestone.Title,
		"description": milestone.Description,
	}
}

// MapIssueRelationships maps relationships for an issue
func MapIssueRelationships(issue *gitea.Issue) map[string]interface{} {
	result := make(map[string]interface{})

	if issue == nil {
		return result
	}

	if issue.Repository != nil {
		result["repository_id"] = issue.Repository.ID
		result["repository_name"] = issue.Repository.Name
	}

	return result
}

// MapPullRequestRelationships maps relationships for a pull request
func MapPullRequestRelationships(pr *gitea.PullRequest) map[string]interface{} {
	result := make(map[string]interface{})

	if pr == nil {
		return result
	}

	// The PullRequest field indicates if this issue is actually a PR
	// This field is only available on Issue struct, not PullRequest struct
	// For PullRequest relationships, we don't have direct linking in the SDK

	return result
}
