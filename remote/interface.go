package remote

import (
	"context"
)

// Issue represents a Git repository issue
type Issue struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Body    string `json:"body,omitempty"`
	User    string `json:"user"`
	Updated string `json:"updated,omitempty"`
	Created string `json:"created,omitempty"`
}

// IssueLister defines the interface for listing issues from a Git repository
type IssueLister interface {
	ListIssues(ctx context.Context, repo string, limit, offset int) ([]Issue, error)
}

// Comment represents a comment on a Git repository issue or pull request
type Comment struct {
	ID      int    `json:"id"`
	Content string `json:"body"`
	Author  string `json:"user"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// IssueCommenter defines the interface for creating comments on Git repository issues
type IssueCommenter interface {
	CreateIssueComment(ctx context.Context, repo string, issueNumber int, comment string) (*Comment, error)
}

// IssueCommentList represents a collection of comments with pagination metadata
type IssueCommentList struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// ListIssueCommentsArgs represents the arguments for listing issue comments
type ListIssueCommentsArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// IssueCommentLister defines the interface for listing comments from Git repository issues
type IssueCommentLister interface {
	ListIssueComments(ctx context.Context, repo string, issueNumber int, limit, offset int) (*IssueCommentList, error)
}

// EditIssueCommentArgs represents the arguments for editing an issue comment
type EditIssueCommentArgs struct {
	Repository  string `json:"repository"`
	IssueNumber int    `json:"issue_number"`
	CommentID   int    `json:"comment_id"`
	NewContent  string `json:"new_content"`
}

// IssueCommentEditor defines the interface for editing comments on Git repository issues
type IssueCommentEditor interface {
	EditIssueComment(ctx context.Context, args EditIssueCommentArgs) (*Comment, error)
}

// CreateIssueArgs represents arguments for creating a new issue
type CreateIssueArgs struct {
	Repository string `json:"repository"`
	Title      string `json:"title"`
	Body       string `json:"body"`
}

// IssueCreator defines the interface for creating issues
type IssueCreator interface {
	CreateIssue(ctx context.Context, args CreateIssueArgs) (*Issue, error)
}

// CreateIssueWithAttachmentsArgs represents arguments for creating a new issue with attachments
type CreateIssueWithAttachmentsArgs struct {
	CreateIssueArgs
	Attachments []ProcessedAttachment
}

// ProcessedAttachment represents a processed file attachment
type ProcessedAttachment struct {
	Data     []byte
	Filename string
	MIMEType string
}

// IssueAttachmentCreator defines the interface for creating issues with attachments
type IssueAttachmentCreator interface {
	CreateIssueWithAttachments(ctx context.Context, args CreateIssueWithAttachmentsArgs) (*Issue, error)
}

// EditIssueArgs represents the arguments for editing an issue
type EditIssueArgs struct {
	Repository  string `json:"repository"`
	Directory   string `json:"directory"`
	IssueNumber int    `json:"issue_number"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	State       string `json:"state"`
}

// IssueEditor defines the interface for editing issues in Git repositories
type IssueEditor interface {
	EditIssue(ctx context.Context, args EditIssueArgs) (*Issue, error)
}

// PullRequestBranch represents a branch reference in a pull request
type PullRequestBranch struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

// PullRequest represents a Git repository pull request
type PullRequest struct {
	ID        int               `json:"id"`
	Number    int               `json:"number"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	State     string            `json:"state"`
	User      string            `json:"user"`
	CreatedAt string            `json:"created"`
	UpdatedAt string            `json:"updated"`
	Head      PullRequestBranch `json:"head"`
	Base      PullRequestBranch `json:"base"`
}

// ListPullRequestsOptions represents the options for listing pull requests
type ListPullRequestsOptions struct {
	State  string `json:"state"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// PullRequestLister defines the interface for listing pull requests from a Git repository
type PullRequestLister interface {
	ListPullRequests(ctx context.Context, repo string, options ListPullRequestsOptions) ([]PullRequest, error)
}

// PullRequestCommentList represents a collection of pull request comments with pagination metadata
type PullRequestCommentList struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// ListPullRequestCommentsArgs represents the arguments for listing pull request comments
type ListPullRequestCommentsArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	Limit             int    `json:"limit"`
	Offset            int    `json:"offset"`
}

// CreatePullRequestCommentArgs represents the arguments for creating a pull request comment
type CreatePullRequestCommentArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	Comment           string `json:"comment"`
}

// PullRequestCommentLister defines the interface for listing comments from Git repository pull requests
type PullRequestCommentLister interface {
	ListPullRequestComments(ctx context.Context, repo string, pullRequestNumber int, limit, offset int) (*PullRequestCommentList, error)
}

