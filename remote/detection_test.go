package remote

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetectRemoteType(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		expectedType   string
		expectError    bool
	}{
		{
			name:           "forgejo detection",
			serverResponse: `{"version": "12.0.0+dev-123-g456def"}`,
			expectedType:   "forgejo",
			expectError:    false,
		},
		{
			name:           "gitea detection",
			serverResponse: `{"version": "1.20.0+dev-123-g456def"}`,
			expectedType:   "gitea",
			expectError:    false,
		},
		{
			name:           "forgejo with forgejo in version",
			serverResponse: `{"version": "forgejo-1.20.0"}`,
			expectedType:   "forgejo",
			expectError:    false,
		},
		{
			name:           "gitea with gitea in version",
			serverResponse: `{"version": "gitea-1.20.0"}`,
			expectedType:   "gitea",
			expectError:    false,
		},
		{
			name:           "ambiguous version defaults to gitea",
			serverResponse: `{"version": "1.20.0"}`,
			expectedType:   "gitea",
			expectError:    false,
		},
		{
			name:           "invalid json response",
			serverResponse: `invalid json`,
			expectedType:   "",
			expectError:    true,
		},
		{
			name:           "server error",
			serverResponse: "",
			expectedType:   "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/version" {
					http.NotFound(w, r)
					return
				}
				if tt.serverResponse == "" {
					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Test detection
			result, err := DetectRemoteType(server.URL, "test-token")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if result != tt.expectedType {
				t.Errorf("Expected remote type '%s', got '%s'", tt.expectedType, result)
			}
		})
	}
}

func TestDetectRemoteType_InvalidURL(t *testing.T) {
	_, err := DetectRemoteType("invalid-url", "test-token")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestDetectRemoteType_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version": "1.20.0"}`))
	}))
	defer server.Close()

	_, err := DetectRemoteType(server.URL, "")
	if err != nil {
		t.Errorf("Expected no error with empty token, got: %v", err)
	}
}
