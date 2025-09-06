package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// MockGiteaClient implements a comprehensive mock Gitea client for testing
type MockGiteaClient struct {
	MockPRs      []*gitea.PullRequest
	MockIssues   []*gitea.Issue
	MockRepos    []*gitea.Repository
	MockUsers    []*gitea.User
	MockBranches []*gitea.Branch
	MockCommits  []*gitea.Commit
	MockError    error
	GetRepoErr   error
}

// Repository operations
func (m *MockGiteaClient) ListMyRepos(opt gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	return m.MockRepos, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetRepo(owner, repo string) (*gitea.Repository, *gitea.Response, error) {
	if m.GetRepoErr != nil {
		return nil, nil, m.GetRepoErr
	}
	for _, r := range m.MockRepos {
		if r.Owner != nil && r.Owner.UserName == owner && r.Name == repo {
			return r, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("repository not found")
}

func (m *MockGiteaClient) CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	repo := &gitea.Repository{
		Name:        opt.Name,
		Description: opt.Description,
		Private:     opt.Private,
	}
	m.MockRepos = append(m.MockRepos, repo)
	return repo, &gitea.Response{}, nil
}

func (m *MockGiteaClient) DeleteRepo(owner, repo string) (*gitea.Response, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	for i, r := range m.MockRepos {
		if r.Owner != nil && r.Owner.UserName == owner && r.Name == repo {
			m.MockRepos = append(m.MockRepos[:i], m.MockRepos[i+1:]...)
			return &gitea.Response{}, nil
		}
	}
	return &gitea.Response{}, fmt.Errorf("repository not found")
}

// Pull Request operations
func (m *MockGiteaClient) ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}

	// Filter PRs by repository
	var filteredPRs []*gitea.PullRequest

	for _, pr := range m.MockPRs {
		// In a real implementation, PRs would be associated with repositories
		// For this mock, we'll assume all PRs belong to the requested repository
		// unless we have repository-specific mock data
		filteredPRs = append(filteredPRs, pr)
	}

	// Apply state filtering if specified
	if opt.State != "" && opt.State != gitea.StateAll {
		var stateFiltered []*gitea.PullRequest
		for _, pr := range filteredPRs {
			if pr.State == opt.State {
				stateFiltered = append(stateFiltered, pr)
			}
		}
		filteredPRs = stateFiltered
	}

	return filteredPRs, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	for _, pr := range m.MockPRs {
		if pr.Index == index {
			return pr, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("pull request not found")
}

func (m *MockGiteaClient) CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	pr := &gitea.PullRequest{
		Index:  int64(len(m.MockPRs) + 1),
		Title:  opt.Title,
		Body:   opt.Body,
		State:  gitea.StateOpen,
		Poster: &gitea.User{UserName: "testuser"},
	}
	m.MockPRs = append(m.MockPRs, pr)
	return pr, &gitea.Response{}, nil
}

func (m *MockGiteaClient) EditPullRequest(owner, repo string, index int64, opt gitea.EditPullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	// Simplified implementation to avoid type issues
	for _, pr := range m.MockPRs {
		if pr.Index == index {
			return pr, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("pull request not found")
}

// Issue operations
func (m *MockGiteaClient) ListIssues(opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}

	// For simplicity, return all mock issues
	return m.MockIssues, &gitea.Response{}, nil
}

func (m *MockGiteaClient) ListRepoIssues(owner, repo string, opt gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}

	// For simplicity, return all mock issues
	return m.MockIssues, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetIssue(owner, repo string, index int64) (*gitea.Issue, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	for _, issue := range m.MockIssues {
		if issue.Index == index {
			return issue, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("issue not found")
}

func (m *MockGiteaClient) CreateIssue(owner, repo string, opt gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	issue := &gitea.Issue{
		Index:  int64(len(m.MockIssues) + 1),
		Title:  opt.Title,
		Body:   opt.Body,
		State:  "open",
		Poster: &gitea.User{UserName: "testuser"},
	}
	m.MockIssues = append(m.MockIssues, issue)
	return issue, &gitea.Response{}, nil
}

func (m *MockGiteaClient) EditIssue(owner, repo string, index int64, opt gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	// Simplified implementation to avoid type issues
	for _, issue := range m.MockIssues {
		if issue.Index == index {
			return issue, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("issue not found")
}

// User operations
func (m *MockGiteaClient) GetMyUserInfo() (*gitea.User, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	if len(m.MockUsers) > 0 {
		return m.MockUsers[0], &gitea.Response{}, nil
	}
	return &gitea.User{UserName: "testuser", Email: "test@example.com"}, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetUserInfo(user string) (*gitea.User, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	for _, u := range m.MockUsers {
		if u.UserName == user {
			return u, &gitea.Response{}, nil
		}
	}
	return &gitea.User{UserName: user, Email: user + "@example.com"}, &gitea.Response{}, nil
}

// Branch operations
func (m *MockGiteaClient) ListRepoBranches(owner, repo string, opt gitea.ListRepoBranchesOptions) ([]*gitea.Branch, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	return m.MockBranches, &gitea.Response{}, nil
}

func (m *MockGiteaClient) GetRepoBranch(owner, repo, branch string) (*gitea.Branch, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	for _, b := range m.MockBranches {
		if b.Name == branch {
			return b, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("branch not found")
}

// Commit operations
func (m *MockGiteaClient) GetSingleCommit(owner, repo, sha string) (*gitea.Commit, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	for _, c := range m.MockCommits {
		if c.SHA == sha {
			return c, &gitea.Response{}, nil
		}
	}
	return nil, &gitea.Response{}, fmt.Errorf("commit not found")
}

func (m *MockGiteaClient) ListRepoCommits(owner, repo string, opt gitea.ListCommitOptions) ([]*gitea.Commit, *gitea.Response, error) {
	if m.MockError != nil {
		return nil, nil, m.MockError
	}
	return m.MockCommits, &gitea.Response{}, nil
}
