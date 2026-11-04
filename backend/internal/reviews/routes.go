// Package reviews stores review history and calculates vocabulary and kanji insights.
package reviews

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes registers review recording, history, and review-derived statistics.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	store := reviewStore{db: db}
	mux.HandleFunc("GET /reviews", store.listReviews)
	mux.HandleFunc("POST /reviews", store.createReview)
	mux.HandleFunc("GET /vocabulary-stats", store.listVocabularyStats)
	mux.HandleFunc("GET /kanji-stats", store.listKanjiStats)
}
