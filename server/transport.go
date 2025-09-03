// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// Transport defines the interface for MCP communication
type Transport interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	Connect() error
	Disconnect() error
	GetState() ConnectionState
	IsConnected() bool
}

// ConnectionState represents the state of a transport connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateClosing
	StateClosed
)

// StdioTransport implements Transport for stdio communication
type StdioTransport struct {
	reader      *bufio.Reader
	writer      io.Writer
	logger      *logrus.Logger
	readTimeout time.Duration
	state       ConnectionState
}

// NewStdioTransport creates a new stdio transport for MCP communication
func NewStdioTransport(cfg *config.Config, logger *logrus.Logger) *StdioTransport {
	transport := &StdioTransport{
		reader:      bufio.NewReader(os.Stdin),
		writer:      os.Stdout,
		logger:      logger,
		readTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
		state:       StateDisconnected,
	}

	logger.Info("Stdio transport created")
	return transport
}

// Connect establishes the connection (for stdio, this is always connected)
func (st *StdioTransport) Connect() error {
	st.logger.Info("Connecting stdio transport")
	st.state = StateConnected
	return nil
}

// Disconnect closes the connection
func (st *StdioTransport) Disconnect() error {
	st.logger.Info("Disconnecting stdio transport")
	st.state = StateClosing

	// For stdio, we don't actually close stdin/stdout
	// but we mark the state as closed
	st.state = StateClosed
	return nil
}

// GetState returns the current connection state
func (st *StdioTransport) GetState() ConnectionState {
	return st.state
}

// IsConnected returns true if the transport is connected
func (st *StdioTransport) IsConnected() bool {
	return st.state == StateConnected
}

// Read reads data from stdin with timeout
func (st *StdioTransport) Read(p []byte) (n int, err error) {
	if !st.IsConnected() {
		return 0, fmt.Errorf("transport is not connected")
	}

	// Create a channel to signal read completion
	done := make(chan struct{})

	var bytesRead int
	var readErr error

	// Perform read in goroutine
	go func() {
		defer close(done)
		bytesRead, readErr = st.reader.Read(p)
	}()

	// Wait for either read completion or timeout
	if st.readTimeout > 0 {
		select {
		case <-done:
			return bytesRead, readErr
		case <-time.After(st.readTimeout):
			return 0, fmt.Errorf("read timeout after %v", st.readTimeout)
		}
	} else {
		// No timeout, wait for read to complete
		<-done
		return bytesRead, readErr
	}
}

// Write writes data to stdout
func (st *StdioTransport) Write(p []byte) (n int, err error) {
	if !st.IsConnected() {
		return 0, fmt.Errorf("transport is not connected")
	}

	st.logger.Debugf("Writing response: %s", string(p))
	return st.writer.Write(p)
}

// Close closes the transport
func (st *StdioTransport) Close() error {
	return st.Disconnect()
}

// SSETransport implements Transport for Server-Sent Events over HTTP
type SSETransport struct {
	config      *config.Config
	logger      *logrus.Logger
	server      *http.Server
	connections map[string]*SSEConnection
	mu          sync.RWMutex
	state       ConnectionState
}

// SSEConnection represents a single SSE connection
type SSEConnection struct {
	id       string
	writer   http.ResponseWriter
	flusher  http.Flusher
	done     chan struct{}
	mu       sync.Mutex
	isClosed bool
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(cfg *config.Config, logger *logrus.Logger) *SSETransport {
	return &SSETransport{
		config:      cfg,
		logger:      logger,
		connections: make(map[string]*SSEConnection),
		state:       StateDisconnected,
	}
}

// Connect starts the HTTP server for SSE transport
func (st *SSETransport) Connect() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.state == StateConnected {
		return nil
	}

	st.logger.Info("Starting SSE transport server")
	st.state = StateConnecting

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", st.handleSSE)
	mux.HandleFunc("/health", st.handleHealth)

	st.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", st.config.Host, st.config.Port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		st.logger.Infof("SSE server listening on %s", st.server.Addr)
		if err := st.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			st.logger.Errorf("SSE server error: %v", err)
		}
	}()

	st.state = StateConnected
	return nil
}

