package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestServerLifecycle(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectedErr error
	}{
		{
			name: "ValidConfig",
			config: &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    "test-token",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "info",
			},
		},
		{
			name:        "NilConfig",
			config:      nil,
			expectedErr: fmt.Errorf("config cannot be nil"),
		},
		{
			name: "InvalidLogLevel",
			config: &config.Config{
				ForgejoURL:   "https://example.forgejo.com",
				AuthToken:    "test-token",
				TeaPath:      "tea",
				Host:         "localhost",
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				LogLevel:     "invalid",
			},
			expectedErr: fmt.Errorf("invalid configuration: LogLevel: must be a valid value."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.config)
			if !cmp.Equal(tt.expectedErr, err) {
				t.Error(cmp.Diff(tt.expectedErr, err))
			}
			if err != nil {
				return
			}
			if server == nil {
				t.Fatal("New() should return non-nil server")
			}
			if server.logger == nil {
				t.Error("server.logger should be initialized")
			}
			if server.config == nil {
				t.Error("server.config should be set")
			}
		})
	}
}

func TestServerStartStop(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server == nil {
		t.Fatal("New() returned nil server")
	}
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(t.Context())
	}()
	time.Sleep(100 * time.Millisecond)
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	select {
	case err := <-startErr:
		if err != nil {
			t.Errorf("Start() error = %v, expected no error after Stop", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Start() did not return after Stop() was called")
	}
}

func TestServerStopWithoutStart(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v, expected no error when called on non-started server", err)
	}
}

func TestServerWithLogger(t *testing.T) {
	cfg := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "debug",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server.logger == nil {
		t.Fatal("expected server.logger to be non-nil")
	}
	if !cmp.Equal(logrus.DebugLevel, server.logger.Level) {
		t.Errorf("expected logger level %v, got %v", logrus.DebugLevel, server.logger.Level)
	}
}
