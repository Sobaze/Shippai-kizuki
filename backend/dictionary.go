package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"unicode"
	"unicode/utf8"
)

type kanjiDictionary struct {
	db      *sql.DB
	client  *http.Client
	baseURL string
}

// The upstream API uses snake_case JSON names; our API uses camelCase.
type dictionaryEntry struct {
	Kanji        string   `json:"kanji"`
	Meanings     []string `json:"meanings"`
	OnReadings   []string `json:"on_readings"`
	KunReadings  []string `json:"kun_readings"`
	NameReadings []string `json:"name_readings"`
	StrokeCount  int      `json:"stroke_count"`
}

type kanjiDetails struct {
	Kanji        string   `json:"kanji"`
	Meanings     []string `json:"meanings"`
	OnReadings   []string `json:"onReadings"`
	KunReadings  []string `json:"kunReadings"`
	NameReadings []string `json:"nameReadings"`
	StrokeCount  int      `json:"strokeCount"`
}

func (d *kanjiDictionary) lookup(w http.ResponseWriter, r *http.Request) {
	character := r.PathValue("character")
	value, _ := utf8.DecodeRuneInString(character)
	if utf8.RuneCountInString(character) != 1 || !unicode.Is(unicode.Han, value) {
		http.Error(w, "provide exactly one kanji character", http.StatusBadRequest)
		return
	}

	details, found, err := getStoredKanji(r.Context(), d.db, character)
	if err != nil {
		log.Printf("dictionary database lookup: %v", err)
		http.Error(w, "could not read the local dictionary database; please try again", http.StatusInternalServerError)
		return
	}
	if found {
		writeKanjiDetails(w, details, "sqlite")
		return
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, d.baseURL+url.PathEscape(character), nil)
	if err != nil {
		log.Printf("create dictionary request: %v", err)
		http.Error(w, "could not prepare dictionary lookup", http.StatusInternalServerError)
		return
	}
	response, err := d.client.Do(request)
	if err != nil {
		if r.Context().Err() != nil {
			return // The browser has cancelled the lookup.
		}
		log.Printf("dictionary lookup: %v", err)
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			http.Error(w, "dictionary lookup timed out; please try again", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "dictionary service is unavailable; please try again", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		http.Error(w, "no dictionary entry found for this kanji", http.StatusNotFound)
		return
	}
	if response.StatusCode != http.StatusOK {
		http.Error(w, "dictionary service is unavailable; please try again", http.StatusBadGateway)
		return
	}

	var entry dictionaryEntry
	if err := json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(&entry); err != nil {
		log.Printf("decode dictionary entry: %v", err)
		http.Error(w, "dictionary service returned an unreadable entry; please try again", http.StatusBadGateway)
		return
	}
	if entry.Kanji != character || entry.StrokeCount <= 0 {
		http.Error(w, "dictionary service returned an incomplete entry; please try again", http.StatusBadGateway)
		return
	}

	// Empty reading lists should be JSON arrays, not null.
	if entry.Meanings == nil {
		entry.Meanings = []string{}
	}
	if entry.OnReadings == nil {
		entry.OnReadings = []string{}
	}
	if entry.KunReadings == nil {
		entry.KunReadings = []string{}
	}
	if entry.NameReadings == nil {
		entry.NameReadings = []string{}
	}
	details = kanjiDetails{
		Kanji: entry.Kanji, Meanings: entry.Meanings,
		OnReadings: entry.OnReadings, KunReadings: entry.KunReadings,
		NameReadings: entry.NameReadings, StrokeCount: entry.StrokeCount,
	}
	if err := saveKanji(r.Context(), d.db, details, request.URL.String()); err != nil {
		log.Printf("dictionary database save: %v", err)
		http.Error(w, "could not save the dictionary entry locally; please try again", http.StatusInternalServerError)
		return
	}
	writeKanjiDetails(w, details, "kanjiapi")
}

func writeKanjiDetails(w http.ResponseWriter, details kanjiDetails, source string) {
	w.Header().Set("X-Dictionary-Source", source)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Printf("write dictionary response: %v", err)
	}
}
