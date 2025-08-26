// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
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
