package servertest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestServer represents a test harness for running the MCP server
type TestServer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	t       *testing.T
	client  *mcp.Client
	session *mcp.ClientSession
	once    *sync.Once
	started bool
}

// MockGiteaServer represents a mock Gitea API server for testing
type MockGiteaServer struct {
	server        *httptest.Server
	issues        map[string][]MockIssue
	comments      map[string][]MockComment
	pullRequests  map[string][]MockPullRequest
	files         map[string][]byte             // File content storage
	notifications map[string][]MockNotification // Add notifications storage
	// Repositories that should return 404
	notFoundRepos map[string]bool
	// Comment IDs that should return 403
	forbiddenCommentIDs map[int]bool
	// Comment IDs that should return 500 error
	serverErrorCommentIDs map[int]bool
	nextID                int
	mu                    sync.Mutex
}

// MockIssue represents a mock issue for testing
type MockIssue struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	Updated string `json:"updated_at"`
	Created string `json:"created_at"`
}

// MockComment represents a mock comment for testing
type MockComment struct {
	ID      int    `json:"id"`
	Content string `json:"body"`
	Author  string `json:"user"`
	Created string `json:"created_at"`
	Updated string `json:"updated_at"`
}

// MockCommentUser represents the user who created the comment
type MockCommentUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// MockPullRequest represents a mock pull request for testing
type MockPullRequest struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	BaseRef   string `json:"base_ref"`
	UpdatedAt string `json:"updated_at"`
}

