// Package tea provides a wrapper for the Gitea SDK to interact with Forgejo repositories
package tea

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
)

type GiteaWrapper struct {
	client *gitea.Client
}

type AuthType int

const (
	AuthTypeToken AuthType = iota
	AuthTypeBasic
)

type AuthConfig struct {
	Type     AuthType
	Token    string
	Username string
	Password string
}

func newGiteaClient(baseURL string, authConfig *AuthConfig) (*gitea.Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if authConfig == nil {
		return nil, fmt.Errorf("authConfig cannot be nil")
	}

	var authOpt gitea.ClientOption
	switch authConfig.Type {
	case AuthTypeToken:
		if authConfig.Token == "" {
			return nil, fmt.Errorf("token cannot be empty for token authentication")
		}
		authOpt = gitea.SetToken(authConfig.Token)
	case AuthTypeBasic:
		if authConfig.Username == "" {
			return nil, fmt.Errorf("username cannot be empty for basic authentication")
		}
		if authConfig.Password == "" {
			return nil, fmt.Errorf("password cannot be empty for basic authentication")
		}
		authOpt = gitea.SetBasicAuth(authConfig.Username, authConfig.Password)
	default:
		return nil, fmt.Errorf("unsupported authentication type: %d", authConfig.Type)
	}

	client, err := gitea.NewClient(baseURL, authOpt,
		gitea.SetHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		gitea.SetUserAgent("forgejo-mcp/1.0.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}
	return client, nil
}

func (w *GiteaWrapper) Initialize(baseURL, token string) error {
	client, err := newGiteaClient(baseURL, &AuthConfig{Type: AuthTypeToken, Token: token})
	if err != nil {
		return err
	}
	w.client = client
	return nil
}

func (w *GiteaWrapper) InitializeWithAuth(baseURL string, authConfig *AuthConfig) error {
	client, err := newGiteaClient(baseURL, authConfig)
	if err != nil {
		return err
	}
	w.client = client
	return nil
}

// InitializeWithFallback initializes the GiteaWrapper with fallback authentication
// It tries the primary auth config first, and if that fails, tries the fallback
func (w *GiteaWrapper) InitializeWithFallback(baseURL string, primaryAuth, fallbackAuth *AuthConfig) error {
	err := w.InitializeWithAuth(baseURL, primaryAuth)
	if err == nil {
		return nil
	}
	if fallbackAuth != nil {
		fallbackErr := w.InitializeWithAuth(baseURL, fallbackAuth)
		if fallbackErr == nil {
			return nil
		}
		return fmt.Errorf("primary auth failed: %w; fallback auth failed: %w", err, fallbackErr)
	}
	return fmt.Errorf("auth failed: %w", err)
}

func (w *GiteaWrapper) IsInitialized() bool {
	return w.client != nil
}
func (w *GiteaWrapper) Ping(ctx context.Context) error {
	if !w.IsInitialized() {
		return fmt.Errorf("wrapper not initialized")
	}
	_, _, err := w.client.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to validate connection: %w", err)
	}
	return nil
}

// ListRepositories lists repositories with optional filters
func (w *GiteaWrapper) ListRepositories(ctx context.Context, filters *RepositoryFilters) ([]*gitea.Repository, *gitea.Response, error) {
	if !w.IsInitialized() {
		return nil, nil, fmt.Errorf("wrapper not initialized")
	}

	// If we have search filters, use SearchRepos instead of ListMyRepos
	if filters != nil && filters.Query != "" {
		opts := buildSearchRepoOptions(filters)
		return w.client.SearchRepos(*opts)
	}

	// Use ListMyRepos for non-search cases
	opts := buildRepoListOptions(filters)
	return w.client.ListMyRepos(*opts)
}

// GetRepository gets a specific repository by owner and name
func (w *GiteaWrapper) GetRepository(ctx context.Context, owner, name string) (*gitea.Repository, *gitea.Response, error) {
	if !w.IsInitialized() {
		return nil, nil, fmt.Errorf("wrapper not initialized")
	}
	return w.client.GetRepo(owner, name)
}
