package servertest

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

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
	notFoundRepos map[string]bool // Repositories that should return 404
	nextID        int
	mu            sync.Mutex
}

// MockIssue represents a mock issue for testing
type MockIssue struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	State string `json:"state"`
}

// MockComment represents a mock comment for testing
type MockComment struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
	Created string `json:"created"`
}

// MockCommentUser represents the user who created the comment
type MockCommentUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// MockPullRequest represents a mock pull request for testing
type MockPullRequest struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// NewMockGiteaServer creates a new mock Gitea server
func NewMockGiteaServer(t *testing.T) *MockGiteaServer {
	mock := &MockGiteaServer{
		issues:        make(map[string][]MockIssue),
		comments:      make(map[string][]MockComment),
		pullRequests:  make(map[string][]MockPullRequest),
		notFoundRepos: make(map[string]bool),
		nextID:        1,
	}

	// Create HTTP handler for mock API with modern routing
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", mock.handleVersion)

	// Register individual handlers with method + path patterns
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/pulls", mock.handlePullRequests)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues", mock.handleIssues)
	handler.HandleFunc("POST /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleCreateComment)
	handler.HandleFunc("GET /api/v1/repos/{owner}/{repo}/issues/{number}/comments", mock.handleListComments)
	handler.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/issues/comments/{id}", mock.handleEditComment)

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
	key := owner + "/" + repo
	m.issues[key] = issues
}

// AddComments adds mock comments for a repository
func (m *MockGiteaServer) AddComments(owner, repo string, comments []MockComment) {
	key := owner + "/" + repo + "/comments"
	m.comments[key] = comments
}

// AddPullRequests adds mock pull requests for a repository
func (m *MockGiteaServer) AddPullRequests(owner, repo string, pullRequests []MockPullRequest) {
	key := owner + "/" + repo
	m.pullRequests[key] = pullRequests
}

// SetNotFoundRepo marks a repository as not found (will return 404)
func (m *MockGiteaServer) SetNotFoundRepo(owner, repo string) {
	key := owner + "/" + repo
	m.notFoundRepos[key] = true
}

// handleVersion handles the version endpoint
func (m *MockGiteaServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
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
	if m.notFoundRepos[repoKey] {
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
		if end > len(pullRequests) || limit == 0 {
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

	issues, exists := m.issues[repoKey]
	if !exists {
		http.NotFound(w, r)
		return
	}

	writeJSONResponse(w, issues, http.StatusOK)
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

	// Create mock comment response that matches Gitea SDK format
	comment := map[string]any{
		"id":      m.nextID,
		"body":    commentReq.Body,
		"created": "2025-09-09T10:30:00Z",
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
		Created: "2025-09-09T10:30:00Z",
	}

	// Store comment
	key := repoKey + "/comments"
	m.comments[key] = append(m.comments[key], mockComment)

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

	// Convert to Gitea SDK format
	comments := make([]map[string]any, len(storedComments))
	for i, mc := range storedComments {
		comments[i] = map[string]any{
			"id":      mc.ID,
			"body":    mc.Content,
			"created": mc.Created,
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

	// Update stored comment
	m.mu.Lock()
	key := repoKey + "/comments"
	if storedComments, exists := m.comments[key]; exists {
		for i := range storedComments {
			if storedComments[i].ID == 1 { // Use fixed ID for testing
				storedComments[i].Content = commentReq.Body
				break
			}
		}
	}
	m.mu.Unlock()

	// Create mock comment response that matches Gitea SDK format
	comment := map[string]any{
		"id":      123, // Use fixed ID for testing
		"body":    commentReq.Body,
		"created": "2025-09-10T10:00:00Z",
		"user": map[string]any{
			"login": "testuser",
		},
	}

	writeJSONResponse(w, comment, http.StatusOK)
}

// NewTestServer creates a new TestServer instance
func NewTestServer(t *testing.T, ctx context.Context, env map[string]string) *TestServer {
	if ctx == nil {
		ctx = t.Context()
	}
	ctx, cancel := context.WithCancel(ctx)
	defaults := map[string]string{
		"FORGEJO_REMOTE_URL": "http://change-me.now.localhost",
		"FORGEJO_AUTH_TOKEN": "test-token",
	}
	maps.Copy(defaults, env)

	// Use real client
	srv, err := server.NewFromConfig(&config.Config{
		RemoteURL: defaults["FORGEJO_REMOTE_URL"],
		AuthToken: defaults["FORGEJO_AUTH_TOKEN"],
	})
	if err != nil {
		t.Fatal(err)
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
		t.Fatal("failed to connect client: ", err)
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
			t.Log(err)
		}
	})
	return ts
}

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
