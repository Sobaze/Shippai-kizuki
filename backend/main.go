package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	store := reviewStore{reviews: demoReviews()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /reviews", store.listReviews)
	mux.HandleFunc("POST /reviews", store.createReview)
	mux.HandleFunc("GET /vocabulary-stats", store.listVocabularyStats)

	log.Println("Shippai Kizuki backend starting at http://localhost:8080")
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := healthResponse{Status: "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write health response: %v", err)
	}
}
