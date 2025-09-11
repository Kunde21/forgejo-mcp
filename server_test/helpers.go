package servertest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getRepoKeyFromRequest extracts the repository key from path values using modern Go 1.22+ routing
func getRepoKeyFromRequest(r *http.Request) (string, error) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	if owner == "" || repo == "" {
		return "", http.ErrMissingFile
	}

	return owner + "/" + repo, nil
}

// validateRepository checks if repository exists and is accessible
func validateRepository(mock *MockGiteaServer, repoKey string) bool {
	if repoKey == "" {
		return false
	}

	// Check if repository key is in valid format (owner/repo)
	owner, repo, ok := strings.Cut(repoKey, "/")
	if !ok || owner == "" || repo == "" {
		return false
	}

	// Check if repository is marked as not found
	if mock.notFoundRepos[repoKey] {
		return false
	}

	// For backward compatibility, check if it's the special "nonexistent/repo" case
	if repoKey == "nonexistent/repo" {
		return false
	}

	// Check if repository has any data (issues, pull requests, or comments)
	if _, exists := mock.issues[repoKey]; exists {
		return true
	}
	if _, exists := mock.pullRequests[repoKey]; exists {
		return true
	}
	if _, exists := mock.comments[repoKey+"/comments"]; exists {
		return true
	}

	// If no data exists but it's not marked as not found, consider it valid
	// This allows for empty repositories
	return !mock.notFoundRepos[repoKey]
}

// parsePagination extracts limit and offset from query parameters
func parsePagination(r *http.Request) (limit, offset int) {
	// Set default values
	limit = 0 // 0 means no limit (return all items)
	offset = 0

	// Parse limit from query parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Parse offset from query parameters
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}

// validateAuthToken validates authentication token from request headers
func validateAuthToken(r *http.Request) bool {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")

	// Empty authorization header should be rejected
	if authHeader == "" {
		// Check if there's a token in query parameters
		queryToken := r.URL.Query().Get("token")
		if queryToken == "" {
			// No authentication at all - acceptable for backward compatibility
			return true
		}
		// There's a query token, validate it
		return queryToken == "mock-token" || queryToken == "test-token"
	}

	var headerToken string
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		headerToken = after
	} else if after, ok := strings.CutPrefix(authHeader, "token "); ok {
		headerToken = after
	} else {
		// Malformed authorization header (not Bearer or token format)
		return false
	}

	// Check token in query parameters
	queryToken := r.URL.Query().Get("token")

	// Reject invalid tokens
	if headerToken == "invalid-token" || queryToken == "invalid-token" {
		return false
	}

	// If there's a header token, it must be valid
	if headerToken != "mock-token" && headerToken != "test-token" {
		return false
	}

	// If there's a query token, it must be valid
	if queryToken != "" && queryToken != "mock-token" && queryToken != "test-token" {
		return false
	}

	return true
}

// writeJSONResponse writes standardized JSON responses
func writeJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
