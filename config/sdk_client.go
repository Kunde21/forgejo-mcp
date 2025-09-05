package config

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// CreateGiteaClient creates a new Gitea SDK client with token-based authentication
func (c *Config) CreateGiteaClient() (*gitea.Client, error) {
	if c.ForgejoURL == "" {
		return nil, fmt.Errorf("ForgejoURL is required for SDK client creation")
	}

	if c.AuthToken == "" {
		return nil, fmt.Errorf("AuthToken is required for SDK client authentication")
	}

	// Create client with token authentication
	client, err := gitea.NewClient(c.ForgejoURL, gitea.SetToken(c.AuthToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea SDK client: %w", err)
	}

	return client, nil
}

// ValidateForSDK validates that the configuration is suitable for SDK client creation
func (c *Config) ValidateForSDK() error {
	if c.ForgejoURL == "" {
		return fmt.Errorf("ForgejoURL cannot be empty")
	}

	if c.AuthToken == "" {
		return fmt.Errorf("AuthToken cannot be empty")
	}

	// Additional validation could be added here
	// For example, URL format validation, token format validation, etc.

	return nil
}