// Disconnect stops the HTTP server
func (st *SSETransport) Disconnect() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.state != StateConnected {
		return nil
	}

	st.logger.Info("Stopping SSE transport server")
	st.state = StateClosing

	// Close all connections
	for id, conn := range st.connections {
		conn.Close()
		delete(st.connections, id)
	}

	// Shutdown server
	if st.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := st.server.Shutdown(ctx); err != nil {
			st.logger.Errorf("Server shutdown error: %v", err)
			return err
		}
	}

	st.state = StateClosed
	return nil
}

// Read reads data from SSE connections (not applicable for SSE transport)
func (st *SSETransport) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("SSE transport does not support reading")
}

// Write sends data to all connected SSE clients
func (st *SSETransport) Write(p []byte) (n int, err error) {
	if !st.IsConnected() {
		return 0, fmt.Errorf("transport is not connected")
	}

	st.mu.RLock()
	defer st.mu.RUnlock()

	message := fmt.Sprintf("data: %s\n\n", string(p))

	// Send to all connections
	for id, conn := range st.connections {
		if err := conn.Send(message); err != nil {
			st.logger.Errorf("Failed to send to connection %s: %v", id, err)
			// Remove failed connection
			go func(connID string) {
				st.mu.Lock()
				delete(st.connections, connID)
				st.mu.Unlock()
			}(id)
		}
	}

	return len(p), nil
}

// Close closes the transport
func (st *SSETransport) Close() error {
	return st.Disconnect()
}

// GetState returns the current connection state
func (st *SSETransport) GetState() ConnectionState {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.state
}

// IsConnected returns true if the transport is connected
func (st *SSETransport) IsConnected() bool {
	return st.GetState() == StateConnected
}

// handleSSE handles SSE connection requests
func (st *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Check if client supports SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create new connection
	connID := fmt.Sprintf("%d", time.Now().UnixNano())
	conn := &SSEConnection{
		id:      connID,
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	// Add connection
	st.mu.Lock()
	st.connections[connID] = conn
	st.mu.Unlock()

	st.logger.Infof("New SSE connection: %s", connID)

	// Send initial connection event
	conn.Send("event: connected\ndata: {\"status\": \"connected\"}\n\n")

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.done:
			st.logger.Infof("SSE connection closed: %s", connID)
			return
		case <-ticker.C:
			// Send keepalive
			conn.Send(": keepalive\n\n")
		case <-r.Context().Done():
			st.logger.Infof("SSE request context done: %s", connID)
			st.removeConnection(connID)
			return
		}
	}
}

// handleHealth provides a health check endpoint
func (st *SSETransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// removeConnection removes a connection from the map
func (st *SSETransport) removeConnection(id string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.connections, id)
}

// Send sends a message to the SSE client
func (sc *SSEConnection) Send(message string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.isClosed {
		return fmt.Errorf("connection is closed")
	}

	_, err := fmt.Fprint(sc.writer, message)
	if err != nil {
		return err
	}

	sc.flusher.Flush()
	return nil
}

// Close closes the SSE connection
func (sc *SSEConnection) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.isClosed {
		sc.isClosed = true
		close(sc.done)
	}
}

// SSETransportAdapter adapts our SSETransport to the MCP SDK Transport interface
type SSETransportAdapter struct {
	transport *SSETransport
	logger    *logrus.Logger
}

// NewSSETransportAdapter creates a new SSE transport adapter for MCP SDK
func NewSSETransportAdapter(cfg *config.Config, logger *logrus.Logger) *SSETransportAdapter {
	return &SSETransportAdapter{
		transport: NewSSETransport(cfg, logger),
		logger:    logger,
	}
}

// Connect implements mcp.Transport.Connect
func (sta *SSETransportAdapter) Connect(ctx context.Context) (mcp.Connection, error) {
	if err := sta.transport.Connect(); err != nil {
		return nil, err
	}

	// Return a connection adapter
	return &SSEConnectionAdapter{
		transport: sta.transport,
		logger:    sta.logger,
	}, nil
}

// SSEConnectionAdapter adapts our SSE connection to MCP SDK Connection interface
type SSEConnectionAdapter struct {
	transport *SSETransport
	logger    *logrus.Logger
}

// Read implements mcp.Connection.Read
func (sca *SSEConnectionAdapter) Read(ctx context.Context) (jsonrpc.Message, error) {
	// SSE transport doesn't support reading (server-sent events are one-way)
	return nil, fmt.Errorf("SSE transport does not support reading")
}

