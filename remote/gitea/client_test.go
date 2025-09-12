package gitea

// This is a compile-time check - if the methods don't exist, this won't compile
var _ GiteaClientInterface = (*GiteaClient)(nil)
