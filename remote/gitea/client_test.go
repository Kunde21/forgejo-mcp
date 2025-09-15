package gitea

import (
	"github.com/kunde21/forgejo-mcp/remote"
)

// This is a compile-time check - if the methods don't exist, this won't compile
var _ remote.ClientInterface = (*GiteaClient)(nil)
