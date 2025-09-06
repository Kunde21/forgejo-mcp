package gitea

import (
	"fmt"
	"strings"
)

// ValidateRepositoryFormat validates that a repository parameter follows the owner/repo format
func ValidateRepositoryFormat(repoParam string) (bool, error) {
	if repoParam == "" {
		return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
	}

	owner, repo, ok := strings.Cut(repoParam, "/")
	if !ok {
		return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
	}
	if owner == "" {
		return false, fmt.Errorf("invalid repository format: owner cannot be empty")
	}
	if repo == "" {
		return false, fmt.Errorf("invalid repository format: repository name cannot be empty")
	}

	// Basic validation - allow common special characters
	// Let the API handle more complex validation and security
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.<>@"
	for _, char := range owner + repo {
		if !strings.ContainsRune(validChars, char) {
			return false, fmt.Errorf("invalid repository format: expected 'owner/repo'")
		}
	}
	return true, nil
}
