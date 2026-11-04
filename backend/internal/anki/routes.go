// Package anki provides read-only AnkiConnect access and local vocabulary imports.
package anki

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes registers Anki endpoints using the supplied upstream client.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, client *http.Client, endpoint string) {
	upstream := ankiClient{client: client, endpoint: endpoint}
	importer := ankiImporter{anki: &upstream, db: db}
	mux.HandleFunc("GET /anki/decks", upstream.listDecks)
	mux.HandleFunc("GET /anki/import-status", importer.importStatus)
	mux.HandleFunc("POST /anki/import-vocabulary", importer.importVocabulary)
}
