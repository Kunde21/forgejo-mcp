package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kunde21/forgejo-mcp/config"
	"github.com/kunde21/forgejo-mcp/server"
)

// mockHTTPClient creates a mock HTTP client that doesn't make real requests
func mockHTTPClient() *http.Client {
	// Create a custom transport that doesn't make real requests
	transport := &mockTransport{}
	return &http.Client{Transport: transport}
}

// mockTransport implements http.RoundTripper but doesn't make real requests
type mockTransport struct{}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Return a mock response for any request
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}

func TestNewFromConfig_GiteaClientType(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "1.20.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		RemoteURL:  testServer.URL,
		AuthToken:  "test-token",
		ClientType: "gitea",
	}

	s, err := server.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}

	if s.Config() == nil {
		t.Fatal("Expected config to be set")
	}

	if s.Config().ClientType != "gitea" {
		t.Errorf("Expected ClientType to be 'gitea', got '%s'", s.Config().ClientType)
	}
}

func TestNewFromConfig_ForgejoClientType(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "12.0.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		RemoteURL:  testServer.URL,
		AuthToken:  "test-token",
		ClientType: "forgejo",
	}

	s, err := server.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}

	if s.Config() == nil {
		t.Fatal("Expected config to be set")
	}

	if s.Config().ClientType != "forgejo" {
		t.Errorf("Expected ClientType to be 'forgejo', got '%s'", s.Config().ClientType)
	}
}

func TestNewFromConfig_AutoClientType(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "1.20.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		RemoteURL:  testServer.URL,
		AuthToken:  "test-token",
		ClientType: "auto",
	}

	s, err := server.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}

	if s.Config() == nil {
		t.Fatal("Expected config to be set")
	}

	// Auto-detection should set the ClientType to either "gitea" or "forgejo"
	clientType := s.Config().ClientType
	if clientType != "gitea" && clientType != "forgejo" {
		t.Errorf("Expected ClientType to be 'gitea' or 'forgejo', got '%s'", clientType)
	}
}

func TestNewFromConfig_EmptyClientTypeDefaultsToAuto(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "12.0.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		RemoteURL:  testServer.URL,
		AuthToken:  "test-token",
		ClientType: "", // Empty should default to auto
	}

	s, err := server.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}

	if s.Config() == nil {
		t.Fatal("Expected config to be set")
	}

	// Empty ClientType should be treated as auto and set to either "gitea" or "forgejo"
	clientType := s.Config().ClientType
	if clientType != "gitea" && clientType != "forgejo" {
		t.Errorf("Expected ClientType to be 'gitea' or 'forgejo', got '%s'", clientType)
	}
}

func TestNewFromConfig_InvalidClientType(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "1.20.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		RemoteURL:  testServer.URL,
		AuthToken:  "test-token",
		ClientType: "invalid",
	}

	_, err := server.NewFromConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid client type")
	}

	if !contains(err.Error(), "ClientType must be one of") {
		t.Errorf("Expected error message to contain 'ClientType must be one of', got: %v", err)
	}
}

func TestNewFromConfig_BackwardCompatibility(t *testing.T) {
	// Create a test server that responds to the version endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version": "1.20.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer testServer.Close()

	// Test that existing behavior (no ClientType specified) still works
	cfg := &config.Config{
		RemoteURL: testServer.URL,
		AuthToken: "test-token",
		// ClientType not set - should default to auto
	}

	s, err := server.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}

	if s.Config() == nil {
		t.Fatal("Expected config to be set")
	}

	// Should auto-detect and set ClientType
	clientType := s.Config().ClientType
	if clientType != "gitea" && clientType != "forgejo" {
		t.Errorf("Expected ClientType to be 'gitea' or 'forgejo', got '%s'", clientType)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
