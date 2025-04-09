// Package main is the entry point for the beacon header verification tool
package main

import (
	"log"
	"os"

	"github.com/Mikelle/beacon-root-verification/beacon-verifier/app"
	"github.com/Mikelle/beacon-root-verification/beacon-verifier/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatalf("Error initializing application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
