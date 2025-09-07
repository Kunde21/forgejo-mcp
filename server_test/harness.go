package servertest

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestServer represents a test harness for running the MCP server
type TestServer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	t       *testing.T
	client  *client.Client
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
	client, err := client.NewStdioMCPClientWithOptions("go", clEnv, []string{"run", "../."})
	if err != nil {
		t.Fatal("failed to create stdio MCP client: ", err)
	}
	ts := &TestServer{
		ctx:    ctx,
		cancel: cancel,
		t:      t,
		client: client,
		once:   &sync.Once{},
	}

	// Use t.Cleanup for resource cleanup
	t.Cleanup(func() {
		cancel()
		if err := client.Close(); err != nil {
			t.Log(err)
		}
	})
	return ts
}
func (ts *TestServer) Client() *client.Client { return ts.client }

// IsRunning checks if the server process is running
func (ts *TestServer) IsRunning() bool {
	return ts != nil && ts.client != nil && ts.started
}

// Start starts the server process with error handling
func (ts *TestServer) Start() error {
	var err error
	ts.once.Do(func() {
		err = ts.client.Start(ts.ctx)
		ts.started = err == nil
	})
	if err != nil {
		return fmt.Errorf("failed to start server process: %w", err)
	}
	return nil
}

// Initialize initializes the MCP client for communication with the server
func (ts *TestServer) Initialize() error {
	if !ts.started {
		if err := ts.Start(); err != nil {
			return err
		}
	}
	// Perform MCP initialization handshake
	_, err := ts.client.Initialize(ts.ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MCP protocol: %w", err)
	}
	return nil
}
