package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRepositoryValidate(t *testing.T) {
	tests := []struct {
		name    string
		repo    *Repository
		wantErr bool
	}{
		{
			name: "valid repository",
			repo: &Repository{
				Owner:    "owner",
				Name:     "repo",
				FullName: "owner/repo",
				Private:  false,
				Fork:     false,
				URL:      "https://example.com/owner/repo",
			},
			wantErr: false,
		},
		{
			name: "missing owner",
			repo: &Repository{
				Owner:    "",
				Name:     "repo",
				FullName: "owner/repo",
				Private:  false,
				Fork:     false,
				URL:      "https://example.com/owner/repo",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			repo: &Repository{
				Owner:    "owner",
				Name:     "",
				FullName: "owner/repo",
				Private:  false,
				Fork:     false,
				URL:      "https://example.com/owner/repo",
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			repo: &Repository{
				Owner:    "owner",
				Name:     "repo",
				FullName: "owner/repo",
				Private:  false,
				Fork:     false,
				URL:      "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Repository.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &User{
				ID:        1,
				Username:  "username",
				Email:     "user@example.com",
				AvatarURL: "https://example.com/avatar.jpg",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			user: &User{
				ID:        1,
				Username:  "",
				Email:     "user@example.com",
				AvatarURL: "https://example.com/avatar.jpg",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.user.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilterOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *FilterOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: &FilterOptions{
				Page:    1,
				PerPage: 10,
				Sort:    "created",
				Order:   "desc",
			},
			wantErr: false,
		},
		{
			name: "negative page",
			options: &FilterOptions{
				Page:    -1,
				PerPage: 10,
				Sort:    "created",
				Order:   "desc",
			},
			wantErr: true,
		},
		{
			name: "zero perPage",
			options: &FilterOptions{
				Page:    1,
				PerPage: 0,
				Sort:    "created",
				Order:   "desc",
			},
			wantErr: true,
		},
		{
			name: "invalid order",
			options: &FilterOptions{
				Page:    1,
				PerPage: 10,
				Sort:    "created",
				Order:   "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.options.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("FilterOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimestampMarshalJSON(t *testing.T) {
	now := time.Now()
	ts := Timestamp{Time: now}

	data, err := ts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var parsed time.Time
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Compare the time strings instead of exact time values to avoid nanosecond precision issues
	if parsed.Format(time.RFC3339) != now.Format(time.RFC3339) {
		t.Errorf("Timestamp.MarshalJSON() = %v, want %v", parsed.Format(time.RFC3339), now.Format(time.RFC3339))
	}
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	now := time.Now()
	jsonData := []byte(fmt.Sprintf(`"%s"`, now.Format(time.RFC3339)))

	var ts Timestamp
	err := ts.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Compare the time strings instead of exact time values to avoid nanosecond precision issues
	if ts.Time.Format(time.RFC3339) != now.Format(time.RFC3339) {
		t.Errorf("Timestamp.UnmarshalJSON() = %v, want %v", ts.Time.Format(time.RFC3339), now.Format(time.RFC3339))
	}
}
