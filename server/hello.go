package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// handleHello handles the "hello" tool request.
// This is a simple demonstration tool that returns a hello world message.
//
// Migration Note: Updated to use the official SDK's handler signature:
// (context.Context, *mcp.CallToolRequest, args) (*mcp.CallToolResult, any, error)
// instead of the previous SDK's handler pattern.
func (s *Server) handleHello(ctx context.Context, request *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	// Validate context - required for proper request handling
	if ctx == nil {
		return TextError("Context is required"), nil, nil
	}
	// Return successful response with hello message
	// Migration: Uses official SDK's CallToolResult structure
	return TextResult("Hello, World!"), nil, nil
}
