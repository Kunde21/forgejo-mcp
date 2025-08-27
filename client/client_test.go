package client

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var exampleCom *url.URL

func init() {
	exampleCom, _ = url.Parse("https://example.com")
}

func TestClientInterface(t *testing.T) {
	// Test that ForgejoClient implements Client interface
	var _ Client = (*ForgejoClient)(nil)

	// Test New function with explicit expected values
	want := &ForgejoClient{
		baseURL:   exampleCom,
		token:     "test-token",
		timeout:   30 * time.Second,
		userAgent: "forgejo-mcp-client/1.0.0",
	}

	got, err := New("https://example.com", "test-token")
	if !cmp.Equal(nil, err) {
		t.Error(cmp.Diff(nil, err))
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreUnexported(ForgejoClient{})) {
		t.Error(cmp.Diff(want, got, cmpopts.IgnoreUnexported(ForgejoClient{})))
	}
}

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		token   string
		want    *ForgejoClient
		wantErr error
	}{
		{
			name:    "valid inputs",
			baseURL: "https://example.com",
			token:   "test-token",
			want: &ForgejoClient{
				baseURL:   exampleCom,
				token:     "test-token",
				timeout:   30 * time.Second,
				userAgent: "forgejo-mcp-client/1.0.0",
			},
			wantErr: nil,
		},
		{
			name:    "empty baseURL",
			baseURL: "",
			token:   "test-token",
			want:    nil,
			wantErr: &ValidationError{Message: "baseURL cannot be empty", Field: "baseURL"},
		},
		{
			name:    "empty token",
			baseURL: "https://example.com",
			token:   "",
			want:    nil,
			wantErr: &ValidationError{Message: "token cannot be empty", Field: "token"},
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			token:   "test-token",
			want:    nil,
			wantErr: &ValidationError{Message: "invalid baseURL format, must be a valid HTTP/HTTPS URL", Field: "baseURL"},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			client, err := New(tst.baseURL, tst.token)
			if !cmp.Equal(tst.wantErr, err) {
				t.Error(cmp.Diff(tst.wantErr, err))
			}
			if !cmp.Equal(tst.want, client, cmpopts.IgnoreUnexported(ForgejoClient{})) {
				t.Error(cmp.Diff(tst.want, client, cmpopts.IgnoreUnexported(ForgejoClient{})))
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ClientConfig
		want    *ForgejoClient
		wantErr error
	}{
		{
			name:   "default",
			config: DefaultConfig(),
			want: &ForgejoClient{
				baseURL:   exampleCom,
				token:     "test-token",
				timeout:   30 * time.Second,
				userAgent: "forgejo-mcp-client/1.0.0",
			},
		},
		{
			name:   "config",
			config: &ClientConfig{Timeout: 60 * time.Second, UserAgent: "test-agent/1.0"},
			want: &ForgejoClient{
				baseURL:   exampleCom,
				token:     "test-token",
				timeout:   60 * time.Second,
				userAgent: "test-agent/1.0",
			},
		},
		{
			name:   "nil",
			config: nil,
			want: &ForgejoClient{
				baseURL:   exampleCom,
				token:     "test-token",
				timeout:   30 * time.Second,
				userAgent: "forgejo-mcp-client/1.0.0",
			},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := NewWithConfig("https://example.com", "test-token", tst.config)
			if !cmp.Equal(tst.wantErr, err) {
				t.Error(cmp.Diff(tst.wantErr, err))
			}
			if !cmp.Equal(tst.want, got, cmp.AllowUnexported(ForgejoClient{})) {
				t.Error(cmp.Diff(tst.want, got, cmp.AllowUnexported(ForgejoClient{})))
			}
		})
	}
}

func TestClientGetters(t *testing.T) {
	config := &ClientConfig{
		Timeout:   45 * time.Second,
		UserAgent: "custom-agent/2.0",
	}

	client, err := NewWithConfig("https://test.com", "test-token", config)
	if !cmp.Equal(nil, err) {
		t.Error(cmp.Diff(nil, err))
	}

	tests := []struct {
		name string
		want any
		got  any
	}{
		{
			name: "GetBaseURL",
			want: "https://test.com",
			got:  client.GetBaseURL(),
		},
		{
			name: "GetTimeout",
			want: 45 * time.Second,
			got:  client.GetTimeout(),
		},
		{
			name: "GetUserAgent",
			want: "custom-agent/2.0",
			got:  client.GetUserAgent(),
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			if !cmp.Equal(tst.want, tst.got) {
				t.Error(cmp.Diff(tst.want, tst.got))
			}
		})
	}
}
