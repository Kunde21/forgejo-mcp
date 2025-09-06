package servertest

import (
	"context"
	"fmt"
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

// NewTestServer creates a new TestServer instance
func NewTestServer(t *testing.T, ctx context.Context) *TestServer {
	if ctx == nil {
		ctx = t.Context()
	}
	ctx, cancel := context.WithCancel(ctx)
	client, err := client.NewStdioMCPClientWithOptions("go", []string{}, []string{"run", "../."})
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
