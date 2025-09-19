package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RepositoryResolution represents the result of resolving a directory to repository information
type RepositoryResolution struct {
	Directory  string `json:"directory"`   // The original directory path
	Repository string `json:"repository"`  // The resolved repository in "owner/repo" format
	RemoteURL  string `json:"remote_url"`  // The full remote URL
	RemoteName string `json:"remote_name"` // The name of the remote (e.g., "origin", "upstream")
}

// RepositoryResolver handles directory-to-repository resolution
type RepositoryResolver struct {
	// Future: Add caching, timeout configuration, etc.
}

// NewRepositoryResolver creates a new RepositoryResolver instance
func NewRepositoryResolver() *RepositoryResolver {
	return &RepositoryResolver{}
}

// ValidateDirectory validates that the directory exists and is a git repository
func (r *RepositoryResolver) ValidateDirectory(directory string) error {
	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}

	// Check if .git directory exists
	gitDir := filepath.Join(directory, ".git")
	if gitInfo, err := os.Stat(gitDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", directory)
		}
		return fmt.Errorf("failed to access .git directory: %w", err)
	} else if !gitInfo.IsDir() {
		return fmt.Errorf("not a git repository: .git is not a directory in %s", directory)
	}

	return nil
}

// ExtractRemoteInfo extracts repository information from git remote configuration
func (r *RepositoryResolver) ExtractRemoteInfo(directory string) (string, error) {
	gitConfigPath := filepath.Join(directory, ".git", "config")

	// Read git config file
	file, err := os.Open(gitConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read git config: %w", err)
	}
	defer file.Close()

	var remoteURL string

	scanner := bufio.NewScanner(file)
	inRemoteSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for remote section start
		if strings.HasPrefix(line, "[remote \"") {
			inRemoteSection = true
			continue
		}

		// Check for section end
		if line == "]" {
			inRemoteSection = false
			continue
		}

		// Look for URL in remote section
		if inRemoteSection && strings.HasPrefix(line, "url = ") {
			remoteURL = strings.TrimPrefix(line, "url = ")
			break // Use the first remote we find
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to parse git config: %w", err)
	}

	if remoteURL == "" {
		return "", fmt.Errorf("no configured remotes found in %s", directory)
	}

	// Parse the remote URL to extract owner/repo
	return r.parseRemoteURL(remoteURL)
}

// parseRemoteURL parses various git remote URL formats to extract owner/repo
func (r *RepositoryResolver) parseRemoteURL(remoteURL string) (string, error) {
	// HTTPS URL: https://forgejo.example.com/owner/repo.git
	httpsPattern := regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := httpsPattern.FindStringSubmatch(remoteURL); matches != nil {
		return fmt.Sprintf("%s/%s", matches[1], matches[2]), nil
	}

	// SSH URL: git@forgejo.example.com:owner/repo.git
	sshPattern := regexp.MustCompile(`^git@[^:]+:([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(remoteURL); matches != nil {
		return fmt.Sprintf("%s/%s", matches[1], matches[2]), nil
	}

	// Git protocol: git://forgejo.example.com/owner/repo.git
	gitPattern := regexp.MustCompile(`^git://[^/]+/([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := gitPattern.FindStringSubmatch(remoteURL); matches != nil {
		return fmt.Sprintf("%s/%s", matches[1], matches[2]), nil
	}

	return "", fmt.Errorf("failed to parse remote URL: %s", remoteURL)
}

// ResolveRepository performs the complete directory to repository resolution
func (r *RepositoryResolver) ResolveRepository(directory string) (*RepositoryResolution, error) {
	// Validate directory
	if err := r.ValidateDirectory(directory); err != nil {
		return nil, err
	}

	// Extract remote information
	repo, err := r.ExtractRemoteInfo(directory)
	if err != nil {
		return nil, err
	}

	// Get the remote URL and name
	gitConfigPath := filepath.Join(directory, ".git", "config")
	file, err := os.Open(gitConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read git config: %w", err)
	}
	defer file.Close()

	var remoteURL string
	var remoteName string

	scanner := bufio.NewScanner(file)
	inRemoteSection := false
	currentRemote := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[remote \"") {
			inRemoteSection = true
			start := strings.Index(line, "\"") + 1
			end := strings.LastIndex(line, "\"")
			if start > 0 && end > start {
				currentRemote = line[start:end]
			}
			continue
		}

		if line == "]" {
			inRemoteSection = false
			continue
		}

		if inRemoteSection && strings.HasPrefix(line, "url = ") {
			url := strings.TrimPrefix(line, "url = ")
			// Check if this URL corresponds to the repo we extracted
			if parsedRepo, err := r.parseRemoteURL(url); err == nil && parsedRepo == repo {
				remoteURL = url
				remoteName = currentRemote
				break
			}
		}
	}

	return &RepositoryResolution{
		Directory:  directory,
		Repository: repo,
		RemoteURL:  remoteURL,
		RemoteName: remoteName,
	}, nil
}

// ValidateParameters validates mutual exclusivity between directory and repository parameters
func (r *RepositoryResolver) ValidateParameters(directory, repository string) error {
	directoryProvided := directory != ""
	repositoryProvided := repository != ""

	if !directoryProvided && !repositoryProvided {
		return fmt.Errorf("exactly one of directory or repository must be provided")
	}

	if directoryProvided && repositoryProvided {
		return fmt.Errorf("exactly one of directory or repository must be provided")
	}

	return nil
}
