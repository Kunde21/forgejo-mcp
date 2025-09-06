package gitea

import "code.gitea.io/sdk/gitea"

// GiteaClientInterface defines the interface for Gitea client operations
type GiteaClientInterface interface {
	// Repository operations
	ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error)
	GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error)
	CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	DeleteRepo(owner, repo string) (*gitea.Response, error)

	// Pull Request operations
	ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
	GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error)
	CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
	EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error)

	// Issue operations
	ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error)
	CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error)
	EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error)

	// User operations
	GetMyUserInfo() (*gitea.User, *gitea.Response, error)
	GetUserInfo(user string) (*gitea.User, *gitea.Response, error)

	// Branch operations
	ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error)
	GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error)

	// Commit operations
	GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error)
	ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error)
}
