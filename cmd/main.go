package main

import (
	"log"
	"os"

	"github.com/Kunde21/forgejo-mcp/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logging
	if err := setupLogging(cfg.Debug, cfg.LogLevel); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	// Application entry point
	if err := run(cfg); err != nil {
		logrus.WithError(err).Error("Application error")
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	// TODO: Implement application logic
	logrus.Info("Starting forgejo-mcp server")
	return nil
}
