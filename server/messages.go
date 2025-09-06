package server

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TextResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func TextError(err error) *mcp.CallToolResult { return TextResult(fmt.Sprintf("Error %v", err)) }
