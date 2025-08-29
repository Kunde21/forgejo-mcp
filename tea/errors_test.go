package tea

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestValidateConnectionErrors(t *testing.T) {
	wrapper := &GiteaWrapper{}
	err := wrapper.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() error = nil, want error")
	}

	wantErr := "wrapper not initialized"
	if !cmp.Equal(wantErr, err.Error()) {
		t.Error(cmp.Diff(wantErr, err.Error()))
	}
}

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		response *gitea.Response
		want     map[string]interface{}
	}{
		{
			name: "Gitea API error with response",
			err:  fmt.Errorf("404 Not Found"),
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
				},
			},
			want: map[string]interface{}{
				"error":       "404 Not Found",
				"status_code": 404,
				"status":      "404 Not Found",
				"type":        "api_error",
			},
		},
		{
			name: "Authentication error",
			err:  fmt.Errorf("authentication failed: invalid token"),
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 401,
					Status:     "401 Unauthorized",
				},
			},
			want: map[string]interface{}{
				"error":       "authentication failed: invalid token",
				"status_code": 401,
				"status":      "401 Unauthorized",
				"type":        "auth_error",
			},
		},
		{
			name: "Rate limit error",
			err:  fmt.Errorf("rate limit exceeded"),
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 429,
					Status:     "429 Too Many Requests",
				},
			},
			want: map[string]interface{}{
				"error":       "rate limit exceeded",
				"status_code": 429,
				"status":      "429 Too Many Requests",
				"type":        "rate_limit_error",
			},
		},
		{
			name:     "Generic error",
			err:      errors.New("something went wrong"),
			response: nil,
			want: map[string]interface{}{
				"error": "something went wrong",
				"type":  "error",
			},
		},
		{
			name:     "nil error",
			err:      nil,
			response: nil,
			want: map[string]interface{}{
				"type": "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAPIError(tt.err, tt.response)
			if !cmp.Equal(tt.want, got) {
				t.Errorf("FormatAPIError() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestHandlePartialSuccess(t *testing.T) {
	// Test data
	issues := []*gitea.Issue{
		{ID: 1, Title: "Issue 1"},
		{ID: 2, Title: "Issue 2"},
	}

	prs := []*gitea.PullRequest{
		{ID: 1, Title: "PR 1"},
	}

	tests := []struct {
		name          string
		issues        []*gitea.Issue
		prs           []*gitea.PullRequest
		issueErr      error
		prErr         error
		wantIssues    []map[string]interface{}
		wantPRs       []map[string]interface{}
		wantHasErrors bool
		wantErrors    []map[string]interface{}
	}{
		{
			name:          "all successful",
			issues:        issues,
			prs:           prs,
			issueErr:      nil,
			prErr:         nil,
			wantIssues:    TransformIssuesToMCP(issues),
			wantPRs:       TransformPullRequestsToMCP(prs),
			wantHasErrors: false,
			wantErrors:    []map[string]interface{}{},
		},
		{
			name:          "issue error only",
			issues:        nil,
			prs:           prs,
			issueErr:      errors.New("failed to fetch issues"),
			prErr:         nil,
			wantIssues:    []map[string]interface{}{},
			wantPRs:       TransformPullRequestsToMCP(prs),
			wantHasErrors: true,
			wantErrors: []map[string]interface{}{
				{
					"error": "failed to fetch issues",
					"type":  "issue_fetch_error",
				},
			},
		},
		{
			name:          "both errors",
			issues:        nil,
			prs:           nil,
			issueErr:      errors.New("failed to fetch issues"),
			prErr:         errors.New("failed to fetch PRs"),
			wantIssues:    []map[string]interface{}{},
			wantPRs:       []map[string]interface{}{},
			wantHasErrors: true,
			wantErrors: []map[string]interface{}{
				{
					"error": "failed to fetch issues",
					"type":  "issue_fetch_error",
				},
				{
					"error": "failed to fetch PRs",
					"type":  "pr_fetch_error",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandlePartialSuccess(tt.issues, tt.prs, tt.issueErr, tt.prErr)

			if !cmp.Equal(tt.wantIssues, result.Issues) {
				t.Errorf("HandlePartialSuccess().Issues mismatch (-want +got):\n%s", cmp.Diff(tt.wantIssues, result.Issues))
			}

			if !cmp.Equal(tt.wantPRs, result.PullRequests) {
				t.Errorf("HandlePartialSuccess().PullRequests mismatch (-want +got):\n%s", cmp.Diff(tt.wantPRs, result.PullRequests))
			}

			if tt.wantHasErrors != result.HasErrors {
				t.Errorf("HandlePartialSuccess().HasErrors = %v, want %v", result.HasErrors, tt.wantHasErrors)
			}

			if !cmp.Equal(tt.wantErrors, result.Errors) {
				t.Errorf("HandlePartialSuccess().Errors mismatch (-want +got):\n%s", cmp.Diff(tt.wantErrors, result.Errors))
			}
		})
	}
}

func TestCheckRateLimit(t *testing.T) {
	reset := "Wed, 29 Aug 2025 10:10:00 GMT"

	tests := []struct {
		name        string
		response    *gitea.Response
		wantLimited bool
		wantInfo    map[string]interface{}
	}{
		{
			name: "rate limited response",
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 429,
					Header: func() http.Header {
						header := http.Header{}
						header.Set("X-RateLimit-Limit", "60")
						header.Set("X-RateLimit-Remaining", "0")
						header.Set("X-RateLimit-Reset", reset)
						return header
					}(),
				},
			},
			wantLimited: true,
			wantInfo: map[string]interface{}{
				"limited":   true,
				"limit":     60,
				"remaining": 0,
				"reset":     reset,
				"type":      "rate_limit_info",
			},
		},
		{
			name: "normal response with rate limit info",
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 200,
					Header: func() http.Header {
						header := http.Header{}
						header.Set("X-RateLimit-Limit", "60")
						header.Set("X-RateLimit-Remaining", "50")
						header.Set("X-RateLimit-Reset", reset)
						return header
					}(),
				},
			},
			wantLimited: false,
			wantInfo: map[string]interface{}{
				"limited":   false,
				"limit":     60,
				"remaining": 50,
				"reset":     reset,
				"type":      "rate_limit_info",
			},
		},
		{
			name: "response without rate limit headers",
			response: &gitea.Response{
				Response: &http.Response{
					StatusCode: 200,
					Header:     http.Header{},
				},
			},
			wantLimited: false,
			wantInfo: map[string]interface{}{
				"limited": false,
				"type":    "rate_limit_info",
			},
		},
		{
			name:        "nil response",
			response:    nil,
			wantLimited: false,
			wantInfo: map[string]interface{}{
				"limited": false,
				"type":    "rate_limit_info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimited := CheckRateLimit(tt.response)
			if tt.wantLimited != gotLimited {
				t.Errorf("CheckRateLimit() = %v, want %v", gotLimited, tt.wantLimited)
			}

			gotInfo := GetRateLimitInfo(tt.response)
			if !cmp.Equal(tt.wantInfo, gotInfo) {
				t.Errorf("GetRateLimitInfo() mismatch (-want +got):\n%s", cmp.Diff(tt.wantInfo, gotInfo))
			}
		})
	}
}
