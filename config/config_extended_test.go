package config

import (
	"os"
	"path/filepath"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) func()
		want    *Config
		wantErr error
	}{
		{
			name: "load config from environment variables",
			setup: func(t *testing.T) func() {
				envVars := map[string]string{
					"FORGEJO_MCP_FORGEJO_URL": "https://test.forgejo.com",
					"FORGEJO_MCP_AUTH_TOKEN":  "test-token",
					"FORGEJO_MCP_DEBUG":       "true",
					"FORGEJO_MCP_LOG_LEVEL":   "debug",
				}
				for key, value := range envVars {
					os.Setenv(key, value)
				}
				return func() {
					for key := range envVars {
						os.Unsetenv(key)
					}
				}
			},
			want: &Config{
				ForgejoURL:   "https://test.forgejo.com",
				AuthToken:    "test-token",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        true,
				LogLevel:     "debug",
			},
			wantErr: nil,
		},
		{
			name: "load config from file",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				t.Cleanup(func() { os.RemoveAll(tempDir) })
				configFile := filepath.Join(tempDir, "config.yaml")
				configContent := `forgejo_url: "https://file-test.forgejo.com"
auth_token: "file-test-token"
tea_path: "/file/test/tea"
debug: true
log_level: "debug"`

				if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				oldDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
				return func() {
					os.Chdir(oldDir)
				}
			},
			want: &Config{
				ForgejoURL:   "https://file-test.forgejo.com",
				AuthToken:    "file-test-token",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        true,
				LogLevel:     "debug",
			},
			wantErr: nil,
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			t.Cleanup(tst.setup(t))
			got, err := Load()
			if !cmp.Equal(tst.wantErr, err, cmpopts.EquateErrors()) {
				t.Error(cmp.Diff(tst.wantErr, err, cmpopts.EquateErrors()))
			}
			if !cmp.Equal(tst.want, got) {
				t.Error(cmp.Diff(tst.want, got))
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name: "valid config",
			config: &Config{
				ForgejoURL:   "https://forgejo.com",
				AuthToken:    "testing-repo-auth-token-123",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        false,
				LogLevel:     "info",
			},
			wantErr: nil,
		},
		{
			name: "missing forgejo URL",
			config: &Config{
				ForgejoURL:   "",
				AuthToken:    "testing-repo-auth-token-123",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        false,
				LogLevel:     "info",
			},
		},
		{
			name: "missing auth token",
			config: &Config{
				ForgejoURL:   "https://forgejo.com",
				AuthToken:    "",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        false,
				LogLevel:     "info",
			},
		},
		{
			name: "invalid port",
			config: &Config{
				ForgejoURL:   "https://forgejo.com",
				AuthToken:    "testing-repo-auth-token-123",
				Host:         "localhost",
				Port:         0,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        false,
				LogLevel:     "info",
			},
			wantErr: validation.Errors{
				"Port": validation.NewError("validation_required", "cannot be blank"),
			},
		},
		{
			name: "invalid log level",
			config: &Config{
				ForgejoURL:   "https://forgejo.com",
				AuthToken:    "testing-repo-auth-token-123",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				Debug:        false,
				LogLevel:     "invalid",
			},
			wantErr: validation.Errors{
				"LogLevel": validation.NewError("validation_in_invalid", "must be a valid value"),
			},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := tst.config.Validate()
			if !cmp.Equal(tst.wantErr, err, cmpopts.IgnoreUnexported(validation.ErrorObject{})) {
				t.Error(cmp.Diff(tst.wantErr, err, cmpopts.IgnoreUnexported(validation.ErrorObject{})))
			}
		})
	}
}