// MockNotification represents a mock notification for testing
type MockNotification struct {
	ID         int    `json:"id"`
	Repository string `json:"repository"`
	Type       string `json:"type"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Unread     bool   `json:"unread"`
	Updated    string `json:"updated_at"`
	URL        string `json:"url"`
}

// GetTextContent extracts text content from MCP content slice
//
// Parameters:
//   - content: []mcp.Content the content slice to extract from
//
// Returns:
//   - string: the extracted text content, or empty string if not found
//
// Example usage:
//
//	text := GetTextContent(result.Content)
//	if strings.Contains(text, "success") {
//	    t.Log("Operation succeeded")
//	}
func GetTextContent(content []mcp.Content) string {
	for _, c := range content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}

// GetStructuredContent extracts structured content from MCP result
//
// Parameters:
//   - result: *mcp.CallToolResult the result to extract structured content from
//
// Returns:
//   - map[string]any: the structured content, or nil if not found
//
// Example usage:
//
//	structured := GetStructuredContent(result)
//	if comment, ok := structured["comment"].(map[string]any); ok {
//	    t.Logf("Comment ID: %v", comment["id"])
//	}
func GetStructuredContent(result *mcp.CallToolResult) map[string]any {
	if result == nil || result.StructuredContent == nil {
		return nil
	}
	if structured, ok := result.StructuredContent.(map[string]any); ok {
		return structured
	}
	return nil
}

// CreateTestContext creates a standardized test context with timeout
//
// Parameters:
//   - t: *testing.T for test context
//   - timeout: time.Duration for the context timeout (defaults to 5 seconds if 0)
//
// Returns:
//   - context.Context: the created context
//   - context.CancelFunc: the cancel function for cleanup
//
// Example usage:
//
//	ctx, cancel := CreateTestContext(t, 10*time.Second)
//	defer cancel()
func CreateTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(t.Context(), timeout)
}

// NewMockGiteaServer creates a new mock Gitea server
func NewMockGiteaServer(t *testing.T) *MockGiteaServer {
	mock := &MockGiteaServer{
		issues:                make(map[string][]MockIssue),
		comments:              make(map[string][]MockComment),
		pullRequests:          make(map[string][]MockPullRequest),
		files:                 make(map[string][]byte),
		notifications:         make(map[string][]MockNotification),
		notFoundRepos:         make(map[string]bool),
		forbiddenCommentIDs:   make(map[int]bool),
		serverErrorCommentIDs: make(map[int]bool),
		nextID:                1,
	}

	// Create HTTP handler for mock API with modern routing
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", mock.handleGiteaVersion)

	// Register individual handlers with method + path patterns
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", mock.handlePullRequests)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls/{number}", mock.handlePullRequest)
	handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/pulls", mock.handleCreatePullRequest)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/pulls/{number}", mock.handleEditPullRequest)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", mock.handleIssues)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/{number}", mock.handleEditIssue)
	handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleCreateComment)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleListComments)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", mock.handleEditComment)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/contents/{path...}", mock.handleGetFileContent)
	handler.HandleFunc("GET /api/v1/notifications", mock.handleNotifications)

	mock.server = httptest.NewServer(handler)
	t.Cleanup(mock.server.Close)
	return mock
}

// URL returns the mock server URL
func (m *MockGiteaServer) URL() string {
	return m.server.URL
}

// AddIssues adds mock issues for a repository
func (m *MockGiteaServer) AddIssues(owner, repo string, issues []MockIssue) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := owner + "/" + repo
	m.issues[key] = issues
}

// AddIssue adds a single mock issue for a repository
func (m *MockGiteaServer) AddIssue(owner, repo string, issue MockIssue) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := owner + "/" + repo
	if m.issues[key] == nil {
		m.issues[key] = []MockIssue{}
	}
	m.issues[key] = append(m.issues[key], issue)
}

// AddComments adds mock comments for a repository
func (m *MockGiteaServer) AddComments(owner, repo string, comments []MockComment) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := owner + "/" + repo + "/comments"
	m.comments[key] = comments
}

// AddPullRequests adds mock pull requests for a repository
func (m *MockGiteaServer) AddPullRequests(owner, repo string, pullRequests []MockPullRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := owner + "/" + repo
	m.pullRequests[key] = pullRequests
}

// AddNotifications adds mock notifications
func (m *MockGiteaServer) AddNotifications(notifications []MockNotification) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications["user"] = notifications
}

// SetNotFoundRepo marks a repository as not found (will return 404)
func (m *MockGiteaServer) SetNotFoundRepo(owner, repo string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := owner + "/" + repo
	m.notFoundRepos[key] = true
}

// SetForbiddenCommentEdit marks a comment ID as forbidden (will return 403)
func (m *MockGiteaServer) SetForbiddenCommentEdit(commentID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.forbiddenCommentIDs[commentID] = true
}

// SetServerErrorCommentEdit marks a comment ID as server error (will return 500)
func (m *MockGiteaServer) SetServerErrorCommentEdit(commentID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.serverErrorCommentIDs[commentID] = true
}

// handleVersion handles the version endpoint
func (m *MockGiteaServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
}

// createGiteaHandler creates an HTTP handler that responds as a Gitea server
func (m *MockGiteaServer) createGiteaHandler() http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", m.handleGiteaVersion)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", m.handlePullRequests)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", m.handleIssues)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/{number}", m.handleEditIssue)
	handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", m.handleCreateComment)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", m.handleListComments)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", m.handleEditComment)
	return handler
}

// createForgejoHandler creates an HTTP handler that responds as a Forgejo server
func (m *MockGiteaServer) createForgejoHandler() http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", m.handleForgejoVersion)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", m.handlePullRequests)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", m.handleIssues)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/{number}", m.handleEditIssue)
	handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", m.handleCreateComment)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", m.handleListComments)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", m.handleEditComment)
	return handler
}

// handleGiteaVersion handles the version endpoint for Gitea
func (m *MockGiteaServer) handleGiteaVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
}

// handleForgejoVersion handles the version endpoint for Forgejo
func (m *MockGiteaServer) handleForgejoVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "12.0.0"})
}

// handlePullRequests handles pull requests endpoint
func (m *MockGiteaServer) handlePullRequests(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		m.mu.Unlock()
		http.NotFound(w, r)
		return
	}

	pullRequests, exists := m.pullRequests[repoKey]
	if !exists {
		pullRequests = []MockPullRequest{}
	}

	// Filter by state if provided in query parameters
	state := r.URL.Query().Get("state")
	if state != "" {
		var filtered []MockPullRequest
		for _, pr := range pullRequests {
			if state == "all" || pr.State == state {
				filtered = append(filtered, pr)
			}
		}
		pullRequests = filtered
	}

	// Handle pagination
	limit, offset := parsePagination(r)

	// Apply pagination
	if offset >= len(pullRequests) {
		pullRequests = []MockPullRequest{}
	} else {
		end := offset + limit
		if end > len(pullRequests) {
			end = len(pullRequests)
		}
		pullRequests = pullRequests[offset:end]
	}

	// Convert to Gitea SDK format
	giteaPRs := make([]map[string]any, len(pullRequests))
	for i, pr := range pullRequests {
		giteaPRs[i] = map[string]any{
			"id":     pr.ID,
			"number": pr.Number,
			"title":  pr.Title,
			"body":   "",
			"state":  pr.State,
			"user": map[string]any{
				"login": "testuser",
			},
			"created_at": "2025-09-11T10:30:00Z",
			"updated_at": "2025-09-11T10:30:00Z",
			"head": map[string]any{
				"ref": "feature-branch",
				"sha": "abc123",
			},
			"base": map[string]any{
				"ref": "main",
				"sha": "def456",
			},
		}
	}

	writeJSONResponse(w, giteaPRs, http.StatusOK)
}

// handlePullRequest handles single pull request endpoint
func (m *MockGiteaServer) handlePullRequest(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Extract PR number from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/repos/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[2] != "pulls" {
		http.NotFound(w, r)
		return
	}
	prNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	pullRequests, exists := m.pullRequests[repoKey]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Find the PR
	var foundPR *MockPullRequest
	for _, pr := range pullRequests {
		if pr.Number == prNumber {
			foundPR = &pr
			break
		}
	}
	if foundPR == nil {
		http.NotFound(w, r)
		return
	}

	// Return the PR in Gitea API format
	giteaPR := map[string]any{
		"id":     foundPR.ID,
		"number": foundPR.Number,
		"title":  foundPR.Title,
		"body":   foundPR.Body,
		"state":  foundPR.State,
		"user": map[string]any{
			"login":     "testuser",
			"username":  "testuser",
			"id":        1,
			"full_name": "Test User",
		},
		"poster": map[string]any{
			"login":     "testuser",
			"username":  "testuser",
			"id":        1,
			"full_name": "Test User",
		},
		"created_at": "2025-09-11T10:30:00Z",
		"updated_at": foundPR.UpdatedAt,
		"closed_at":  nil,
		"merged_at":  nil,
		"due_date":   nil,
		"head": map[string]any{
			"ref": "feature-branch",
			"sha": "abc123",
		},
		"base": map[string]any{
			"ref": "main",
			"sha": "def456",
		},
		"html_url":              fmt.Sprintf("https://example.com/%s/pull/%d", repoKey, foundPR.Number),
		"diff_url":              fmt.Sprintf("https://example.com/%s/pull/%d.diff", repoKey, foundPR.Number),
		"patch_url":             fmt.Sprintf("https://example.com/%s/pull/%d.patch", repoKey, foundPR.Number),
		"comments":              0,
		"mergeable":             true,
		"has_merged":            false,
		"allow_maintainer_edit": true,
		"assignee":              nil,
		"assignees":             []map[string]any{},
		"merged_by":             nil,
		"merged_commit_id":      nil,
		"labels":                []map[string]any{},
		"milestone":             nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(giteaPR)
}

// handleEditPullRequest handles pull request edit endpoint
func (m *MockGiteaServer) handleEditPullRequest(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "PATCH" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key and PR number from path
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Extract PR number from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/repos/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[2] != "pulls" {
		http.NotFound(w, r)
		return
	}
	prNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		m.mu.Unlock()
		http.NotFound(w, r)
		return
	}

	pullRequests, exists := m.pullRequests[repoKey]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Find the PR to edit
	prIndex := -1
	for i, pr := range pullRequests {
		if pr.Number == prNumber {
			prIndex = i
			break
		}
	}
	if prIndex == -1 {
		http.NotFound(w, r)
		return
	}

	// Parse request body for edit options
	var editOptions struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		State string `json:"state"`
		Base  string `json:"base"`
	}
	if err := json.NewDecoder(r.Body).Decode(&editOptions); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update the PR
	pr := &pullRequests[prIndex]
	if editOptions.Title != "" {
		pr.Title = editOptions.Title
	}
	if editOptions.Body != "" {
		pr.Body = editOptions.Body
	}
	if editOptions.State != "" {
		pr.State = editOptions.State
	}
	if editOptions.Base != "" {
		pr.BaseRef = editOptions.Base
	}
	pr.UpdatedAt = "2025-10-04T12:00:00Z"

	// Return the updated PR
	giteaPR := map[string]any{
		"id":     pr.ID,
		"number": pr.Number,
		"title":  pr.Title,
		"body":   pr.Body,
		"state":  pr.State,
		"user": map[string]any{
			"login": "testuser",
		},
		"created_at": "2025-09-11T10:30:00Z",
		"updated_at": pr.UpdatedAt,
		"head": map[string]any{
			"ref": "feature-branch",
			"sha": "abc123",
		},
		"base": map[string]any{
			"ref": pr.BaseRef,
			"sha": "def456",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(giteaPR)
}

// handleCreatePullRequest handles pull request creation endpoint
func (m *MockGiteaServer) handleCreatePullRequest(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	// Parse request body
	var createRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  string `json:"head"`
		Base  string `json:"base"`
		Draft bool   `json:"draft"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if createRequest.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if createRequest.Head == "" {
		http.Error(w, "Head branch is required", http.StatusBadRequest)
		return
	}
	if createRequest.Base == "" {
		http.Error(w, "Base branch is required", http.StatusBadRequest)
		return
	}

	// Initialize pull requests slice if it doesn't exist
	if m.pullRequests[repoKey] == nil {
		m.pullRequests[repoKey] = []MockPullRequest{}
	}

	// Create new pull request
	newPR := MockPullRequest{
		ID:        m.nextID,
		Number:    m.nextID,
		Title:     createRequest.Title,
		Body:      createRequest.Body,
		State:     "open",
		BaseRef:   createRequest.Base,
		UpdatedAt: "2025-10-07T12:00:00Z",
	}

	if createRequest.Draft {
		newPR.Title = "[DRAFT] " + newPR.Title
	}

	// Add to pull requests
	m.pullRequests[repoKey] = append(m.pullRequests[repoKey], newPR)
	m.nextID++

	// Return response in Gitea API format
	giteaPR := map[string]any{
		"id":     newPR.Number,
		"number": newPR.Number,
		"title":  newPR.Title,
		"body":   newPR.Body,
		"state":  newPR.State,
		"user": map[string]any{
			"login": "testuser",
		},
		"created_at": "2025-10-07T12:00:00Z",
		"updated_at": newPR.UpdatedAt,
		"head": map[string]any{
			"ref": createRequest.Head,
			"sha": "abc123",
		},
		"base": map[string]any{
			"ref": createRequest.Base,
			"sha": "def456",
		},
		"draft": createRequest.Draft,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(giteaPR)
}

// handleIssues handles issues endpoint
func (m *MockGiteaServer) handleIssues(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	issues, exists := m.issues[repoKey]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Filter by state if provided in query parameters (default to open like Gitea client)
	state := r.URL.Query().Get("state")
	if state == "" {
		state = "open" // Default to open issues only
	}
	var filteredIssues []MockIssue
	for _, issue := range issues {
		if state == "all" || issue.State == state {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	// Handle pagination
	limit, offset := parsePagination(r)

	// Apply pagination
	if offset >= len(filteredIssues) {
		filteredIssues = []MockIssue{}
	} else {
		end := offset + limit
		if end > len(filteredIssues) {
			end = len(filteredIssues)
		}
		filteredIssues = filteredIssues[offset:end]
	}

	// Convert to Gitea SDK format
	giteaIssues := make([]map[string]any, len(filteredIssues))
	for i, issue := range filteredIssues {
		giteaIssues[i] = map[string]any{
			"id":     issue.Index,
			"number": issue.Index,
			"title":  issue.Title,
			"body":   "",
			"state":  issue.State,
			"user": map[string]any{
				"login": "testuser",
			},
			"created_at": "2025-09-14T10:30:00Z",
			"updated_at": "2025-09-14T10:30:00Z",
		}
	}

	writeJSONResponse(w, giteaIssues, http.StatusOK)
}

// handleCreateComment handles comment creation endpoint
func (m *MockGiteaServer) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository exists (simulate API error for nonexistent repos)
	if repoKey == "nonexistent/repo" {
		http.NotFound(w, r)
		return
	}

	// Parse comment from request body
	var commentReq struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&commentReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Store comment with proper synchronization
	m.mu.Lock()

	// Create mock comment response that matches Gitea SDK format
	comment := map[string]any{
		"id":         m.nextID,
		"body":       commentReq.Body,
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"user": map[string]any{
			"login": "testuser",
		},
	}
	m.nextID++

	// Store comment for listing
	mockComment := MockComment{
		ID:      m.nextID - 1, // Use the ID we just assigned
		Content: commentReq.Body,
		Author:  "testuser",
		Created: "2024-01-01T00:00:00Z",
		Updated: "2024-01-01T00:00:00Z",
	}

	key := repoKey + "/comments"
	m.comments[key] = append(m.comments[key], mockComment)
	m.mu.Unlock()

	writeJSONResponse(w, comment, http.StatusCreated)
}

// handleListComments handles comment listing endpoint
func (m *MockGiteaServer) handleListComments(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	// Check if repository exists (simulate API error for nonexistent repos)
	if repoKey == "nonexistent/repo" {
		http.NotFound(w, r)
		return
	}

	// Get stored comments for this repository
	key := repoKey + "/comments"
	storedComments, exists := m.comments[key]
	if !exists {
		storedComments = []MockComment{}
	}
	limit, offset := parsePagination(r)
	// Apply pagination
	if offset >= len(storedComments) {
		storedComments = []MockComment{}
	} else {
		end := offset + limit
		if end > len(storedComments) || limit == 0 {
			end = len(storedComments)
		}
		storedComments = storedComments[offset:end]
	}

	// Convert to Gitea SDK format
	comments := make([]map[string]any, len(storedComments))
	for i, mc := range storedComments {
		comments[i] = map[string]any{
			"id":         mc.ID,
			"body":       mc.Content,
			"created_at": mc.Created,
			"updated_at": mc.Updated,
			"user": map[string]any{
				"login": mc.Author,
			},
		}
	}

	writeJSONResponse(w, comments, http.StatusOK)
}

// handleEditComment handles comment editing endpoint
func (m *MockGiteaServer) handleEditComment(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "PATCH" {
		http.NotFound(w, r)
		return
	}

	// Check authentication token
	if !validateAuthToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if repository is marked as not found
	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	// Parse comment ID from URL
	commentIDStr := r.PathValue("id")
	if commentIDStr == "" {
		http.NotFound(w, r)
		return
	}

	commentID := 123 // Default ID for testing
	if commentIDStr != "" {
		if parsedID, err := strconv.Atoi(commentIDStr); err == nil {
			commentID = parsedID
		}
	}

	// Parse comment from request body
	var commentReq struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&commentReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if repository exists (simulate API error for nonexistent repos)
	if repoKey == "nonexistent/repo" {
		http.NotFound(w, r)
		return
	}

	// Check if comment is forbidden (simulate permission error)
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forbiddenCommentIDs[commentID] {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if comment should return server error
	if m.serverErrorCommentIDs[commentID] {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Update stored comment
	key := repoKey + "/comments"
	commentFound := false
	if storedComments, exists := m.comments[key]; exists {
		for i := range storedComments {
			if storedComments[i].ID == commentID {
				storedComments[i].Content = commentReq.Body
				commentFound = true
				break
			}
		}
	}

	// If comment not found, return 404
	if !commentFound {
		http.NotFound(w, r)
		return
	}

	// Create mock comment response that matches Gitea SDK format
	comment := map[string]any{
		"id":         commentID,
		"body":       commentReq.Body,
		"created_at": "2025-09-10T10:00:00Z",
		"updated_at": "2025-09-10T10:00:00Z",
		"user": map[string]any{
			"login": "testuser",
		},
	}

	writeJSONResponse(w, comment, http.StatusOK)
}

// handleEditIssue handles issue editing endpoint
func (m *MockGiteaServer) handleEditIssue(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "PATCH" {
		fmt.Printf("DEBUG: Method not PATCH, returning 404\n")
		http.NotFound(w, r)
		return
	}

	// Check authentication token
	if !validateAuthToken(r) {
		fmt.Printf("DEBUG: Auth validation failed\n")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract repository key and issue number from path (same pattern as handleEditPullRequest)
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Extract issue number from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/repos/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[2] != "issues" {
		http.NotFound(w, r)
		return
	}
	issueNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Parse request body
	var editReq struct {
		Title *string `json:"title"`
		Body  *string `json:"body"`
		State *string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&editReq); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Check if repository is marked as not found
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	// Find and update the issue
	storedIssues, exists := m.issues[repoKey]
	if !exists {
		http.NotFound(w, r)
		return
	}

	issueFound := false
	var updatedIssue MockIssue
	for i := range storedIssues {
		if storedIssues[i].Index == issueNumber {
			// Update fields if provided (non-empty)
			if editReq.Title != nil && *editReq.Title != "" {
				storedIssues[i].Title = *editReq.Title
			}
			if editReq.Body != nil && *editReq.Body != "" {
				storedIssues[i].Body = *editReq.Body
			}
			if editReq.State != nil && *editReq.State != "" {
				storedIssues[i].State = *editReq.State
			}
			// Update timestamp
			storedIssues[i].Updated = "2025-10-06T12:00:00Z"
			updatedIssue = storedIssues[i]
			issueFound = true
			break
		}
	}

	// If issue not found, return 404
	if !issueFound {
		http.NotFound(w, r)
		return
	}

	// Create mock issue response that matches Forgejo SDK format
	issue := map[string]any{
		"id":         updatedIssue.Index,
		"number":     updatedIssue.Index,
		"title":      updatedIssue.Title,
		"body":       updatedIssue.Body,
		"state":      updatedIssue.State,
		"created_at": updatedIssue.Created,
		"updated_at": updatedIssue.Updated,
		"user": map[string]any{
			"login": "testuser",
		},
	}

	writeJSONResponse(w, issue, http.StatusOK)
}

// NewTestServer creates a new TestServer instance with standardized setup
// This is the primary constructor for most tests, providing a clean API
// while maintaining backward compatibility.
//
// Example usage:
//	ts := NewTestServer(t, ctx, map[string]string{
//		"FORGEJO_REMOTE_URL": mock.URL(),
//		"FORGEJO_AUTH_TOKEN": "mock-token",
//	})
func NewTestServer(t *testing.T, ctx context.Context, env map[string]string) *TestServer {
	return NewTestServerWithCompatAndDebug(t, ctx, env, false, false)
}

// NewTestServerWithDebug creates a new test server with optional debug mode
func NewTestServerWithDebug(t *testing.T, ctx context.Context, env map[string]string, debug bool) *TestServer {
	return NewTestServerWithCompatAndDebug(t, ctx, env, debug, false)
}

// NewTestServerWithCompat creates a new test server with optional compatibility mode
func NewTestServerWithCompat(t *testing.T, ctx context.Context, env map[string]string, compat bool) *TestServer {
	return NewTestServerWithCompatAndDebug(t, ctx, env, false, compat)
}

// NewTestServerWithCompatAndDebug creates a new test server with optional debug and compatibility modes
func NewTestServerWithCompatAndDebug(t *testing.T, ctx context.Context, env map[string]string, debug, compat bool) *TestServer {
	if ctx == nil {
		ctx = t.Context()
	}

	// Create a context with timeout for safety
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

	// Set environment variables for config loading
	for key, value := range env {
		t.Setenv(key, value)
	}

	// Use real client
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	srv, err := server.NewFromConfigWithDebugAndCompat(cfg, debug, compat)
	if err != nil {
		t.Fatalf("Failed to create server from config: %v", err)
	}

	// Create client with in-memory transport
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	// Create in-memory transports for client-server communication
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Start server in background
	go func() {
		if err := srv.MCPServer().Run(ctx, serverTransport); err != nil {
			if !errors.Is(context.Canceled, err) {
				t.Logf("Server error: %v", err)
			}
		}
	}()

	// Connect client
	session, err := client.Connect(ctx, clientTransport, &mcp.ClientSessionOptions{})
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}

	ts := &TestServer{
		ctx:     ctx,
		cancel:  cancel,
		t:       t,
		client:  client,
		session: session,
		once:    &sync.Once{},
	}

	// Use t.Cleanup for resource cleanup
	t.Cleanup(func() {
		cancel()
		if err := session.Close(); err != nil {
			t.Logf("Error closing session: %v", err)
		}
	})
	return ts
}

// Client returns the MCP client session for tool calls
func (ts *TestServer) Client() *mcp.ClientSession { return ts.session }

// IsRunning checks if the server process is running
func (ts *TestServer) IsRunning() bool {
	return ts != nil && ts.client != nil && ts.started
}

// Start starts the server process with error handling
func (ts *TestServer) Start() error {
	// In the new SDK, the server is already started when we created the session
	ts.started = true
	return nil
}

// Initialize initializes the MCP client for communication with the server
func (ts *TestServer) Initialize() error {
	// In the new SDK, initialization happens automatically during connection
	return nil
}

// CallToolWithValidation calls a tool with standardized error handling and validation
//
// Parameters:
//   - ctx: context.Context for the tool call
//   - toolName: string name of the tool to call
//   - arguments: map[string]any arguments for the tool call
//
// Returns:
//   - *mcp.CallToolResult: the result of the tool call
//   - error: any error that occurred during the call
//
// Example usage:
//
//	result, err := ts.CallToolWithValidation(ctx, "issue_list", map[string]any{
//	    "repository": "testuser/testrepo",
//	    "limit":      10,
//	})
func (ts *TestServer) CallToolWithValidation(ctx context.Context, toolName string, arguments map[string]any) (*mcp.CallToolResult, error) {
	if ctx == nil {
		ctx = ts.ctx
	}

	result, err := ts.Client().CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call tool '%s': %w", toolName, err)
	}

	return result, nil
}

// ValidateToolResult compares an actual tool result with expected result using deep equality
//
// Parameters:
//   - expected: *mcp.CallToolResult the expected result
//   - actual: *mcp.CallToolResult the actual result
//   - t: *testing.T for reporting errors
//
// Returns:
//   - bool: true if results match, false otherwise
//
// Example usage:
//
//	if !ts.ValidateToolResult(tc.expect, result, t) {
//	    t.Errorf("Tool result validation failed")
//	}
func (ts *TestServer) ValidateToolResult(expected, actual *mcp.CallToolResult, t *testing.T) bool {
	t.Helper()

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(mcp.TextContent{})) {
		t.Errorf("Tool result mismatch (-expected +actual):\n%s",
			cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(mcp.TextContent{})))
		return false
	}

	return true
}

// ValidateErrorResult validates that a tool result contains an expected error
//
// Parameters:
//   - result: *mcp.CallToolResult the result to validate
//   - expectedErrorText: string the expected error text (partial match allowed)
//   - t: *testing.T for reporting errors
//
// Returns:
//   - bool: true if result contains expected error, false otherwise
//
// Example usage:
//
//	if !ts.ValidateErrorResult(result, "Invalid request", t) {
//	    t.Errorf("Expected error result not found")
//	}
func (ts *TestServer) ValidateErrorResult(result *mcp.CallToolResult, expectedErrorText string, t *testing.T) bool {
	t.Helper()

	if !result.IsError {
		t.Errorf("Expected error result, but got success")
		return false
	}

	actualText := GetTextContent(result.Content)
	if !strings.Contains(actualText, expectedErrorText) {
		t.Errorf("Expected error text '%s' not found in result: '%s'", expectedErrorText, actualText)
		return false
	}

	return true
}

// ValidateSuccessResult validates that a tool result is successful and contains expected text
//
// Parameters:
//   - result: *mcp.CallToolResult the result to validate
//   - expectedSuccessText: string the expected success text (partial match allowed)
//   - t: *testing.T for reporting errors
//
// Returns:
//   - bool: true if result is successful and contains expected text, false otherwise
//
// Example usage:
//
//	if !ts.ValidateSuccessResult(result, "Comment created successfully", t) {
//	    t.Errorf("Expected success result not found")
//	}

// AddFile adds mock file content for a repository
func (m *MockGiteaServer) AddFile(owner, repo, ref, filepath string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s/%s", owner, repo, ref, filepath)
	m.files[key] = content
}

// handleGetFileContent handles file content retrieval endpoint
func (m *MockGiteaServer) handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Extract repository key from path values
	repoKey, err := getRepoKeyFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Extract file path from path values
	filepath := r.PathValue("path")
	if filepath == "" {
		http.NotFound(w, r)
		return
	}

	// Get reference from query parameters (default to "main")
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "main"
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if repository is marked as not found
	if m.notFoundRepos[repoKey] {
		http.NotFound(w, r)
		return
	}

	// Look for file content
	key := fmt.Sprintf("%s/%s/%s", repoKey, ref, filepath)
	content, exists := m.files[key]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Return file content in Gitea API format
	response := map[string]any{
		"content":  string(content),
		"encoding": "none", // We're storing raw content, not base64
		"name":     filepath[strings.LastIndex(filepath, "/")+1:],
		"path":     filepath,
		"sha":      "mock-sha-123",
		"size":     len(content),
		"type":     "file",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ts *TestServer) ValidateSuccessResult(result *mcp.CallToolResult, expectedSuccessText string, t *testing.T) bool {
	t.Helper()

	if result.IsError {
		t.Errorf("Expected success result, but got error: %s", GetTextContent(result.Content))
		return false
	}

	actualText := GetTextContent(result.Content)
	if !strings.Contains(actualText, expectedSuccessText) {
		t.Errorf("Expected success text '%s' not found in result: '%s'", expectedSuccessText, actualText)
		return false
	}

	return true
}

// handleNotifications handles notification endpoint
func (m *MockGiteaServer) handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	notifications, exists := m.notifications["user"]
	if !exists {
		notifications = []MockNotification{}
	}

	// Debug: log the query parameters
	// fmt.Printf("DEBUG: Query params: %v\n", r.URL.Query())

	// Filter by status if provided (simulate SDK behavior)
	// The SDK might send status as "status-types" or "status", check both
	statuses := r.URL.Query()["status"]
	if len(statuses) == 0 {
		statuses = r.URL.Query()["status-types"]
	}

	var filtered []MockNotification
	for _, notif := range notifications {
		// If no status specified, return all (SDK default behavior)
		if len(statuses) == 0 {
			filtered = append(filtered, notif)
			continue
		}

		// Check if notification matches any of the requested statuses
		shouldInclude := false
		for _, status := range statuses {
			if (status == "read" && !notif.Unread) || (status == "unread" && notif.Unread) {
				shouldInclude = true
				break
			}
		}

		if shouldInclude {
			filtered = append(filtered, notif)
		}
	}

	// Convert to SDK format
	sdkNotifications := make([]map[string]any, len(filtered))
	for i, notif := range filtered {
		sdkNotifications[i] = map[string]any{
			"id":         notif.ID,
			"unread":     notif.Unread,
			"updated_at": notif.Updated,
			"repository": map[string]any{
				"full_name": notif.Repository,
			},
			"subject": map[string]any{
				"title": notif.Title,
				"type":  strings.Title(notif.Type),
				"url":   notif.URL,
			},
		}
	}

	writeJSONResponse(w, sdkNotifications, http.StatusOK)
}
