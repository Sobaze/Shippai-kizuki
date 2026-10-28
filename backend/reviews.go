package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type review struct {
	Vocabulary string `json:"vocabulary"`
	Rating     string `json:"rating"`
	Source     string `json:"source,omitempty"`
	ReviewedAt int64  `json:"reviewedAt,omitempty"`
}

type reviewStore struct {
	db *sql.DB
}

func (s *reviewStore) loadReviews(ctx context.Context, limit int) ([]review, error) {
	query := `SELECT vocabulary, rating, source, reviewed_at_ms FROM review_events WHERE rating IS NOT NULL`
	var args []any
	if limit > 0 {
		query += ` ORDER BY reviewed_at_ms DESC, id DESC LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reviews := []review{}
	for rows.Next() {
		var entry review
		if err := rows.Scan(&entry.Vocabulary, &entry.Rating, &entry.Source, &entry.ReviewedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, entry)
	}
	return reviews, rows.Err()
}

func (s *reviewStore) listReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := s.loadReviews(r.Context(), 100)
	if err != nil {
		log.Printf("load reviews: %v", err)
		http.Error(w, "Could not load saved reviews.", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		log.Printf("write reviews response: %v", err)
	}
}

func (s *reviewStore) createReview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var entry struct {
		Vocabulary string `json:"vocabulary"`
		Rating     string `json:"rating"`
	}
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

	saved := review{Vocabulary: entry.Vocabulary, Rating: entry.Rating, Source: "manual", ReviewedAt: time.Now().UnixMilli()}
	_, err := s.db.ExecContext(r.Context(), `INSERT INTO review_events
		(vocabulary, rating, reviewed_at_ms, source, raw_json) VALUES (?, ?, ?, 'manual', '{}')`,
		saved.Vocabulary, saved.Rating, saved.ReviewedAt)
	if err != nil {
		log.Printf("save manual review: %v", err)
		http.Error(w, "Could not save the review.", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(saved); err != nil {
		log.Printf("write created review response: %v", err)
	}
}
