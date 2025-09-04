// Package server provides tests for MCP server authentication integration
package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/auth"
	"github.com/Kunde21/forgejo-mcp/config"
)

// MockTokenValidator implements auth.TokenValidator for testing
type MockTokenValidator struct {
	validateFunc func(baseURL, token string) error
}

func (m *MockTokenValidator) ValidateToken(baseURL, token string) error {
	if m.validateFunc != nil {
		return m.validateFunc(baseURL, token)
	}
	return nil
}

// TestAuthIntegration_ServerInitialization tests that the server properly initializes with authentication components
func TestAuthIntegration_ServerInitialization(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectError   bool
		expectedError string
	}{
		{
			name: "valid config with auth token",
			config: &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    "valid-testing-auth-token-12345",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			},
			expectError: false,
		},
		{
			name: "config with empty auth token",
			config: &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    "a",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			},
			expectError:   true,
			expectedError: "AuthToken: the length must be between 20 and 100",
		},
		{
			name: "config with invalid forgejo URL",
			config: &config.Config{
				ForgejoURL:   ":invalid-url::",
				AuthToken:    "valid-testing-auth-token-12345",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			},
			expectError:   true,
			expectedError: "invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if server == nil {
				t.Error("expected non-nil server")
				return
			}

			// Verify server has authentication state management
			if server.authState == nil {
				t.Error("expected server to have authentication state management initialized")
			}
		})
	}
}

// TestAuthIntegration_ToolExecutionWithAuthValidation tests that tool execution validates authentication
func TestAuthIntegration_ToolExecutionWithAuthValidation(t *testing.T) {
	tests := []struct {
		name          string
		authToken     string
		mockValidator *MockTokenValidator
		toolName      string
		toolArgs      map[string]interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:      "successful tool execution with valid auth",
			authToken: "valid-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					if token == "valid-testing-auth-token-12345-abcdef-very-long-token" {
						return nil
					}
					return errors.New("invalid token")
				},
			},
			toolName:    "pr_list",
			toolArgs:    map[string]interface{}{"owner": "test", "repo": "test"},
			expectError: false,
		},
		{
			name:      "tool execution fails with invalid auth",
			authToken: "invalid-token-2-short",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthServerError(401, "Unauthorized")
				},
			},
			toolName:      "pr_list",
			toolArgs:      map[string]interface{}{"owner": "test", "repo": "test"},
			expectError:   true,
			expectedError: "authentication failed",
		},
		{
			name:      "tool execution fails with network error",
			authToken: "valid-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthNetworkError("connection failed", "https://example.forgejo.com")
				},
			},
			toolName:      "pr_list",
			toolArgs:      map[string]interface{}{"owner": "test", "repo": "test"},
			expectError:   true,
			expectedError: "network error",
		},
		{
			name:      "tool execution fails with timeout",
			authToken: "valid-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthTimeoutError(5 * time.Second)
				},
			},
			toolName:      "pr_list",
			toolArgs:      map[string]interface{}{"owner": "test", "repo": "test"},
			expectError:   true,
			expectedError: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    tt.authToken,
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			}

			server, err := New(cfg)
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			// Set up mock validator
			if server.authState != nil {
				server.authState.validator = tt.mockValidator
			}

			// Create authenticated tool handler
			handler := NewAuthenticatedToolHandler(server.toolRegistry, server.authState, server, nil, server.logger)

			// Execute tool
			ctx := context.Background()
			result, err := handler.HandleRequest(ctx, "tools/call", map[string]interface{}{
				"name":      tt.toolName,
				"arguments": tt.toolArgs,
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

// TestAuthIntegration_AuthStateManagement tests authentication state management
func TestAuthIntegration_AuthStateManagement(t *testing.T) {
	tests := []struct {
		name            string
		authToken       string
		mockValidator   *MockTokenValidator
		concurrentCalls int
		expectCacheHit  bool
	}{
		{
			name:      "successful auth state caching",
			authToken: "valid-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					if token == "valid-testing-auth-token-12345-abcdef-very-long-token" {
						return nil
					}
					return errors.New("invalid token")
				},
			},
			concurrentCalls: 5,
			expectCacheHit:  true,
		},
		{
			name:      "failed auth not cached",
			authToken: "invalid-token-2-short",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthServerError(401, "Unauthorized")
				},
			},
			concurrentCalls: 3,
			expectCacheHit:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    tt.authToken,
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			}

			server, err := New(cfg)
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			// Set up mock validator
			if server.authState != nil {
				server.authState.validator = tt.mockValidator
			}

			// Track validation calls
			var validationCallCount int
			var mu sync.Mutex

			originalValidateFunc := tt.mockValidator.validateFunc
			tt.mockValidator.validateFunc = func(baseURL, token string) error {
				mu.Lock()
				validationCallCount++
				mu.Unlock()
				return originalValidateFunc(baseURL, token)
			}

			// Execute multiple concurrent authentication validations
			var wg sync.WaitGroup
			results := make([]error, tt.concurrentCalls)

			for i := 0; i < tt.concurrentCalls; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					results[index] = server.authState.ValidateToken(context.Background(), cfg.ForgejoURL, tt.authToken)
				}(i)
			}

			wg.Wait()

			// Check results
			for i, result := range results {
				if tt.expectCacheHit && i > 0 {
					// For successful validations, only first call should hit validator
					continue
				}
				if result != nil && !tt.expectCacheHit {
					// Expected failure
					continue
				}
				if result == nil && tt.expectCacheHit && i > 0 {
					// Should have used cache
					continue
				}
				// Validate individual result
				if tt.expectCacheHit && result != nil {
					t.Errorf("expected successful validation but got error: %v", result)
				}
				if !tt.expectCacheHit && result == nil {
					t.Errorf("expected validation error but got success")
				}
			}

			// Verify caching behavior
			if tt.expectCacheHit {
				mu.Lock()
				if validationCallCount != 1 {
					t.Errorf("expected exactly 1 validation call for cached successful validation, got %d", validationCallCount)
				}
				mu.Unlock()
			}
		})
	}
}

