package config

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// SDKClientFactory handles creation of Gitea SDK clients
type SDKClientFactory struct {
	config *Config
}

// NewSDKClientFactory creates a new SDK client factory
func NewSDKClientFactory(config *Config) *SDKClientFactory {
	return &SDKClientFactory{
		config: config,
	}
}

// CreateClient creates a new Gitea SDK client with token-based authentication
func (f *SDKClientFactory) CreateClient() (*gitea.Client, error) {
	if f.config.ForgejoURL == "" {
		return nil, fmt.Errorf("ForgejoURL is required for SDK client creation")
	}

	if f.config.AuthToken == "" {
		return nil, fmt.Errorf("AuthToken is required for SDK client authentication")
	}

	// Create client with token authentication
	client, err := gitea.NewClient(f.config.ForgejoURL, gitea.SetToken(f.config.AuthToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea SDK client: %w", err)
	}

	return client, nil
}

// ValidateConfiguration validates that the configuration is suitable for SDK client creation
func (f *SDKClientFactory) ValidateConfiguration() error {
	if f.config.ForgejoURL == "" {
		return fmt.Errorf("ForgejoURL cannot be empty")
	}

	if f.config.AuthToken == "" {
		return fmt.Errorf("AuthToken cannot be empty")
	}

	// Additional validation could be added here
	// For example, URL format validation, token format validation, etc.

	return nil
}
