package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kunde21/forgejo-mcp/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- run(ctx)
	}()

	// Wait for either the application to finish or an interrupt signal
	select {
	case err := <-errChan:
		if err != nil {
			logrus.WithError(err).Error("Application error")
			os.Exit(1)
		}
		logrus.Info("Application exited successfully")
	case <-sigChan:
		logrus.Info("Received interrupt signal, shutting down...")
		cancel()

		// Give the application some time to shut down gracefully
		select {
		case <-errChan:
			logrus.Info("Application shut down gracefully")
		case <-time.After(30 * time.Second):
			logrus.Warn("Shutdown timeout exceeded, forcing exit")
			os.Exit(1)
		}
	}
}

func run(ctx context.Context) error {
	if err := cmd.Execute(ctx); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}