// TestAuthIntegration_ThreadSafeAuthHandling tests thread-safe authentication handling
func TestAuthIntegration_ThreadSafeAuthHandling(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "valid-testing-auth-token-12345-abcdef-very-long-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Set up mock validator
	var validationCount int
	var mu sync.Mutex
	mockValidator := &MockTokenValidator{
		validateFunc: func(baseURL, token string) error {
			mu.Lock()
			validationCount++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Simulate network delay
			return nil
		},
	}
	if server.authState != nil {
		server.authState.validator = mockValidator
	}

	// Execute concurrent authentication validations
	const numGoroutines = 10
	const callsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*callsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				err := server.authState.ValidateToken(context.Background(), cfg.ForgejoURL, cfg.AuthToken)
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Collect results
	var successCount, errorCount int
	for err := range errors {
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Verify results
	expectedTotal := numGoroutines * callsPerGoroutine
	if successCount != expectedTotal {
		t.Errorf("expected %d successful validations, got %d", expectedTotal, successCount)
	}
	if errorCount != 0 {
		t.Errorf("expected 0 errors, got %d", errorCount)
	}

	// Verify thread safety - should validate at most once per goroutine due to caching
	mu.Lock()
	if validationCount > numGoroutines {
		t.Errorf("expected at most %d validation calls (one per goroutine), got %d", numGoroutines, validationCount)
	}
	if validationCount < 1 {
		t.Errorf("expected at least 1 validation call, got %d", validationCount)
	}
	mu.Unlock()
}

// TestAuthIntegration_AuthStateInitialization tests authentication state initialization
func TestAuthIntegration_AuthStateInitialization(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "testing-auth-token-123",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Verify authentication state is properly initialized
	if server.authState == nil {
		t.Fatal("expected authState to be initialized")
	}

	if server.authState.cache == nil {
		t.Error("expected authState.cache to be initialized")
	}

	// Note: validator may be nil if Gitea client creation failed (network issues)
	// This is acceptable behavior - auth will fail gracefully

	if server.authState.logger == nil {
		t.Error("expected authState.logger to be initialized")
	}
}

// TestAuthIntegration_CompleteAuthenticationFlow tests the complete authentication flow from server startup to tool execution
func TestAuthIntegration_CompleteAuthenticationFlow(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		authToken      string
		mockValidator  *MockTokenValidator
		toolName       string
		toolArgs       map[string]interface{}
		expectSuccess  bool
		expectedResult interface{}
	}{
		{
			name: "complete successful authentication flow",
			setupEnv: func() {
				t.Setenv("GITEA_TOKEN", "env-testing-auth-token-12345-abcdef-very-long-token")
			},
			cleanupEnv: func() {
				// Environment cleanup is handled by t.Setenv
			},
			authToken: "env-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					if token == "env-testing-auth-token-12345-abcdef-very-long-token" {
						return nil
					}
					return auth.NewAuthServerError(401, "Unauthorized")
				},
			},
			toolName:      "pr_list",
			toolArgs:      map[string]interface{}{"owner": "testuser", "repo": "testrepo"},
			expectSuccess: true,
			expectedResult: map[string]interface{}{
				"tool":      "pr_list",
				"status":    "executed",
				"arguments": map[string]interface{}{"owner": "testuser", "repo": "testrepo"},
			},
		},
		{
			name: "authentication flow with token validation failure",
			setupEnv: func() {
				t.Setenv("GITEA_TOKEN", "invalid-short-token")
			},
			cleanupEnv: func() {},
			authToken:  "invalid-2-short-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthServerError(401, "Unauthorized")
				},
			},
			toolName:      "issue_list",
			toolArgs:      map[string]interface{}{"owner": "testuser", "repo": "testrepo"},
			expectSuccess: false,
		},
		{
			name: "authentication flow with network timeout",
			setupEnv: func() {
				t.Setenv("GITEA_TOKEN", "timeout-testing-auth-token-12345-abcdef-very-long-token")
			},
			cleanupEnv: func() {
				// Environment cleanup is handled by t.Setenv
			},
			authToken: "timeout-testing-auth-token-12345-abcdef-very-long-token",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthTimeoutError(2 * time.Second)
				},
			},
			toolName:      "pr_list",
			toolArgs:      map[string]interface{}{"owner": "testuser", "repo": "testrepo"},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				t.Cleanup(tt.cleanupEnv)
			}

			// Create server configuration
			cfg := &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    tt.authToken,
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			}

			// Create server
			server, err := New(cfg)
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			// Set up mock validator
			if server.authState != nil {
				server.authState.validator = tt.mockValidator
			}

			// Create authenticated tool handler
			handler := NewAuthenticatedToolHandler(server.toolRegistry, server.authState, server, nil, server.logger)

			// Execute tool request
			ctx := context.Background()
			result, err := handler.HandleRequest(ctx, "tools/call", map[string]interface{}{
				"name":      tt.toolName,
				"arguments": tt.toolArgs,
			})

			// Verify results
			if tt.expectSuccess {
				if err != nil {
					t.Errorf("expected successful execution but got error: %v", err)
					return
				}
				if result == nil {
					t.Error("expected non-nil result for successful execution")
					return
				}

				// Verify result structure
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Errorf("expected result to be a map, got %T", result)
					return
				}

				if resultMap["tool"] != tt.toolName {
					t.Errorf("expected tool name %q, got %q", tt.toolName, resultMap["tool"])
				}

				if resultMap["status"] != "executed" {
					t.Errorf("expected status 'executed', got %q", resultMap["status"])
				}
			} else {
				if err == nil {
					t.Error("expected error but got successful execution")
				}
			}
		})
	}
}

