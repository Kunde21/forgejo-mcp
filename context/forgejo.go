package context

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// IsForgejoRemote checks if the given URL points to a Forgejo instance
// Supports both known Forgejo hosts and custom instances
func IsForgejoRemote(urlStr string) bool {
	// Remove protocol and get host
	host := extractHost(urlStr)
	if host == "" {
		return false
	}

	// Check for known Forgejo instances
	knownHosts := []string{
		"codeberg.org",
		"forgejo.org",
		"git.sr.ht", // SourceHut also uses Forgejo
	}

	for _, knownHost := range knownHosts {
		if host == knownHost || strings.HasSuffix(host, "."+knownHost) {
			return true
		}
	}

	// Check for known non-Forgejo hosts that should be rejected
	nonForgejoHosts := []string{
		"github.com",
		"gitlab.com",
		"bitbucket.org",
	}

	for _, nonForgejoHost := range nonForgejoHosts {
		if host == nonForgejoHost || strings.HasSuffix(host, "."+nonForgejoHost) {
			return false
		}
	}

	// For custom instances, accept any other valid host
	return isValidHost(host)
}

// ParseRepository extracts owner and repository name from a git URL
// Supports both SSH and HTTPS formats
func ParseRepository(urlStr string) (owner, repo string, err error) {
	// Handle SSH format: git@host:user/repo.git
	if strings.Contains(urlStr, "@") && !strings.HasPrefix(urlStr, "http") {
		return parseSSHURL(urlStr)
	}

	// Handle HTTPS format: https://host/user/repo.git
	if strings.HasPrefix(urlStr, "http") {
		return parseHTTPSURL(urlStr)
	}

	return "", "", fmt.Errorf("unsupported URL format: %s", urlStr)
}

// extractHost extracts the host from a URL string
func extractHost(urlStr string) string {
	// Handle SSH format with ssh:// protocol
	if strings.HasPrefix(urlStr, "ssh://") {
		if parsed, err := url.Parse(urlStr); err == nil {
			return parsed.Host
		}
	}

	// Handle SSH format git@host:path
	if strings.Contains(urlStr, "@") && strings.Contains(urlStr, ":") && !strings.HasPrefix(urlStr, "ssh://") {
		parts := strings.Split(urlStr, "@")
		if len(parts) == 2 {
			hostPart := strings.Split(parts[1], ":")[0]
			return hostPart
		}
	}

	// Handle HTTPS format
	if strings.HasPrefix(urlStr, "http") {
		if parsed, err := url.Parse(urlStr); err == nil {
			return parsed.Host
		}
	}

	return ""
}

// parseSSHURL parses SSH git URLs (git@host:user/repo.git)
func parseSSHURL(urlStr string) (owner, repo string, err error) {
	// Format: git@host:user/repo.git or ssh://git@host/user/repo.git
	var path string

	if strings.HasPrefix(urlStr, "ssh://") {
		// ssh://git@host/user/repo.git
		if parsed, err := url.Parse(urlStr); err == nil {
			path = strings.TrimPrefix(parsed.Path, "/")
		} else {
			return "", "", fmt.Errorf("invalid SSH URL format: %s", urlStr)
		}
	} else {
		// git@host:user/repo.git
		parts := strings.Split(urlStr, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL format: %s", urlStr)
		}
		path = parts[1]
	}

	// Extract owner and repo from path
	return extractOwnerRepo(path)
}

// parseHTTPSURL parses HTTPS git URLs (https://host/user/repo.git)
func parseHTTPSURL(urlStr string) (owner, repo string, err error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid HTTPS URL: %w", err)
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	return extractOwnerRepo(path)
}

// extractOwnerRepo extracts owner and repository from a path like "user/repo.git"
func extractOwnerRepo(path string) (owner, repo string, err error) {
	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository path format: %s", path)
	}

	owner = parts[0]
	repo = parts[1]

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("empty owner or repository name in path: %s", path)
	}

	return owner, repo, nil
}

// isValidHost checks if a host string is valid
func isValidHost(host string) bool {
	if host == "" {
		return false
	}

	// Basic validation: contains at least one dot or is localhost
	if strings.Contains(host, ".") || host == "localhost" {
		return true
	}

	// Allow IP addresses
	ipRegex := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)
	return ipRegex.MatchString(host)
}
