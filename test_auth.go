package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	servertest "github.com/kunde21/forgejo-mcp/server_test"
)

func TestMockServerAuth(t *testing.T) {
	// Create a mock server
	mock := servertest.NewMockGiteaServer(t)

	// Test with invalid token
	req, _ := http.NewRequest("PATCH", mock.URL()+"/api/v1/repos/testuser/testrepo/issues/comments/123", strings.NewReader(`{"body": "test"}`))
	req.Header.Set("Authorization", "token invalid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status code: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}
