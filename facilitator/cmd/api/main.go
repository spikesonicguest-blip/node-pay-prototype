package main

import (
	"log"
	"net/http"
	
	"nodepay-facilitator/internal/handlers"
)

func main() {
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
