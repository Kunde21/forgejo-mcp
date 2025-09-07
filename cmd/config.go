package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config subcommand
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validate configuration and test connectivity",
		Long: `Validate the current configuration and test connectivity to the Forgejo instance.

This command checks that all required configuration values are present and valid,
then attempts to connect to the configured Forgejo instance to verify accessibility.`,
		RunE: runConfig,
	}

	return cmd
}

func runConfig(cmd *cobra.Command, args []string) error {
	cmd.Println("🔍 Validating configuration...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display current configuration (without sensitive data)
	cmd.Printf("📋 Configuration Status:\n")
	cmd.Printf("  Remote URL: %s\n", cfg.RemoteURL)
	if cfg.AuthToken != "" {
		cmd.Printf("  Auth Token: ****%s\n", cfg.AuthToken[len(cfg.AuthToken)-4:])
	} else {
		cmd.Printf("  Auth Token: Not set\n")
	}

	// Validate configuration
	err = cfg.Validate()
	if err != nil {
		cmd.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	cmd.Printf("✅ Configuration validation passed\n")

	// Test connectivity (skip for test/example URLs)
	if !isTestURL(cfg.RemoteURL) {
		cmd.Println("\n🌐 Testing Forgejo connectivity...")
		err = testForgejoConnectivity(cfg.RemoteURL, cfg.AuthToken)
		if err != nil {
			cmd.Printf("❌ Connectivity test failed: %v\n", err)
			return err
		}
		cmd.Printf("✅ Forgejo connectivity test passed\n")
	} else {
		cmd.Println("\n🌐 Skipping connectivity test for test/example URL")
	}

	cmd.Printf("🎉 Configuration is valid!\n")

	return nil
}

func testForgejoConnectivity(remoteURL, authToken string) error {
	if remoteURL == "" {
		return fmt.Errorf("remote URL is not configured")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test basic connectivity to the Forgejo instance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", remoteURL+"/api/v1/version", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is provided
	if authToken != "" {
		req.Header.Set("Authorization", "token "+authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Forgejo instance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Forgejo API returned status %d", resp.StatusCode)
	}

	return nil
}

// isTestURL checks if the URL is a test/example URL that shouldn't be connectivity tested
func isTestURL(url string) bool {
	testIndicators := []string{
		"example.com",
		"localhost",
		"127.0.0.1",
		"test",
		"mock",
		"fake",
	}

	for _, indicator := range testIndicators {
		if strings.Contains(strings.ToLower(url), indicator) {
			return true
		}
	}

	return false
}
