package servertest

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

	// Create HTTP handler for mock API
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", mock.handleVersion)
	handler.HandleFunc("/api/v1/repos/", mock.handleRepoRequests)

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

// handleRepoRequests handles repository issues and comments endpoints
func (m *MockGiteaServer) handleRepoRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle pull requests endpoint
	if strings.Contains(path, "/pulls") && r.Method == "GET" {
		parts := strings.Split(strings.TrimPrefix(path, "/api/v1/repos/"), "/pulls")
		if len(parts) != 2 || parts[0] == "" || strings.Contains(parts[0], "/") == false {
			http.NotFound(w, r)
			return
		}
		repoKey := parts[0]

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
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := len(pullRequests) // Default to all items
		offset := 0                // Default to start from beginning

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(giteaPRs)
		return
	}

	// Handle issues endpoint
	if strings.Contains(path, "/issues") && !strings.Contains(path, "/comments") && r.Method == "GET" {
		parts := strings.Split(strings.TrimPrefix(path, "/api/v1/repos/"), "/issues")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		repoKey := parts[0]

		issues, exists := m.issues[repoKey]
		if !exists {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issues)
		return
	}

	// Handle comment creation endpoint
	if strings.Contains(path, "/comments") && r.Method == "POST" {
		// Parse path: /api/v1/repos/{owner}/{repo}/issues/{number}/comments
		pathParts := strings.Split(strings.TrimPrefix(path, "/api/v1/repos/"), "/")
		if len(pathParts) < 5 {
			http.NotFound(w, r)
			return
		}

		owner := pathParts[0]
		repo := pathParts[1]
		repoKey := owner + "/" + repo

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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
		return
	}
	// Handle comment listing endpoint
	if strings.Contains(path, "/comments") && r.Method == "GET" {
		// Parse path: /api/v1/repos/{owner}/{repo}/issues/{number}/comments
		pathParts := strings.Split(strings.TrimPrefix(path, "/api/v1/repos/"), "/")
		if len(pathParts) < 5 {
			http.NotFound(w, r)
			return
		}

		owner := pathParts[0]
		repo := pathParts[1]
		repoKey := owner + "/" + repo

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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
		return
	}

	// Handle comment editing endpoint
	if strings.Contains(path, "/comments/") && r.Method == "PATCH" {
		// Check authentication token
		authHeader := r.Header.Get("Authorization")
		token := r.URL.Query().Get("token")

		// Extract token value from header
		var headerToken string
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			headerToken = after
		} else if after0, ok0 := strings.CutPrefix(authHeader, "token "); ok0 {
			headerToken = after0
		}

		// Reject invalid-token
		if headerToken == "invalid-token" || token == "invalid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Accept valid tokens or no token (for backward compatibility)
		if headerToken != "" && headerToken != "mock-token" && headerToken != "test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse path: /api/v1/repos/{owner}/{repo}/issues/comments/{id}
		pathParts := strings.Split(strings.TrimPrefix(path, "/api/v1/repos/"), "/")
		if len(pathParts) < 5 {
			http.NotFound(w, r)
			return
		}

		owner := pathParts[0]
		repo := pathParts[1]
		repoKey := owner + "/" + repo
		// Check if repository is marked as not found
		if m.notFoundRepos[repoKey] {
			http.NotFound(w, r)
			return
		}

		// Parse comment ID from URL
		commentIDStr := pathParts[4]
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
		key := repoKey + "/comments"
		if storedComments, exists := m.comments[key]; exists {
			for i := range storedComments {
				if storedComments[i].ID == 1 { // Use fixed ID for testing
					storedComments[i].Content = commentReq.Body
					break
				}
			}
		}

		// Create mock comment response that matches Gitea SDK format
		comment := map[string]any{
			"id":      123, // Use fixed ID for testing
			"body":    commentReq.Body,
			"created": "2025-09-10T10:00:00Z",
			"user": map[string]any{
				"login": "testuser",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
		return
	}

	http.NotFound(w, r)
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
