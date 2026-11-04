// Package dictionary looks up kanji through the local cache and dictionary API.
package dictionary

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes registers dictionary lookups using the supplied upstream client.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, client *http.Client, baseURL string) {
	dictionary := kanjiDictionary{db: db, client: client, baseURL: baseURL}
	mux.HandleFunc("GET /kanji/{character}", dictionary.lookup)
}
