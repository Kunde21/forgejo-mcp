package server

import (
	"fmt"
	"slices"
	"strings"

	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
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

// ValidateRepositoryExistence checks if a repository exists via Gitea API
func ValidateRepositoryExistence(client giteasdk.GiteaClientInterface, repoParam string) (bool, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return false, err
	}
	owner, repo, _ := strings.Cut(repoParam, "/")
	if _, _, err = client.GetRepo(owner, repo); err != nil {
		return false, fmt.Errorf("failed to validate repository existence: %w", err)
	}
	return true, nil
}

// ValidateRepositoryAccess checks if the user has access to the repository
func ValidateRepositoryAccess(client giteasdk.GiteaClientInterface, repoParam string) (bool, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return false, err
	}
	owner, repo, _ := strings.Cut(repoParam, "/")
	_, _, err = client.GetRepo(owner, repo)
	if err != nil {
		return false, fmt.Errorf("failed to validate repository access: %w", err)
	}
	return true, nil
}

// ValidatePRListArgs validates arguments for PR list requests
func ValidatePRListArgs(args PRListArgs) error {
	if args.Repository == "" && args.CWD == "" {
		return fmt.Errorf("repository parameter or cwd parameter is required")
	}

	if args.Repository != "" {
		if valid, err := ValidateRepositoryFormat(args.Repository); !valid {
			return err
		}
	}

	// Validate state parameter
	if args.State != "" {
		validStates := []string{"open", "closed", "all"}
		if !slices.Contains(validStates, args.State) {
			return fmt.Errorf("invalid state parameter: must be one of %v", validStates)
		}
	}

	return nil
}

// ValidateIssueListArgs validates arguments for issue list requests
func ValidateIssueListArgs(args IssueListArgs) error {
	if args.Repository == "" && args.CWD == "" {
		return fmt.Errorf("repository parameter or cwd parameter is required")
	}

	if args.Repository != "" {
		if valid, err := ValidateRepositoryFormat(args.Repository); !valid {
			return err
		}
	}

	// Validate state parameter
	if args.State != "" {
		validStates := []string{"open", "closed", "all"}
		found := false
		for _, state := range validStates {
			if args.State == state {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid state parameter: must be one of %v", validStates)
		}
	}

	return nil
}

// ValidateRepositoryListArgs validates arguments for repository list requests
func ValidateRepositoryListArgs(args RepoListArgs) error {
	// Repository list args are simple, just validate limit if provided
	if args.Limit < 0 {
		return fmt.Errorf("limit parameter must be non-negative")
	}
	return nil
}