// PullRequestCommenter defines the interface for creating comments on Git repository pull requests
type PullRequestCommenter interface {
	CreatePullRequestComment(ctx context.Context, repo string, pullRequestNumber int, comment string) (*Comment, error)
}

// EditPullRequestCommentArgs represents the arguments for editing a pull request comment
type EditPullRequestCommentArgs struct {
	Repository        string `json:"repository"`
	PullRequestNumber int    `json:"pull_request_number"`
	CommentID         int    `json:"comment_id"`
	NewContent        string `json:"new_content"`
}

// PullRequestCommentEditor defines the interface for editing comments on Git repository pull requests
type PullRequestCommentEditor interface {
	EditPullRequestComment(ctx context.Context, args EditPullRequestCommentArgs) (*Comment, error)
}

// EditPullRequestArgs represents the arguments for editing a pull request
type EditPullRequestArgs struct {
	Repository        string `json:"repository"`
	Directory         string `json:"directory"`
	PullRequestNumber int    `json:"pull_request_number"`
	Title             string `json:"title"`
	Body              string `json:"body"`
	State             string `json:"state"`
	BaseBranch        string `json:"base_branch"`
}

// PullRequestEditor defines the interface for editing pull requests in Git repositories
type PullRequestEditor interface {
	EditPullRequest(ctx context.Context, args EditPullRequestArgs) (*PullRequest, error)
}

// CreatePullRequestArgs represents arguments for creating a new pull request
type CreatePullRequestArgs struct {
	Repository string `json:"repository"`
	Head       string `json:"head"` // Source branch
	Base       string `json:"base"` // Target branch
	Title      string `json:"title"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Assignee   string `json:"assignee"` // Single reviewer
}

// PullRequestCreator defines the interface for creating pull requests
type PullRequestCreator interface {
	CreatePullRequest(ctx context.Context, args CreatePullRequestArgs) (*PullRequest, error)
}

// PullRequestGetter defines the interface for fetching a single pull request
type PullRequestGetter interface {
	GetPullRequest(ctx context.Context, repo string, number int) (*PullRequestDetails, error)
}

// PullRequestDetails represents comprehensive pull request information
type PullRequestDetails struct {
	// Basic fields (matching PullRequest for compatibility)
	ID        int               `json:"id"`
	Number    int               `json:"number"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	State     string            `json:"state"`
	User      string            `json:"user"`
	CreatedAt string            `json:"created"`
	UpdatedAt string            `json:"updated"`
	Head      PullRequestBranch `json:"head"`
	Base      PullRequestBranch `json:"base"`

	// Additional metadata fields
	HTMLURL             string     `json:"html_url"`
	DiffURL             string     `json:"diff_url"`
	PatchURL            string     `json:"patch_url"`
	Labels              []Label    `json:"labels,omitempty"`
	Milestone           *Milestone `json:"milestone,omitempty"`
	Assignee            string     `json:"assignee,omitempty"`
	Assignees           []string   `json:"assignees,omitempty"`
	Comments            int        `json:"comments"`
	IsLocked            bool       `json:"is_locked"`
	Mergeable           bool       `json:"mergeable"`
	HasMerged           bool       `json:"has_merged"`
	MergedAt            string     `json:"merged_at,omitempty"`
	MergeCommitSHA      string     `json:"merge_commit_sha,omitempty"`
	MergedBy            string     `json:"merged_by,omitempty"`
	AllowMaintainerEdit bool       `json:"allow_maintainer_edit"`
	ClosedAt            string     `json:"closed_at,omitempty"`
	Deadline            string     `json:"deadline,omitempty"`
}

// Label represents a repository label
type Label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// Milestone represents a repository milestone
type Milestone struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	State        string `json:"state"`
	OpenIssues   int    `json:"open_issues"`
	ClosedIssues int    `json:"closed_issues"`
}

// FileContentFetcher defines interface for fetching repository file contents
type FileContentFetcher interface {
	GetFileContent(ctx context.Context, owner, repo, ref, filepath string) ([]byte, error)
}

// ClientInterface combines IssueLister, IssueCommenter, IssueCommentLister, IssueCommentEditor, IssueCreator, IssueAttachmentCreator, IssueEditor, PullRequestLister, PullRequestCommentLister, PullRequestCommenter, PullRequestCommentEditor, PullRequestEditor, PullRequestCreator, PullRequestGetter, and FileContentFetcher for complete Git operations
type ClientInterface interface {
	IssueLister
	IssueCommenter
	IssueCommentLister
	IssueCommentEditor
	IssueCreator
	IssueAttachmentCreator
	IssueEditor
	PullRequestLister
	PullRequestCommentLister
	PullRequestCommenter
	PullRequestCommentEditor
	PullRequestEditor
	PullRequestCreator
	PullRequestGetter
	FileContentFetcher
}
