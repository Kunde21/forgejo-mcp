package tea

import (
	"strings"
	"testing"
)

func TestNewGiteaClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "invalid url",
			baseURL:     "://invalid",
			token:       "test-token",
			wantErr:     true,
			errContains: "missing protocol scheme",
		},
		{
			name:        "empty url",
			baseURL:     "",
			token:       "test-token",
			wantErr:     true,
			errContains: "baseURL cannot be empty",
		},
		{
			name:        "empty token",
			baseURL:     "https://example.com",
			token:       "",
			wantErr:     true,
			errContains: "token cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authConfig := &AuthConfig{
				Type:  AuthTypeToken,
				Token: tt.token,
			}
			client, err := newGiteaClient(tt.baseURL, authConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("newGiteaClient() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("newGiteaClient() error = %v, wantErrContains %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("newGiteaClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client == nil {
				t.Error("newGiteaClient() client = nil, want not nil")
				return
			}
		})
	}
}
