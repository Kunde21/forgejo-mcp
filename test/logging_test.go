package test

import (
	"bytes"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestLoggingConfiguration validates that logging can be configured properly
func TestLoggingConfiguration(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr)

	// Test default logging level
	if logrus.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected default log level to be Info, got %v", logrus.GetLevel())
	}

	// Test logging at different levels
	logrus.Info("Test info message")
	if !bytes.Contains(buf.Bytes(), []byte("Test info message")) {
		t.Error("Info message was not logged")
	}

	logrus.Debug("Test debug message")
	// Debug message should not appear with Info level
	if bytes.Contains(buf.Bytes(), []byte("Test debug message")) {
		t.Error("Debug message should not be logged at Info level")
	}

	// Test setting debug level
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("Test debug message 2")
	if !bytes.Contains(buf.Bytes(), []byte("Test debug message 2")) {
		t.Error("Debug message was not logged at Debug level")
	}
}

// TestStructuredLogging validates structured logging functionality
func TestStructuredLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr)

	// Reset to info level
	logrus.SetLevel(logrus.InfoLevel)

	// Test structured logging
	logrus.WithFields(logrus.Fields{
		"component": "test",
		"version":   "1.0",
	}).Info("Structured log message")

	if !bytes.Contains(buf.Bytes(), []byte("Structured log message")) {
		t.Error("Structured message was not logged")
	}

	if !bytes.Contains(buf.Bytes(), []byte("component=test")) {
		t.Error("Component field was not included in structured log")
	}

	if !bytes.Contains(buf.Bytes(), []byte("version=1.0")) {
		t.Error("Version field was not included in structured log")
	}
}
