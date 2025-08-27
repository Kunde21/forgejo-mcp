// Package client provides a Gitea SDK client for Forgejo repositories
package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"code.gitea.io/sdk/gitea"
)

// RepositoryFilters holds filter parameters for repository operations
type RepositoryFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// Search
	Query string `json:"query,omitempty"`

	// Ownership and access
	OwnerID       int64  `json:"owner_id,omitempty"`
	StarredByUser int64  `json:"starred_by_user,omitempty"`
	Type          string `json:"type,omitempty"`

	// Visibility and status
	IsPrivate  *bool `json:"is_private,omitempty"`
	IsArchived *bool `json:"is_archived,omitempty"`

	// Sorting
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`

	// Additional filters
	ExcludeTemplate bool `json:"exclude_template,omitempty"`
}

// IssueFilters holds filter parameters for issue operations
type IssueFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// State and type filters
	State StateType `json:"state,omitempty"`
	Type  string    `json:"type,omitempty"`

	// Content filters
	Labels     []string `json:"labels,omitempty"`
	Milestones []string `json:"milestones,omitempty"`
	KeyWord    string   `json:"keyword,omitempty"`

	// Time filters
	Since  *time.Time `json:"since,omitempty"`
	Before *time.Time `json:"before,omitempty"`

	// User filters
	CreatedBy   string `json:"created_by,omitempty"`
	AssignedBy  string `json:"assigned_by,omitempty"`
	MentionedBy string `json:"mentioned_by,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Team        string `json:"team,omitempty"`
}

// PullRequestFilters holds filter parameters for pull request operations
type PullRequestFilters struct {
	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// State filter
	State StateType `json:"state,omitempty"`

	// Sorting
	Sort string `json:"sort,omitempty"`

	// Milestone filter
	Milestone int64 `json:"milestone,omitempty"`
}

// Client defines the interface for interacting with Forgejo repositories
type RepositoryLister interface {
	ListRepositories(filters *RepositoryFilters) ([]Repository, error)
	GetRepository(owner, name string) (*Repository, error)
}

// Client defines the interface for interacting with Forgejo repositories
type Client interface {
	// ListPRs retrieves pull requests for a repository with optional filters
	ListPRs(owner, repo string, filters *PullRequestFilters) ([]PullRequest, error)

	// ListIssues retrieves issues for a repository with optional filters
	ListIssues(owner, repo string, filters *IssueFilters) ([]Issue, error)

	// Repository operations
	RepositoryLister
}

// ClientConfig holds configuration options for the ForgejoClient
type ClientConfig struct {
	Timeout   time.Duration
	UserAgent string
}

