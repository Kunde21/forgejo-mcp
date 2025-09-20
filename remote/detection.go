package remote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// VersionResponse represents the response from the /api/v1/version endpoint
type VersionResponse struct {
	Version string `json:"version"`
}

// DetectRemoteType detects whether a remote server is Forgejo or Gitea
// by calling the /api/v1/version endpoint and parsing the response
func DetectRemoteType(remoteURL, authToken string) (string, error) {
	if remoteURL == "" {
		return "", fmt.Errorf("remote URL cannot be empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build the version endpoint URL
	versionURL := strings.TrimSuffix(remoteURL, "/") + "/api/v1/version"

	// Create request
	req, err := http.NewRequest("GET", versionURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is provided
	if authToken != "" {
		req.Header.Set("Authorization", "token "+authToken)
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call version endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version endpoint returned status %d", resp.StatusCode)
	}

	// Parse the response
	var versionResp VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
		return "", fmt.Errorf("failed to parse version response: %w", err)
	}

	// Analyze the version string to determine the remote type
	return analyzeVersionString(versionResp.Version), nil
}

// analyzeVersionString analyzes a version string to determine if it's Forgejo or Gitea
func analyzeVersionString(version string) string {
	if version == "" {
		return "gitea" // Default to Gitea if version is empty
	}

	// Convert to lowercase for case-insensitive matching
	versionLower := strings.ToLower(version)

	// Check for explicit Forgejo indicators
	forgejoIndicators := []string{"forgejo", "12."}
	for _, indicator := range forgejoIndicators {
		if strings.Contains(versionLower, indicator) {
			return "forgejo"
		}
	}

	// Check for explicit Gitea indicators
	giteaIndicators := []string{"gitea", "1."}
	for _, indicator := range giteaIndicators {
		if strings.Contains(versionLower, indicator) {
			return "gitea"
		}
	}

	// Use regex to detect version patterns
	// Forgejo typically uses 12.x.x pattern
	forgejoPattern := regexp.MustCompile(`^12\.\d+\.\d+`)
	if forgejoPattern.MatchString(version) {
		return "forgejo"
	}

	// Gitea typically uses 1.x.x pattern
	giteaPattern := regexp.MustCompile(`^1\.\d+\.\d+`)
	if giteaPattern.MatchString(version) {
		return "gitea"
	}

	// Default to Gitea for ambiguous cases
	return "gitea"
}
