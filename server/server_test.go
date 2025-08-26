package server

import (
	"testing"
	"time"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestServerLifecycle_ValidConfig(t *testing.T) {
	config := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "info",
	}

	server, err := New(config)

	if err != nil {
		t.Fatalf("New() with valid config should not return error, got: %v", err)
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
}

func TestServerLifecycle_NilConfig(t *testing.T) {
	server, err := New(nil)

	if err == nil {
		t.Error("New() with nil config should return error")
	}
	if server != nil {
		t.Error("New() with nil config should return nil server")
	}
}

func TestServerLifecycle_InvalidLogLevel(t *testing.T) {
	config := &config.Config{
		ForgejoURL:   "https://example.forgejo.com",
		AuthToken:    "test-token",
		TeaPath:      "tea",
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		LogLevel:     "invalid",
	}

	server, err := New(config)

	if err == nil {
		t.Error("New() with invalid log level should return error")
	}
	if server != nil {
		t.Error("New() with invalid log level should return nil server")
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
