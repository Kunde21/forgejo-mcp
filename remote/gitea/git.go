package gitea

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ResolveCWDToRepository attempts to resolve a CWD path to a repository identifier
// Uses git remote -v to query remote URLs and extract repository information
func ResolveCWDToRepository(cwd string) (string, error) {
	if cwd == "" {
		return "", fmt.Errorf("CWD is empty")
	}

	// Check if the directory is a git repository and get remote URLs
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		// If git command fails, fall back to path-based resolution
		return resolveCWDFromPath(cwd)
	}

	// Parse the git remote output to extract repository identifier
	repoID, err := parseGitRemoteOutput(string(output))
	if err != nil {
		// If parsing fails, fall back to path-based resolution
		return resolveCWDFromPath(cwd)
	}

	return repoID, nil
}

// parseGitRemoteOutput parses the output of 'git remote -v' to extract repository identifier
func parseGitRemoteOutput(output string) (string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no remote URLs found")
	}

	// Regex to match common git remote URL patterns
	// Supports: https://host/owner/repo, git@host:owner/repo, ssh://git@host/owner/repo
	re := regexp.MustCompile(`(?:https?://|git@|ssh://git@)([^:/]+)[:/]([^/]+)/([^/\s]+)`)

	for _, line := range lines {
		// Skip lines that don't contain fetch URLs (look for (fetch) at the end)
		if !strings.HasSuffix(strings.TrimSpace(line), "(fetch)") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			owner := matches[2]
			repo := strings.TrimSuffix(matches[3], ".git") // Remove .git extension if present

			repoParam := owner + "/" + repo
			if valid, _ := ValidateRepositoryFormat(repoParam); valid {
				return repoParam, nil
			}
		}
	}

	return "", fmt.Errorf("no valid repository identifier found in remote URLs")
}

// resolveCWDFromPath is the fallback method that uses path-based resolution
func resolveCWDFromPath(cwd string) (string, error) {
	// Remove any trailing slashes
	cwd = strings.TrimSuffix(cwd, "/")

	// Look for patterns like /owner/repo or owner/repo in the path
	parts := strings.Split(cwd, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("CWD does not contain a valid repository path")
	}

	// Find the last two parts that could be owner/repo
	for i := len(parts) - 1; i >= 1; i-- {
		potentialOwner := parts[i-1]
		potentialRepo := parts[i]

		// Skip empty parts
		if potentialOwner == "" || potentialRepo == "" {
			continue
		}

		// Basic validation - should not contain spaces or special chars that aren't allowed in repo names
		if strings.ContainsAny(potentialOwner, " \t\n\r") || strings.ContainsAny(potentialRepo, " \t\n\r") {
			continue
		}

		repoParam := potentialOwner + "/" + potentialRepo
		if valid, _ := ValidateRepositoryFormat(repoParam); valid {
			return repoParam, nil
		}
	}

	return "", fmt.Errorf("could not resolve repository from CWD: %s", cwd)
}