// ForgejoClient implements the Client interface using the Gitea SDK
type ForgejoClient struct {
	baseURL     *url.URL
	token       string
	timeout     time.Duration
	userAgent   string
	giteaClient *gitea.Client
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
func (c *ForgejoClient) ListPRs(owner, repo string, filters *PullRequestFilters) ([]PullRequest, error) {
	if c.giteaClient == nil {
		return nil, fmt.Errorf("Gitea client not initialized")
	}

	opts := &gitea.ListPullRequestsOptions{
		State:     gitea.StateType(filters.State),
		Sort:      filters.Sort,
		Milestone: filters.Milestone,
	}

	if filters != nil {
		if filters.Page > 0 {
			opts.Page = filters.Page
		}
		if filters.PageSize > 0 {
			opts.PageSize = filters.PageSize
		}
	}

	giteaPRs, _, err := c.giteaClient.ListRepoPullRequests(owner, repo, *opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	prs := make([]PullRequest, len(giteaPRs))
	for i, giteaPR := range giteaPRs {
		prs[i] = transformPullRequest(giteaPR)
	}

	return prs, nil
}

// ListIssues retrieves issues for a repository with optional filters
func (c *ForgejoClient) ListIssues(owner, repo string, filters *IssueFilters) ([]Issue, error) {
	if c.giteaClient == nil {
		return nil, fmt.Errorf("Gitea client not initialized")
	}

	opts := &gitea.ListIssueOption{
		State:       gitea.StateType(filters.State),
		KeyWord:     filters.KeyWord,
		CreatedBy:   filters.CreatedBy,
		AssignedBy:  filters.AssignedBy,
		MentionedBy: filters.MentionedBy,
		Owner:       filters.Owner,
		Team:        filters.Team,
	}

	if filters != nil {
		if filters.Page > 0 {
			opts.Page = filters.Page
		}
		if filters.PageSize > 0 {
			opts.PageSize = filters.PageSize
		}
		if len(filters.Labels) > 0 {
			opts.Labels = filters.Labels
		}
		if len(filters.Milestones) > 0 {
			opts.Milestones = filters.Milestones
		}
		if filters.Since != nil {
			opts.Since = *filters.Since
		}
		if filters.Before != nil {
			opts.Before = *filters.Before
		}
		if filters.Type != "" {
			opts.Type = gitea.IssueType(filters.Type)
		}
	}

	giteaIssues, _, err := c.giteaClient.ListRepoIssues(owner, repo, *opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	issues := make([]Issue, len(giteaIssues))
	for i, giteaIssue := range giteaIssues {
		issues[i] = transformIssue(giteaIssue)
	}

	return issues, nil
}

// ListRepositories retrieves repositories with optional filters
func (c *ForgejoClient) ListRepositories(filters *RepositoryFilters) ([]Repository, error) {
	if c.giteaClient == nil {
		return nil, fmt.Errorf("Gitea client not initialized")
	}

	page := 1
	pageSize := 30

	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.PageSize > 0 {
			pageSize = filters.PageSize
		}

		if filters.Query != "" {
			searchOpts := &gitea.SearchRepoOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: pageSize,
				},
				Keyword: filters.Query,
			}

			if filters.OwnerID > 0 {
				searchOpts.OwnerID = filters.OwnerID
			}
			if filters.StarredByUser > 0 {
				searchOpts.StarredByUserID = filters.StarredByUser
			}
			if filters.IsPrivate != nil {
				searchOpts.IsPrivate = filters.IsPrivate
			}
			if filters.IsArchived != nil {
				searchOpts.IsArchived = filters.IsArchived
			}
			if filters.Type != "" {
				searchOpts.Type = gitea.RepoType(filters.Type)
			}
			if filters.Sort != "" {
				searchOpts.Sort = filters.Sort
			}
			if filters.Order != "" {
				searchOpts.Order = filters.Order
			}
			searchOpts.ExcludeTemplate = filters.ExcludeTemplate

			giteaRepos, _, err := c.giteaClient.SearchRepos(*searchOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to search repositories: %w", err)
			}

			repos := make([]Repository, len(giteaRepos))
			for i, giteaRepo := range giteaRepos {
				repos[i] = transformRepository(giteaRepo)
			}

			return repos, nil
		}
	}

	opts := &gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}

	giteaRepos, _, err := c.giteaClient.ListMyRepos(*opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	repos := make([]Repository, len(giteaRepos))
	for i, giteaRepo := range giteaRepos {
		repos[i] = transformRepository(giteaRepo)
	}

	return repos, nil
}

