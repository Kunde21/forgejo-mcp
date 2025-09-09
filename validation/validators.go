package validation

import (
	"errors"
	"regexp"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

// RepositoryRegex defines the pattern for valid repository names
var RepositoryRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

// ValidateRepository validates repository format using ozzo-validation patterns
func ValidateRepository(value interface{}) error {
	repo, ok := value.(string)
	if !ok {
		return errors.New("repository must be a string")
	}

	if repo == "" {
		return errors.New("repository cannot be empty")
	}

	if !RepositoryRegex.MatchString(repo) {
		return errors.New("repository must be in format 'owner/repo'")
	}

	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(limit, offset int) error {
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	if offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}

// ValidateIssueNumber validates issue number
func ValidateIssueNumber(value interface{}) error {
	issueNumber, ok := value.(int)
	if !ok {
		return errors.New("issue number must be an integer")
	}

	if issueNumber < 1 {
		return errors.New("issue number must be positive")
	}

	return nil
}

// ValidateCommentContent validates comment content
func ValidateCommentContent(value interface{}) error {
	comment, ok := value.(string)
	if !ok {
		return errors.New("comment must be a string")
	}

	if comment == "" {
		return errors.New("comment content cannot be empty")
	}

	// Trim whitespace and check again
	if len(strings.TrimSpace(comment)) == 0 {
		return errors.New("comment content cannot be only whitespace")
	}

	return nil
}

// RepositoryRule returns an ozzo-validation rule for repository validation
func RepositoryRule() v.Rule {
	return v.By(ValidateRepository)
}

// IssueNumberRule returns an ozzo-validation rule for issue number validation
func IssueNumberRule() v.Rule {
	return v.By(ValidateIssueNumber)
}

// CommentContentRule returns an ozzo-validation rule for comment content validation
func CommentContentRule() v.Rule {
	return v.By(ValidateCommentContent)
}

// PaginationLimitRule returns an ozzo-validation rule for pagination limit
func PaginationLimitRule() v.Rule {
	return v.Min(1).Error("limit must be at least 1")
}

// PaginationOffsetRule returns an ozzo-validation rule for pagination offset
func PaginationOffsetRule() v.Rule {
	return v.Min(0).Error("offset must be non-negative")
}

// CombinedPaginationLimitRule returns a rule that combines min and max for limit
func CombinedPaginationLimitRule() v.Rule {
	return v.By(func(value interface{}) error {
		limit, ok := value.(int)
		if !ok {
			return errors.New("limit must be an integer")
		}
		if limit < 1 || limit > 100 {
			return errors.New("limit must be between 1 and 100")
		}
		return nil
	})
}
