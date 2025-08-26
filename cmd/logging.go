package main

import (
	"github.com/sirupsen/logrus"
)

// setupLogging configures the logging system based on the configuration
func setupLogging(debug bool, logLevel string) error {
	// Set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
	}

	// Set formatter
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Log that logging is configured
	logrus.WithFields(logrus.Fields{
		"debug":    debug,
		"logLevel": logLevel,
	}).Debug("Logging configured")

	return nil
}