// Write implements mcp.Connection.Write
func (sca *SSEConnectionAdapter) Write(ctx context.Context, msg jsonrpc.Message) error {
	// Convert message to JSON bytes
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = sca.transport.Write(data)
	return err
}

// Close implements mcp.Connection.Close
func (sca *SSEConnectionAdapter) Close() error {
	return sca.transport.Close()
}

// SessionID implements mcp.Connection.SessionID
func (sca *SSEConnectionAdapter) SessionID() string {
	return "sse-session"
}

// Request represents an MCP JSON-RPC request
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Response represents an MCP JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an MCP JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RequestHandler defines the interface for handling MCP requests
type RequestHandler interface {
	HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error)
}

// RequestDispatcher handles routing of MCP requests to appropriate handlers
type RequestDispatcher struct {
	handlers map[string]RequestHandler
	logger   *logrus.Logger
}

// NewRequestDispatcher creates a new request dispatcher
func NewRequestDispatcher(logger *logrus.Logger) *RequestDispatcher {
	return &RequestDispatcher{
		handlers: make(map[string]RequestHandler),
		logger:   logger,
	}
}

// RegisterHandler registers a handler for a specific method
func (rd *RequestDispatcher) RegisterHandler(method string, handler RequestHandler) {
	rd.handlers[method] = handler
	rd.logger.Infof("Registered handler for method: %s", method)
}

// Dispatch dispatches a request to the appropriate handler
func (rd *RequestDispatcher) Dispatch(ctx context.Context, req *Request) *Response {
	rd.logger.Debugf("Dispatching request: %s", req.Method)

	handler, exists := rd.handlers[req.Method]
	if !exists {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}

	result, err := handler.HandleRequest(ctx, req.Method, req.Params)
	if err != nil {
		rd.logger.Errorf("Handler error for method %s: %v", req.Method, err)
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// MessageProcessor handles the processing of MCP messages
type MessageProcessor struct {
	dispatcher *RequestDispatcher
	transport  Transport
	logger     *logrus.Logger
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(dispatcher *RequestDispatcher, transport Transport, logger *logrus.Logger) *MessageProcessor {
	return &MessageProcessor{
		dispatcher: dispatcher,
		transport:  transport,
		logger:     logger,
	}
}

// ProcessMessages processes incoming MCP messages in a loop
func (mp *MessageProcessor) ProcessMessages(ctx context.Context) error {
	mp.logger.Info("Starting MCP message processing")

	scanner := bufio.NewScanner(mp.transport)
	for {
		select {
		case <-ctx.Done():
			mp.logger.Info("Message processing stopped due to context cancellation")
			return ctx.Err()
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					mp.logger.Errorf("Scanner error: %v", err)
					return fmt.Errorf("scanner error: %w", err)
				}
				// EOF reached
				mp.logger.Info("EOF reached, stopping message processing")
				return nil
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			if err := mp.processMessage(ctx, line); err != nil {
				mp.logger.Errorf("Error processing message: %v", err)
				// Continue processing other messages
			}
		}
	}
}

// processMessage processes a single MCP message
func (mp *MessageProcessor) processMessage(ctx context.Context, message string) error {
	mp.logger.Debugf("Processing message: %s", message)

	var req Request
	if err := json.Unmarshal([]byte(message), &req); err != nil {
		mp.logger.Errorf("Failed to unmarshal request: %v", err)
		return mp.sendErrorResponse(nil, -32700, "Parse error", err.Error())
	}

	if req.JSONRPC != "2.0" {
		return mp.sendErrorResponse(req.ID, -32600, "Invalid Request", "jsonrpc version must be 2.0")
	}

	if req.Method == "" {
		return mp.sendErrorResponse(req.ID, -32600, "Invalid Request", "method is required")
	}

	response := mp.dispatcher.Dispatch(ctx, &req)

	return mp.sendResponse(response)
}

// sendResponse sends a response back through the transport
func (mp *MessageProcessor) sendResponse(resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		mp.logger.Errorf("Failed to marshal response: %v", err)
		return err
	}

	data = append(data, '\n') // Add newline for line-based protocol
	_, err = mp.transport.Write(data)
	return err
}

// sendErrorResponse sends an error response
func (mp *MessageProcessor) sendErrorResponse(id interface{}, code int, message string, data ...interface{}) error {
	resp := &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	if len(data) > 0 {
		resp.Error.Data = data[0]
	}

	return mp.sendResponse(resp)
}
