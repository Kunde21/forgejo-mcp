package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"

	giteasdk "github.com/Kunde21/forgejo-mcp/remote/gitea"
)

// Handler defines the interface for MCP tool handlers
type Handler interface {
	// Common interface for all handlers
}

// PRListArgs represents arguments for PR list requests
type PRListArgs struct {
	Repository string `json:"repository,omitempty"`
	CWD        string `json:"cwd,omitempty"`
	State      string `json:"state,omitempty"`
	Author     string `json:"author,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// RepoListArgs represents arguments for repository list requests
type RepoListArgs struct {
	Limit int `json:"limit,omitempty"`
}

// IssueListArgs represents arguments for issue list requests
type IssueListArgs struct {
	Repository string   `json:"repository,omitempty"`
	CWD        string   `json:"cwd,omitempty"`
	State      string   `json:"state,omitempty"`
	Author     string   `json:"author,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// MCPResult represents the result of an MCP tool call
type MCPResult struct {
	ToolResult *mcp.CallToolResult
	Data       any
	Error      error
}

// HandlerDependencies contains all dependencies needed by handlers
type HandlerDependencies struct {
	Logger *logrus.Logger
	Client giteasdk.GiteaClientInterface
}

// RepositoryInfo represents repository information
type RepositoryInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	Owner       string `json:"owner"`
	URL         string `json:"url,omitempty"`
}

// PRInfo represents pull request information
type PRInfo struct {
	Number     int64           `json:"number"`
	Title      string          `json:"title"`
	State      string          `json:"state"`
	Author     string          `json:"author"`
	URL        string          `json:"url,omitempty"`
	Type       string          `json:"type"`
	CreatedAt  string          `json:"createdAt,omitempty"`
	UpdatedAt  string          `json:"updatedAt,omitempty"`
	Repository *RepositoryInfo `json:"repository,omitempty"`
}

// IssueInfo represents issue information
type IssueInfo struct {
	Number     int64           `json:"number"`
	Title      string          `json:"title"`
	State      string          `json:"state"`
	Author     string          `json:"author"`
	URL        string          `json:"url,omitempty"`
	Type       string          `json:"type"`
	CreatedAt  string          `json:"createdAt,omitempty"`
	UpdatedAt  string          `json:"updatedAt,omitempty"`
	Repository *RepositoryInfo `json:"repository,omitempty"`
}

// ListResponse represents a generic list response
type ListResponse struct {
	Items []map[string]any `json:"items"`
	Total int              `json:"total"`
}
