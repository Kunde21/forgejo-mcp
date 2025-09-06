package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	config    *Config
}

func NewServer() (*Server, error) {
	s := &Server{
		config: LoadConfig(),
	}

	// Initialize MCP server
	mcpServer := server.NewMCPServer("forgejo-mcp", "1.0.0")

	// Create and register the hello tool
	helloTool := mcp.NewTool("hello",
		mcp.WithDescription("Returns a hello world message"),
	)

	mcpServer.AddTool(helloTool, s.handleHello)

	s.mcpServer = mcpServer
	return s, nil
}

func (s *Server) Start() error {
	// Start the MCP server using stdio transport
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) Stop() error {
	// MCP server doesn't have a direct stop method for stdio
	// It runs until the process ends
	return nil
}

func (s *Server) handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Basic validation - check if request is valid
	if ctx == nil {
		return mcp.NewToolResultError("Context is required"), nil
	}

	// Return the hello world message
	return mcp.NewToolResultText("Hello, World!"), nil
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Println("Starting MCP server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
