package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
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

var (
	repoReg  = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
	emptyReg = regexp.MustCompilePOSIX(`[^[:space:]]+`)
)

// ValidateAttachment validates file data, filename, and size
func ValidateAttachment(data []byte, filename string, maxSize int64, allowedTypes []string) error {
	// Size validation
	if int64(len(data)) > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed %d", len(data), maxSize)
	}

	// MIME type validation
	mimeType := http.DetectContentType(data)
	if !isAllowedMimeType(mimeType, allowedTypes) {
		return fmt.Errorf("MIME type %s not allowed", mimeType)
	}

	// Filename validation
	if !isValidFilename(filename) {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	return nil
}

func isAllowedMimeType(mimeType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if allowed == "*" || strings.HasPrefix(mimeType, allowed) {
			return true
		}
	}
	return false
}

func isValidFilename(filename string) bool {
	// Basic filename validation - no path traversal, reasonable length
	clean := filepath.Base(filename)
	return clean == filename && len(clean) > 0 && len(clean) < 255
}

// generateInputSchema generates a JSON schema for the given argument type
func generateInputSchema[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		// Fallback to empty schema if generation fails
		return &jsonschema.Schema{}
	}
	return schema
}

// generateOutputSchema generates a JSON schema for the given result type
func generateOutputSchema[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		// Fallback to empty schema if generation fails
		return &jsonschema.Schema{}
	}
	return schema
}
