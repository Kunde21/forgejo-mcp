package forgejo

import (
	"fmt"
	"net/http"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/kunde21/forgejo-mcp/remote"
)

// ForgejoClient implements the ClientInterface using the Forgejo SDK
type ForgejoClient struct {
	client *forgejo.Client
}

// NewForgejoClient creates a new Forgejo client
func NewForgejoClient(url, token string) (*ForgejoClient, error) {
	return NewForgejoClientWithHTTPClient(url, token, nil)
}

// NewForgejoClientWithHTTPClient creates a new Forgejo client with a custom HTTP client
func NewForgejoClientWithHTTPClient(url, token string, httpClient *http.Client) (*ForgejoClient, error) {
	var client *forgejo.Client
	var err error

	if httpClient != nil {
		client, err = forgejo.NewClient(url, forgejo.SetToken(token), forgejo.SetHTTPClient(httpClient))
	} else {
		client, err = forgejo.NewClient(url, forgejo.SetToken(token))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Forgejo client: %w", err)
	}

	return &ForgejoClient{
		client: client,
	}, nil
}

// Compile-time interface check
var _ remote.ClientInterface = (*ForgejoClient)(nil)
