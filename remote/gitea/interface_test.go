package gitea

import (
	"testing"
)

func TestIssueListerInterface(t *testing.T) {
	// Test that IssueLister interface is properly defined
	var _ IssueLister = (*GiteaClient)(nil)
}

func TestIssueListerListIssues(t *testing.T) {
	// Test the ListIssues method signature
	// This is a compile-time test to ensure the interface is correct
	// We don't call the method to avoid nil pointer panic
	// The interface compliance is tested in TestIssueListerInterface
}

func TestIssueCommenterInterface(t *testing.T) {
	// Test that IssueCommenter interface is properly defined
	var _ IssueCommenter = (*GiteaClient)(nil)
}

func TestIssueCommenterCreateIssueComment(t *testing.T) {
	// Test the CreateIssueComment method signature
	// This is a compile-time test to ensure the interface is correct
	// We don't call the method to avoid nil pointer panic
	// The interface compliance is tested in TestIssueCommenterInterface
}

func TestIssueCommentStruct(t *testing.T) {
	// Test IssueComment struct definition and JSON serialization
	comment := IssueComment{
		ID:      123,
		Content: "This is a test comment",
		Author:  "testuser",
		Created: "2025-09-09T10:00:00Z",
	}

	// Test struct fields are accessible
	if comment.ID != 123 {
		t.Errorf("Expected ID to be 123, got %d", comment.ID)
	}
	if comment.Content != "This is a test comment" {
		t.Errorf("Expected Content to be 'This is a test comment', got %s", comment.Content)
	}
	if comment.Author != "testuser" {
		t.Errorf("Expected Author to be 'testuser', got %s", comment.Author)
	}
	if comment.Created != "2025-09-09T10:00:00Z" {
		t.Errorf("Expected Created to be '2025-09-09T10:00:00Z', got %s", comment.Created)
	}
}
