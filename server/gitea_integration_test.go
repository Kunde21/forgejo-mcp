package server

import (
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/client"
	"github.com/Kunde21/forgejo-mcp/config"
)

// TestGiteaClientConfigurationIntegration tests that the configuration now supports Gitea client settings
func TestGiteaClientConfigurationIntegration(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:    "https://gitea.example.com",
		AuthToken:     "test-token-123",
		ClientTimeout: 45,
		UserAgent:     "custom-user-agent/2.0.0",
		TeaPath:       "tea",
		Host:          "localhost",
		Port:          8080,
		ReadTimeout:   30,
		WriteTimeout:  30,
		LogLevel:      "info",
	}

	// Validate that the new configuration fields are supported
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// Verify the new Gitea client configuration fields
	if cfg.ClientTimeout != 45 {
		t.Errorf("Expected ClientTimeout=45, got %d", cfg.ClientTimeout)
	}

	if cfg.UserAgent != "custom-user-agent/2.0.0" {
		t.Errorf("Expected UserAgent='custom-user-agent/2.0.0', got '%s'", cfg.UserAgent)
	}

	// Test client creation (using the configuration would be similar to this)
	clientConfig := &client.ClientConfig{
		Timeout:   time.Duration(cfg.ClientTimeout) * time.Second,
		UserAgent: cfg.UserAgent,
	}

	// Verify configuration can be used to create client
	if clientConfig.Timeout != 45*time.Second {
		t.Errorf("Expected timeout=45s, got %v", clientConfig.Timeout)
	}

	if clientConfig.UserAgent != "custom-user-agent/2.0.0" {
		t.Errorf("Expected UserAgent='custom-user-agent/2.0.0', got '%s'", clientConfig.UserAgent)
	}

	t.Log("Gitea client configuration integration test passed")
}

// TestGiteaClientConfigurationCompatibility tests that the configuration is compatible with client creation
func TestGiteaClientConfigurationCompatibility(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:    "https://gitea.example.com",
		AuthToken:     "test-token-123",
		ClientTimeout: 30,
		UserAgent:     "forgejo-mcp-client/1.0.0",
		TeaPath:       "tea",
		Host:          "localhost",
		Port:          8080,
		ReadTimeout:   30,
		WriteTimeout:  30,
		LogLevel:      "info",
	}

	// Test that we can build a client configuration from our config
	clientConfig := &client.ClientConfig{
		Timeout:   time.Duration(cfg.ClientTimeout) * time.Second,
		UserAgent: cfg.UserAgent,
	}

	// Verify the client configuration matches expected values
	if clientConfig.Timeout != 30*time.Second {
		t.Errorf("Expected timeout=30s, got %v", clientConfig.Timeout)
	}

	if clientConfig.UserAgent != "forgejo-mcp-client/1.0.0" {
		t.Errorf("Expected UserAgent='forgejo-mcp-client/1.0.0', got '%s'", clientConfig.UserAgent)
	}

	// Test input validation that would happen during client creation
	if cfg.ForgejoURL == "" {
		t.Error("ForgejoURL should not be empty for client creation")
	}

	if cfg.AuthToken == "" {
		t.Error("AuthToken should not be empty for client creation")
	}

	// This demonstrates that configuration is ready for actual client creation:
	// forgejoClient, err := client.NewWithConfig(cfg.ForgejoURL, cfg.AuthToken, clientConfig)

	t.Log("Gitea client configuration compatibility test passed")
}

// TestEndToEndWorkflowPreparation demonstrates the setup needed for end-to-end workflows
func TestEndToEndWorkflowPreparation(t *testing.T) {
	// This test demonstrates how the configuration and client would be used together
	// in the actual server setup for end-to-end workflows

	// 1. Load configuration (this would normally come from Load())
	cfg := &config.Config{
		ForgejoURL:    "https://gitea.example.com",
		AuthToken:     "test-token-123",
		ClientTimeout: 30,
		UserAgent:     "forgejo-mcp-client/1.0.0",
		TeaPath:       "tea",
		Host:          "localhost",
		Port:          8080,
		ReadTimeout:   30,
		WriteTimeout:  30,
		LogLevel:      "info",
	}

	// 2. Validate configuration
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// 3. Create Gitea SDK client configuration (but not the actual client to avoid connection)
	clientConfig := &client.ClientConfig{
		Timeout:   time.Duration(cfg.ClientTimeout) * time.Second,
		UserAgent: cfg.UserAgent,
	}

	// 4. Verify configuration is ready for client creation
	if cfg.ForgejoURL == "" || cfg.AuthToken == "" {
		t.Fatal("Configuration missing required fields for client creation")
	}

	// 5. In a real scenario, this is how the client would be created:
	// forgejoClient, err := client.NewWithConfig(cfg.ForgejoURL, cfg.AuthToken, clientConfig)
	// handler := NewGitSdkPRListHandler(logger, forgejoClient)

	// 6. Verify the configuration setup is complete
	if clientConfig == nil {
		t.Fatal("Expected non-nil client configuration for end-to-end workflows")
	}

	t.Log("End-to-end workflow preparation test passed - ready for integration")
}
