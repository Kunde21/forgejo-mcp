package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	// Basic check that we got a config with default values
	if cfg.ForgejoURL == "" {
		t.Error("Load() failed to populate ForgejoURL")
	}
	if cfg.TeaPath == "" {
		t.Error("Load() failed to populate TeaPath")
	}
}

func TestValidate(t *testing.T) {
	// Test valid config
	validConfig := &Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		Debug:        true,
		LogLevel:     "info",
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}

	// Test invalid config - missing Forgejo URL
	invalidConfig1 := &Config{
		ForgejoURL:   "",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		Debug:        true,
		LogLevel:     "info",
	}

	if err := invalidConfig1.Validate(); err == nil {
		t.Error("Validate() expected error for empty Forgejo URL, got nil")
	}

	// Test invalid config - missing auth token
	invalidConfig2 := &Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		Debug:        true,
		LogLevel:     "info",
	}

	if err := invalidConfig2.Validate(); err == nil {
		t.Error("Validate() expected error for empty auth token, got nil")
	}
}

func TestValidateForGiteaClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid gitea client config",
			config: &Config{
				ForgejoURL:    "https://gitea.example.com",
				AuthToken:     "valid_token",
				ClientTimeout: 30,
				UserAgent:     "test-agent/1.0",
				Host:          "localhost",
				Port:          8080,
				LogLevel:      "info",
			},
			wantErr: false,
		},
		{
			name: "missing user agent",
			config: &Config{
				ForgejoURL:    "https://gitea.example.com",
				AuthToken:     "valid_token",
				ClientTimeout: 30,
				UserAgent:     "",
				Host:          "localhost",
				Port:          8080,
				LogLevel:      "info",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				ForgejoURL:    "https://gitea.example.com",
				AuthToken:     "valid_token",
				ClientTimeout: -1,
				UserAgent:     "test-agent/1.0",
				Host:          "localhost",
				Port:          8080,
				LogLevel:      "info",
			},
			wantErr: true,
		},
		{
			name: "invalid URL scheme",
			config: &Config{
				ForgejoURL:    "ftp://gitea.example.com",
				AuthToken:     "valid_token",
				ClientTimeout: 30,
				UserAgent:     "test-agent/1.0",
				Host:          "localhost",
				Port:          8080,
				LogLevel:      "info",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateForGiteaClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateForGiteaClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
