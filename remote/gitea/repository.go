package gitea

import (
	"fmt"
	"strings"
)

// extractRepositoryMetadata extracts and caches repository metadata
func extractRepositoryMetadata(client GiteaClientInterface, repoParam string) (map[string]any, error) {
	valid, err := ValidateRepositoryFormat(repoParam)
	if !valid {
		return nil, err
	}

	owner, repo, _ := strings.Cut(repoParam, "/")
	giteaRepo, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repository metadata: %w", err)
	}

	metadata := map[string]any{
		"id":          giteaRepo.ID,
		"name":        giteaRepo.Name,
		"fullName":    giteaRepo.FullName,
		"description": giteaRepo.Description,
		"private":     giteaRepo.Private,
		"fork":        giteaRepo.Fork,
		"archived":    giteaRepo.Archived,
		"stars":       giteaRepo.Stars,
		"forks":       giteaRepo.Forks,
		"size":        giteaRepo.Size,
		"url":         giteaRepo.HTMLURL,
		"sshUrl":      giteaRepo.SSHURL,
		"cloneUrl":    giteaRepo.CloneURL,
	}

	if giteaRepo.Owner != nil {
		metadata["owner"] = map[string]any{
			"id":       giteaRepo.Owner.ID,
			"username": giteaRepo.Owner.UserName,
			"fullName": giteaRepo.Owner.FullName,
			"email":    giteaRepo.Owner.Email,
		}
	}

	return metadata, nil
}
