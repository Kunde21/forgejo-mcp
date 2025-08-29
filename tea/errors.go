package tea

import (
	"strconv"

	"code.gitea.io/sdk/gitea"
)

// PartialSuccessResult represents the result of an operation that may have partial success
type PartialSuccessResult struct {
	Issues       []map[string]interface{}
	PullRequests []map[string]interface{}
	HasErrors    bool
	Errors       []map[string]interface{}
}

// FormatAPIError formats an API error into MCP-compatible response format
func FormatAPIError(err error, response *gitea.Response) map[string]interface{} {
	result := make(map[string]interface{})

	if err == nil {
		result["type"] = "success"
		return result
	}

	result["error"] = err.Error()

	// Determine error type based on response status code
	if response != nil && response.Response != nil {
		statusCode := response.Response.StatusCode
		result["status_code"] = statusCode
		result["status"] = response.Response.Status

		switch {
		case statusCode == 401 || statusCode == 403:
			result["type"] = "auth_error"
		case statusCode == 429:
			result["type"] = "rate_limit_error"
		case statusCode >= 400 && statusCode < 500:
			result["type"] = "api_error"
		case statusCode >= 500:
			result["type"] = "server_error"
		default:
			result["type"] = "api_error"
		}
	} else {
		// Network or other errors
		result["type"] = "error"
	}

	return result
}

// HandlePartialSuccess handles scenarios where some operations succeed and others fail
func HandlePartialSuccess(issues []*gitea.Issue, prs []*gitea.PullRequest, issueErr error, prErr error) PartialSuccessResult {
	result := PartialSuccessResult{
		Issues:       []map[string]interface{}{},
		PullRequests: []map[string]interface{}{},
		HasErrors:    false,
		Errors:       []map[string]interface{}{},
	}

	// Handle successful issues
	if issues != nil {
		result.Issues = TransformIssuesToMCP(issues)
	}

	// Handle successful PRs
	if prs != nil {
		result.PullRequests = TransformPullRequestsToMCP(prs)
	}

	// Handle issue errors
	if issueErr != nil {
		result.HasErrors = true
		result.Errors = append(result.Errors, map[string]interface{}{
			"error": issueErr.Error(),
			"type":  "issue_fetch_error",
		})
	}

	// Handle PR errors
	if prErr != nil {
		result.HasErrors = true
		result.Errors = append(result.Errors, map[string]interface{}{
			"error": prErr.Error(),
			"type":  "pr_fetch_error",
		})
	}

	return result
}

// CheckRateLimit checks if the response indicates rate limiting
func CheckRateLimit(response *gitea.Response) bool {
	if response == nil || response.Response == nil {
		return false
	}

	// Check if status code indicates rate limiting
	if response.Response.StatusCode == 429 {
		return true
	}

	// Check rate limit headers
	remaining := response.Response.Header.Get("X-RateLimit-Remaining")
	if remaining != "" {
		if remainingVal, err := strconv.Atoi(remaining); err == nil {
			return remainingVal <= 0
		}
	}

	return false
}

// GetRateLimitInfo extracts rate limit information from the response
func GetRateLimitInfo(response *gitea.Response) map[string]interface{} {
	result := map[string]interface{}{
		"type":    "rate_limit_info",
		"limited": false,
	}

	if response == nil || response.Response == nil {
		return result
	}

	// Check if currently rate limited
	result["limited"] = CheckRateLimit(response)

	// Extract rate limit headers
	if limit := response.Response.Header.Get("X-RateLimit-Limit"); limit != "" {
		if limitVal, err := strconv.Atoi(limit); err == nil {
			result["limit"] = limitVal
		}
	}

	if remaining := response.Response.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if remainingVal, err := strconv.Atoi(remaining); err == nil {
			result["remaining"] = remainingVal
		}
	}

	if reset := response.Response.Header.Get("X-RateLimit-Reset"); reset != "" {
		result["reset"] = reset
	}

	return result
}
