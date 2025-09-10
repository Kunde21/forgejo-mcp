package servertest

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
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
	server   *httptest.Server
	issues   map[string][]MockIssue
	comments map[string][]MockComment
	nextID   int
}

// MockIssue represents a mock issue for testing
type MockIssue struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	State string `json:"state"`
}

// MockComment represents a mock comment for testing
type MockComment struct {
	ID        int             `json:"id"`
	Body      string          `json:"body"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	User      MockCommentUser `json:"user"`
}

// MockCommentUser represents the user who created the comment
type MockCommentUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// NewMockGiteaServer creates a new mock Gitea server
func NewMockGiteaServer(t *testing.T) *MockGiteaServer {
	mock := &MockGiteaServer{
		issues:   make(map[string][]MockIssue),
		comments: make(map[string][]MockComment),
		nextID:   1,
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

// handleVersion handles the version endpoint
func (m *MockGiteaServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
}

// handleRepoRequests handles repository issues and comments endpoints
func (m *MockGiteaServer) handleRepoRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle issues endpoint
	if strings.Contains(path, "/issues") && r.Method == "GET" {
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

		// Create mock comment
		comment := MockComment{
			ID:        m.nextID,
			Body:      commentReq.Body,
			CreatedAt: "2025-09-09T10:30:00Z",
			UpdatedAt: "2025-09-09T10:30:00Z",
			User: MockCommentUser{
				ID:       1,
				Username: "testuser",
			},
		}
		m.nextID++

		// Store comment
		key := repoKey + "/comments"
		m.comments[key] = append(m.comments[key], comment)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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
