package validation

import (
	"testing"
)

func TestValidateRepository(t *testing.T) {
	testCases := []struct {
		name        string
		repo        interface{}
		expectError bool
	}{
		{"valid repo", "owner/repo", false},
		{"valid with numbers", "user123/repo456", false},
		{"valid with underscores", "user_name/repo_name", false},
		{"valid with dots", "user.name/repo.name", false},
		{"valid with hyphens", "user-name/repo-name", false},
		{"empty repo", "", true},
		{"missing slash", "ownerrepo", true},
		{"multiple slashes", "owner/repo/extra", true},
		{"leading slash", "/owner/repo", true},
		{"trailing slash", "owner/repo/", true},
		{"invalid characters", "owner/repo!", true},
		{"spaces in owner", "own er/repo", true},
		{"spaces in repo", "owner/rep o", true},
		{"non-string type", 123, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRepository(tc.repo)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for repo %v", tc.repo)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for repo %v, got %v", tc.repo, err)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	testCases := []struct {
		name        string
		limit       int
		offset      int
		expectError bool
	}{
		{"valid pagination", 10, 0, false},
		{"valid limit max", 100, 0, false},
		{"valid limit min", 1, 0, false},
		{"valid offset positive", 10, 50, false},
		{"limit too low", 0, 0, true},
		{"limit too high", 101, 0, true},
		{"limit negative", -1, 0, true},
		{"offset negative", 10, -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePagination(tc.limit, tc.offset)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for limit=%d, offset=%d", tc.limit, tc.offset)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for limit=%d, offset=%d, got %v", tc.limit, tc.offset, err)
			}
		})
	}
}

func TestValidateIssueNumber(t *testing.T) {
	testCases := []struct {
		name        string
		issueNumber interface{}
		expectError bool
	}{
		{"valid issue number", 1, false},
		{"valid large number", 1000, false},
		{"zero issue number", 0, true},
		{"negative issue number", -1, true},
		{"non-integer type", "1", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIssueNumber(tc.issueNumber)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for issue number %v", tc.issueNumber)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for issue number %v, got %v", tc.issueNumber, err)
			}
		})
	}
}

func TestValidateCommentContent(t *testing.T) {
	testCases := []struct {
		name        string
		comment     interface{}
		expectError bool
	}{
		{"valid comment", "Valid comment", false},
		{"another valid comment", "Another valid comment", false},
		{"empty comment", "", true},
		{"whitespace comment", "   ", true},
		{"tab and newline", "\t\n", true},
		{"non-string type", 123, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCommentContent(tc.comment)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for comment %v", tc.comment)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for comment %v, got %v", tc.comment, err)
			}
		})
	}
}

func TestRepositoryRegex(t *testing.T) {
	validRepos := []string{
		"owner/repo",
		"user123/repo456",
		"user_name/repo_name",
		"user.name/repo.name",
		"user-name/repo-name",
	}

	invalidRepos := []string{
		"",
		"ownerrepo",
		"owner/repo/extra",
		"/owner/repo",
		"owner/repo/",
		"owner/repo!",
		"own er/repo",
		"owner/rep o",
	}

	for _, repo := range validRepos {
		if !RepositoryRegex.MatchString(repo) {
			t.Errorf("Expected %s to be valid", repo)
		}
	}

	for _, repo := range invalidRepos {
		if RepositoryRegex.MatchString(repo) {
			t.Errorf("Expected %s to be invalid", repo)
		}
	}
}
