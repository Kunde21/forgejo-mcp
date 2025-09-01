package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestServerLifecycle_ValidConfig(t *testing.T) {
	config := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(config)

	if err != nil {
		t.Fatalf("New() with valid config should not return error, got: %v", err)
	}
	if server == nil {
		t.Fatal("New() should return non-nil server")
	}
	if server.logger == nil {
		t.Error("server.logger should be initialized")
	}
	if server.config == nil {
		t.Error("server.config should be set")
	}
}

func TestServerLifecycle_NilConfig(t *testing.T) {
	server, err := New(nil)

	if err == nil {
		t.Error("New() with nil config should return error")
	}
	if server != nil {
		t.Error("New() with nil config should return nil server")
	}
}

func TestServerLifecycle_InvalidLogLevel(t *testing.T) {
	config := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "invalid",
	}

	server, err := New(config)

	if err == nil {
		t.Error("New() with invalid log level should return error")
	}
	if server != nil {
		t.Error("New() with invalid log level should return nil server")
	}
}

func TestServerStartStop(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server == nil {
		t.Fatal("New() returned nil server")
	}
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(t.Context())
	}()
	time.Sleep(100 * time.Millisecond)
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	select {
	case err := <-startErr:
		if err != nil {
			t.Errorf("Start() error = %v, expected no error after Stop", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Start() did not return after Stop() was called")
	}
}

func TestServerStopWithoutStart(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v, expected no error when called on non-started server", err)
	}
}

func TestServerWithLogger(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "debug",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server.logger == nil {
		t.Fatal("expected server.logger to be non-nil")
	}
	if !cmp.Equal(logrus.DebugLevel, server.logger.Level) {
		t.Errorf("expected logger level %v, got %v", logrus.DebugLevel, server.logger.Level)
	}
}

// Test AuthState functionality
func TestNewAuthState(t *testing.T) {
	logger := logrus.New()
	validator := &mockTokenValidator{}

	authState := NewAuthState(validator, logger)

	if authState == nil {
		t.Fatal("NewAuthState() returned nil")
	}
	if authState.validator != validator {
		t.Error("validator not set correctly")
	}
	if authState.logger != logger {
		t.Error("logger not set correctly")
	}
	if authState.cache == nil {
		t.Error("cache should be initialized")
	}
}

func TestAuthStateValidateToken(t *testing.T) {
	logger := logrus.New()
	validator := &mockTokenValidator{shouldSucceed: true}

	authState := NewAuthState(validator, logger)

	// Test successful validation
	err := authState.ValidateToken(context.Background(), "https://example.com", "test-token")
	if err != nil {
		t.Errorf("ValidateToken() should succeed, got error: %v", err)
	}

	// Test failed validation
	validator.shouldSucceed = false
	err = authState.ValidateToken(context.Background(), "https://example.com", "bad-token")
	if err == nil {
		t.Error("ValidateToken() should fail with bad token")
	}
}

func TestAuthStateCache(t *testing.T) {
	logger := logrus.New()
	validator := &mockTokenValidator{shouldSucceed: true, callCount: 0}

	authState := NewAuthState(validator, logger)

	// First call should validate and cache
	err1 := authState.ValidateToken(context.Background(), "https://example.com", "test-token")
	if err1 != nil {
		t.Errorf("First ValidateToken() should succeed, got error: %v", err1)
	}
	if validator.callCount != 1 {
		t.Errorf("Expected validator to be called once, got %d", validator.callCount)
	}

	// Second call should use cache
	err2 := authState.ValidateToken(context.Background(), "https://example.com", "test-token")
	if err2 != nil {
		t.Errorf("Second ValidateToken() should succeed, got error: %v", err2)
	}
	if validator.callCount != 1 {
		t.Errorf("Expected validator to still be called once (cached), got %d", validator.callCount)
	}
}

func TestAuthStateClearCache(t *testing.T) {
	logger := logrus.New()
	validator := &mockTokenValidator{shouldSucceed: true, callCount: 0}

	authState := NewAuthState(validator, logger)

	// Cache a successful validation
	authState.ValidateToken(context.Background(), "https://example.com", "test-token")

	// Clear cache
	authState.ClearCache()

	// Next call should validate again
	authState.ValidateToken(context.Background(), "https://example.com", "test-token")
	if validator.callCount != 2 {
		t.Errorf("Expected validator to be called twice after cache clear, got %d", validator.callCount)
	}
}

// Test GiteaTokenValidator
func TestGiteaTokenValidator(t *testing.T) {
	// Test with nil client
	validator := &GiteaTokenValidator{}
	err := validator.ValidateToken("https://example.com", "test-token")
	if err == nil {
		t.Error("ValidateToken() should fail with nil client")
	}
	expectedErr := "Gitea client not initialized"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message %q, got: %v", expectedErr, err)
	}
}

