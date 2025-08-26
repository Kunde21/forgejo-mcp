// Package server implements the MCP server functionality for Forgejo repositories
package server

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ToolDefinition represents a tool definition with its metadata and schema
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolRegistry manages the registration and discovery of MCP tools
type ToolRegistry struct {
	tools  map[string]*ToolDefinition
	logger *logrus.Logger
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(logger *logrus.Logger) *ToolRegistry {
	return &ToolRegistry{
		tools:  make(map[string]*ToolDefinition),
		logger: logger,
	}
}

// RegisterTool registers a new tool in the registry
func (tr *ToolRegistry) RegisterTool(tool *ToolDefinition) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if tool.Description == "" {
		return fmt.Errorf("tool description cannot be empty")
	}
	if tool.InputSchema == nil {
		return fmt.Errorf("tool input schema cannot be nil")
	}

	tr.tools[tool.Name] = tool
	tr.logger.Infof("Registered tool: %s", tool.Name)
	return nil
}

// GetTool retrieves a tool definition by name
func (tr *ToolRegistry) GetTool(name string) (*ToolDefinition, bool) {
	tool, exists := tr.tools[name]
	return tool, exists
}

// GetAllTools returns all registered tools
func (tr *ToolRegistry) GetAllTools() []*ToolDefinition {
	tools := make([]*ToolDefinition, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolNames returns a list of all registered tool names
func (tr *ToolRegistry) GetToolNames() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}

// ToolValidator validates tool parameters against their schema
type ToolValidator struct {
	logger *logrus.Logger
}

// NewToolValidator creates a new tool validator
func NewToolValidator(logger *logrus.Logger) *ToolValidator {
	return &ToolValidator{
		logger: logger,
	}
}

// ValidateParameters validates the parameters for a given tool
func (tv *ToolValidator) ValidateParameters(toolName string, params map[string]interface{}) error {
	tv.logger.Debugf("Validating parameters for tool %s: %v", toolName, params)

	// Basic validation - can be extended with more sophisticated schema validation
	switch toolName {
	case "pr_list":
		return tv.validatePRListParams(params)
	case "issue_list":
		return tv.validateIssueListParams(params)
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

// validatePRListParams validates parameters for pr_list tool
func (tv *ToolValidator) validatePRListParams(params map[string]interface{}) error {
	// Validate state parameter if provided
	if state, exists := params["state"]; exists {
		if stateStr, ok := state.(string); ok {
			validStates := []string{"open", "closed", "merged", "all"}
			for _, validState := range validStates {
				if stateStr == validState {
					return nil
				}
			}
			return fmt.Errorf("invalid state parameter: %s, must be one of: %v", stateStr, validStates)
		}
		return fmt.Errorf("state parameter must be a string")
	}

	// Validate author parameter if provided
	if author, exists := params["author"]; exists {
		if authorStr, ok := author.(string); ok {
			if authorStr == "" {
				return fmt.Errorf("author parameter cannot be empty")
			}
			// Basic validation for username format
			if len(authorStr) < 1 || len(authorStr) > 255 {
				return fmt.Errorf("author parameter must be between 1 and 255 characters")
			}
		} else {
			return fmt.Errorf("author parameter must be a string")
		}
	}

	// Validate limit parameter if provided
	if limit, exists := params["limit"]; exists {
		if limitNum, ok := limit.(float64); ok {
			if limitNum < 1 || limitNum > 100 {
				return fmt.Errorf("limit parameter must be between 1 and 100")
			}
		} else {
			return fmt.Errorf("limit parameter must be a number")
		}
	}

	return nil
}

// validateIssueListParams validates parameters for issue_list tool
func (tv *ToolValidator) validateIssueListParams(params map[string]interface{}) error {
	// Validate state parameter if provided
	if state, exists := params["state"]; exists {
		if stateStr, ok := state.(string); ok {
			validStates := []string{"open", "closed", "all"}
			for _, validState := range validStates {
				if stateStr == validState {
					return nil
				}
			}
			return fmt.Errorf("invalid state parameter: %s, must be one of: %v", stateStr, validStates)
		}
		return fmt.Errorf("state parameter must be a string")
	}

	// Validate labels parameter if provided
	if labels, exists := params["labels"]; exists {
		if labelsArray, ok := labels.([]interface{}); ok {
			if len(labelsArray) > 10 {
				return fmt.Errorf("labels parameter cannot contain more than 10 labels")
			}
			for _, label := range labelsArray {
				if labelStr, ok := label.(string); ok {
					if labelStr == "" {
						return fmt.Errorf("label cannot be empty")
					}
					if len(labelStr) > 50 {
						return fmt.Errorf("label cannot be longer than 50 characters")
					}
				} else {
					return fmt.Errorf("labels parameter must contain only strings")
				}
			}
		} else {
			return fmt.Errorf("labels parameter must be an array")
		}
	}

	// Validate author parameter if provided
	if author, exists := params["author"]; exists {
		if authorStr, ok := author.(string); ok {
			if authorStr == "" {
				return fmt.Errorf("author parameter cannot be empty")
			}
			if len(authorStr) < 1 || len(authorStr) > 255 {
				return fmt.Errorf("author parameter must be between 1 and 255 characters")
			}
		} else {
			return fmt.Errorf("author parameter must be a string")
		}
	}

	// Validate limit parameter if provided
	if limit, exists := params["limit"]; exists {
		if limitNum, ok := limit.(float64); ok {
			if limitNum < 1 || limitNum > 100 {
				return fmt.Errorf("limit parameter must be between 1 and 100")
			}
		} else {
			return fmt.Errorf("limit parameter must be a number")
		}
	}

	return nil
}

// CreateDefaultTools creates and returns the default tool definitions
func CreateDefaultTools() []*ToolDefinition {
	return []*ToolDefinition{
		{
			Name:        "pr_list",
			Description: "List pull requests from the Forgejo repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Filter by PR state (open, closed, merged, all)",
						"enum":        []string{"open", "closed", "merged", "all"},
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Filter by PR author username",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of PRs to return",
						"minimum":     1,
						"maximum":     100,
					},
				},
			},
		},
		{
			Name:        "issue_list",
			Description: "List issues from the Forgejo repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Filter by issue state (open, closed, all)",
						"enum":        []string{"open", "closed", "all"},
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"description": "Filter by issue labels",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Filter by issue author username",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of issues to return",
						"minimum":     1,
						"maximum":     100,
					},
				},
			},
		},
	}
}

