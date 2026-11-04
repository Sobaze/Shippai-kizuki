package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Sobaze/Shippai-kizuki/backend/internal/anki"
	"github.com/Sobaze/Shippai-kizuki/backend/internal/dictionary"
	"github.com/Sobaze/Shippai-kizuki/backend/internal/quiz"
	"github.com/Sobaze/Shippai-kizuki/backend/internal/reviews"
	"github.com/Sobaze/Shippai-kizuki/backend/internal/storage"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := storage.Open(filepath.Join("data", "shippai-kizuki.sqlite"))
	if err != nil {
		return err
	}
	defer db.Close()

	mux, err := newMux(db)
	if err != nil {
		return err
	}
	log.Println("Shippai Kizuki backend starting at http://localhost:8080")
	return http.ListenAndServe("localhost:8080", mux)
}

func newMux(db *sql.DB) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	reviews.RegisterRoutes(mux, db)
	dictionary.RegisterRoutes(mux, db, &http.Client{Timeout: 8 * time.Second}, "https://kanjiapi.dev/v1/kanji/")
	anki.RegisterRoutes(mux, db, &http.Client{Timeout: 5 * time.Second}, "http://127.0.0.1:8765/")
	if err := quiz.RegisterRoutes(mux, db); err != nil {
		return nil, err
	}
	return mux, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := healthResponse{Status: "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write health response: %v", err)
	}
}
