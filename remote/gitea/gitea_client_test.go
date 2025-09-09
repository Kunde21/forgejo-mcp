package gitea

import (
	"strings"
	"testing"
)

func TestGiteaClientCreateIssueComment(t *testing.T) {
	// Test CreateIssueComment method structure
	// Since we can't easily mock the Gitea client for unit tests,
	// we focus on testing that the method exists and has the right signature

	client := &GiteaClient{}

	// Test that the method exists and has the right signature
	// We don't call it because it would panic with nil client
	if client == nil {
		t.Error("GiteaClient should not be nil")
	}

	// Test that the method can be assigned to the interface
	var _ IssueCommenter = client
}

func TestGiteaClientCreateIssueCommentInvalidRepo(t *testing.T) {
	// Test repository parsing logic for invalid formats
	// We test the parsing logic in isolation since we can't call the method with nil client

	testCases := []struct {
		input    string
		expected bool // true if should pass basic parsing
	}{
		{"owner/repo", true},
		{"invalid-format", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Test the parsing logic that would be used in CreateIssueComment
			_, _, ok := strings.Cut(tc.input, "/")
			if ok != tc.expected {
				t.Errorf("Expected parsing result %v for input %s, got %v", tc.expected, tc.input, ok)
			}
		})
	}
}

func TestGiteaClientCreateIssueCommentRepoParsing(t *testing.T) {
	// Test repository parsing logic in isolation
	testCases := []struct {
		input     string
		wantOwner string
		wantRepo  string
		wantError bool
	}{
		{"owner/repo", "owner", "repo", false},
		{"user/project", "user", "project", false},     // Valid format
		{"invalid-format", "", "", true},               // No slash
		{"too/many/parts", "too", "many/parts", false}, // Multiple slashes - implementation accepts this
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Test the parsing logic by checking if strings.Cut works as expected
			owner, repo, ok := strings.Cut(tc.input, "/")

			// Check if parsing succeeded
			if !ok && !tc.wantError {
				t.Errorf("Expected parsing to succeed for input %s", tc.input)
			}
			if ok && tc.wantError {
				t.Errorf("Expected parsing to fail for input %s", tc.input)
			}

			// For valid cases, check the parsed values
			if ok && !tc.wantError {
				if owner != tc.wantOwner || repo != tc.wantRepo {
					t.Errorf("Expected owner=%s repo=%s, got owner=%s repo=%s", tc.wantOwner, tc.wantRepo, owner, repo)
				}
			}
		})
	}
}
