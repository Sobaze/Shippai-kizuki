package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type review struct {
	Vocabulary string `json:"vocabulary"`
	Rating     string `json:"rating"`
}

type reviewStore struct {
	mu      sync.Mutex
	reviews []review
}

func (s *reviewStore) listReviews(w http.ResponseWriter, r *http.Request) {
	// Copy while locked so encoding does not hold up other requests.
	s.mu.Lock()
	reviews := make([]review, len(s.reviews))
	copy(reviews, s.reviews)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		log.Printf("write reviews response: %v", err)
	}
}

func (s *reviewStore) createReview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var entry review
	if err := decoder.Decode(&entry); err != nil {
		http.Error(w, "body must be a JSON object with vocabulary and rating fields", http.StatusBadRequest)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "body must contain exactly one JSON object", http.StatusBadRequest)
		return
	}

	entry.Vocabulary = strings.TrimSpace(entry.Vocabulary)
	if entry.Vocabulary == "" {
		http.Error(w, "vocabulary must not be empty", http.StatusBadRequest)
		return
	}
	switch entry.Rating {
	case "again", "hard", "good", "easy":
		// These are the four supported Anki-style ratings.
	default:
		http.Error(w, "rating must be again, hard, good, or easy", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.reviews = append(s.reviews, entry)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		log.Printf("write created review response: %v", err)
	}
}
