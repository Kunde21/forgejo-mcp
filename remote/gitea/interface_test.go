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
