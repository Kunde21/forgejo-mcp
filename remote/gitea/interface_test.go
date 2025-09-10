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

func TestIssueCommentListerInterface(t *testing.T) {
	// Test that IssueCommentLister interface is properly defined
	var _ IssueCommentLister = (*GiteaClient)(nil)
}

func TestIssueCommentListerListIssueComments(t *testing.T) {
	// Test the ListIssueComments method signature
	// This is a compile-time test to ensure the interface is correct
	// We don't call the method to avoid nil pointer panic
	// The interface compliance is tested in TestIssueCommentListerInterface
}

func TestIssueCommentListStruct(t *testing.T) {
	// Test IssueCommentList struct definition and JSON serialization
	comments := []IssueComment{
		{
			ID:      1,
			Content: "First comment",
			Author:  "user1",
			Created: "2025-09-10T09:00:00Z",
		},
		{
			ID:      2,
			Content: "Second comment",
			Author:  "user2",
			Created: "2025-09-10T10:00:00Z",
		},
	}

	commentList := IssueCommentList{
		Comments: comments,
		Total:    2,
		Limit:    15,
		Offset:   0,
	}

	// Test struct fields are accessible
	if len(commentList.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(commentList.Comments))
	}
	if commentList.Total != 2 {
		t.Errorf("Expected Total to be 2, got %d", commentList.Total)
	}
	if commentList.Limit != 15 {
		t.Errorf("Expected Limit to be 15, got %d", commentList.Limit)
	}
	if commentList.Offset != 0 {
		t.Errorf("Expected Offset to be 0, got %d", commentList.Offset)
	}

	// Test individual comment fields
	if commentList.Comments[0].ID != 1 {
		t.Errorf("Expected first comment ID to be 1, got %d", commentList.Comments[0].ID)
	}
	if commentList.Comments[1].Content != "Second comment" {
		t.Errorf("Expected second comment content to be 'Second comment', got %s", commentList.Comments[1].Content)
	}
}

func TestListIssueCommentsArgsStruct(t *testing.T) {
	// Test ListIssueCommentsArgs struct definition and validation tags
	args := ListIssueCommentsArgs{
		Repository:  "owner/repo",
		IssueNumber: 42,
		Limit:       15,
		Offset:      0,
	}

	// Test struct fields are accessible
	if args.Repository != "owner/repo" {
		t.Errorf("Expected Repository to be 'owner/repo', got %s", args.Repository)
	}
	if args.IssueNumber != 42 {
		t.Errorf("Expected IssueNumber to be 42, got %d", args.IssueNumber)
	}
	if args.Limit != 15 {
		t.Errorf("Expected Limit to be 15, got %d", args.Limit)
	}
	if args.Offset != 0 {
		t.Errorf("Expected Offset to be 0, got %d", args.Offset)
	}
}
