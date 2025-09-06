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
	if cfg.Host == "" {
		t.Error("Load() failed to populate Host default")
	}
}