// Test AuthenticatedToolHandler
func TestNewAuthenticatedToolHandler(t *testing.T) {
	logger := logrus.New()
	registry := &ToolRegistry{}
	authState := NewAuthState(nil, logger)
	server := &Server{}
	innerHandler := &mockRequestHandler{}

	handler := NewAuthenticatedToolHandler(registry, authState, server, innerHandler, logger)

	if handler == nil {
		t.Fatal("NewAuthenticatedToolHandler() returned nil")
	}
	if handler.registry != registry {
		t.Error("registry not set correctly")
	}
	if handler.authState != authState {
		t.Error("authState not set correctly")
	}
	if handler.server != server {
		t.Error("server not set correctly")
	}
	if handler.innerHandler != innerHandler {
		t.Error("innerHandler not set correctly")
	}
}

func TestAuthenticatedToolHandlerHandleRequest(t *testing.T) {
	logger := logrus.New()
	registry := &ToolRegistry{}
	authState := NewAuthState(&mockTokenValidator{shouldSucceed: true}, logger)

	cfg := &config.Config{
		ForgejoURL: "https://example.com",
		AuthToken:  "test-token",
	}
	server, _ := New(cfg)
	server.authState = authState

	innerHandler := &mockRequestHandler{}
	handler := NewAuthenticatedToolHandler(registry, authState, server, innerHandler, logger)

	// Test successful request
	params := map[string]interface{}{
		"name":      "test-tool",
		"arguments": map[string]interface{}{"arg1": "value1"},
	}

	result, err := handler.HandleRequest(context.Background(), "tools/call", params)
	if err != nil {
		t.Errorf("HandleRequest() should succeed, got error: %v", err)
	}
	if result == nil {
		t.Error("HandleRequest() should return a result")
	}

	// Test with authentication failure
	authState.validator = &mockTokenValidator{shouldSucceed: false}
	_, err = handler.HandleRequest(context.Background(), "tools/call", params)
	if err == nil {
		t.Error("HandleRequest() should fail with authentication error")
	}
}

// Mock implementations for testing
type mockTokenValidator struct {
	shouldSucceed bool
	callCount     int
}

func (m *mockTokenValidator) ValidateToken(baseURL, token string) error {
	m.callCount++
	if !m.shouldSucceed {
		return errors.New("mock validation failed")
	}
	return nil
}

type mockRequestHandler struct{}

func (m *mockRequestHandler) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"status": "success"}, nil
}

// Test server initialization with tool system
func TestServerInitializeToolSystem(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Check that tool system components are initialized
	if server.toolRegistry == nil {
		t.Error("toolRegistry should be initialized")
	}
	if server.dispatcher == nil {
		t.Error("dispatcher should be initialized")
	}
	if server.processor == nil {
		t.Error("processor should be initialized")
	}
}

// Test server with debug logging
func TestServerWithDebugLogging(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "debug",
		Debug:        true,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if server.logger == nil {
		t.Fatal("expected server.logger to be non-nil")
	}
	if server.logger.Level != logrus.DebugLevel {
		t.Errorf("expected logger level %v, got %v", logrus.DebugLevel, server.logger.Level)
	}

	// Check that debug formatter is used when Debug is true
	// Note: We can't easily test the formatter type, but we can verify the logger is configured
}

// Test server with invalid configuration
func TestServerInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "empty forgejo URL",
			config: &config.Config{
				ForgejoURL: "",
				AuthToken:  "test-token",
				LogLevel:   "info",
			},
		},
		{
			name: "empty auth token",
			config: &config.Config{
				ForgejoURL: "https://example.com",
				AuthToken:  "",
				LogLevel:   "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.config)
			if err == nil {
				t.Errorf("New() should fail with invalid config: %s", tt.name)
			}
			if server != nil {
				t.Errorf("New() should return nil server for invalid config: %s", tt.name)
			}
		})
	}
}

// Test server start and stop with context cancellation
func TestServerStartStopWithContext(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in goroutine
	startErrCh := make(chan error, 1)
	go func() {
		startErrCh <- server.Start(ctx)
	}()

	// Let it start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-startErrCh:
		if err != nil {
			t.Errorf("Start() should not return error on context cancellation, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not return after context cancellation")
	}
}

// Test server stop without start (alternative test)
func TestServerStopWithoutStartAlternative(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL: "https://example.forgejo.com",
		AuthToken:  "test-token",
		LogLevel:   "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Stop should not panic or error when called on non-started server
	err = server.Stop()
	if err != nil {
		t.Errorf("Stop() on non-started server should not error, got: %v", err)
	}
}

// Test server with different log levels
func TestServerLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected logrus.Level
	}{
		{"debug level", "debug", logrus.DebugLevel},
		{"info level", "info", logrus.InfoLevel},
		{"warn level", "warn", logrus.WarnLevel},
		{"error level", "error", logrus.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ForgejoURL: "https://example.forgejo.com",
				AuthToken:  "test-token",
				LogLevel:   tt.logLevel,
			}

			server, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if server.logger.Level != tt.expected {
				t.Errorf("Expected log level %v, got %v", tt.expected, server.logger.Level)
			}
		})
	}
}
