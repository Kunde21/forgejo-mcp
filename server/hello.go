package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HelloArgs represents the arguments for the hello tool
type HelloArgs struct {
	// No arguments required for hello tool
}

// HelloResult represents the result data for the hello tool
type HelloResult struct {
	Message string `json:"message,omitempty"`
}

// handleHello handles the "hello" tool request.
// This is a simple demonstration tool that returns a hello world message.
//
// Migration Note: Updated to use the official SDK's handler signature:
// (context.Context, *mcp.CallToolRequest, args) (*mcp.CallToolResult, any, error)
// instead of the previous SDK's handler pattern.
func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args HelloArgs) (*mcp.CallToolResult, *HelloResult, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}
	// Return successful response with hello message
	// Migration: Uses official SDK's CallToolResult structure
	return TextResult("Hello, World!"), &HelloResult{Message: "Hello, World!"}, nil
}
