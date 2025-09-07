package servertest

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestMockGiteaServer tests the mock Gitea server functionality
func TestMockGiteaServer(t *testing.T) {
	mock := NewMockGiteaServer()
	defer mock.Close()

	// Add test issues
	mock.AddIssues("testuser", "testrepo", []MockIssue{
		{Index: 1, Title: "Test Issue 1", State: "open"},
		{Index: 2, Title: "Test Issue 2", State: "closed"},
	})

	// Test version endpoint
	resp, err := http.Get(mock.URL() + "/api/v1/version")
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test issues endpoint
	resp, err = http.Get(mock.URL() + "/api/v1/repos/testuser/testrepo/issues")
	if err != nil {
		t.Fatalf("Failed to get issues: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var issues []MockIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		t.Fatalf("Failed to decode issues: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
}
