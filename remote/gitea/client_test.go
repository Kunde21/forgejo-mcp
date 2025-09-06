package gitea

import (
	"testing"

	"code.gitea.io/sdk/gitea"
)

// mockGiteaClient is a mock implementation of GiteaClientInterface for testing
type mockGiteaClient struct {
	listMyReposFunc          func(gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error)
	getRepoFunc              func(string, string) (*gitea.Repository, *gitea.Response, error)
	createRepoFunc           func(gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	deleteRepoFunc           func(string, string) (*gitea.Response, error)
	listRepoPullRequestsFunc func(string, string, gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
	getPullRequestFunc       func(string, string, int64) (*gitea.PullRequest, *gitea.Response, error)
	createPullRequestFunc    func(string, string, gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
	editPullRequestFunc      func(string, string, int64, gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
	listIssuesFunc           func(gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	listRepoIssuesFunc       func(string, string, gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error)
	getIssueFunc             func(string, string, int64) (*gitea.Issue, *gitea.Response, error)
	createIssueFunc          func(string, string, gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error)
	editIssueFunc            func(string, string, int64, gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error)
	getMyUserInfoFunc        func() (*gitea.User, *gitea.Response, error)
	getUserInfoFunc          func(string) (*gitea.User, *gitea.Response, error)
	listRepoBranchesFunc     func(string, string, gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error)
	getRepoBranchFunc        func(string, string, string) (*gitea.Branch, *gitea.Response, error)
	getSingleCommitFunc      func(string, string, string) (*gitea.Commit, *gitea.Response, error)
	listRepoCommitsFunc      func(string, string, gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error)
}

func (m *mockGiteaClient) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	if m.listMyReposFunc != nil {
		return m.listMyReposFunc(opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
	if m.getRepoFunc != nil {
		return m.getRepoFunc(owner, repo)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	if m.createRepoFunc != nil {
		return m.createRepoFunc(opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) DeleteRepo(owner, repo string) (*gitea.Response, error) {
	if m.deleteRepoFunc != nil {
		return m.deleteRepoFunc(owner, repo)
	}
	return nil, nil
}

func (m *mockGiteaClient) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	if m.listRepoPullRequestsFunc != nil {
		return m.listRepoPullRequestsFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error) {
	if m.getPullRequestFunc != nil {
		return m.getPullRequestFunc(owner, repo, index)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.createPullRequestFunc != nil {
		return m.createPullRequestFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.editPullRequestFunc != nil {
		return m.editPullRequestFunc(owner, repo, index, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.listIssuesFunc != nil {
		return m.listIssuesFunc(opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.listRepoIssuesFunc != nil {
		return m.listRepoIssuesFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error) {
	if m.getIssueFunc != nil {
		return m.getIssueFunc(owner, repo, index)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.createIssueFunc != nil {
		return m.createIssueFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.editIssueFunc != nil {
		return m.editIssueFunc(owner, repo, index, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetMyUserInfo() (*gitea.User, *gitea.Response, error) {
	if m.getMyUserInfoFunc != nil {
		return m.getMyUserInfoFunc()
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetUserInfo(user string) (*gitea.User, *gitea.Response, error) {
	if m.getUserInfoFunc != nil {
		return m.getUserInfoFunc(user)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error) {
	if m.listRepoBranchesFunc != nil {
		return m.listRepoBranchesFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error) {
	if m.getRepoBranchFunc != nil {
		return m.getRepoBranchFunc(owner, repo, branch)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error) {
	if m.getSingleCommitFunc != nil {
		return m.getSingleCommitFunc(owner, repo, sha)
	}
	return nil, nil, nil
}

func (m *mockGiteaClient) ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error) {
	if m.listRepoCommitsFunc != nil {
		return m.listRepoCommitsFunc(owner, repo, opt)
	}
	return nil, nil, nil
}

func TestGiteaClientInterface(t *testing.T) {
	// Test that mockGiteaClient implements GiteaClientInterface
	var _ GiteaClientInterface = &mockGiteaClient{}

	// Test that MockGiteaClient implements GiteaClientInterface
	var _ GiteaClientInterface = &MockGiteaClient{}

	// Test that interface methods can be called
	mock := &mockGiteaClient{
		listMyReposFunc: func(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
			return []*gitea.Repository{}, nil, nil
		},
		getRepoFunc: func(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
			return &gitea.Repository{Name: repo}, nil, nil
		},
	}

	// Test ListMyRepos
	repos, _, err := mock.ListMyRepos(gitea.ListReposOptions{})
	if err != nil {
		t.Errorf("ListMyRepos failed: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}

	// Test GetRepo
	repo, _, err := mock.GetRepo("testowner", "testrepo")
	if err != nil {
		t.Errorf("GetRepo failed: %v", err)
	}
	if repo.Name != "testrepo" {
		t.Errorf("Expected repo name 'testrepo', got '%s'", repo.Name)
	}
}
