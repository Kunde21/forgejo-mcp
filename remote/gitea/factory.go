package gitea

import (
	"fmt"
	"net/url"

	"code.gitea.io/sdk/gitea"
)

// ClientConfig holds configuration for creating Gitea clients
type ClientConfig struct {
	BaseURL string
	Token   string
}

// Validate validates the client configuration
func (c *ClientConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("client config cannot be nil")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("BaseURL cannot be empty")
	}

	if c.Token == "" {
		return fmt.Errorf("Token cannot be empty")
	}

	// Validate URL format
	if _, err := url.Parse(c.BaseURL); err != nil {
		return fmt.Errorf("invalid BaseURL format: %w", err)
	}

	return nil
}

// giteaClientWrapper wraps the Gitea SDK client to implement GiteaClientInterface
type giteaClientWrapper struct {
	client *gitea.Client
}

// NewGiteaClient creates a new Gitea client that implements GiteaClientInterface
func NewGiteaClient(baseURL, token string) (GiteaClientInterface, error) {
	return NewGiteaClientFromConfig(&ClientConfig{BaseURL: baseURL, Token: token})
}

// NewGiteaClientFromConfig creates a new Gitea client from configuration
func NewGiteaClientFromConfig(config *ClientConfig) (GiteaClientInterface, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid client configuration: %w", err)
	}
	client, err := gitea.NewClient(config.BaseURL, gitea.SetToken(config.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea SDK client: %w", err)
	}
	return &giteaClientWrapper{client: client}, nil
}

// Implement GiteaClientInterface methods by delegating to the wrapped client

func (w *giteaClientWrapper) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	return w.client.ListMyRepos(opt)
}

func (w *giteaClientWrapper) GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
	return w.client.GetRepo(owner, repo)
}

func (w *giteaClientWrapper) CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	return w.client.CreateRepo(opt)
}

func (w *giteaClientWrapper) DeleteRepo(owner, repo string) (*gitea.Response, error) {
	return w.client.DeleteRepo(owner, repo)
}

func (w *giteaClientWrapper) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	return w.client.ListRepoPullRequests(owner, repo, opt)
}

func (w *giteaClientWrapper) GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error) {
	return w.client.GetPullRequest(owner, repo, index)
}

func (w *giteaClientWrapper) CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	return w.client.CreatePullRequest(owner, repo, opt)
}

func (w *giteaClientWrapper) EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	return w.client.EditPullRequest(owner, repo, index, opt)
}

func (w *giteaClientWrapper) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	return w.client.ListIssues(opt)
}

func (w *giteaClientWrapper) ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	return w.client.ListRepoIssues(owner, repo, opt)
}

func (w *giteaClientWrapper) GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error) {
	return w.client.GetIssue(owner, repo, index)
}

func (w *giteaClientWrapper) CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error) {
	return w.client.CreateIssue(owner, repo, opt)
}

func (w *giteaClientWrapper) EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error) {
	return w.client.EditIssue(owner, repo, index, opt)
}

func (w *giteaClientWrapper) GetMyUserInfo() (*gitea.User, *gitea.Response, error) {
	return w.client.GetMyUserInfo()
}

func (w *giteaClientWrapper) GetUserInfo(user string) (*gitea.User, *gitea.Response, error) {
	return w.client.GetUserInfo(user)
}

func (w *giteaClientWrapper) ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error) {
	return w.client.ListRepoBranches(owner, repo, opt)
}

func (w *giteaClientWrapper) GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error) {
	return w.client.GetRepoBranch(owner, repo, branch)
}

func (w *giteaClientWrapper) GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error) {
	return w.client.GetSingleCommit(owner, repo, sha)
}

func (w *giteaClientWrapper) ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error) {
	return w.client.ListRepoCommits(owner, repo, opt)
}
