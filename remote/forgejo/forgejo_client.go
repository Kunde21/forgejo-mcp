package forgejo

import (
	"fmt"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/kunde21/forgejo-mcp/remote"
)

// ForgejoClient implements the ClientInterface using the Forgejo SDK
type ForgejoClient struct {
	client *forgejo.Client
}

// NewForgejoClient creates a new Forgejo client
func NewForgejoClient(url, token string) (*ForgejoClient, error) {
	client, err := forgejo.NewClient(url, forgejo.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Forgejo client: %w", err)
	}

	return &ForgejoClient{
		client: client,
	}, nil
}

// Compile-time interface check
var _ remote.ClientInterface = (*ForgejoClient)(nil)
