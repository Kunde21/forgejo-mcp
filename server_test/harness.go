package servertest

import (
	"context"
	"encoding/json"
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
	server *httptest.Server
	issues map[string][]MockIssue
}

// MockIssue represents a mock issue for testing
type MockIssue struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	State string `json:"state"`
}

// NewMockGiteaServer creates a new mock Gitea server
func NewMockGiteaServer(t *testing.T) *MockGiteaServer {
	mock := &MockGiteaServer{
		issues: make(map[string][]MockIssue),
	}

	// Create HTTP handler for mock API
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/version", mock.handleVersion)
	handler.HandleFunc("/api/v1/repos/", mock.handleRepoIssues)

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

// handleVersion handles the version endpoint
func (m *MockGiteaServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
}

// handleRepoIssues handles repository issues endpoint
func (m *MockGiteaServer) handleRepoIssues(w http.ResponseWriter, r *http.Request) {
	// Parse path to get owner/repo
	path := r.URL.Path
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
	clEnv := []string{}
	for key, value := range env {
		clEnv = append(clEnv, key+"="+value)
	}
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
			t.Logf("Server error: %v", err)
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