// GetRepository retrieves a specific repository by owner and name
func (c *ForgejoClient) GetRepository(owner, name string) (*Repository, error) {
	if c.giteaClient == nil {
		return nil, fmt.Errorf("Gitea client not initialized")
	}

	giteaRepo, _, err := c.giteaClient.GetRepo(owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	repo := transformRepository(giteaRepo)
	return &repo, nil
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

	// Create Gitea client
	giteaClient, err := gitea.NewClient(baseURL,
		gitea.SetToken(token),
		gitea.SetUserAgent(config.UserAgent),
		gitea.SetHTTPClient(&http.Client{Timeout: config.Timeout}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &ForgejoClient{
		baseURL:     parsedURL,
		token:       token,
		timeout:     config.Timeout,
		userAgent:   config.UserAgent,
		giteaClient: giteaClient,
	}, nil
}

// GetBaseURL returns the client's base URL
func (c *ForgejoClient) GetBaseURL() string {
	return c.baseURL.String()
}

func (c *ForgejoClient) GetTimeout() time.Duration {
	return c.timeout
}

func (c *ForgejoClient) GetUserAgent() string {
	return c.userAgent
}

func transformRepository(giteaRepo *gitea.Repository) Repository {
	if giteaRepo == nil {
		return Repository{}
	}

	return Repository{
		ID:          giteaRepo.ID,
		Name:        giteaRepo.Name,
		FullName:    giteaRepo.FullName,
		Description: giteaRepo.Description,
		HTMLURL:     giteaRepo.HTMLURL,
		CloneURL:    giteaRepo.CloneURL,
	}
}

func transformPullRequest(giteaPR *gitea.PullRequest) PullRequest {
	if giteaPR == nil {
		return PullRequest{}
	}

	pr := PullRequest{
		ID:        giteaPR.ID,
		URL:       giteaPR.URL,
		Index:     giteaPR.Index,
		Title:     giteaPR.Title,
		Body:      giteaPR.Body,
		State:     StateType(giteaPR.State),
		IsLocked:  giteaPR.IsLocked,
		Comments:  giteaPR.Comments,
		HTMLURL:   giteaPR.HTMLURL,
		DiffURL:   giteaPR.DiffURL,
		PatchURL:  giteaPR.PatchURL,
		Mergeable: giteaPR.Mergeable,
		HasMerged: giteaPR.HasMerged,
		Created:   giteaPR.Created,
		Updated:   giteaPR.Updated,
		Closed:    giteaPR.Closed,
	}

	if giteaPR.Poster != nil {
		pr.Poster = &User{
			ID:        giteaPR.Poster.ID,
			UserName:  giteaPR.Poster.UserName,
			FullName:  giteaPR.Poster.FullName,
			Email:     giteaPR.Poster.Email,
			AvatarURL: giteaPR.Poster.AvatarURL,
		}
	}

	if giteaPR.Labels != nil {
		pr.Labels = make([]*Label, len(giteaPR.Labels))
		for i, label := range giteaPR.Labels {
			pr.Labels[i] = &Label{
				ID:    label.ID,
				Name:  label.Name,
				Color: label.Color,
			}
		}
	}

	if giteaPR.Base != nil {
		repo := transformRepository(giteaPR.Base.Repository)
		pr.Base = &PRBranchInfo{
			Ref:    giteaPR.Base.Ref,
			SHA:    giteaPR.Base.Sha,
			RepoID: giteaPR.Base.RepoID,
			Repo:   &repo,
		}
	}

	if giteaPR.Head != nil {
		repo := transformRepository(giteaPR.Head.Repository)
		pr.Head = &PRBranchInfo{
			Ref:    giteaPR.Head.Ref,
			SHA:    giteaPR.Head.Sha,
			RepoID: giteaPR.Head.RepoID,
			Repo:   &repo,
		}
	}

	if giteaPR.Merged != nil {
		pr.Merged = giteaPR.Merged
	}

	if giteaPR.MergedCommitID != nil {
		pr.MergedCommitID = giteaPR.MergedCommitID
	}

	if giteaPR.MergedBy != nil {
		pr.MergedBy = &User{
			ID:        giteaPR.MergedBy.ID,
			UserName:  giteaPR.MergedBy.UserName,
			FullName:  giteaPR.MergedBy.FullName,
			Email:     giteaPR.MergedBy.Email,
			AvatarURL: giteaPR.MergedBy.AvatarURL,
		}
	}

	return pr
}

func transformIssue(giteaIssue *gitea.Issue) Issue {
	if giteaIssue == nil {
		return Issue{}
	}

	issue := Issue{
		ID:               giteaIssue.ID,
		URL:              giteaIssue.URL,
		HTMLURL:          giteaIssue.HTMLURL,
		Index:            giteaIssue.Index,
		OriginalAuthor:   giteaIssue.OriginalAuthor,
		OriginalAuthorID: giteaIssue.OriginalAuthorID,
		Title:            giteaIssue.Title,
		Body:             giteaIssue.Body,
		Ref:              giteaIssue.Ref,
		State:            StateType(giteaIssue.State),
		IsLocked:         giteaIssue.IsLocked,
		Comments:         giteaIssue.Comments,
		Created:          giteaIssue.Created,
		Updated:          giteaIssue.Updated,
	}

	if giteaIssue.Poster != nil {
		issue.Poster = &User{
			ID:        giteaIssue.Poster.ID,
			UserName:  giteaIssue.Poster.UserName,
			FullName:  giteaIssue.Poster.FullName,
			Email:     giteaIssue.Poster.Email,
			AvatarURL: giteaIssue.Poster.AvatarURL,
		}
	}

	if giteaIssue.Labels != nil {
		issue.Labels = make([]*Label, len(giteaIssue.Labels))
		for i, label := range giteaIssue.Labels {
			issue.Labels[i] = &Label{
				ID:    label.ID,
				Name:  label.Name,
				Color: label.Color,
			}
		}
	}

	if giteaIssue.Closed != nil {
		issue.Closed = giteaIssue.Closed
	}

	return issue
}
