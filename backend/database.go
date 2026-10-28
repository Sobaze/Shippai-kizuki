package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func openDatabase(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	// database/sql is safe for concurrent use. One connection serializes SQLite
	// operations for this small application; network lookups never hold it open.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS kanji_entries (
			kanji TEXT PRIMARY KEY NOT NULL,
			meanings TEXT NOT NULL,
			on_readings TEXT NOT NULL,
			kun_readings TEXT NOT NULL,
			name_readings TEXT NOT NULL,
			stroke_count INTEGER NOT NULL CHECK (stroke_count > 0),
			source TEXT NOT NULL,
			fetched_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS anki_cards (
			card_id INTEGER PRIMARY KEY,
			note_id INTEGER NOT NULL,
			deck_name TEXT NOT NULL,
			model_name TEXT NOT NULL,
			vocabulary TEXT NOT NULL,
			reading TEXT NOT NULL,
			answer_text TEXT NOT NULL,
			fields_json TEXT NOT NULL,
			imported_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS review_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			anki_review_id INTEGER UNIQUE,
			anki_card_id INTEGER,
			vocabulary TEXT NOT NULL,
			rating TEXT,
			reviewed_at_ms INTEGER NOT NULL,
			source TEXT NOT NULL,
			raw_json TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS review_events_time
			ON review_events(reviewed_at_ms DESC, id DESC);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return db, nil
}

func getStoredKanji(ctx context.Context, db *sql.DB, character string) (kanjiDetails, bool, error) {
	var details kanjiDetails
	var meanings, onReadings, kunReadings, nameReadings string
	err := db.QueryRowContext(ctx, `
		SELECT kanji, meanings, on_readings, kun_readings, name_readings, stroke_count
		FROM kanji_entries WHERE kanji = ?
	`, character).Scan(
		&details.Kanji, &meanings, &onReadings, &kunReadings, &nameReadings, &details.StrokeCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return kanjiDetails{}, false, nil
	}
	if err != nil {
		return kanjiDetails{}, false, fmt.Errorf("read stored kanji: %w", err)
	}

	// Decode each JSON column back into its corresponding Go slice.
	encoded := []string{meanings, onReadings, kunReadings, nameReadings}
	lists := []*[]string{&details.Meanings, &details.OnReadings, &details.KunReadings, &details.NameReadings}
	for i, list := range lists {
		if err := json.Unmarshal([]byte(encoded[i]), list); err != nil {
			return kanjiDetails{}, false, fmt.Errorf("decode stored kanji list: %w", err)
		}
		if *list == nil {
			*list = []string{}
		}
	}
	return details, true, nil
}

func saveKanji(ctx context.Context, db *sql.DB, details kanjiDetails, source string) error {
	lists := [][]string{details.Meanings, details.OnReadings, details.KunReadings, details.NameReadings}
	encoded := make([]string, len(lists))
	for i, list := range lists {
		if list == nil {
			list = []string{}
		}
		data, err := json.Marshal(list)
		if err != nil {
			return fmt.Errorf("encode kanji list: %w", err)
		}
		encoded[i] = string(data)
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO kanji_entries
			(kanji, meanings, on_readings, kun_readings, name_readings, stroke_count, source, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(kanji) DO UPDATE SET
			meanings = excluded.meanings,
			on_readings = excluded.on_readings,
			kun_readings = excluded.kun_readings,
			name_readings = excluded.name_readings,
			stroke_count = excluded.stroke_count,
			source = excluded.source,
			fetched_at = excluded.fetched_at
	`, details.Kanji, encoded[0], encoded[1], encoded[2], encoded[3], details.StrokeCount,
		source, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save kanji: %w", err)
	}
	return nil
}
