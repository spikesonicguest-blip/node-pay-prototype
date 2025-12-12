package main

import (
	"log"
	"net/http"
	"os"
	
	"nodepay-facilitator/internal/handlers"
	
	"github.com/joho/godotenv"
)

func main() {
	// Debug: Print CWD
	cwd, _ := os.Getwd()
	log.Printf("Starting Facilitator in directory: %s\n", cwd)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults or system environment variables")
	}

	// Initialize Handlers
	h := handlers.New()

	// Router (using standard ServeMux)
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/v1/supported", h.Supported)
	mux.HandleFunc("/v1/verify", h.Verify)
	mux.HandleFunc("/v1/settle", h.Settle)

	log.Println("Facilitator starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
