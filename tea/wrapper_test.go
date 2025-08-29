package tea

import (
	"context"
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

func TestGiteaWrapper_InitializeWithFallback(t *testing.T) {
	w := &GiteaWrapper{}
	err := w.InitializeWithFallback("https://example.com", &AuthConfig{Type: AuthTypeToken, Token: "valid-token"}, &AuthConfig{Type: AuthTypeBasic, Username: "user", Password: "pass"})
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	w2 := &GiteaWrapper{}
	err2 := w2.InitializeWithFallback("https://example.com", &AuthConfig{Type: AuthTypeToken, Token: "valid-token"}, nil)
	if err2 == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestGiteaWrapper_ListRepositories(t *testing.T) {
	w := &GiteaWrapper{}
	_, _, err := w.ListRepositories(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for uninitialized wrapper, got nil")
	}
}

func TestGiteaWrapper_GetRepository(t *testing.T) {
	w := &GiteaWrapper{}
	_, _, err := w.GetRepository(context.Background(), "owner", "repo")
	if err == nil {
		t.Error("Expected error for uninitialized wrapper, got nil")
	}
}

func TestGiteaWrapper_ListIssues(t *testing.T) {
	w := &GiteaWrapper{}
	_, _, err := w.ListIssues(context.Background(), "owner", "repo", nil)
	if err == nil {
		t.Error("Expected error for uninitialized wrapper, got nil")
	}
}

func TestGiteaWrapper_ListPullRequests(t *testing.T) {
	w := &GiteaWrapper{}
	_, _, err := w.ListPullRequests(context.Background(), "owner", "repo", nil)
	if err == nil {
		t.Error("Expected error for uninitialized wrapper, got nil")
	}
}

func TestGiteaWrapper_Initialize(t *testing.T) {
	w := &GiteaWrapper{}
	err := w.Initialize("https://example.com", "test-token")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with empty token
	w2 := &GiteaWrapper{}
	err2 := w2.Initialize("https://example.com", "")
	if err2 == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestGiteaWrapper_InitializeWithAuth(t *testing.T) {
	w := &GiteaWrapper{}

	// Test with nil auth config
	err := w.InitializeWithAuth("https://example.com", nil)
	if err == nil {
		t.Error("Expected error for nil auth config, got nil")
	}

	// Test with empty URL
	err2 := w.InitializeWithAuth("", &AuthConfig{Type: AuthTypeToken, Token: "test-token"})
	if err2 == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	// Test with valid token auth
	err3 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: AuthTypeToken, Token: "test-token"})
	if err3 == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with valid basic auth
	err4 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: AuthTypeBasic, Username: "user", Password: "pass"})
	if err4 == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with empty token
	err5 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: AuthTypeToken, Token: ""})
	if err5 == nil {
		t.Error("Expected error for empty token, got nil")
	}

	// Test with empty username
	err6 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: AuthTypeBasic, Username: "", Password: "pass"})
	if err6 == nil {
		t.Error("Expected error for empty username, got nil")
	}

	// Test with empty password
	err7 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: AuthTypeBasic, Username: "user", Password: ""})
	if err7 == nil {
		t.Error("Expected error for empty password, got nil")
	}

	// Test with unsupported auth type
	err8 := w.InitializeWithAuth("https://example.com", &AuthConfig{Type: 999})
	if err8 == nil {
		t.Error("Expected error for unsupported auth type, got nil")
	}
}

func TestGiteaWrapper_IsInitialized(t *testing.T) {
	w := &GiteaWrapper{}

	// Should be false when not initialized
	if w.IsInitialized() {
		t.Error("Expected IsInitialized to return false for uninitialized wrapper")
	}

	// Note: We can't easily test the true case without a real Gitea client
}

func TestGiteaWrapper_Ping(t *testing.T) {
	w := &GiteaWrapper{}

	// Should fail when not initialized
	err := w.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for uninitialized wrapper, got nil")
	}

	if !strings.Contains(err.Error(), "wrapper not initialized") {
		t.Errorf("Expected 'wrapper not initialized' error, got %v", err)
	}
}

func TestGiteaWrapper_ProcessBatch(t *testing.T) {
	// Test the processSingleRequest function indirectly through ProcessBatch
	processor := NewBatchProcessor(1)

	// Test with invalid request (missing owner/repo)
	requests := []BatchRequest{
		{ID: "1", Method: "listPRs", Owner: "", Repo: ""}, // Invalid
	}

	ctx := context.Background()
	responses, err := processor.ProcessBatch(ctx, requests)

	if err != nil {
		t.Errorf("ProcessBatch failed: %v", err)
	}

	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}

	if responses[0].Error == nil {
		t.Error("Expected error for invalid request")
	}
}
