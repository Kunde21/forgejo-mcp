// Package client provides a Gitea SDK client for Forgejo repositories
package client

import (
	"net/url"
	"time"
)

// Client defines the interface for interacting with Forgejo repositories
type Client interface {
	// ListPRs retrieves pull requests for a repository with optional filters
	ListPRs(owner, repo string, filters map[string]interface{}) ([]PullRequest, error)

	// ListIssues retrieves issues for a repository with optional filters
	ListIssues(owner, repo string, filters map[string]interface{}) ([]Issue, error)
}

// ClientConfig holds configuration options for the ForgejoClient
type ClientConfig struct {
	Timeout   time.Duration
	UserAgent string
}

// ForgejoClient implements the Client interface using the Gitea SDK
type ForgejoClient struct {
	baseURL   *url.URL
	token     string
	timeout   time.Duration
	userAgent string
}

// PullRequest represents a pull request from the Gitea API
type PullRequest struct {
	ID             int64         `json:"id"`
	URL            string        `json:"url"`
	Index          int64         `json:"number"`
	Poster         *User         `json:"user"`
	Title          string        `json:"title"`
	Body           string        `json:"body"`
	Labels         []*Label      `json:"labels"`
	State          StateType     `json:"state"`
	IsLocked       bool          `json:"is_locked"`
	Comments       int           `json:"comments"`
	HTMLURL        string        `json:"html_url"`
	DiffURL        string        `json:"diff_url"`
	PatchURL       string        `json:"patch_url"`
	Mergeable      bool          `json:"mergeable"`
	HasMerged      bool          `json:"merged"`
	Merged         *time.Time    `json:"merged_at"`
	MergedCommitID *string       `json:"merge_commit_sha"`
	MergedBy       *User         `json:"merged_by"`
	Base           *PRBranchInfo `json:"base"`
	Head           *PRBranchInfo `json:"head"`
	Created        *time.Time    `json:"created_at"`
	Updated        *time.Time    `json:"updated_at"`
	Closed         *time.Time    `json:"closed_at"`
}

// Issue represents an issue from the Gitea API
type Issue struct {
	ID               int64      `json:"id"`
	URL              string     `json:"url"`
	HTMLURL          string     `json:"html_url"`
	Index            int64      `json:"number"`
	Poster           *User      `json:"user"`
	OriginalAuthor   string     `json:"original_author"`
	OriginalAuthorID int64      `json:"original_author_id"`
	Title            string     `json:"title"`
	Body             string     `json:"body"`
	Ref              string     `json:"ref"`
	Labels           []*Label   `json:"labels"`
	State            StateType  `json:"state"`
	IsLocked         bool       `json:"is_locked"`
	Comments         int        `json:"comments"`
	Created          time.Time  `json:"created_at"`
	Updated          time.Time  `json:"updated_at"`
	Closed           *time.Time `json:"closed_at"`
}

// User represents a user from the Gitea API
type User struct {
	ID        int64  `json:"id"`
	UserName  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// Label represents a label from the Gitea API
type Label struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// StateType represents the state of an issue or pull request
type StateType string

const (
	StateOpen   StateType = "open"
	StateClosed StateType = "closed"
	StateAll    StateType = "all"
)

// PRBranchInfo represents branch information for a pull request
type PRBranchInfo struct {
	Ref    string      `json:"ref"`
	SHA    string      `json:"sha"`
	RepoID int64       `json:"repo_id"`
	Repo   *Repository `json:"repo"`
}

// Repository represents a repository from the Gitea API
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
}

// ListPRs retrieves pull requests for a repository with optional filters
func (c *ForgejoClient) ListPRs(owner, repo string, filters map[string]interface{}) ([]PullRequest, error) {
	// TODO: Implement using Gitea SDK
	// This is a placeholder implementation for now
	return []PullRequest{}, nil
}

// ListIssues retrieves issues for a repository with optional filters
func (c *ForgejoClient) ListIssues(owner, repo string, filters map[string]interface{}) ([]Issue, error) {
	// TODO: Implement using Gitea SDK
	// This is a placeholder implementation for now
	return []Issue{}, nil
}

// DefaultConfig returns a default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:   30 * time.Second,
		UserAgent: "forgejo-mcp-client/1.0.0",
	}
}

// New creates a new ForgejoClient with the given base URL and token using default configuration
func New(baseURL, token string) (*ForgejoClient, error) {
	return NewWithConfig(baseURL, token, DefaultConfig())
}

// NewWithConfig creates a new ForgejoClient with custom configuration
func NewWithConfig(baseURL, token string, config *ClientConfig) (*ForgejoClient, error) {
	if baseURL == "" {
		return nil, &ValidationError{
			Message: "baseURL cannot be empty",
			Field:   "baseURL",
		}
	}

	if token == "" {
		return nil, &ValidationError{
			Message: "token cannot be empty",
			Field:   "token",
		}
	}

	if config == nil {
		config = DefaultConfig()
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, &ValidationError{
			Message: "invalid baseURL format, must be a valid HTTP/HTTPS URL",
			Field:   "baseURL",
		}
	}

	return &ForgejoClient{
		baseURL:   parsedURL,
		token:     token,
		timeout:   config.Timeout,
		userAgent: config.UserAgent,
	}, nil
}

// GetBaseURL returns the client's base URL
func (c *ForgejoClient) GetBaseURL() string {
	return c.baseURL.String()
}

// GetTimeout returns the client's timeout
func (c *ForgejoClient) GetTimeout() time.Duration {
	return c.timeout
}

// GetUserAgent returns the client's user agent
func (c *ForgejoClient) GetUserAgent() string {
	return c.userAgent
}
