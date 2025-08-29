package tea

import (
	"code.gitea.io/sdk/gitea"
)

// TransformRepositoryToMCP transforms a Gitea repository to MCP response format
func TransformRepositoryToMCP(repo *gitea.Repository) map[string]interface{} {
	result := map[string]interface{}{
		"type": "repository",
	}

	if repo == nil {
		return result
	}

	result["id"] = repo.ID
	result["name"] = repo.Name
	result["full_name"] = repo.FullName

	if repo.Description != "" {
		result["description"] = repo.Description
	}

	if repo.HTMLURL != "" {
		result["html_url"] = repo.HTMLURL
	}

	if repo.CloneURL != "" {
		result["clone_url"] = repo.CloneURL
	}

	if repo.SSHURL != "" {
		result["ssh_url"] = repo.SSHURL
	}

	if !repo.Created.IsZero() {
		result["created_at"] = repo.Created
	}

	if !repo.Updated.IsZero() {
		result["updated_at"] = repo.Updated
	}

	result["private"] = repo.Private
	result["archived"] = repo.Archived

	return result
}

// TransformIssueToMCP transforms a Gitea issue to MCP response format
func TransformIssueToMCP(issue *gitea.Issue) map[string]interface{} {
	result := map[string]interface{}{
		"type": "issue",
	}

	if issue == nil {
		return result
	}

	result["id"] = issue.ID
	result["number"] = issue.Index
	result["title"] = issue.Title

	if issue.Body != "" {
		result["body"] = issue.Body
	}

	result["state"] = string(issue.State)

	if issue.Poster != nil {
		result["author"] = issue.Poster.UserName
		result["author_name"] = issue.Poster.FullName
	}

	// Transform labels
	labels := make([]map[string]interface{}, len(issue.Labels))
	for i, label := range issue.Labels {
		if label != nil {
			labels[i] = map[string]interface{}{
				"id":    label.ID,
				"name":  label.Name,
				"color": label.Color,
			}
		}
	}
	result["labels"] = labels

	if !issue.Created.IsZero() {
		result["created_at"] = issue.Created
	}

	if !issue.Updated.IsZero() {
		result["updated_at"] = issue.Updated
	}

	if issue.Closed != nil && !issue.Closed.IsZero() {
		result["closed_at"] = issue.Closed
	}

	return result
}

// TransformPullRequestToMCP transforms a Gitea pull request to MCP response format
func TransformPullRequestToMCP(pr *gitea.PullRequest) map[string]interface{} {
	result := map[string]interface{}{
		"type": "pull_request",
	}

	if pr == nil {
		return result
	}

	result["id"] = pr.ID
	result["number"] = pr.Index
	result["title"] = pr.Title

	if pr.Body != "" {
		result["body"] = pr.Body
	}

	result["state"] = string(pr.State)

	if pr.Poster != nil {
		result["author"] = pr.Poster.UserName
		result["author_name"] = pr.Poster.FullName
	}

	// Transform labels
	labels := make([]map[string]interface{}, len(pr.Labels))
	for i, label := range pr.Labels {
		if label != nil {
			labels[i] = map[string]interface{}{
				"id":    label.ID,
				"name":  label.Name,
				"color": label.Color,
			}
		}
	}
	result["labels"] = labels

	if pr.Created != nil && !pr.Created.IsZero() {
		result["created_at"] = pr.Created
	}

	if pr.Updated != nil && !pr.Updated.IsZero() {
		result["updated_at"] = pr.Updated
	}

	if pr.Closed != nil && !pr.Closed.IsZero() {
		result["closed_at"] = pr.Closed
	}

	if pr.Merged != nil && !pr.Merged.IsZero() {
		result["merged_at"] = pr.Merged
	}

	result["has_merged"] = pr.HasMerged

	if pr.Base != nil {
		result["base_branch"] = pr.Base.Ref
		result["base_sha"] = pr.Base.Sha
	}

	if pr.Head != nil {
		result["head_branch"] = pr.Head.Ref
		result["head_sha"] = pr.Head.Sha
	}

	return result
}

// TransformRepositoriesToMCP transforms a slice of Gitea repositories to MCP response format
func TransformRepositoriesToMCP(repos []*gitea.Repository) []map[string]interface{} {
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		result[i] = TransformRepositoryToMCP(repo)
	}
	return result
}

// TransformIssuesToMCP transforms a slice of Gitea issues to MCP response format
func TransformIssuesToMCP(issues []*gitea.Issue) []map[string]interface{} {
	result := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		result[i] = TransformIssueToMCP(issue)
	}
	return result
}

// TransformPullRequestsToMCP transforms a slice of Gitea pull requests to MCP response format
func TransformPullRequestsToMCP(prs []*gitea.PullRequest) []map[string]interface{} {
	result := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		result[i] = TransformPullRequestToMCP(pr)
	}
	return result
}
