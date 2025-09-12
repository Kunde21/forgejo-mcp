package server

import (
	"fmt"
	"regexp"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TextResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}

func TextResultf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}
}

func TextError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}, IsError: true}
}

func TextErrorf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}, IsError: true}
}

var repoReg = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

// nonEmptyString validates that a string is not empty and not only whitespace
func nonEmptyString() v.Rule {
	return v.By(func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("cannot be blank")
		}
		return nil
	})
}
