package cmd

import (
	"strings"
	"testing"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/Kunde21/forgejo-mcp/server"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestServeCommand_ValidConfig(t *testing.T) {
	// Test that serve command can be created and has expected flags
	cmd := serveCmd

	// Check command properties
	if diff := cmp.Diff("serve", cmd.Use); diff != "" {
		t.Errorf("cmd.Use mismatch (-want +got):\n%s", diff)
	}
	if !containsString(cmd.Short, "MCP server") {
		t.Errorf("cmd.Short does not contain 'MCP server', got: %s", cmd.Short)
	}
	if !containsString(cmd.Long, "Model Context Protocol") {
		t.Errorf("cmd.Long does not contain 'Model Context Protocol', got: %s", cmd.Long)
	}

	// Check flags exist
	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Fatal("host flag not found")
	}
	if diff := cmp.Diff("localhost", hostFlag.DefValue); diff != "" {
		t.Errorf("host flag DefValue mismatch (-want +got):\n%s", diff)
	}

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("port flag not found")
	}
	if diff := cmp.Diff("3000", portFlag.DefValue); diff != "" {
		t.Errorf("port flag DefValue mismatch (-want +got):\n%s", diff)
	}

	logLevelFlag := cmd.Flags().Lookup("log-level")
	if logLevelFlag == nil {
		t.Fatal("log-level flag not found")
	}
	if diff := cmp.Diff("info", logLevelFlag.DefValue); diff != "" {
		t.Errorf("log-level flag DefValue mismatch (-want +got):\n%s", diff)
	}

	debugFlag := cmd.Flags().Lookup("debug")
	if debugFlag == nil {
		t.Fatal("debug flag not found")
	}
	if diff := cmp.Diff("false", debugFlag.DefValue); diff != "" {
		t.Errorf("debug flag DefValue mismatch (-want +got):\n%s", diff)
	}

	transportFlag := cmd.Flags().Lookup("transport")
	if transportFlag == nil {
		t.Fatal("transport flag not found")
	}
	if diff := cmp.Diff("stdio", transportFlag.DefValue); diff != "" {
		t.Errorf("transport flag DefValue mismatch (-want +got):\n%s", diff)
	}
}

func TestServeCommand_PreRunE_ValidInputs(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      int
		transport string
		logLevel  string
		wantErr   bool
	}{
		{
			name:      "valid localhost",
			host:      "localhost",
			port:      8080,
			transport: "stdio",
			logLevel:  "info",
			wantErr:   false,
		},
		{
			name:      "valid IP address",
			host:      "127.0.0.1",
			port:      3000,
			transport: "sse",
			logLevel:  "debug",
			wantErr:   false,
		},
		{
			name:      "valid hostname",
			host:      "example.com",
			port:      1,
			transport: "stdio",
			logLevel:  "warn",
			wantErr:   false,
		},
		{
			name:      "valid max port",
			host:      "localhost",
			port:      65535,
			transport: "sse",
			logLevel:  "error",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables to test values
			host = tt.host
			port = tt.port
			transport = tt.transport
			logLevel = tt.logLevel

			// Test PreRunE validation
			err := serveCmd.PreRunE(serveCmd, []string{})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestServeCommand_PreRunE_InvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      int
		transport string
		logLevel  string
	}{
		{
			name:      "empty host",
			host:      "",
			port:      8080,
			transport: "stdio",
			logLevel:  "info",
		},
		{
			name:      "port too low",
			host:      "localhost",
			port:      0,
			transport: "stdio",
			logLevel:  "info",
		},
		{
			name:      "port too high",
			host:      "localhost",
			port:      65536,
			transport: "stdio",
			logLevel:  "info",
		},
		{
			name:      "invalid transport",
			host:      "localhost",
			port:      8080,
			transport: "invalid",
			logLevel:  "info",
		},
		{
			name:      "invalid log level",
			host:      "localhost",
			port:      8080,
			transport: "stdio",
			logLevel:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables to test values
			host = tt.host
			port = tt.port
			transport = tt.transport
			logLevel = tt.logLevel

			// Test PreRunE validation
			err := serveCmd.PreRunE(serveCmd, []string{})
			if err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestServerCreation(t *testing.T) {
	// Test that we can create a server with valid configuration
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	srv, err := server.NewMCPServer(cfg)
	if err != nil {
		t.Fatalf("expected no error creating MCP server, got: %v", err)
	}
	if srv == nil {
		t.Fatal("expected server to be non-nil")
	}

	// Test that we can get the MCP server
	mcpSrv := srv.GetMCPServer()
	if mcpSrv == nil {
		t.Fatal("expected MCP server to be non-nil")
	}

	// Test that we can get the logger
	logger := srv.GetLogger()
	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}

	// Test that we can get the config
	serverCfg := srv.GetConfig()
	if serverCfg == nil {
		t.Fatal("expected config to be non-nil")
	}
}

func TestHealthCheckHandler(t *testing.T) {
	// Test health check handler directly
	logger := logrus.New()
	handler := server.NewHealthCheckHandler(logger)

	result, err := handler.HandleRequest(t.Context(), "health/check", map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be map[string]interface{}, got %T", result)
	}

	if diff := cmp.Diff("healthy", response["status"]); diff != "" {
		t.Errorf("status mismatch (-want +got):\n%s", diff)
	}

	// Check that timestamp exists
	if _, exists := response["timestamp"]; !exists {
		t.Error("expected timestamp in response")
	}

	// Check that version exists
	if _, exists := response["version"]; !exists {
		t.Error("expected version in response")
	}

	if diff := cmp.Diff("1.0.0", response["version"]); diff != "" {
		t.Errorf("version mismatch (-want +got):\n%s", diff)
	}
}