// RegisterDefaultTools registers all default tools in the registry
func (tr *ToolRegistry) RegisterDefaultTools() error {
	tools := CreateDefaultTools()
	for _, tool := range tools {
		if err := tr.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}
	return nil
}

// ToolManifestGenerator generates tool manifests for client discovery
type ToolManifestGenerator struct {
	registry *ToolRegistry
	logger   *logrus.Logger
}

// NewToolManifestGenerator creates a new tool manifest generator
func NewToolManifestGenerator(registry *ToolRegistry, logger *logrus.Logger) *ToolManifestGenerator {
	return &ToolManifestGenerator{
		registry: registry,
		logger:   logger,
	}
}

// GenerateManifest generates a complete tool manifest
func (tmg *ToolManifestGenerator) GenerateManifest() map[string]interface{} {
	tools := tmg.registry.GetAllTools()

	toolManifests := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		toolManifests = append(toolManifests, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	tmg.logger.Debugf("Generated manifest with %d tools", len(toolManifests))
	return map[string]interface{}{
		"tools": toolManifests,
	}
}

// InitializeServerTools initializes the tool system for the MCP server
func (s *Server) InitializeToolSystem() error {
	s.logger.Info("Initializing tool system...")

	// Create tool registry
	toolRegistry := NewToolRegistry(s.logger)

	// Register default tools
	if err := toolRegistry.RegisterDefaultTools(); err != nil {
		return fmt.Errorf("failed to register default tools: %w", err)
	}

	// Create tool validator
	toolValidator := NewToolValidator(s.logger)

	// Create tool manifest generator
	manifestGenerator := NewToolManifestGenerator(toolRegistry, s.logger)

	// Store tool registry in server
	s.toolRegistry = toolRegistry

	// Create a composite handler that uses these components
	toolHandler := &ToolSystemHandler{
		registry:          toolRegistry,
		validator:         toolValidator,
		manifestGenerator: manifestGenerator,
		prHandler:         NewPRListHandler(s.logger),
		issueHandler:      NewIssueListHandler(s.logger),
		logger:            s.logger,
	}

	// Register the tool system handler
	s.dispatcher.RegisterHandler("tools/call", toolHandler)
	s.dispatcher.RegisterHandler("tools/list", NewToolManifestHandler(s.logger))

	s.logger.Info("Tool system initialized successfully")
	return nil
}

// RegisterTool registers a new tool in the server's tool registry
func (s *Server) RegisterTool(tool *ToolDefinition) error {
	if s.toolRegistry == nil {
		return fmt.Errorf("tool system not initialized")
	}
	return s.toolRegistry.RegisterTool(tool)
}

// RegisterTools registers multiple tools in the server's tool registry
func (s *Server) RegisterTools(tools []*ToolDefinition) error {
	if s.toolRegistry == nil {
		return fmt.Errorf("tool system not initialized")
	}

	for _, tool := range tools {
		if err := s.toolRegistry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}
	return nil
}

// GetTool retrieves a tool definition by name
func (s *Server) GetTool(name string) (*ToolDefinition, bool) {
	if s.toolRegistry == nil {
		return nil, false
	}
	return s.toolRegistry.GetTool(name)
}

// GetAllTools returns all registered tools
func (s *Server) GetAllTools() []*ToolDefinition {
	if s.toolRegistry == nil {
		return nil
	}
	return s.toolRegistry.GetAllTools()
}

// GetToolNames returns a list of all registered tool names
func (s *Server) GetToolNames() []string {
	if s.toolRegistry == nil {
		return nil
	}
	return s.toolRegistry.GetToolNames()
}

// ToolManifest returns the complete tool manifest for client discovery
func (s *Server) ToolManifest() map[string]interface{} {
	if s.toolRegistry == nil {
		s.logger.Warn("Tool registry not initialized, returning empty manifest")
		return map[string]interface{}{
			"tools": []map[string]interface{}{},
		}
	}

	manifestGenerator := NewToolManifestGenerator(s.toolRegistry, s.logger)
	return manifestGenerator.GenerateManifest()
}

// GetToolManifest returns the tool manifest in the format expected by MCP clients
func (s *Server) GetToolManifest() map[string]interface{} {
	return s.ToolManifest()
}

// ToolSystemHandler handles tool calls using the tool registry system
type ToolSystemHandler struct {
	registry          *ToolRegistry
	validator         *ToolValidator
	manifestGenerator *ToolManifestGenerator
	prHandler         *PRListHandler
	issueHandler      *IssueListHandler
	logger            *logrus.Logger
}

// HandleRequest handles a tool call request
func (tsh *ToolSystemHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	tsh.logger.Debugf("Tool system handling request: %s", method)

	// Extract tool name from params
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required and must be a string")
	}

	// Extract tool arguments
	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Validate tool exists
	_, exists := tsh.registry.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Validate parameters
	if err := tsh.validator.ValidateParameters(toolName, arguments); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Route to appropriate handler
	switch toolName {
	case "pr_list":
		return tsh.prHandler.HandleRequest(ctx, toolName, arguments)
	case "issue_list":
		return tsh.issueHandler.HandleRequest(ctx, toolName, arguments)
	default:
		return nil, fmt.Errorf("tool handler not implemented: %s", toolName)
	}
}
