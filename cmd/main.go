package main

import (
	"log"
	"os"
)

func main() {
	// Application entry point
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	// TODO: Implement application logic
	return nil
}
