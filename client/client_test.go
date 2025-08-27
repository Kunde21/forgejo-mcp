package client

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var _ RepositoryLister = (*ForgejoClient)(nil)
var _ Client = (*ForgejoClient)(nil)
var exampleCom *url.URL

func init() {
	exampleCom, _ = url.Parse("https://example.com")
}

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		token   string
		wantErr error
	}{
		{
			name:    "empty baseURL",
			baseURL: "",
			token:   "test-token",
			wantErr: &ValidationError{Message: "baseURL cannot be empty", Field: "baseURL"},
		},
		{
			name:    "empty token",
			baseURL: "https://example.com",
			token:   "",
			wantErr: &ValidationError{Message: "token cannot be empty", Field: "token"},
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			token:   "test-token",
			wantErr: &ValidationError{Message: "invalid baseURL format, must be a valid HTTP/HTTPS URL", Field: "baseURL"},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			client, err := New(tst.baseURL, tst.token)
			if !cmp.Equal(tst.wantErr, err) {
				t.Error(cmp.Diff(tst.wantErr, err))
			}
			if tst.wantErr == nil && client == nil {
				t.Error("expected client to be created, got nil")
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	// Skip this test since it requires a real connection
	t.Skip("Skipping test that requires real connection to Gitea API")
}

func TestClientGetters(t *testing.T) {
	// Skip this test since it requires a real connection
	t.Skip("Skipping test that requires real connection to Gitea API")
}