// TestAuthIntegration_ErrorScenarios tests various error scenarios in authentication
func TestAuthIntegration_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		authToken     string
		mockValidator *MockTokenValidator
		toolName      string
		toolArgs      map[string]interface{}
		expectError   bool
		errorType     string
	}{
		{
			name:      "server error during validation",
			authToken: "test-testing-auth-token-12345",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthServerError(500, "Internal Server Error")
				},
			},
			toolName:    "pr_list",
			toolArgs:    map[string]interface{}{"owner": "test", "repo": "test"},
			expectError: true,
			errorType:   "AuthServerError",
		},
		{
			name:      "network error during validation",
			authToken: "test-testing-auth-token-12345",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthNetworkError("connection refused", "https://example.forgejo.com")
				},
			},
			toolName:    "pr_list",
			toolArgs:    map[string]interface{}{"owner": "test", "repo": "test"},
			expectError: true,
			errorType:   "AuthNetworkError",
		},
		{
			name:      "timeout error during validation",
			authToken: "test-testing-auth-token-12345",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return auth.NewAuthTimeoutError(5 * time.Second)
				},
			},
			toolName:    "pr_list",
			toolArgs:    map[string]interface{}{"owner": "test", "repo": "test"},
			expectError: true,
			errorType:   "AuthTimeoutError",
		},
		{
			name:      "token validation error",
			authToken: "test-testing-auth-token-12345",
			mockValidator: &MockTokenValidator{
				validateFunc: func(baseURL, token string) error {
					return &auth.TokenValidationError{
						Message: "token format invalid",
						Field:   "token",
					}
				},
			},
			toolName:    "pr_list",
			toolArgs:    map[string]interface{}{"owner": "test", "repo": "test"},
			expectError: true,
			errorType:   "TokenValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    tt.authToken,
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			}

			server, err := New(cfg)
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			// Set up mock validator
			if server.authState != nil {
				server.authState.validator = tt.mockValidator
			}

			// Create authenticated tool handler
			handler := NewAuthenticatedToolHandler(server.toolRegistry, server.authState, server, nil, server.logger)

			// Execute tool
			ctx := context.Background()
			_, err = handler.HandleRequest(ctx, "tools/call", map[string]interface{}{
				"name":      tt.toolName,
				"arguments": tt.toolArgs,
			})

			// Verify error
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Verify error contains expected error type in message
				errorMsg := err.Error()
				var expectedInMessage string
				switch tt.errorType {
				case "AuthServerError":
					expectedInMessage = "authentication server error"
				case "AuthNetworkError":
					expectedInMessage = "authentication network error"
				case "AuthTimeoutError":
					expectedInMessage = "authentication timed out"
				case "TokenValidationError":
					expectedInMessage = "token validation failed"
				default:
					expectedInMessage = tt.errorType
				}

				if !containsString(errorMsg, expectedInMessage) {
					t.Errorf("expected error message to contain %q, got %q", expectedInMessage, errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i < len(s)-len(substr)+1; i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
