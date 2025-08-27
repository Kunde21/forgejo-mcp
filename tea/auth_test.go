package tea

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInitializeWithAuth(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		authConfig *AuthConfig
		wantErr    error
	}{
		{
			name:    "invalid URL with token auth",
			baseURL: "://invalid",
			authConfig: &AuthConfig{
				Type:  AuthTypeToken,
				Token: "test-token",
			},
			wantErr: errors.New("failed to create Gitea client: parse \"://invalid/api/v1/version\": missing protocol scheme"),
		},
		{
			name:    "invalid URL with basic auth",
			baseURL: "://invalid",
			authConfig: &AuthConfig{
				Type:     AuthTypeBasic,
				Username: "test-user",
				Password: "test-password",
			},
			wantErr: errors.New("failed to create Gitea client: parse \"://invalid/api/v1/version\": missing protocol scheme"),
		},
		{
			name:       "nil auth config",
			baseURL:    "https://example.com",
			authConfig: nil,
			wantErr:    errors.New("authConfig cannot be nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &GiteaWrapper{}
			err := wrapper.InitializeWithAuth(tt.baseURL, tt.authConfig)
			if !cmp.Equal(tt.wantErr.Error(), err.Error()) {
				t.Error(cmp.Diff(tt.wantErr.Error(), err.Error()))
			}
		})
	}
}

func TestInitializeWithFallback(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		primaryAuth  *AuthConfig
		fallbackAuth *AuthConfig
		wantErr      error
	}{
		{
			name:    "invalid URL with both auth methods",
			baseURL: "://invalid",
			primaryAuth: &AuthConfig{
				Type:  AuthTypeToken,
				Token: "invalid-token",
			},
			fallbackAuth: &AuthConfig{
				Type:     AuthTypeBasic,
				Username: "invalid-user",
				Password: "invalid-password",
			},
			wantErr: errors.New("primary auth failed: failed to create Gitea client: parse \"://invalid/api/v1/version\": missing protocol scheme; fallback auth failed: failed to create Gitea client: parse \"://invalid/api/v1/version\": missing protocol scheme"),
		},
		{
			name:    "invalid URL with primary auth only",
			baseURL: "://invalid",
			primaryAuth: &AuthConfig{
				Type:  AuthTypeToken,
				Token: "invalid-token",
			},
			fallbackAuth: nil,
			wantErr:      errors.New("auth failed: failed to create Gitea client: parse \"://invalid/api/v1/version\": missing protocol scheme"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &GiteaWrapper{}
			err := wrapper.InitializeWithFallback(tt.baseURL, tt.primaryAuth, tt.fallbackAuth)
			if !cmp.Equal(tt.wantErr.Error(), err.Error()) {
				t.Error(cmp.Diff(tt.wantErr.Error(), err.Error()))
			}
		})
	}
}

func TestAuthConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		authConfig  *AuthConfig
		baseURL     string
		wantErr     error
		errContains string
	}{
		{
			name:        "nil auth config",
			authConfig:  nil,
			baseURL:     "https://example.com",
			wantErr:     errors.New("authConfig cannot be nil"),
			errContains: "authConfig cannot be nil",
		},
		{
			name:        "token auth with empty token",
			authConfig:  &AuthConfig{Type: AuthTypeToken},
			baseURL:     "https://example.com",
			wantErr:     errors.New("token cannot be empty for token authentication"),
			errContains: "token cannot be empty for token authentication",
		},
		{
			name:        "basic auth with empty username",
			authConfig:  &AuthConfig{Type: AuthTypeBasic, Password: "password"},
			baseURL:     "https://example.com",
			wantErr:     errors.New("username cannot be empty for basic authentication"),
			errContains: "username cannot be empty for basic authentication",
		},
		{
			name:        "basic auth with empty password",
			authConfig:  &AuthConfig{Type: AuthTypeBasic, Username: "username"},
			baseURL:     "https://example.com",
			wantErr:     errors.New("password cannot be empty for basic authentication"),
			errContains: "password cannot be empty for basic authentication",
		},
		{
			name:        "unsupported auth type",
			authConfig:  &AuthConfig{Type: 99},
			baseURL:     "https://example.com",
			wantErr:     errors.New("unsupported authentication type: 99"),
			errContains: "unsupported authentication type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newGiteaClient(tt.baseURL, tt.authConfig)
			if !cmp.Equal(tt.wantErr.Error(), err.Error()) {
				t.Error(cmp.Diff(tt.wantErr.Error(), err.Error()))
			}
		})
	}
}
