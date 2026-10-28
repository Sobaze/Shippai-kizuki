package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"
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
	db, err := openDatabase(filepath.Join("data", "shippai-kizuki.sqlite"))
	if err != nil {
		return err
	}
	defer db.Close()

	store := reviewStore{db: db}
	dictionary := kanjiDictionary{
		db:      db,
		client:  &http.Client{Timeout: 8 * time.Second},
		baseURL: "https://kanjiapi.dev/v1/kanji/",
	}
	anki := ankiClient{
		client:   &http.Client{Timeout: 5 * time.Second},
		endpoint: "http://127.0.0.1:8765/",
	}
	importer := ankiImporter{anki: &anki, db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /reviews", store.listReviews)
	mux.HandleFunc("POST /reviews", store.createReview)
	mux.HandleFunc("GET /vocabulary-stats", store.listVocabularyStats)
	mux.HandleFunc("GET /kanji-stats", store.listKanjiStats)
	mux.HandleFunc("GET /kanji/{character}", dictionary.lookup)
	mux.HandleFunc("GET /anki/decks", anki.listDecks)
	mux.HandleFunc("GET /anki/import-status", importer.importStatus)
	mux.HandleFunc("POST /anki/import-vocabulary", importer.importVocabulary)

	log.Println("Shippai Kizuki backend starting at http://localhost:8080")
	return http.ListenAndServe("localhost:8080", mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := healthResponse{Status: "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write health response: %v", err)
	}
}
