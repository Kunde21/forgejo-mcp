package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RepositoryError represents a base error type for repository resolution operations
type RepositoryError struct {
	Op   string // Operation that failed
	Path string // Path involved in the error
	Err  error  // Underlying error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("repository %s failed for %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("repository %s failed for %s", e.Op, e.Path)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// DirectoryNotFoundError indicates that a directory does not exist
type DirectoryNotFoundError struct {
	RepositoryError
}

func NewDirectoryNotFoundError(path string) error {
	return &DirectoryNotFoundError{
		RepositoryError: RepositoryError{
			Op:   "validate",
			Path: path,
			Err:  fmt.Errorf("directory does not exist"),
		},
	}
}

func (e *DirectoryNotFoundError) Is(target error) bool {
	_, ok := target.(*DirectoryNotFoundError)
	return ok
}

// NotGitRepositoryError indicates that a directory is not a git repository
type NotGitRepositoryError struct {
	RepositoryError
	Reason string
}

func NewNotGitRepositoryError(path, reason string) error {
	return &NotGitRepositoryError{
		RepositoryError: RepositoryError{
			Op:   "validate",
			Path: path,
		},
		Reason: reason,
	}
}

func (e *NotGitRepositoryError) Error() string {
	return fmt.Sprintf("not a git repository: %s (%s)", e.Path, e.Reason)
}

func (e *NotGitRepositoryError) Is(target error) bool {
	_, ok := target.(*NotGitRepositoryError)
	return ok
}

// NoRemotesConfiguredError indicates that no git remotes are configured
type NoRemotesConfiguredError struct {
	RepositoryError
}

func NewNoRemotesConfiguredError(path string) error {
	return &NoRemotesConfiguredError{
		RepositoryError: RepositoryError{
			Op:   "extract",
			Path: path,
			Err:  fmt.Errorf("no configured remotes"),
		},
	}
}

func (e *NoRemotesConfiguredError) Is(target error) bool {
	_, ok := target.(*NoRemotesConfiguredError)
	return ok
}

// InvalidRemoteURLError indicates that a remote URL cannot be parsed
type InvalidRemoteURLError struct {
	RepositoryError
	URL string
}

func NewInvalidRemoteURLError(url string) error {
	return &InvalidRemoteURLError{
		RepositoryError: RepositoryError{
			Op:   "parse",
			Path: url,
		},
		URL: url,
	}
}

func (e *InvalidRemoteURLError) Error() string {
	return fmt.Sprintf("failed to parse remote URL: %s", e.URL)
}

func (e *InvalidRemoteURLError) Is(target error) bool {
	_, ok := target.(*InvalidRemoteURLError)
	return ok
}

// RepositoryResolution represents the result of resolving a directory to repository information
type RepositoryResolution struct {
	Directory  string `json:"directory,omitzero"`  // The original directory path
	Repository string `json:"repository,omitzero"` // The resolved repository in "owner/repo" format
	RemoteURL  string `json:"remote_url"`          // The full remote URL
	RemoteName string `json:"remote_name"`         // The name of the remote (e.g., "origin", "upstream")
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
		return NewDirectoryNotFoundError(directory)
	}

	// Check if .git directory exists
	gitDir := filepath.Join(directory, ".git")
	gitInfo, err := os.Stat(gitDir)
	switch {
	case os.IsNotExist(err):
		return NewNotGitRepositoryError(directory, "no .git directory found")
	case err != nil:
		return &RepositoryError{
			Op:   "validate",
			Path: gitDir,
			Err:  fmt.Errorf("failed to access .git directory: %w", err),
		}
	case !gitInfo.IsDir():
		return NewNotGitRepositoryError(directory, ".git is not a directory")
	}
	return nil
}

// ExtractRemoteInfo extracts repository information from git remote configuration
func (r *RepositoryResolver) ExtractRemoteInfo(directory string) (string, error) {
	gitConfigPath := filepath.Join(directory, ".git", "config")

	// Read git config file
	file, err := os.Open(gitConfigPath)
	if err != nil {
		return "", &RepositoryError{
			Op:   "extract",
			Path: gitConfigPath,
			Err:  fmt.Errorf("failed to read git config: %w", err),
		}
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
		return "", &RepositoryError{
			Op:   "extract",
			Path: gitConfigPath,
			Err:  fmt.Errorf("failed to parse git config: %w", err),
		}
	}

	if remoteURL == "" {
		return "", NewNoRemotesConfiguredError(directory)
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

	return "", NewInvalidRemoteURLError(remoteURL)
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
