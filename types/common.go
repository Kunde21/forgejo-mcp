package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Repository represents a Git repository
type Repository struct {
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Private  bool   `json:"private"`
	Fork     bool   `json:"fork"`
	URL      string `json:"url"`
}

// User represents a user in the system
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// Timestamp wraps time.Time with custom JSON marshaling for RFC3339 format
type Timestamp struct {
	time.Time
}

// MarshalJSON implements json.Marshaler interface for Timestamp
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler interface for Timestamp
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

// FilterOptions represents query parameters for filtering
type FilterOptions struct {
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"perPage,omitempty"`
	Sort    string `json:"sort,omitempty"`
	Order   string `json:"order,omitempty"`
}

// SortOrder represents the sorting order
type SortOrder string

const (
	Ascending  SortOrder = "asc"
	Descending SortOrder = "desc"
)

// Validate checks if a Repository has required fields
func (r *Repository) Validate() error {
	if r.Owner == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}
	if r.Name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if r.URL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}
	return nil
}

// Validate checks if a User has required fields
func (u *User) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("user username cannot be empty")
	}
	return nil
}

// Validate checks if FilterOptions are valid
func (f *FilterOptions) Validate() error {
	if f.Page < 0 {
		return fmt.Errorf("page cannot be negative")
	}
	if f.PerPage <= 0 {
		return fmt.Errorf("perPage must be positive")
	}
	if f.Order != "" && f.Order != string(Ascending) && f.Order != string(Descending) {
		return fmt.Errorf("order must be 'asc' or 'desc'")
	}
	return nil
}
